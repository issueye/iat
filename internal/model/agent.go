package model

type Agent struct {
	Base
	Name         string `json:"name"`
	Description  string `json:"description"`
	SystemPrompt string `json:"systemPrompt"`
	Type         string `json:"type"` // "builtin" or "custom"
	ModeID       uint   `json:"modeId"`
	Mode         Mode   `json:"mode"`
	ModelID      uint   `json:"modelId"`
	Model        AIModel `json:"-"`
	Tools        []Tool  `json:"tools" gorm:"many2many:agent_tools;"`
}
