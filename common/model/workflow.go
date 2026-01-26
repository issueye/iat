package model

import (
	"encoding/json"
	"time"
)

type WorkflowStatus string

const (
	WorkflowPending   WorkflowStatus = "pending"
	WorkflowRunning   WorkflowStatus = "running"
	WorkflowCompleted WorkflowStatus = "completed"
	WorkflowFailed    WorkflowStatus = "failed"
	WorkflowCancelled WorkflowStatus = "cancelled"
)

type Workflow struct {
	Base
	SessionID   uint           `json:"sessionId" gorm:"index"`
	Goal        string         `json:"goal" gorm:"type:text"`
	Status      WorkflowStatus `json:"status" gorm:"default:'pending'"`
	TaskTree    string         `json:"taskTree" gorm:"type:text"` // JSON representation of the DAG
	Result      string         `json:"result" gorm:"type:text"`
	StartedAt   *time.Time     `json:"startedAt"`
	CompletedAt *time.Time     `json:"completedAt"`
}

type WorkflowTask struct {
	Base
	WorkflowID  uint       `json:"workflowId" gorm:"index"`
	TaskID      string     `json:"taskId" gorm:"index"` // ID from the Planner
	Title       string     `json:"title"`
	Description string     `json:"description" gorm:"type:text"`
	Capability  string     `json:"capability"`
	AgentID     uint       `json:"agentId"`
	Status      TaskStatus `json:"status" gorm:"default:'pending'"`
	Input       string     `json:"input" gorm:"type:text"`
	Output      string     `json:"output" gorm:"type:text"`
	Error       string     `json:"error" gorm:"type:text"`
	DependsOn   string     `json:"dependsOn" gorm:"type:text"` // JSON array of task IDs
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

type SubTask struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DependsOn   []string `json:"dependsOn,omitempty"`
	Capability  string   `json:"capability"`
}

type TaskTree struct {
	Goal  string    `json:"goal"`
	Tasks []SubTask `json:"tasks"`
}

func (w *Workflow) GetTaskTree() (map[string]any, error) {
	var tree map[string]any
	err := json.Unmarshal([]byte(w.TaskTree), &tree)
	return tree, err
}

func (w *Workflow) SetTaskTree(tree map[string]any) error {
	data, err := json.Marshal(tree)
	if err != nil {
		return err
	}
	w.TaskTree = string(data)
	return nil
}
