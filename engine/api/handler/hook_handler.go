package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"iat/common/model"
	"iat/engine/internal/service"
)

type HookHandler struct {
	svc *service.HookService
}

func NewHookHandler(svc *service.HookService) *HookHandler {
	return &HookHandler{svc: svc}
}

func (h *HookHandler) List(w http.ResponseWriter, r *http.Request) {
	hooks, err := h.svc.ListHooks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(hooks)
}

func (h *HookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var hook model.Hook
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateHook(&hook); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *HookHandler) Update(w http.ResponseWriter, r *http.Request) {
	var hook model.Hook
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateHook(&hook); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *HookHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.svc.DeleteHook(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
