package repo

import (
	"iat/common/model"
	"iat/common/pkg/db"
)

type AgentRepo struct{}

func NewAgentRepo() *AgentRepo {
	return &AgentRepo{}
}

func (r *AgentRepo) Create(a *model.Agent) error {
	return db.DB.Create(a).Error
}

func (r *AgentRepo) Update(a *model.Agent) error {
	// Update associations explicitly if needed, but GORM Save should handle it if struct is set up correctly
	// However, for many2many, we often need to replace associations.
	// Let's use Association().Replace() for tools to be safe and clean.
	
	// First save the agent basic info
	if err := db.DB.Omit("Tools").Save(a).Error; err != nil {
		return err
	}
	
	// Then replace tools association
	return db.DB.Model(a).Association("Tools").Replace(a.Tools)
}

func (r *AgentRepo) Delete(id uint) error {
	// GORM handles many2many join table cleanup automatically if configured, 
	// but explicit cleanup is sometimes safer. 
	// Default behavior is usually fine for deleting the agent (join rows deleted).
	return db.DB.Delete(&model.Agent{}, id).Error
}

func (r *AgentRepo) List() ([]model.Agent, error) {
	var agents []model.Agent
	err := db.DB.Preload("Model").Preload("Tools").Preload("MCPServers").Preload("Mode").Find(&agents).Error
	return agents, err
}

func (r *AgentRepo) GetByID(id uint) (*model.Agent, error) {
	var a model.Agent
	err := db.DB.Preload("Model").Preload("Tools").Preload("MCPServers").Preload("Mode").First(&a, id).Error
	return &a, err
}
