package handler

import (
	"encoding/json"
	"iat/engine/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type SubAgentTaskHandler struct {
	svc *service.SubAgentTaskService
}

func NewSubAgentTaskHandler(svc *service.SubAgentTaskService) *SubAgentTaskHandler {
	return &SubAgentTaskHandler{svc: svc}
}

func (h *SubAgentTaskHandler) List(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.URL.Query().Get("sessionId")
	if sessionIDStr == "" {
		http.Error(w, "sessionId is required", http.StatusBadRequest)
		return
	}
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid sessionId", http.StatusBadRequest)
		return
	}

	tasks, err := h.svc.ListBySessionID(uint(sessionID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *SubAgentTaskHandler) Abort(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	// /api/subagent-tasks/{id}/abort
	if len(parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	taskID := parts[3]

	if err := h.svc.AbortTask(taskID, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *SubAgentTaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	taskID := parts[3]

	task, err := h.svc.GetTask(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
