package model

type Mode struct {
	Base
	Key          string `json:"key" gorm:"uniqueIndex"` // chat, plan, build
	Name         string `json:"name"`
	Description  string `json:"description"`
	SystemPrompt string `json:"systemPrompt"`
}
