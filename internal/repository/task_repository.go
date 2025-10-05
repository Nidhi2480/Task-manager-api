package repository

import (
	"context"
	"database/sql"
	"log"
	"task-manager/internal/models"
	"time"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id int64) (*models.Task, error)
	GetAll(ctx context.Context) ([]*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	MarkComplete(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	GetDueTasks(ctx context.Context, from, to time.Time) ([]*models.Task, error)
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	query := `INSERT INTO tasks (title, description, due_date, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id
			`

	now := time.Now()
	err := r.db.QueryRowContext(
		ctx,
		query,
		task.Title,
		task.Description,
		task.DueDate,
		now,
		now,
	).Scan(&task.ID)

	if err != nil {
		return err
	}

	task.CreatedAt = now
	task.UpdatedAt = now

	return nil
}

func (r *taskRepository) GetByID(ctx context.Context, id int64) (*models.Task, error) {
	query := `SELECT id, title, description, due_date, is_completed, created_at, updated_at
				FROM tasks
				WHERE id = $1`

	task := &models.Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.DueDate,
		&task.IsCompleted,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return task, nil
}

func (r *taskRepository) GetAll(ctx context.Context) ([]*models.Task, error) {
	query := `SELECT id, title, description, due_date, is_completed, created_at, updated_at
				FROM tasks
				ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.DueDate,
			&task.IsCompleted,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *taskRepository) Update(ctx context.Context, task *models.Task) error {
	query := `UPDATE tasks 
				SET title = $1, description = $2, due_date = $3, updated_at = $4
				WHERE id = $5
			`

	task.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(
		ctx,
		query,
		task.Title,
		task.Description,
		task.DueDate,
		task.UpdatedAt,
		task.ID,
	)

	return err
}

func (r *taskRepository) MarkComplete(ctx context.Context, id int64) error {
	query := `UPDATE tasks 
				SET is_completed = true, updated_at = $1
				WHERE id = $2
			`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)

	return err
}

func (r *taskRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)

	return err
}

func (r *taskRepository) GetDueTasks(ctx context.Context, from, to time.Time) ([]*models.Task, error) {
	query := `SELECT id, title, description, due_date, is_completed, created_at, updated_at
			FROM tasks
			WHERE due_date BETWEEN $1 AND $2 
			AND is_completed = false
			ORDER BY due_date ASC
		`
	log.Printf("Handler triggered:%v - %v", from, to)

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.DueDate,
			&task.IsCompleted,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
