package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"iat/engine/internal/service"
)

type AgentHandler struct {
	svc *service.AgentService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
	agents, err := h.svc.ListAgents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(agents)
}

func (h *AgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		SystemPrompt   string `json:"systemPrompt"`
		Type           string `json:"type"`
		ModelID        uint   `json:"modelId"`
		ToolIDs        []uint `json:"toolIds"`
		MCPServerIDs   []uint `json:"mcpServerIds"`
		ModeIDs        []uint `json:"modeIds"`
		ExternalURL    string `json:"externalUrl"`
		ExternalType   string `json:"externalType"`
		ExternalParams string `json:"externalParams"`
		Status         string `json:"status"`
		Capabilities   string `json:"capabilities"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateAgent(req.Name, req.Description, req.SystemPrompt, req.Type, req.ExternalURL, req.ExternalType, req.ExternalParams, req.ModelID, req.ToolIDs, req.MCPServerIDs, req.ModeIDs, req.Status, req.Capabilities); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
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
		Name           string `json:"name"`
		Description    string `json:"description"`
		SystemPrompt   string `json:"systemPrompt"`
		Type           string `json:"type"`
		ModelID        uint   `json:"modelId"`
		ToolIDs        []uint `json:"toolIds"`
		MCPServerIDs   []uint `json:"mcpServerIds"`
		ModeIDs        []uint `json:"modeIds"`
		ExternalURL    string `json:"externalUrl"`
		ExternalType   string `json:"externalType"`
		ExternalParams string `json:"externalParams"`
		Status         string `json:"status"`
		Capabilities   string `json:"capabilities"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateAgent(uint(id), req.Name, req.Description, req.SystemPrompt, req.Type, req.ExternalURL, req.ExternalType, req.ExternalParams, req.ModelID, req.ToolIDs, req.MCPServerIDs, req.ModeIDs, req.Status, req.Capabilities); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AgentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
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

	if err := h.svc.DeleteAgent(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
