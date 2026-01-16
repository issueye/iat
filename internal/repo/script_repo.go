package repo

import (
	"iat/internal/model"
	"iat/internal/pkg/db"
)

type ScriptRepo struct{}

func NewScriptRepo() *ScriptRepo {
	return &ScriptRepo{}
}

func (r *ScriptRepo) Create(s *model.Script) error {
	return db.DB.Create(s).Error
}

func (r *ScriptRepo) Update(s *model.Script) error {
	return db.DB.Save(s).Error
}

func (r *ScriptRepo) Delete(id uint) error {
	return db.DB.Delete(&model.Script{}, id).Error
}

func (r *ScriptRepo) List() ([]model.Script, error) {
	var scripts []model.Script
	err := db.DB.Find(&scripts).Error
	return scripts, err
}

func (r *ScriptRepo) GetByID(id uint) (*model.Script, error) {
	var s model.Script
	err := db.DB.First(&s, id).Error
	return &s, err
}
