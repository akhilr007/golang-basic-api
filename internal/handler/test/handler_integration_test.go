package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/akhilr007/tasks/internal/db"
	h "github.com/akhilr007/tasks/internal/handler"
	"github.com/akhilr007/tasks/internal/store"

	"github.com/go-chi/chi/v5"
)

func setupWithDB(t *testing.T) *chi.Mux {
	t.Helper()

	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Fatal("TEST_DB_URL not set")
	}

	pool, err := db.NewPool(dsn)
	if err != nil {
		t.Fatal(err)
	}

	// cleanup DB before test
	_, err = pool.Exec(context.Background(), "TRUNCATE tasks RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	pgStore := store.NewPGStore(pool)
	handler := h.NewHandler(pgStore, newTestLogger())

	r := chi.NewRouter()
	handler.Routes(r)

	return r
}

func TestCreateTask_Integration(t *testing.T) {
	mux := setupWithDB(t)

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"task"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d", rr.Code)
	}
}

func TestGetTaskByID_Integration(t *testing.T) {
	mux := setupWithDB(t)

	// create first
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"task"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// fetch
	req = httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}

func TestUpdateTask_Integration(t *testing.T) {
	mux := setupWithDB(t)

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"task"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	req = httptest.NewRequest(http.MethodPut, "/tasks/1", strings.NewReader(`{"done":true}`))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}

func TestDeleteTask_Integration(t *testing.T) {
	mux := setupWithDB(t)

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"task"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	req = httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 got %d", rr.Code)
	}
}