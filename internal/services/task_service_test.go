package services_test

import (
	"context"
	"errors"
	"task-manager/internal/models"
	"task-manager/internal/services"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id int64) (*models.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Task, int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, 0, args.Error(1)
	}
	return args.Get(0).([]*models.Task), 0, args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) MarkComplete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) GetDueTasks(ctx context.Context, from, to time.Time) ([]*models.Task, error) {
	args := m.Called(ctx, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Task), args.Error(1)
}

// -------------------- Tests --------------------

func TestTaskServiceMethods(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := services.NewTaskService(mockRepo)

	now := time.Now()
	task := &models.Task{ID: 1, Title: "Test", Description: "Desc", DueDate: now, IsCompleted: false}

	t.Run("CreateTask", func(t *testing.T) {
		req := &models.CreateTaskRequest{Title: "Test", Description: "Desc", DueDate: now}

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil).Once()

		createdTask, err := service.CreateTask(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, req.Title, createdTask.Title)
		assert.Equal(t, req.Description, createdTask.Description)

		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateTask error", func(t *testing.T) {
		req := &models.CreateTaskRequest{Title: "Test", Description: "Desc", DueDate: now}
		internalErr := errors.New("internal error")
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(internalErr).Once()

		got, err := service.CreateTask(context.Background(), req)
		assert.ErrorIs(t, err, internalErr)
		assert.Nil(t, got)

		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateTask empty input", func(t *testing.T) {
		req := &models.CreateTaskRequest{Description: "Desc", DueDate: now}

		got, err := service.CreateTask(context.Background(), req)
		assert.ErrorIs(t, err, services.ErrInvalidInput)
		assert.Nil(t, got)

		mockRepo.AssertExpectations(t)
	})

	t.Run("GetTask", func(t *testing.T) {
		testCases := map[string]struct {
			id       int64
			mockTask *models.Task
			mockErr  error
			wantErr  error
		}{
			"found":     {id: 1, mockTask: task, mockErr: nil, wantErr: nil},
			"not found": {id: 2, mockTask: nil, mockErr: nil, wantErr: services.ErrTaskNotFound},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				mockRepo.On("GetByID", mock.Anything, tc.id).Return(tc.mockTask, tc.mockErr).Once()
				got, err := service.GetTask(context.Background(), tc.id)
				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
					assert.Nil(t, got)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tc.mockTask, got)
				}
				mockRepo.AssertExpectations(t)
			})
		}
	})

	t.Run("GetAllTasks", func(t *testing.T) {
		mockRepo.On("GetAll", mock.Anything).Return([]*models.Task{task}, nil).Once()
		tasks, _, err := service.GetAllTasks(context.Background(), 1, 1)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateTask", func(t *testing.T) {
		req := &models.UpdateTaskRequest{Title: "Updated", Description: "Updated Desc", DueDate: now}

		mockRepo.On("GetByID", mock.Anything, int64(1)).Return(task, nil).Once()
		mockRepo.On("Update", mock.Anything, task).Return(nil).Once()

		updatedTask, err := service.UpdateTask(context.Background(), 1, req)
		assert.NoError(t, err)
		assert.Equal(t, "Updated", updatedTask.Title)
		assert.Equal(t, "Updated Desc", updatedTask.Description)

		mockRepo.AssertExpectations(t)
	})

	t.Run("MarkTaskComplete", func(t *testing.T) {
		mockRepo.On("MarkComplete", mock.Anything, int64(1)).Return(nil).Once()

		err := service.MarkTaskComplete(context.Background(), 1)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteTask", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, int64(1)).Return(nil).Once()

		err := service.DeleteTask(context.Background(), 1)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetDueTasks", func(t *testing.T) {
		from := now.Unix()
		to := now.Add(time.Hour).Unix()

		mockRepo.On("GetDueTasks", mock.Anything, time.Unix(from, 0), time.Unix(to, 0)).
			Return([]*models.Task{task}, nil).Once()

		tasks, err := service.GetDueTasks(context.Background(), from, to)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		mockRepo.AssertExpectations(t)
	})
}
