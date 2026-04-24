package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	h "github.com/akhilr007/tasks/internal/handler"
	"github.com/akhilr007/tasks/internal/model"
	"github.com/akhilr007/tasks/internal/store"
	"github.com/go-chi/chi/v5"
)

// ---------- Mock Store ----------

type MockStore struct {
	GetAllFunc  func(ctx context.Context) ([]model.Task, error)
	GetByIDFunc func(ctx context.Context, id int) (model.Task, error)
	CreateFunc  func(ctx context.Context, title string) (model.Task, error)
	UpdateFunc  func(ctx context.Context, id int, title *string, done *bool) (model.Task, error)
	DeleteFunc  func(ctx context.Context, id int) error
}

func (m *MockStore) GetAll(ctx context.Context) ([]model.Task, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	panic("GetAllFunc not implemented")
}

func (m *MockStore) GetByID(ctx context.Context, id int) (model.Task, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return model.Task{}, errors.New("GetByIDFunc not implemented")
}

func (m *MockStore) Create(ctx context.Context, title string) (model.Task, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, title)
	}
	panic("CreateFunc not implemented")
}

func (m *MockStore) Update(ctx context.Context, id int, title *string, done *bool) (model.Task, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, title, done)
	}
	return model.Task{}, errors.New("UpdateFunc not implemented")
}

func (m *MockStore) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("DeleteFunc not implemented")
}

// ---------- Factory ----------

func newMockStore() *MockStore {
	return &MockStore{
		GetAllFunc: func(ctx context.Context) ([]model.Task, error) {
			return []model.Task{}, nil
		},
		GetByIDFunc: func(ctx context.Context, id int) (model.Task, error) {
			return model.Task{}, nil
		},
		CreateFunc: func(ctx context.Context, title string) (model.Task, error) {
			return model.Task{ID: 1, Title: title}, nil
		},
		UpdateFunc: func(ctx context.Context, id int, title *string, done *bool) (model.Task, error) {
			return model.Task{ID: id}, nil
		},
		DeleteFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}
}

// ---------- Helper ----------

func newHandlerWithMock(mock store.TaskStore) *h.Handler {
	return h.NewHandler(mock, newTestLogger())
}

func addChiParam(r *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

//
// ---------- CREATE ----------
//

func TestHandleCreateTask(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mock       func() store.TaskStore
		wantStatus int
	}{
		{
			name: "success",
			body: `{"title":"new task"}`,
			mock: func() store.TaskStore {
				m := newMockStore()
				m.CreateFunc = func(ctx context.Context, title string) (model.Task, error) {
					return model.Task{ID: 1, Title: title}, nil
				}
				return m
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid json",
			body: "invalid",
			mock: func() store.TaskStore {
				return newMockStore()
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty title",
			body: `{"title":""}`,
			mock: func() store.TaskStore {
				return newMockStore()
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newHandlerWithMock(tt.mock())

			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.HandleCreateTask(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

//
// ---------- GET ----------
//

func TestHandleGetTaskByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mock       func() store.TaskStore
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			mock: func() store.TaskStore {
				m := newMockStore()
				m.GetByIDFunc = func(ctx context.Context, id int) (model.Task, error) {
					return model.Task{ID: 1}, nil
				}
				return m
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "1",
			mock: func() store.TaskStore {
				m := newMockStore()
				m.GetByIDFunc = func(ctx context.Context, id int) (model.Task, error) {
					return model.Task{}, store.ErrNotFound
				}
				return m
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid id",
			id:   "abc",
			mock: func() store.TaskStore {
				return newMockStore()
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newHandlerWithMock(tt.mock())

			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.id, nil)
			req = addChiParam(req, "id", tt.id)

			rr := httptest.NewRecorder()
			handler.HandleGetTaskByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}
