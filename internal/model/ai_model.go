package model

type AIModel struct {
	Base
	Name       string `json:"name"`
	Provider   string `json:"provider"` // openai, deepseek, ollama
	BaseURL    string `json:"baseUrl"`
	APIKey     string `json:"apiKey"`
	ConfigJSON string `json:"configJson"` // Extra config
	IsDefault  bool   `json:"isDefault"`
}
