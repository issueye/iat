package model

type Message struct {
	Base
	SessionID uint   `json:"sessionId" gorm:"index"`
	Role      string `json:"role"`                  // consts.RoleSystem, consts.RoleUser, consts.RoleAssistant, consts.RoleTool
	Category  string `json:"category" gorm:"index"` // e.g. "tool"

	Content    string `json:"content"`
	TokenCount int    `json:"tokenCount"`
	Prompt     string `json:"prompt" gorm:"type:longtext"` // The full prompt sent to AI for this message

	// Tool message fields (when Role == consts.RoleTool)
	ToolCallID    string `json:"toolCallId" gorm:"index"`
	ToolName      string `json:"toolName"`
	ToolArgs      string `json:"toolArguments" gorm:"type:longtext"`
	ToolOutput    string `json:"toolOutput" gorm:"type:longtext"`
	ToolHasResult bool   `json:"toolHasResult"`
	ToolOk        bool   `json:"toolOk"`
}
