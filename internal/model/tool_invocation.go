package model

type ToolInvocation struct {
	Base
	SessionID  uint   `json:"sessionId" gorm:"index"`
	ToolCallID string `json:"toolCallId" gorm:"index"`
	Name       string `json:"name"`
	Arguments  string `json:"arguments"`
	Output     string `json:"output"`
	HasResult  bool   `json:"hasResult"`
	Ok         bool   `json:"ok"`
}

