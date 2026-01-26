package service

import (
	"iat/common/model"
	"iat/engine/internal/repo"
)

type ProjectService struct {
	repo *repo.ProjectRepo
}

func NewProjectService() *ProjectService {
	return &ProjectService{
		repo: repo.NewProjectRepo(),
	}
}

func (s *ProjectService) CreateProject(name, description, path string) error {
	project := &model.Project{
		Name:        name,
		Description: description,
		Path:        path,
	}
	return s.repo.Create(project)
}

func (s *ProjectService) UpdateProject(id uint, name, description, path string) error {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	project.Name = name
	project.Description = description
	project.Path = path
	return s.repo.Update(project)
}

func (s *ProjectService) DeleteProject(id uint) error {
	return s.repo.Delete(id)
}

func (s *ProjectService) ListProjects() ([]model.Project, error) {
	return s.repo.List()
}
