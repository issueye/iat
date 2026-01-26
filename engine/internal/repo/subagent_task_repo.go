package repo

import (
	"iat/common/model"
	"iat/common/pkg/db"
)

type SubAgentTaskRepo struct{}

func NewSubAgentTaskRepo() *SubAgentTaskRepo {
	return &SubAgentTaskRepo{}
}

// Create 创建子任务
func (r *SubAgentTaskRepo) Create(task *model.SubAgentTask) error {
	return db.DB.Create(task).Error
}

// Update 更新子任务
func (r *SubAgentTaskRepo) Update(task *model.SubAgentTask) error {
	return db.DB.Save(task).Error
}

// GetByTaskID 根据 TaskID (UUID) 获取
func (r *SubAgentTaskRepo) GetByTaskID(taskID string) (*model.SubAgentTask, error) {
	var task model.SubAgentTask
	err := db.DB.Where("task_id = ?", taskID).First(&task).Error
	return &task, err
}

// UpdateStatus 更新任务状态
func (r *SubAgentTaskRepo) UpdateStatus(taskID string, status model.SubAgentTaskStatus, result, errMsg string) error {
	return db.DB.Model(&model.SubAgentTask{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"status": status,
			"result": result,
			"error":  errMsg,
		}).Error
}

// ListBySessionID 根据 SessionID 列出所有子任务
func (r *SubAgentTaskRepo) ListBySessionID(sessionID uint) ([]model.SubAgentTask, error) {
	var tasks []model.SubAgentTask
	err := db.DB.Where("session_id = ?", sessionID).Order("created_at asc").Find(&tasks).Error
	return tasks, err
}

// ListByParentTaskID 根据父任务 ID 列出子任务
func (r *SubAgentTaskRepo) ListByParentTaskID(parentTaskID string) ([]model.SubAgentTask, error) {
	var tasks []model.SubAgentTask
	err := db.DB.Where("parent_task_id = ?", parentTaskID).Order("created_at asc").Find(&tasks).Error
	return tasks, err
}

// DeleteBySessionID 删除会话下所有子任务
func (r *SubAgentTaskRepo) DeleteBySessionID(sessionID uint) error {
	return db.DB.Where("session_id = ?", sessionID).Delete(&model.SubAgentTask{}).Error
}

// ListRunningBySessionID 获取会话中正在运行的子任务
func (r *SubAgentTaskRepo) ListRunningBySessionID(sessionID uint) ([]model.SubAgentTask, error) {
	var tasks []model.SubAgentTask
	err := db.DB.Where("session_id = ? AND status IN ?", sessionID, []model.SubAgentTaskStatus{
		model.SubAgentTaskPending,
		model.SubAgentTaskRunning,
	}).Find(&tasks).Error
	return tasks, err
}
