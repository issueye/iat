package repo

import (
	"iat/internal/model"
	"iat/internal/pkg/db"
)

type TaskRepo struct{}

func NewTaskRepo() *TaskRepo {
	return &TaskRepo{}
}

func (r *TaskRepo) Create(task *model.Task) error {
	return db.DB.Create(task).Error
}

func (r *TaskRepo) Update(task *model.Task) error {
	return db.DB.Save(task).Error
}

func (r *TaskRepo) Delete(id uint) error {
	return db.DB.Delete(&model.Task{}, id).Error
}

func (r *TaskRepo) GetByID(id uint) (*model.Task, error) {
	var task model.Task
	err := db.DB.First(&task, id).Error
	return &task, err
}

func (r *TaskRepo) ListBySessionID(sessionID uint) ([]model.Task, error) {
	var tasks []model.Task
	err := db.DB.Where("session_id = ?", sessionID).Order("created_at asc").Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepo) DeleteBySessionID(sessionID uint) error {
	return db.DB.Where("session_id = ?", sessionID).Delete(&model.Task{}).Error
}
