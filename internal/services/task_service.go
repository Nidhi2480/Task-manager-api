package services

import (
	"context"
	"errors"
	"task-manager/internal/models"
	"task-manager/internal/repository"
	"time"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidInput = errors.New("invalid input")
)

type TaskService interface {
	CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.Task, error)
	GetTask(ctx context.Context, id int64) (*models.Task, error)
	GetAllTasks(ctx context.Context, limit, offset int) ([]*models.Task, int, error)
	UpdateTask(ctx context.Context, id int64, req *models.UpdateTaskRequest) (*models.Task, error)
	MarkTaskComplete(ctx context.Context, id int64) error
	DeleteTask(ctx context.Context, id int64) error
	GetDueTasks(ctx context.Context, from, to int64) ([]*models.Task, error)
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.Task, error) {
	if req.Title == "" {
		return nil, ErrInvalidInput
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate, //UTC
		IsCompleted: false,
	}

	err := s.repo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, id int64) (*models.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *taskService) GetAllTasks(ctx context.Context, limit, offset int) ([]*models.Task, int, error) {
	return s.repo.GetAll(ctx, limit, offset)
}

func (s *taskService) UpdateTask(ctx context.Context, id int64, req *models.UpdateTaskRequest) (*models.Task, error) {
	existingTask, err := s.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		existingTask.Title = req.Title
	}
	if req.Description != "" {
		existingTask.Description = req.Description
	}
	if !req.DueDate.IsZero() {
		existingTask.DueDate = req.DueDate //UTC
	}

	err = s.repo.Update(ctx, existingTask)
	if err != nil {
		return nil, err
	}

	return existingTask, nil
}

func (s *taskService) MarkTaskComplete(ctx context.Context, id int64) error {
	return s.repo.MarkComplete(ctx, id)
}

func (s *taskService) DeleteTask(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *taskService) GetDueTasks(ctx context.Context, from, to int64) ([]*models.Task, error) {
	fromTime := time.Unix(from, 0)
	toTime := time.Unix(to, 0)

	return s.repo.GetDueTasks(ctx, fromTime, toTime)
}
