package repo

import (
	"iat/common/model"
	"iat/common/pkg/db"
)

type MCPServerRepo struct{}

func NewMCPServerRepo() *MCPServerRepo {
	return &MCPServerRepo{}
}

func (r *MCPServerRepo) Create(s *model.MCPServer) error {
	return db.DB.Create(s).Error
}

func (r *MCPServerRepo) Update(s *model.MCPServer) error {
	return db.DB.Save(s).Error
}

func (r *MCPServerRepo) Delete(id uint) error {
	return db.DB.Delete(&model.MCPServer{}, id).Error
}

func (r *MCPServerRepo) GetByID(id uint) (*model.MCPServer, error) {
	var s model.MCPServer
	err := db.DB.First(&s, id).Error
	return &s, err
}

func (r *MCPServerRepo) List() ([]model.MCPServer, error) {
	var list []model.MCPServer
	err := db.DB.Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *MCPServerRepo) ListEnabled() ([]model.MCPServer, error) {
	var list []model.MCPServer
	err := db.DB.Where("enabled = ?", true).Order("created_at desc").Find(&list).Error
	return list, err
}
