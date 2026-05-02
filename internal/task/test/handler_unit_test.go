package task_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akhilr007/tasks/internal/auth"
	"github.com/akhilr007/tasks/internal/task"
	chi "github.com/go-chi/chi/v5"
)

//
// ===============================
// MOCK STORE
// ===============================
//

type MockStore struct {
	GetAllFunc  func(ctx context.Context, userID int) ([]task.Task, error)
	GetByIDFunc func(ctx context.Context, id, userID int) (task.Task, error)
	CreateFunc  func(ctx context.Context, userID int, title string) (task.Task, error)
	UpdateFunc  func(ctx context.Context, id, userID int, title *string, done *bool) (task.Task, error)
	DeleteFunc  func(ctx context.Context, id, userID int) error
}

func (m *MockStore) GetAll(ctx context.Context, userID int) ([]task.Task, error) {
	return m.GetAllFunc(ctx, userID)
}

func (m *MockStore) GetByID(ctx context.Context, id, userID int) (task.Task, error) {
	return m.GetByIDFunc(ctx, id, userID)
}

func (m *MockStore) Create(ctx context.Context, userID int, title string) (task.Task, error) {
	return m.CreateFunc(ctx, userID, title)
}

func (m *MockStore) Update(ctx context.Context, id, userID int, title *string, done *bool) (task.Task, error) {
	return m.UpdateFunc(ctx, id, userID, title, done)
}

func (m *MockStore) Delete(ctx context.Context, id, userID int) error {
	return m.DeleteFunc(ctx, id, userID)
}

//
// ===============================
// HELPERS
// ===============================
//

func newHandlerWithMock(mock *MockStore) *task.Handler {
	logger := newTestLogger()
	service := task.NewService(mock, logger)
	return task.NewHandler(service, logger)
}

func runRequest(t *testing.T, handler *task.Handler, method, url, body string) *httptest.ResponseRecorder {

	t.Helper()

	r := chi.NewRouter()
	r.Route("/tasks", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	ctx := context.WithValue(req.Context(), auth.UserIDKey, 1)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	return rr
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

//
// ===============================
// CREATE TASK
// ===============================
//

func TestHandleCreateTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		mock       func() *MockStore
		wantStatus int
		wantError  bool
	}{
		{
			name: "success",
			body: `{"title":"task"}`,
			mock: func() *MockStore {
				return &MockStore{
					CreateFunc: func(ctx context.Context, userID int, title string) (task.Task, error) {
						return task.Task{ID: 1, UserID: userID, Title: title}, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `invalid`,
			mock:       func() *MockStore { return &MockStore{} },
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "empty title",
			body:       `{"title":""}`,
			mock:       func() *MockStore { return &MockStore{} },
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "db error",
			body: `{"title":"task"}`,
			mock: func() *MockStore {
				return &MockStore{
					CreateFunc: func(ctx context.Context, userID int, title string) (task.Task, error) {
						return task.Task{}, errors.New("db error")
					},
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newHandlerWithMock(tt.mock())
			rr := runRequest(t, handler, http.MethodPost, "/tasks", tt.body)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rr.Code)
			}

			resp := decodeBody(t, rr)

			if tt.wantError {
				if resp["error"] == nil {
					t.Error("expected error in response")
				}
			} else {
				if resp["data"] == nil {
					t.Error("expected data in response")
				}
			}
		})
	}
}

//
// ===============================
// GET TASK BY ID
// ===============================
//

func TestHandleGetTaskByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		mock       func() *MockStore
		wantStatus int
		wantError  bool
	}{
		{
			name: "success",
			url:  "/tasks/1",
			mock: func() *MockStore {
				return &MockStore{
					GetByIDFunc: func(ctx context.Context, id, userID int) (task.Task, error) {
						return task.Task{ID: id, UserID: userID, Title: "task"}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			url:  "/tasks/1",
			mock: func() *MockStore {
				return &MockStore{
					GetByIDFunc: func(ctx context.Context, id, userID int) (task.Task, error) {
						return task.Task{}, task.ErrNotFound
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
		{
			name:       "invalid id",
			url:        "/tasks/abc",
			mock:       func() *MockStore { return &MockStore{} },
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "db error",
			url:  "/tasks/1",
			mock: func() *MockStore {
				return &MockStore{
					GetByIDFunc: func(ctx context.Context, id, userID int) (task.Task, error) {
						return task.Task{}, errors.New("db error")
					},
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newHandlerWithMock(tt.mock())
			rr := runRequest(t, handler, http.MethodGet, tt.url, "")

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

//
// ===============================
// UPDATE TASK
// ===============================
//

func TestHandleUpdateTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		body       string
		mock       func() *MockStore
		wantStatus int
		wantError  bool
	}{
		{
			name: "success",
			url:  "/tasks/1",
			body: `{"title":"updated","done":true}`,
			mock: func() *MockStore {
				return &MockStore{
					UpdateFunc: func(ctx context.Context, id, userID int, title *string, done *bool) (task.Task, error) {
						var t string
						var d bool
						if title != nil {
							t = *title
						}
						if done != nil {
							d = *done
						}
						return task.Task{ID: id, UserID: userID, Title: t, Done: d}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			url:        "/tasks/1",
			body:       `invalid`,
			mock:       func() *MockStore { return &MockStore{} },
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "not found",
			url:  "/tasks/1",
			body: `{"title":"x"}`,
			mock: func() *MockStore {
				return &MockStore{
					UpdateFunc: func(ctx context.Context, id, userID int, title *string, done *bool) (task.Task, error) {
						return task.Task{}, task.ErrNotFound
					},
				}
			},
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newHandlerWithMock(tt.mock())
			rr := runRequest(t, handler, http.MethodPut, tt.url, tt.body)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

//
// ===============================
// DELETE TASK
// ===============================
//

func TestHandleDeleteTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		mock       func() *MockStore
		wantStatus int
	}{
		{
			name: "success",
			url:  "/tasks/1",
			mock: func() *MockStore {
				return &MockStore{
					DeleteFunc: func(ctx context.Context, id, userID int) error {
						return nil
					},
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			url:  "/tasks/1",
			mock: func() *MockStore {
				return &MockStore{
					DeleteFunc: func(ctx context.Context, id, userID int) error {
						return task.ErrNotFound
					},
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newHandlerWithMock(tt.mock())
			rr := runRequest(t, handler, http.MethodDelete, tt.url, "")

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}
