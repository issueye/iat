package repo

import (
	"iat/common/model"
	"iat/common/pkg/db"
)

type ToolRepo struct{}

func NewToolRepo() *ToolRepo {
	return &ToolRepo{}
}

func (r *ToolRepo) Create(tool *model.Tool) error {
	return db.DB.Create(tool).Error
}

func (r *ToolRepo) Update(tool *model.Tool) error {
	return db.DB.Save(tool).Error
}

func (r *ToolRepo) Delete(id uint) error {
	return db.DB.Delete(&model.Tool{}, id).Error
}

func (r *ToolRepo) Get(id uint) (*model.Tool, error) {
	var tool model.Tool
	err := db.DB.First(&tool, id).Error
	return &tool, err
}

func (r *ToolRepo) List() ([]model.Tool, error) {
	var tools []model.Tool
	err := db.DB.Order("created_at desc").Find(&tools).Error
	return tools, err
}
