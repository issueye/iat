package model

type MCPServer struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`    // "stdio", "sse"
	Command     string `json:"command"` // for stdio
	Args        string `json:"args"`    // for stdio (JSON array of strings)
	Env         string `json:"env"`     // for stdio (JSON map[string]string)
	Url         string `json:"url"`     // for sse
	Enabled     bool   `json:"enabled"`
}
