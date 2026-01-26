package handler

import (
	"encoding/json"
	"iat/engine/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type SessionHandler struct {
	svc     *service.SessionService
	chatSvc *service.ChatService
}

func NewSessionHandler(svc *service.SessionService, chatSvc *service.ChatService) *SessionHandler {
	return &SessionHandler{svc: svc, chatSvc: chatSvc}
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	projectIdStr := r.URL.Query().Get("projectId")
	if projectIdStr == "" {
		http.Error(w, "projectId is required", http.StatusBadRequest)
		return
	}
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		http.Error(w, "Invalid projectId", http.StatusBadRequest)
		return
	}

	sessions, err := h.svc.ListSessions(uint(projectId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		ProjectID uint   `json:"projectId"`
		AgentID   uint   `json:"agentId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, err := h.svc.CreateSession(req.Name, req.ProjectID, req.AgentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /api/sessions/{id}
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateSession(uint(id), req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /api/sessions/{id}
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Also clear messages
	_ = h.chatSvc.ClearMessages(uint(id))

	if err := h.svc.DeleteSession(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *SessionHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /api/sessions/{id}/messages
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodDelete {
		if err := h.chatSvc.ClearMessages(uint(id)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	msgs, err := h.chatSvc.ListMessages(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(msgs)
}

func (h *SessionHandler) Abort(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /api/sessions/{id}/abort
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	h.chatSvc.AbortSession(uint(id))
	w.WriteHeader(http.StatusOK)
}
