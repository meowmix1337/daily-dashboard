package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/service"
)

func newTasksRouter() (chi.Router, *service.TasksService) {
	svc := service.NewTasksService()
	h := NewTasksHandler(svc)
	r := chi.NewRouter()
	h.AddRoutes(r)
	return r, svc
}

func TestTasksHandler_List(t *testing.T) {
	r, _ := newTasksRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}
	var tasks []model.Task
	if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) == 0 {
		t.Error("expected at least one task in the default list")
	}
}

func TestTasksHandler_Create(t *testing.T) {
	r, _ := newTasksRouter()

	body := `{"text":"write tests","priority":"high"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusCreated)
	}
	var task model.Task
	if err := json.NewDecoder(w.Body).Decode(&task); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if task.ID == "" {
		t.Error("expected ID to be assigned")
	}
	if task.Text != "write tests" {
		t.Errorf("got text %q, want %q", task.Text, "write tests")
	}
}

func TestTasksHandler_Create_BadBody(t *testing.T) {
	r, _ := newTasksRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestTasksHandler_Update(t *testing.T) {
	r, svc := newTasksRouter()

	// Get an existing task ID
	tasks, _ := svc.List(context.Background())
	id := tasks[0].ID

	body := `{"done":true}`
	req := httptest.NewRequest(http.MethodPatch, "/api/tasks/"+id, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}
	var task model.Task
	if err := json.NewDecoder(w.Body).Decode(&task); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !task.Done {
		t.Error("expected task to be marked done")
	}
}

func TestTasksHandler_Update_NotFound(t *testing.T) {
	r, _ := newTasksRouter()

	body := `{"done":true}`
	req := httptest.NewRequest(http.MethodPatch, "/api/tasks/nonexistent-id", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestTasksHandler_Delete(t *testing.T) {
	r, svc := newTasksRouter()

	tasks, _ := svc.List(context.Background())
	id := tasks[0].ID

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/"+id, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestTasksHandler_Delete_NotFound(t *testing.T) {
	r, _ := newTasksRouter()

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/nonexistent-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusNotFound)
	}
}
