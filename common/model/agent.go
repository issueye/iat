package model

type Agent struct {
	Base
	Name         string `json:"name"`
	Description  string `json:"description"`
	SystemPrompt string `json:"systemPrompt"`
	Type         string `json:"type"`
	ModeID       uint   `json:"modeId"`
	Mode         Mode   `json:"mode"`
	ModelID      uint   `json:"modelId"`
	Model        AIModel `json:"-"`
	Tools        []Tool  `json:"tools" gorm:"many2many:agent_tools;"`
	MCPServers   []MCPServer `json:"mcpServers" gorm:"many2many:agent_mcp_servers;"`
	ExternalURL  string `json:"externalUrl"`
	ExternalType string `json:"externalType"`
	ExternalParams string `json:"externalParams" gorm:"type:text"`
}
