package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"iat/engine/internal/service"
)

type ProjectHandler struct {
	svc      *service.ProjectService
	indexSvc *service.IndexService
}

func NewProjectHandler(svc *service.ProjectService, indexSvc *service.IndexService) *ProjectHandler {
	return &ProjectHandler{svc: svc, indexSvc: indexSvc}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.svc.ListProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(projects)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Path        string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateProject(req.Name, req.Description, req.Path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /api/projects/{id}
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
		Name        string `json:"name"`
		Description string `json:"description"`
		Path        string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateProject(uint(id), req.Name, req.Description, req.Path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /api/projects/{id}
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

	if err := h.svc.DeleteProject(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) Index(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /api/projects/{id}/index
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

	res, err := h.indexSvc.IndexProject(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func (h *ProjectHandler) IndexAll(w http.ResponseWriter, r *http.Request) {
	res, err := h.indexSvc.IndexAllProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}
