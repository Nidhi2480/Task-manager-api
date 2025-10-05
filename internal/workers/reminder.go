package worker

import (
	"context"
	"log"
	"task-manager/internal/services"
	"time"
)

type ReminderWorker struct {
	taskService services.TaskService
	interval    time.Duration
	lookahead   time.Duration
}

func NewReminderWorker(taskService services.TaskService, interval, lookahead time.Duration) *ReminderWorker {
	return &ReminderWorker{
		taskService: taskService,
		interval:    interval,
		lookahead:   lookahead,
	}
}

func (w *ReminderWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkDueTasks(ctx)
		case <-ctx.Done():
			log.Println("Reminder worker stopped")
			return
		}
	}
}

func (w *ReminderWorker) checkDueTasks(ctx context.Context) {
	now := time.Now()
	from := now.Unix()
	to := now.Add(w.lookahead).Unix()

	tasks, err := w.taskService.GetDueTasks(ctx, from, to)
	if err != nil {
		log.Printf("Error getting due tasks: %v", err)
		return
	}

	for _, task := range tasks {
		log.Printf("Reminder: Task %d (%s) is due at %s", task.ID, task.Title, task.DueDate.Format(time.RFC3339))
	}
}
