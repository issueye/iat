package model

type Agent struct {
	Base
	Name         string `json:"name"`
	Description  string `json:"description"`
	SystemPrompt string `json:"systemPrompt"`
	ModelID      uint   `json:"modelId"`
	Model        AIModel `json:"-"`
}
