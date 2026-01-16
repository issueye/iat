package repo

import (
	"iat/internal/model"
	"iat/internal/pkg/db"
)

type AgentRepo struct{}

func NewAgentRepo() *AgentRepo {
	return &AgentRepo{}
}

func (r *AgentRepo) Create(a *model.Agent) error {
	return db.DB.Create(a).Error
}

func (r *AgentRepo) Update(a *model.Agent) error {
	return db.DB.Save(a).Error
}

func (r *AgentRepo) Delete(id uint) error {
	return db.DB.Delete(&model.Agent{}, id).Error
}

func (r *AgentRepo) List() ([]model.Agent, error) {
	var agents []model.Agent
	err := db.DB.Preload("Model").Find(&agents).Error
	return agents, err
}

func (r *AgentRepo) GetByID(id uint) (*model.Agent, error) {
	var a model.Agent
	err := db.DB.Preload("Model").First(&a, id).Error
	return &a, err
}
