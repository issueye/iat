package service

import (
	"fmt"
	"iat/internal/model"
	"iat/internal/repo"
)

type AgentService struct {
	repo *repo.AgentRepo
}

func NewAgentService() *AgentService {
	return &AgentService{
		repo: repo.NewAgentRepo(),
	}
}

func (s *AgentService) CreateAgent(name, description, systemPrompt string, modelID uint) error {
	agent := &model.Agent{
		Name:         name,
		Description:  description,
		SystemPrompt: systemPrompt,
		ModelID:      modelID,
	}
	return s.repo.Create(agent)
}

func (s *AgentService) UpdateAgent(id uint, name, description, systemPrompt string, modelID uint) error {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	agent.Name = name
	agent.Description = description
	agent.SystemPrompt = systemPrompt
	agent.ModelID = modelID
	return s.repo.Update(agent)
}

func (s *AgentService) DeleteAgent(id uint) error {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if agent.Type == "builtin" {
		return fmt.Errorf("cannot delete builtin agent")
	}
	return s.repo.Delete(id)
}

func (s *AgentService) ListAgents() ([]model.Agent, error) {
	return s.repo.List()
}
