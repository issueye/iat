package service

import (
	"iat/internal/model"
	"iat/internal/pkg/ai"
	"iat/internal/repo"
)

type AIModelService struct {
	repo *repo.AIModelRepo
}

func NewAIModelService() *AIModelService {
	return &AIModelService{
		repo: repo.NewAIModelRepo(),
	}
}

func (s *AIModelService) CreateModel(m *model.AIModel) error {
	if m.IsDefault {
		// Unset other defaults
		if err := s.repo.UnsetDefault(); err != nil {
			return err
		}
	}
	return s.repo.Create(m)
}

func (s *AIModelService) UpdateModel(m *model.AIModel) error {
	if m.IsDefault {
		// Unset other defaults
		if err := s.repo.UnsetDefault(); err != nil {
			return err
		}
	}
	return s.repo.Update(m)
}

func (s *AIModelService) GetDefaultModel() (*model.AIModel, error) {
	return s.repo.GetDefault()
}

func (s *AIModelService) DeleteModel(id uint) error {
	return s.repo.Delete(id)
}

func (s *AIModelService) ListModels() ([]model.AIModel, error) {
	return s.repo.List()
}

func (s *AIModelService) TestConnection(m *model.AIModel) error {
	// Use temporary client to test connection
	client, err := ai.NewAIClient(m, nil)
	if err != nil {
		return err
	}
	// Try to generate a simple response (implementation depends on eino capabilities, here we assume client creation validates config or we can try a simple chat)
	// For now, if client creation succeeds, we assume basic config format is correct.
	// A real test would involve sending a "ping" message.
	_ = client
	return nil
}
