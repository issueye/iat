package repo

import (
	"iat/internal/model"
	"iat/internal/pkg/db"
)

type ModeRepo struct{}

func NewModeRepo() *ModeRepo {
	return &ModeRepo{}
}

func (r *ModeRepo) Create(m *model.Mode) error {
	return db.DB.Create(m).Error
}

func (r *ModeRepo) Update(m *model.Mode) error {
	return db.DB.Save(m).Error
}

func (r *ModeRepo) Delete(id uint) error {
	return db.DB.Delete(&model.Mode{}, id).Error
}

func (r *ModeRepo) List() ([]model.Mode, error) {
	var modes []model.Mode
	err := db.DB.Find(&modes).Error
	return modes, err
}

func (r *ModeRepo) GetByID(id uint) (*model.Mode, error) {
	var m model.Mode
	err := db.DB.First(&m, id).Error
	return &m, err
}

func (r *ModeRepo) GetByKey(key string) (*model.Mode, error) {
	var m model.Mode
	err := db.DB.Where("key = ?", key).First(&m).Error
	return &m, err
}
