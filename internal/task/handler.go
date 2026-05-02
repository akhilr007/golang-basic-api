package task

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/akhilr007/tasks/internal/auth"
	"github.com/akhilr007/tasks/internal/utils"

	chi "github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  log,
	}
}

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.HandleGetAllTasks)
	r.Post("/", h.HandleCreateTask)
	r.Get("/{id}", h.HandleGetTaskByID)
	r.Put("/{id}", h.HandleUpdateTaskByID)
	r.Delete("/{id}", h.HandleDeleteTaskByID)
}

func (h *Handler) HandleGetAllTasks(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid user")
		return
	}

	p := utils.ParsePagination(r)

	tasks, hasMore, err := h.service.GetAll(r.Context(), userID, p)
	if err != nil {
		log.Error("failed to fetch tasks", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	meta := utils.BuildMeta(p, hasMore)
	if tasks == nil {
		tasks = []Task{}
	}

	resp := utils.PaginatedResponse[Task]{
		Data: tasks,
		Meta: meta,
	}

	log.Info("fetched tasks", "count", len(tasks))
	utils.WriteSuccess(w, http.StatusOK, resp)
}

func (h *Handler) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid user")
		return
	}

	defer func() {
		_ = r.Body.Close()
	}() // over engineering - go handles itself
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		log.Warn("invalid content type", "content_type", r.Header.Get("Content-Type"))
		utils.WriteError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
		return
	}

	var req CreateTaskRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		log.Warn("invalid request body", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		log.Warn("extra data in request body")
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	} // over engineering

	if strings.TrimSpace(req.Title) == "" {
		log.Warn("empty title provided")
		utils.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}

	task, err := h.service.Create(r.Context(), userID, req.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			log.Warn("empty title provided")
			utils.WriteError(w, http.StatusBadRequest, ErrInvalidTitle.Error())
		} else {
			log.Error("failed to create task", "error", err)
			utils.WriteError(w, http.StatusInternalServerError, utils.ErrInternal)
		}
		return
	}

	log.Info("task created", "task_id", task.ID)
	utils.WriteSuccess(w, http.StatusCreated, task)
}

func (h *Handler) HandleGetTaskByID(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	id, err := utils.ParseIDFromRequest(r)
	if err != nil {
		log.Warn("invalid task id")
		utils.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid user")
		return
	}

	task, err := h.service.GetByID(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.Info("task not found", "task_id", id)
			utils.WriteError(w, http.StatusNotFound, ErrNotFound.Error())
		} else {
			log.Error("failed to fetch task", "task_id", id, "error", err)
			utils.WriteError(w, http.StatusInternalServerError, utils.ErrInternal)
		}
		return
	}

	log.Info("task fetched", "task_id", id)
	utils.WriteSuccess(w, http.StatusOK, task)
}

func (h *Handler) HandleUpdateTaskByID(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	defer func() {
		_ = r.Body.Close()
	}()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		log.Warn("invalid content type")
		utils.WriteError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
		return
	}

	id, err := utils.ParseIDFromRequest(r)
	if err != nil {
		log.Warn("invalid task id")
		utils.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req UpdateTaskRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		log.Warn("invalid request body", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		log.Warn("extra data in request body")
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.Done == nil {
		log.Warn("empty update request", "task_id", id)
		utils.WriteError(w, http.StatusBadRequest, "empty request")
		return
	}

	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid user")
		return
	}

	task, err := h.service.Update(r.Context(), id, userID, req.Title, req.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.Info("task not found for update", "task_id", id)
			utils.WriteError(w, http.StatusNotFound, ErrNotFound.Error())
		} else if errors.Is(err, ErrInvalidTitle) {
			log.Warn("empty title provided", "task_id", id)
			utils.WriteError(w, http.StatusBadRequest, ErrInvalidTitle.Error())
		} else {
			log.Error("failed to update task", "task_id", id, "error", err)
			utils.WriteError(w, http.StatusInternalServerError, utils.ErrInternal)
		}
		return
	}

	log.Info("task updated", "task_id", id)
	utils.WriteSuccess(w, http.StatusOK, task)
}

func (h *Handler) HandleDeleteTaskByID(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With("method", r.Method, "path", r.URL.Path)

	id, err := utils.ParseIDFromRequest(r)
	if err != nil {
		log.Warn("invalid task id")
		utils.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid user")
		return
	}

	err = h.service.Delete(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.Info("task not found for delete", "task_id", id)
			utils.WriteError(w, http.StatusNotFound, ErrNotFound.Error())
		} else {
			log.Error("failed to delete task", "task_id", id, "error", err)
			utils.WriteError(w, http.StatusInternalServerError, utils.ErrInternal)
		}
		return
	}

	log.Info("task deleted", "task_id", id)
	w.WriteHeader(http.StatusNoContent)
}
