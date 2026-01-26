package handler

import (
	"encoding/json"
	"iat/common/pkg/result"
	"iat/common/protocol"
	"iat/engine/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type RegistryHandler struct {
	svc *service.RegistryService
}

func NewRegistryHandler(svc *service.RegistryService) *RegistryHandler {
	return &RegistryHandler{svc: svc}
}

type RegisterRequest struct {
	AgentID      uint                  `json:"agentId"`
	Capabilities []protocol.Capability `json:"capabilities"`
	Endpoint     string                `json:"endpoint"`
}

func (h *RegistryHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		result.Error(err.Error()).JSON(w)
		return
	}

	if err := h.svc.Register(req.AgentID, req.Capabilities, req.Endpoint); err != nil {
		result.Error(err.Error()).JSON(w)
		return
	}

	result.Success("registered").JSON(w)
}

func (h *RegistryHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/registry/heartbeat/")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		result.Error("invalid agent id").JSON(w)
		return
	}

	if err := h.svc.Heartbeat(uint(id)); err != nil {
		result.Error(err.Error()).JSON(w)
		return
	}

	result.Success("ok").JSON(w)
}

func (h *RegistryHandler) Discover(w http.ResponseWriter, r *http.Request) {
	capability := r.URL.Query().Get("capability")
	if capability == "" {
		result.Error("capability query param required").JSON(w)
		return
	}

	agents, err := h.svc.Discover(capability)
	if err != nil {
		result.Error(err.Error()).JSON(w)
		return
	}

	result.Success(agents).JSON(w)
}
