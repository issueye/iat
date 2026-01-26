package handler

import (
	"encoding/json"
	"net/http"
	"iat/engine/internal/service"
)

type ModeHandler struct {
	svc *service.ModeService
}

func NewModeHandler(svc *service.ModeService) *ModeHandler {
	return &ModeHandler{svc: svc}
}

func (h *ModeHandler) List(w http.ResponseWriter, r *http.Request) {
	modes, err := h.svc.ListModes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(modes)
}
