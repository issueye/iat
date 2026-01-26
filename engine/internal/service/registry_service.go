package service

import (
	"encoding/json"
	"fmt"
	"iat/common/model"
	"iat/common/protocol"
	"iat/engine/internal/repo"
	"sync"
	"time"
)

type RegistryService struct {
	agentRepo *repo.AgentRepo
	mu        sync.RWMutex
	// Cache for online agents to avoid frequent DB lookups
	onlineAgents map[uint]*model.Agent
}

func NewRegistryService() *RegistryService {
	return &RegistryService{
		agentRepo:    repo.NewAgentRepo(),
		onlineAgents: make(map[uint]*model.Agent),
	}
}

// Register an agent (local or remote)
func (s *RegistryService) Register(agentID uint, capabilities []protocol.Capability, endpoint string) error {
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	capJSON, _ := json.Marshal(capabilities)
	agent.Capabilities = string(capJSON)
	agent.Status = "online"
	agent.LastHeartbeat = time.Now().Unix()
	if endpoint != "" {
		agent.ExternalURL = endpoint
	}

	err = s.agentRepo.Update(agent)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.onlineAgents[agent.ID] = agent
	s.mu.Unlock()

	return nil
}

// Heartbeat from an agent
func (s *RegistryService) Heartbeat(agentID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, ok := s.onlineAgents[agentID]
	if !ok {
		var err error
		agent, err = s.agentRepo.GetByID(agentID)
		if err != nil {
			return err
		}
	}

	agent.LastHeartbeat = time.Now().Unix()
	agent.Status = "online"
	s.onlineAgents[agentID] = agent

	return s.agentRepo.Update(agent)
}

// Discover agents by capability
func (s *RegistryService) Discover(capabilityName string) ([]model.Agent, error) {
	agents, err := s.agentRepo.List()
	if err != nil {
		return nil, err
	}

	var matched []model.Agent
	for _, a := range agents {
		if a.Status != "online" {
			continue
		}

		var caps []protocol.Capability
		if err := json.Unmarshal([]byte(a.Capabilities), &caps); err != nil {
			continue
		}

		for _, c := range caps {
			if c.Name == capabilityName {
				matched = append(matched, a)
				break
			}
		}
	}

	return matched, nil
}

// Cleanup offline agents (call periodically)
func (s *RegistryService) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	timeout := int64(60) // 60 seconds timeout

	for id, agent := range s.onlineAgents {
		if now-agent.LastHeartbeat > timeout {
			agent.Status = "offline"
			s.agentRepo.Update(agent)
			delete(s.onlineAgents, id)
		}
	}
}
