package repo

import (
	"iat/internal/model"
	"iat/internal/pkg/db"
)

type AIModelRepo struct{}

func NewAIModelRepo() *AIModelRepo {
	return &AIModelRepo{}
}

func (r *AIModelRepo) Create(m *model.AIModel) error {
	return db.DB.Create(m).Error
}

func (r *AIModelRepo) Update(m *model.AIModel) error {
	return db.DB.Save(m).Error
}

func (r *AIModelRepo) Delete(id uint) error {
	return db.DB.Delete(&model.AIModel{}, id).Error
}

func (r *AIModelRepo) GetByID(id uint) (*model.AIModel, error) {
	var m model.AIModel
	err := db.DB.First(&m, id).Error
	return &m, err
}

func (r *AIModelRepo) GetDefault() (*model.AIModel, error) {
	var m model.AIModel
	err := db.DB.Where("is_default = ?", true).First(&m).Error
	return &m, err
}

func (r *AIModelRepo) UnsetDefault() error {
	return db.DB.Model(&model.AIModel{}).Where("is_default = ?", true).Update("is_default", false).Error
}

func (r *AIModelRepo) List() ([]model.AIModel, error) {
	var models []model.AIModel
	err := db.DB.Find(&models).Error
	return models, err
}
