//go:build integration

package task_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akhilr007/tasks/internal/auth"
	"github.com/akhilr007/tasks/internal/task"

	"github.com/go-chi/chi/v5"
)

func setupWithTx(t *testing.T) *chi.Mux {
	t.Helper()

	if testpool == nil {
		t.Skip("TEST_DB_URL is not set; skipping integration test")
	}

	tx, err := testpool.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	email := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "@example.test"
	var userID int
	err = tx.QueryRow(
		context.Background(),
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email,
		"test-password-hash",
	).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	repo := task.NewPGRepository(tx)
	logger := newTestLogger()
	service := task.NewService(repo, logger)
	handler := task.NewHandler(service, logger)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Route("/tasks", func(r chi.Router) {
		handler.Routes(r)
	})

	return r
}

type createResp struct {
	Data struct {
		ID int `json:"id"`
	} `json:"data"`
}

func createTask(t *testing.T, mux *chi.Mux) int {

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"task"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("create failed: %d", rr.Code)
	}

	var resp createResp
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	return resp.Data.ID
}

func TestCreateTask_Integration(t *testing.T) {

	t.Parallel()

	mux := setupWithTx(t)

	id := createTask(t, mux)
	if id == 0 {
		t.Fatal("expected valid id")
	}
}

func TestGetTaskByID_Integration(t *testing.T) {

	t.Parallel()

	mux := setupWithTx(t)

	id := createTask(t, mux)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tasks/%d", id), nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}

func TestUpdateTask_Integration(t *testing.T) {

	t.Parallel()

	mux := setupWithTx(t)

	id := createTask(t, mux)

	req := httptest.NewRequest(http.MethodPut,
		fmt.Sprintf("/tasks/%d", id),
		strings.NewReader(`{"done":true}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}

func TestDeleteTask_Integration(t *testing.T) {

	t.Parallel()

	mux := setupWithTx(t)

	id := createTask(t, mux)

	req := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/tasks/%d", id),
		nil,
	)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 got %d", rr.Code)
	}
}
