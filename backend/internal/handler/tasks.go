package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-playground/validator/v10"

	"github.com/daily-dashboard/backend/internal/middleware"
	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/service"
)

// TasksHandler handles CRUD for user-scoped tasks.
type TasksHandler struct {
	service  *service.TasksService
	validate *validator.Validate
}

// NewTasksHandler creates a new TasksHandler.
func NewTasksHandler(svc *service.TasksService, v *validator.Validate) *TasksHandler {
	return &TasksHandler{service: svc, validate: v}
}

// AddRoutes registers task routes on the given router.
func (h *TasksHandler) AddRoutes(r chi.Router) {
	r.Get("/api/tasks", h.List)
	r.With(httprate.LimitByIP(10, time.Second)).Post("/api/tasks", h.Create)
	r.With(httprate.LimitByIP(10, time.Second)).Patch("/api/tasks/{id}", h.Update)
	r.With(httprate.LimitByIP(10, time.Second)).Delete("/api/tasks/{id}", h.Delete)
}

func (h *TasksHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}

	tasks, err := h.service.List(r.Context(), userID)
	if err != nil {
		slog.Error("failed to list tasks", "error", err, "user_id", userID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasksToResponse(tasks))
}

func (h *TasksHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096) // 4 KB is generous for these small JSON bodies
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.Create(r.Context(), userID, req.Text, req.Priority)
	if err != nil {
		if errors.Is(err, service.ErrTaskValidation) {
			http.Error(w, "invalid request body", http.StatusBadRequest)
		} else {
			slog.Error("failed to create task", "error", err, "user_id", userID)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskToResponse(task))
}

func (h *TasksHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}

	id := chi.URLParam(r, "id")
	r.Body = http.MaxBytesReader(w, r.Body, 4096) // 4 KB is generous for these small JSON bodies
	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.Update(r.Context(), id, userID, req.Done, req.Text, req.Priority)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
		} else if errors.Is(err, service.ErrTaskValidation) {
			http.Error(w, "invalid request body", http.StatusBadRequest)
		} else {
			slog.Error("failed to update task", "error", err, "user_id", userID, "task_id", id)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taskToResponse(task))
}

func (h *TasksHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
		} else {
			slog.Error("failed to delete task", "error", err, "user_id", userID, "task_id", id)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// userIDFromRequest extracts the authenticated user ID from the request context.
func userIDFromRequest(r *http.Request) (string, bool) {
	sess, ok := middleware.SessionFromContext(r.Context())
	if !ok {
		return "", false
	}
	return sess.UserID, true
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized"}`))
}

func taskToResponse(t model.Task) TaskResponse {
	return TaskResponse{
		ID:       t.ID,
		Text:     t.Text,
		Done:     t.Done,
		Priority: t.Priority,
	}
}

func tasksToResponse(tasks []model.Task) []TaskResponse {
	resp := make([]TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, taskToResponse(t))
	}
	return resp
}
