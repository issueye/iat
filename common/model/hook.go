package model

type Hook struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`       // e.g., "pre_run", "post_run", "pre_tool", "post_tool"
	TargetType  string `json:"targetType"` // e.g., "agent", "global"
	TargetID    uint   `json:"targetId"`   // Agent ID if targetType is agent
	Action      string `json:"action"`     // e.g., "script", "http"
	Content     string `json:"content"`    // Script content or URL
	Enabled     bool   `json:"enabled" gorm:"default:true"`
}
