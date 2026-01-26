package repo

import (
	"iat/common/model"
	"iat/common/pkg/db"
)

type WorkflowRepo struct{}

func NewWorkflowRepo() *WorkflowRepo {
	return &WorkflowRepo{}
}

func (r *WorkflowRepo) Create(w *model.Workflow) error {
	return db.DB.Create(w).Error
}

func (r *WorkflowRepo) Update(w *model.Workflow) error {
	return db.DB.Save(w).Error
}

func (r *WorkflowRepo) GetBySessionID(sessionID uint) ([]model.Workflow, error) {
	var workflows []model.Workflow
	err := db.DB.Where("session_id = ?", sessionID).Find(&workflows).Error
	return workflows, err
}

func (r *WorkflowRepo) CreateTask(t *model.WorkflowTask) error {
	return db.DB.Create(t).Error
}

func (r *WorkflowRepo) UpdateTask(t *model.WorkflowTask) error {
	return db.DB.Save(t).Error
}

func (r *WorkflowRepo) GetTasksByWorkflowID(workflowID uint) ([]model.WorkflowTask, error) {
	var tasks []model.WorkflowTask
	err := db.DB.Where("workflow_id = ?", workflowID).Find(&tasks).Error
	return tasks, err
}

func (r *WorkflowRepo) GetTaskByTaskID(workflowID uint, taskID string) (*model.WorkflowTask, error) {
	var task model.WorkflowTask
	err := db.DB.Where("workflow_id = ? AND task_id = ?", workflowID, taskID).First(&task).Error
	return &task, err
}
