package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/akhilr007/tasks/internal/model"
	"github.com/akhilr007/tasks/internal/store"

	"github.com/go-chi/chi/v5"
)

const errInternal = "internal server error"

type Handler struct {
	store  store.TaskStore
	logger *slog.Logger
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
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}

	return id, nil
}

func NewHandler(store store.TaskStore, log *slog.Logger) *Handler {
	return &Handler{
		store:  store,
		logger: log,
	}
}

func (h *Handler) Routes(r chi.Router) {

	r.Get("/health", h.HandlePing)

	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", h.HandleGetAllTasks)
		r.Post("/", h.HandleCreateTask)
		r.Get("/{id}", h.HandleGetTaskByID)
		r.Put("/{id}", h.HandleUpdateTaskByID)
		r.Delete("/{id}", h.HandleDeleteTaskByID)
	})
}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("health check")
	writeSuccess(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (h *Handler) HandleGetAllTasks(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	tasks, err := h.store.GetAll(r.Context())
	if err != nil {
		log.Error("failed to fetch tasks", "error", err)
		writeError(w, http.StatusInternalServerError, errInternal)
		return
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	log.Info("fetched tasks", "count", len(tasks))
	writeSuccess(w, http.StatusOK, tasks)
}

func (h *Handler) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		log.Warn("invalid content type", "content_type", r.Header.Get("Content-Type"))
		writeError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
		return
	}

	var req model.CreateTaskRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		log.Warn("invalid request body", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		log.Warn("extra data in request body")
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		log.Warn("empty title provided")
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	task, err := h.store.Create(r.Context(), req.Title)
	if err != nil {
		log.Error("failed to create task", "error", err)
		writeError(w, http.StatusInternalServerError, errInternal)
		return
	}

	log.Info("task created", "task_id", task.ID)
	writeSuccess(w, http.StatusCreated, task)
}

func (h *Handler) HandleGetTaskByID(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	id, err := parseIDFromRequest(r)
	if err != nil {
		log.Warn("invalid task id")
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			log.Info("task not found", "task_id", id)
			writeError(w, http.StatusNotFound, store.ErrNotFound.Error())
		} else {
			log.Error("failed to fetch task", "task_id", id, "error", err)
			writeError(w, http.StatusInternalServerError, errInternal)
		}
		return
	}

	log.Info("task fetched", "task_id", id)
	writeSuccess(w, http.StatusOK, task)
}

func (h *Handler) HandleUpdateTaskByID(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		log.Warn("invalid content type")
		writeError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
		return
	}

	id, err := parseIDFromRequest(r)
	if err != nil {
		log.Warn("invalid task id")
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req model.UpdateTaskRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		log.Warn("invalid request body", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		log.Warn("extra data in request body")
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.Done == nil {
		log.Warn("empty update request", "task_id", id)
		writeError(w, http.StatusBadRequest, "empty request")
		return
	}

	task, err := h.store.Update(r.Context(), id, req.Title, req.Done)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			log.Info("task not found for update", "task_id", id)
			writeError(w, http.StatusNotFound, store.ErrNotFound.Error())
		} else {
			log.Error("failed to update task", "task_id", id, "error", err)
			writeError(w, http.StatusInternalServerError, errInternal)
		}
		return
	}

	log.Info("task updated", "task_id", id)
	writeSuccess(w, http.StatusOK, task)
}

func (h *Handler) HandleDeleteTaskByID(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	id, err := parseIDFromRequest(r)
	if err != nil {
		log.Warn("invalid task id")
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	err = h.store.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			log.Info("task not found for delete", "task_id", id)
			writeError(w, http.StatusNotFound, store.ErrNotFound.Error())
		} else {
			log.Error("failed to delete task", "task_id", id, "error", err)
			writeError(w, http.StatusInternalServerError, errInternal)
		}
		return
	}

	log.Info("task deleted", "task_id", id)
	w.WriteHeader(http.StatusNoContent)
}
