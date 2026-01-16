package model

type Agent struct {
	Base
	Name         string `json:"name"`
	Description  string `json:"description"`
	SystemPrompt string `json:"systemPrompt"`
	Type         string `json:"type"` // "builtin" or "custom"
	ModelID      uint   `json:"modelId"`
	Model        AIModel `json:"-"`
}
