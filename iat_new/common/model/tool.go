package model

type Tool struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // e.g., "script", "api", "function"
	Content     string `json:"content"` // Script content or API config
	Parameters  string `json:"parameters"` // JSON schema for parameters
}
