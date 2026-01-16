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

func (s *AgentService) CreateAgent(name, description, systemPrompt string, modelID uint, toolIDs []uint) error {
	var tools []model.Tool
	for _, tid := range toolIDs {
		tools = append(tools, model.Tool{Base: model.Base{ID: tid}})
	}

	agent := &model.Agent{
		Name:         name,
		Description:  description,
		SystemPrompt: systemPrompt,
		ModelID:      modelID,
		Tools:        tools,
	}
	return s.repo.Create(agent)
}

func (s *AgentService) UpdateAgent(id uint, name, description, systemPrompt string, modelID uint, toolIDs []uint) error {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	
	var tools []model.Tool
	for _, tid := range toolIDs {
		tools = append(tools, model.Tool{Base: model.Base{ID: tid}})
	}

	agent.Name = name
	agent.Description = description
	agent.SystemPrompt = systemPrompt
	agent.ModelID = modelID
	agent.Tools = tools
	
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
