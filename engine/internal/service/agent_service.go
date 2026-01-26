package service

import (
	"fmt"
	"iat/common/model"
	"iat/engine/internal/repo"
)

type AgentService struct {
	repo *repo.AgentRepo
}

func NewAgentService() *AgentService {
	return &AgentService{
		repo: repo.NewAgentRepo(),
	}
}

func (s *AgentService) CreateAgent(name, description, systemPrompt, agentType, externalURL, externalType, externalParams string, modelID uint, toolIDs []uint, mcpServerIDs []uint, modeIDs []uint, status string, capabilities string) error {
	var tools []model.Tool
	for _, tid := range toolIDs {
		tools = append(tools, model.Tool{Base: model.Base{ID: tid}})
	}

	var mcpServers []model.MCPServer
	for _, mid := range mcpServerIDs {
		mcpServers = append(mcpServers, model.MCPServer{Base: model.Base{ID: mid}})
	}

	var modes []model.Mode
	for _, mid := range modeIDs {
		modes = append(modes, model.Mode{Base: model.Base{ID: mid}})
	}

	if agentType == "" {
		agentType = "custom"
	}

	agent := &model.Agent{
		Name:           name,
		Description:    description,
		SystemPrompt:   systemPrompt,
		Type:           agentType,
		ModelID:        modelID,
		Tools:          tools,
		MCPServers:     mcpServers,
		Modes:          modes,
		ExternalURL:    externalURL,
		ExternalType:   externalType,
		ExternalParams: externalParams,
		Status:         status,
		Capabilities:   capabilities,
	}
	return s.repo.Create(agent)
}

func (s *AgentService) UpdateAgent(id uint, name, description, systemPrompt, agentType, externalURL, externalType, externalParams string, modelID uint, toolIDs []uint, mcpServerIDs []uint, modeIDs []uint, status string, capabilities string) error {
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

	var modes []model.Mode
	for _, mid := range modeIDs {
		modes = append(modes, model.Mode{Base: model.Base{ID: mid}})
	}

	if agentType != "" {
		agent.Type = agentType
	}

	agent.Name = name
	agent.Description = description
	agent.SystemPrompt = systemPrompt
	agent.ModelID = modelID
	agent.Tools = tools
	agent.MCPServers = mcpServers
	agent.Modes = modes
	agent.ExternalURL = externalURL
	agent.ExternalType = externalType
	agent.ExternalParams = externalParams
	
	if status != "" {
		agent.Status = status
	}
	if capabilities != "" {
		agent.Capabilities = capabilities
	}
	
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
