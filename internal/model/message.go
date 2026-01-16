package model

type Message struct {
	Base
	SessionID  uint   `json:"sessionId" gorm:"index"`
	Role       string `json:"role"` // consts.RoleSystem, consts.RoleUser, consts.RoleAssistant, consts.RoleTool
	Content    string `json:"content"`
	TokenCount int    `json:"tokenCount"`
	Prompt     string `json:"prompt" gorm:"type:longtext"` // The full prompt sent to AI for this message
}
