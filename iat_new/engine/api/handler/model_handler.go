package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"iat/common/model"
	"iat/engine/internal/service"
)

type AIModelHandler struct {
	svc *service.AIModelService
}

func NewAIModelHandler(svc *service.AIModelService) *AIModelHandler {
	return &AIModelHandler{svc: svc}
}

func (h *AIModelHandler) List(w http.ResponseWriter, r *http.Request) {
	models, err := h.svc.ListModels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(models)
}

func (h *AIModelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var m model.AIModel
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateModel(&m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *AIModelHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.svc.DeleteModel(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AIModelHandler) Test(w http.ResponseWriter, r *http.Request) {
	var m model.AIModel
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.TestConnection(&m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
