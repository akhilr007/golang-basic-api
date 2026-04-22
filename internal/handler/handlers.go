package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang/tasks/internal/model"
	"golang/tasks/internal/store"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store store.TaskStore
}

type SuccessResponse struct {
	Data any `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error: message,
	})
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, SuccessResponse{
		Data: data,
	})
}

func parseIDFromRequest(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}

	return id, nil
}

func NewHandler(store store.TaskStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) Routes(r chi.Router) {
	
	r.Get("/", h.HandlePing)
	
	r.Route("/tasks", func (r chi.Router) {
		r.Get("/", h.HandleGetAllTasks)
		r.Post("/", h.HandleCreateTask)
		r.Get("/{id}", h.HandleGetTaskByID)
		r.Put("/{id}", h.HandleUpdateTaskByID)
		r.Delete("/{id}", h.HandleDeleteTaskByID)
	}) 
}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (h *Handler) HandleGetAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.store.GetAll()
	writeSuccess(w, http.StatusOK, tasks)
}

func (h *Handler) HandleCreateTask(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	var req model.CreateTaskRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	task := h.store.Create(req.Title)
	writeSuccess(w, http.StatusCreated, task)
}

func (h *Handler) HandleGetTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.store.GetByID(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeSuccess(w, http.StatusOK, task)
}

func (h *Handler) HandleUpdateTaskByID(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	id, err := parseIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req model.UpdateTaskRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.Done == nil {
		writeError(w, http.StatusBadRequest, "empty request")
		return
	}

	task, err := h.store.Update(id, req.Title, req.Done)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeSuccess(w, http.StatusOK, task)
}

func (h *Handler) HandleDeleteTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	err = h.store.Delete(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
