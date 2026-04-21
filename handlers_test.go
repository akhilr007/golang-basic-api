package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setup() *http.ServeMux {
	store := NewStore()
	handler := NewHandler(store)

	mux := http.NewServeMux()
	handler.Routes(mux)
	return mux
}

func TestCreateTask_TableDriven(t *testing.T) {

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid", `{"title": "Task1"}`, http.StatusCreated},
		{"invalid json", `invalid`, http.StatusBadRequest},
		{"empty title", `{"title": ""}`, http.StatusBadRequest},
		{"missing title", `{}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setup()

			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestGetTaskByID_TableDriven(t *testing.T) {
	
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"valid id", "/tasks/1", http.StatusOK},
		{"not found", "/tasks/999", http.StatusNotFound},
		{"invalid id", "/tasks/abc", http.StatusBadRequest},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setup()
			if tt.name == "valid id" {
				req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"new task"}`))
				req.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				mux.ServeHTTP(rr, req)
			}
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rr := httptest.NewRecorder()
			
			mux.ServeHTTP(rr, req)
			
			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}

}

func TestUpdateTask_TableDriven(t *testing.T) {
	
	tests := []struct{
		name string
		body string
		wantStatus int
	}{
		{"update title and done", `{"title": "New Task", "done":true}`, http.StatusOK},
		{"update title", `{"title": "New"}`, http.StatusOK},
		{"update done", `{"done": true}`, http.StatusOK},
		{"invalid json", `invalid`, http.StatusBadRequest},
		{"empty body", `{}`, http.StatusBadRequest},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setup()

			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"new task"}`))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			
			req = httptest.NewRequest(http.MethodPut, "/tasks/1", strings.NewReader(tt.body))
			rr = httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestDeleteTask_TableDriven(t *testing.T) {
	tests := []struct{
		name string
		url string
		wantStatus int
	}{
		{"delete valid", "/tasks/1", http.StatusNoContent},
		{"delete again", "/tasks/1", http.StatusNotFound},
		{"invalid id", "/tasks/abc", http.StatusBadRequest},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setup()
			
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title": "new task"}`))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			
			mux.ServeHTTP(rr, req)
			
			if rr.Code != http.StatusCreated {
				t.Fatalf("failed to create task, got %d", rr.Code)
			}
			
			// special case: delete twice
			if tt.name == "delete again" {
				req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
				rr := httptest.NewRecorder()
				mux.ServeHTTP(rr, req)
			}
			
			req = httptest.NewRequest(http.MethodDelete, tt.url, nil)
			rr = httptest.NewRecorder()
			
			mux.ServeHTTP(rr, req)
			
			if rr.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}