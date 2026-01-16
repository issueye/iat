package repo

import (
	"iat/internal/model"
	"iat/internal/pkg/db"
)

type ProjectRepo struct{}

func NewProjectRepo() *ProjectRepo {
	return &ProjectRepo{}
}

func (r *ProjectRepo) Create(project *model.Project) error {
	return db.DB.Create(project).Error
}

func (r *ProjectRepo) Update(project *model.Project) error {
	return db.DB.Save(project).Error
}

func (r *ProjectRepo) Delete(id uint) error {
	return db.DB.Delete(&model.Project{}, id).Error
}

func (r *ProjectRepo) GetByID(id uint) (*model.Project, error) {
	var project model.Project
	err := db.DB.First(&project, id).Error
	return &project, err
}

func (r *ProjectRepo) List() ([]model.Project, error) {
	var projects []model.Project
	err := db.DB.Find(&projects).Error
	return projects, err
}

func (r *ProjectRepo) ListByIDs(ids []uint) ([]model.Project, error) {
	var projects []model.Project
	if len(ids) == 0 {
		return projects, nil
	}
	err := db.DB.Where("id IN ?", ids).Find(&projects).Error
	return projects, err
}
