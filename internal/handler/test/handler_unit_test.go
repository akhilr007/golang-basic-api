package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	h "github.com/akhilr007/tasks/internal/handler"
	"github.com/akhilr007/tasks/internal/model"
	"github.com/akhilr007/tasks/internal/store"
	chi "github.com/go-chi/chi/v5"
)

//
// ===============================
// MOCK STORE
// ===============================
//

type MockStore struct {
	GetAllFunc  func(ctx context.Context) ([]model.Task, error)
	GetByIDFunc func(ctx context.Context, id int) (model.Task, error)
	CreateFunc  func(ctx context.Context, title string) (model.Task, error)
	UpdateFunc  func(ctx context.Context, id int, title *string, done *bool) (model.Task, error)
	DeleteFunc  func(ctx context.Context, id int) error
}

func (m *MockStore) GetAll(ctx context.Context) ([]model.Task, error) {
	return m.GetAllFunc(ctx)
}

func (m *MockStore) GetByID(ctx context.Context, id int) (model.Task, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockStore) Create(ctx context.Context, title string) (model.Task, error) {
	return m.CreateFunc(ctx, title)
}

func (m *MockStore) Update(ctx context.Context, id int, title *string, done *bool) (model.Task, error) {
	return m.UpdateFunc(ctx, id, title, done)
}

func (m *MockStore) Delete(ctx context.Context, id int) error {
	return m.DeleteFunc(ctx, id)
}

//
// ===============================
// HELPERS
// ===============================
//

func newHandlerWithMock(mock *MockStore) *h.Handler {
	return h.NewHandler(mock, newTestLogger())
}

func runRequest(t *testing.T, handler *h.Handler, method, url, body string) *httptest.ResponseRecorder {

	t.Helper()

	r := chi.NewRouter()
	handler.Routes(r)

	req := httptest.NewRequest(method, url, strings.NewReader(body))
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
					CreateFunc: func(ctx context.Context, title string) (model.Task, error) {
						return model.Task{ID: 1, Title: title}, nil
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
					CreateFunc: func(ctx context.Context, title string) (model.Task, error) {
						return model.Task{}, errors.New("db error")
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
					GetByIDFunc: func(ctx context.Context, id int) (model.Task, error) {
						return model.Task{ID: id, Title: "task"}, nil
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
					GetByIDFunc: func(ctx context.Context, id int) (model.Task, error) {
						return model.Task{}, store.ErrNotFound
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
					GetByIDFunc: func(ctx context.Context, id int) (model.Task, error) {
						return model.Task{}, errors.New("db error")
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
					UpdateFunc: func(ctx context.Context, id int, title *string, done *bool) (model.Task, error) {
						var t string
						var d bool
						if title != nil {
							t = *title
						}
						if done != nil {
							d = *done
						}
						return model.Task{ID: id, Title: t, Done: d}, nil
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
					UpdateFunc: func(ctx context.Context, id int, title *string, done *bool) (model.Task, error) {
						return model.Task{}, store.ErrNotFound
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
					DeleteFunc: func(ctx context.Context, id int) error {
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
					DeleteFunc: func(ctx context.Context, id int) error {
						return store.ErrNotFound
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
