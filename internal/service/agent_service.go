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

func (s *AgentService) CreateAgent(name, description, systemPrompt string, modelID uint, toolIDs []uint, mcpServerIDs []uint, modeID uint) error {
	var tools []model.Tool
	for _, tid := range toolIDs {
		tools = append(tools, model.Tool{Base: model.Base{ID: tid}})
	}

	var mcpServers []model.MCPServer
	for _, mid := range mcpServerIDs {
		mcpServers = append(mcpServers, model.MCPServer{Base: model.Base{ID: mid}})
	}

	agent := &model.Agent{
		Name:         name,
		Description:  description,
		SystemPrompt: systemPrompt,
		ModelID:      modelID,
		Tools:        tools,
		MCPServers:   mcpServers,
		ModeID:       modeID,
	}
	return s.repo.Create(agent)
}

func (s *AgentService) UpdateAgent(id uint, name, description, systemPrompt string, modelID uint, toolIDs []uint, mcpServerIDs []uint, modeID uint) error {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	
	var tools []model.Tool
	for _, tid := range toolIDs {
		tools = append(tools, model.Tool{Base: model.Base{ID: tid}})
	}

	var mcpServers []model.MCPServer
	for _, mid := range mcpServerIDs {
		mcpServers = append(mcpServers, model.MCPServer{Base: model.Base{ID: mid}})
	}

	agent.Name = name
	agent.Description = description
	agent.SystemPrompt = systemPrompt
	agent.ModelID = modelID
	agent.Tools = tools
	agent.MCPServers = mcpServers
	agent.ModeID = modeID
	
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
