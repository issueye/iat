package service

import (
	"iat/internal/model"
	"iat/internal/repo"
)

type ToolService struct {
	repo *repo.ToolRepo
}

func NewToolService() *ToolService {
	return &ToolService{
		repo: repo.NewToolRepo(),
	}
}

func (s *ToolService) ListTools() ([]model.Tool, error) {
	return s.repo.List()
}

func (s *ToolService) CreateTool(tool *model.Tool) error {
	return s.repo.Create(tool)
}

func (s *ToolService) UpdateTool(tool *model.Tool) error {
	return s.repo.Update(tool)
}

func (s *ToolService) DeleteTool(id uint) error {
	return s.repo.Delete(id)
}
