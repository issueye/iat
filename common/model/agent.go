package model

type Agent struct {
	Base
	Name         string `json:"name"`
	Description  string `json:"description"`
	SystemPrompt string `json:"systemPrompt"`
	Type         string `json:"type"`
	Modes        []Mode `json:"modes" gorm:"many2many:agent_modes;"`
	ModelID      uint   `json:"modelId"`
	Model        AIModel `json:"-"`
	Tools        []Tool  `json:"tools" gorm:"many2many:agent_tools;"`
	MCPServers   []MCPServer `json:"mcpServers" gorm:"many2many:agent_mcp_servers;"`
	ExternalURL  string `json:"externalUrl"`
	ExternalType string `json:"externalType"`
	ExternalParams string `json:"externalParams" gorm:"type:text"`
	Status         string `json:"status" gorm:"default:'offline'"` // online, offline, busy
	Capabilities   string `json:"capabilities" gorm:"type:text"`   // JSON array of Capability
	ConfigSchema   string `json:"configSchema" gorm:"type:text"`   // JSON Schema for agent-specific config
	MemoryPolicy   string `json:"memoryPolicy" gorm:"type:text"`   // JSON for memory retention/sharing policy
	LastHeartbeat  int64  `json:"lastHeartbeat"`
}
