package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"task-manager/internal/handlers"
	"task-manager/internal/models"
	"task-manager/internal/services"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.Task, error) {
	args := m.Called(ctx, req)
	if t := args.Get(0); t != nil {
		return t.(*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskService) GetTask(ctx context.Context, id int64) (*models.Task, error) {
	args := m.Called(ctx, id)
	if t := args.Get(0); t != nil {
		return t.(*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskService) GetAllTasks(ctx context.Context, limit int, offset int) ([]*models.Task, int, error) {
	args := m.Called(ctx, limit, offset)
	if t := args.Get(0); t != nil {
		return t.([]*models.Task), args.Int(1), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockTaskService) UpdateTask(ctx context.Context, id int64, req *models.UpdateTaskRequest) (*models.Task, error) {
	args := m.Called(ctx, id, req)
	if t := args.Get(0); t != nil {
		return t.(*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskService) GetDueTasks(ctx context.Context, from int64, to int64) ([]*models.Task, error) {
	args := m.Called(ctx, from, to)
	if t := args.Get(0); t != nil {
		return t.([]*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskService) MarkTaskComplete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockTaskService) DeleteTask(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func makeRequest(t *testing.T, handlerFunc http.HandlerFunc, method, url string, body any) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handlerFunc(rr, req)
	return rr
}

func TestCreateTask(t *testing.T) {
	mockSvc := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockSvc)

	reqBody := &models.CreateTaskRequest{
		Title:       "Task 1",
		Description: "Desc 1",
		DueDate:     time.Now(),
	}
	expected := &models.Task{ID: 1, Title: reqBody.Title, Description: reqBody.Description, DueDate: reqBody.DueDate}

	mockSvc.On("CreateTask", mock.Anything, mock.MatchedBy(func(req *models.CreateTaskRequest) bool {
		return req.Title == "Task 1" && req.Description == "Desc 1"
	})).Return(expected, nil)

	rr := makeRequest(t, handler.CreateTask, "POST", "/tasks", reqBody)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var res models.Task
	json.Unmarshal(rr.Body.Bytes(), &res)
	assert.Equal(t, expected.Title, res.Title)
	mockSvc.AssertExpectations(t)
}

func TestGetTask_NotFound(t *testing.T) {
	mockSvc := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockSvc)

	mockSvc.On("GetTask", mock.Anything, int64(1)).
		Return(nil, services.ErrTaskNotFound)

	req := httptest.NewRequest("GET", "/tasks/1", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id}", handler.GetTask)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestGetTask_InternalServerErr(t *testing.T) {
	mockSvc := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockSvc)

	mockSvc.On("GetTask", mock.Anything, int64(1)).
		Return(nil, errors.New("internal error"))

	req := httptest.NewRequest("GET", "/tasks/1", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id}", handler.GetTask)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestGetTask_ParseErr(t *testing.T) {
	mockSvc := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockSvc)

	req := httptest.NewRequest("GET", "/tasks/abc", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id}", handler.GetTask)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestGetTask(t *testing.T) {
	mockSvc := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockSvc)

	task := &models.Task{ID: 1, Title: "Task 1"}
	mockSvc.On("GetTask", mock.Anything, int64(1)).Return(task, nil)

	req := httptest.NewRequest("GET", "/tasks/1", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id}", handler.GetTask)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var res models.Task
	json.Unmarshal(rr.Body.Bytes(), &res)
	assert.Equal(t, task.ID, res.ID)
	mockSvc.AssertExpectations(t)
}
func TestGetAllTasks(t *testing.T) {
	mockSvc := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockSvc)

	tasks := []*models.Task{{ID: 1}, {ID: 2}}
	total := 2

	mockSvc.On("GetAllTasks", mock.Anything, 10, 0).Return(tasks, total, nil)

	req := httptest.NewRequest("GET", "/tasks?page=1&limit=10", nil)
	rr := httptest.NewRecorder()

	handler.GetAllTasks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var res map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &res)
	assert.NoError(t, err)

	assert.Equal(t, float64(total), res["total"])
	assert.Len(t, res["data"].([]interface{}), 2)

	mockSvc.AssertExpectations(t)
}

func TestUpdateTask(t *testing.T) {
	mockSvc := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockSvc)

	reqBody := &models.UpdateTaskRequest{Title: "Updated"}
	updated := &models.Task{ID: 1, Title: "Updated"}

	mockSvc.On("UpdateTask", mock.Anything, int64(1), reqBody).Return(updated, nil)

	req := httptest.NewRequest("PUT", "/tasks/1", bytes.NewBufferString(`{"title":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id}", handler.UpdateTask)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var res models.Task
	json.Unmarshal(rr.Body.Bytes(), &res)
	assert.Equal(t, "Updated", res.Title)
	mockSvc.AssertExpectations(t)
}

func TestMarkTaskComplete(t *testing.T) {
	mockSvc := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockSvc)

	mockSvc.On("MarkTaskComplete", mock.Anything, int64(1)).Return(nil)

	req := httptest.NewRequest("PATCH", "/tasks/1/complete", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id}/complete", handler.MarkTaskComplete)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var res map[string]bool
	json.Unmarshal(rr.Body.Bytes(), &res)
	assert.True(t, res["status"])
	mockSvc.AssertExpectations(t)
}

func TestDeleteTask(t *testing.T) {
	mockSvc := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockSvc)

	mockSvc.On("DeleteTask", mock.Anything, int64(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/tasks/1", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id}", handler.DeleteTask)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var res map[string]bool
	json.Unmarshal(rr.Body.Bytes(), &res)
	assert.True(t, res["status"])
	mockSvc.AssertExpectations(t)
}
