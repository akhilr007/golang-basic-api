package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockStore struct {
	GetAllFunc  func() []Task
	GetByIDFunc func(int) (Task, error)
	CreateFunc  func(string) Task
	UpdateFunc  func(int, *string, *bool) (Task, error)
	DeleteFunc  func(int) error
}

// ---------- Safe implementations ----------

func (m *MockStore) GetAll() []Task {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	panic("GetAllFunc not implemented")
}

func (m *MockStore) GetByID(id int) (Task, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return Task{}, errors.New("GetByIDFunc not implemented")
}

func (m *MockStore) Create(title string) Task {
	if m.CreateFunc != nil {
		return m.CreateFunc(title)
	}
	panic("CreateFunc not implemented")
}

func (m *MockStore) Update(id int, title *string, done *bool) (Task, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(id, title, done)
	}
	return Task{}, errors.New("UpdateFunc not implemented")
}

func (m *MockStore) Delete(id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return errors.New("DeleteFunc not implemented")
}

// ---------- Factory ----------

func newMockStore() *MockStore {
	return &MockStore{
		GetAllFunc: func() []Task {
			return []Task{}
		},
		GetByIDFunc: func(id int) (Task, error) {
			return Task{}, nil
		},
		CreateFunc: func(title string) Task {
			return Task{ID: 1, Title: title}
		},
		UpdateFunc: func(id int, title *string, done *bool) (Task, error) {
			return Task{ID: id}, nil
		},
		DeleteFunc: func(id int) error {
			return nil
		},
	}
}

// ---------- CREATE ----------

func TestHandleCreateTask(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mock       func() TaskStore
		wantStatus int
	}{
		{
			name: "success",
			body: `{"title":"new task"}`,
			mock: func() TaskStore {
				m := newMockStore()
				m.CreateFunc = func(title string) Task {
					return Task{ID: 1, Title: title}
				}
				return m
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid json",
			body: "invalid",
			mock: func() TaskStore {
				return newMockStore()
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty title",
			body: `{"title":""}`,
			mock: func() TaskStore {
				return newMockStore()
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mock())

			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.handleCreateTask(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

// ---------- GET BY ID ----------

func TestHandleGetTaskByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mock       func() TaskStore
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			mock: func() TaskStore {
				m := newMockStore()
				m.GetByIDFunc = func(id int) (Task, error) {
					return Task{ID: 1, Title: "Task"}, nil
				}
				return m
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "1",
			mock: func() TaskStore {
				m := newMockStore()
				m.GetByIDFunc = func(id int) (Task, error) {
					return Task{}, ErrNotFound
				}
				return m
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid id",
			id:   "abc",
			mock: func() TaskStore {
				return newMockStore()
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mock())

			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.id, nil)
			req.SetPathValue("id", tt.id)

			rr := httptest.NewRecorder()
			handler.handleGetTaskByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

// ---------- UPDATE ----------

func TestHandleUpdateTaskByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		mock       func() TaskStore
		wantStatus int
	}{
		{
			name: "update title",
			id:   "1",
			body: `{"title":"updated"}`,
			mock: func() TaskStore {
				m := newMockStore()
				m.UpdateFunc = func(id int, title *string, done *bool) (Task, error) {
					return Task{ID: id, Title: *title}, nil
				}
				return m
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "update done",
			id:   "1",
			body: `{"done":true}`,
			mock: func() TaskStore {
				m := newMockStore()
				m.UpdateFunc = func(id int, title *string, done *bool) (Task, error) {
					return Task{ID: id, Done: *done}, nil
				}
				return m
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "1",
			body: `{"done":true}`,
			mock: func() TaskStore {
				m := newMockStore()
				m.UpdateFunc = func(id int, title *string, done *bool) (Task, error) {
					return Task{}, ErrNotFound
				}
				return m
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid json",
			id:   "1",
			body: "invalid",
			mock: func() TaskStore {
				return newMockStore()
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mock())

			req := httptest.NewRequest(http.MethodPut, "/tasks/"+tt.id, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", tt.id)

			rr := httptest.NewRecorder()
			handler.handleUpdateTaskByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

// ---------- DELETE ----------

func TestHandleDeleteTaskByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mock       func() TaskStore
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			mock: func() TaskStore {
				m := newMockStore()
				m.DeleteFunc = func(id int) error {
					return nil
				}
				return m
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			id:   "1",
			mock: func() TaskStore {
				m := newMockStore()
				m.DeleteFunc = func(id int) error {
					return ErrNotFound
				}
				return m
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid id",
			id:   "abc",
			mock: func() TaskStore {
				return newMockStore()
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mock())

			req := httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.id, nil)
			req.SetPathValue("id", tt.id)

			rr := httptest.NewRecorder()
			handler.handleDeleteTaskByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}