package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task-manager/internal/handlers"
	"task-manager/internal/middleware"
	"task-manager/internal/repository"
	"task-manager/internal/services"
	worker "task-manager/internal/workers"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	taskRepo := repository.NewTaskRepository(db)
	taskService := services.NewTaskService(taskRepo)
	taskHandler := handlers.NewTaskHandler(taskService)

	router := mux.NewRouter()

	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	router.Handle("/tasks", middleware.JWTMiddleware(http.HandlerFunc(taskHandler.CreateTask))).Methods("POST")
	router.Handle("/tasks", middleware.JWTMiddleware(http.HandlerFunc(taskHandler.GetAllTasks))).Methods("GET")
	router.Handle("/tasks/{id}", middleware.JWTMiddleware(http.HandlerFunc(taskHandler.GetTask))).Methods("GET")
	router.Handle("/tasks/{id}", middleware.JWTMiddleware(http.HandlerFunc(taskHandler.UpdateTask))).Methods("PUT")
	router.Handle("/tasks/{id}/complete", middleware.JWTMiddleware(http.HandlerFunc(taskHandler.MarkTaskComplete))).Methods("PATCH")
	router.Handle("/tasks/{id}", middleware.JWTMiddleware(http.HandlerFunc(taskHandler.DeleteTask))).Methods("DELETE")

	reminderWorker := worker.NewReminderWorker(
		taskService,
		time.Minute,
		5*time.Minute,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go reminderWorker.Start(ctx)

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
