package service

import (
	"iat/common/model"
	"iat/engine/internal/repo"
)

type ModeService struct {
	repo *repo.ModeRepo
}

func NewModeService() *ModeService {
	return &ModeService{
		repo: repo.NewModeRepo(),
	}
}

func (s *ModeService) CreateMode(key, name, description, systemPrompt string) error {
	mode := &model.Mode{
		Key:          key,
		Name:         name,
		Description:  description,
		SystemPrompt: systemPrompt,
	}
	return s.repo.Create(mode)
}

func (s *ModeService) UpdateMode(id uint, key, name, description, systemPrompt string) error {
	mode, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	mode.Key = key
	mode.Name = name
	mode.Description = description
	mode.SystemPrompt = systemPrompt
	return s.repo.Update(mode)
}

func (s *ModeService) DeleteMode(id uint) error {
	return s.repo.Delete(id)
}

func (s *ModeService) ListModes() ([]model.Mode, error) {
	return s.repo.List()
}
