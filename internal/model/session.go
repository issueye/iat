package model

type Session struct {
	Base
	ProjectID uint    `json:"projectId"`
	Project   Project `json:"-"`
	Name      string  `json:"name"`
	AgentID   uint    `json:"agentId"`
	Agent     Agent   `json:"-"`
	Compressed bool   `json:"compressed"`
	Summary    string `json:"summary"`
}
