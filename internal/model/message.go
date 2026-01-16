package model

type Message struct {
	Base
	SessionID  uint   `json:"sessionId" gorm:"index"`
	Role       string `json:"role"` // system, user, assistant
	Content    string `json:"content"`
	TokenCount int    `json:"tokenCount"`
}
