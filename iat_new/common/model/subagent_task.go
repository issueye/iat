package model

// SubAgentTaskStatus 子任务状态
type SubAgentTaskStatus string

const (
	SubAgentTaskPending   SubAgentTaskStatus = "pending"
	SubAgentTaskRunning   SubAgentTaskStatus = "running"
	SubAgentTaskCompleted SubAgentTaskStatus = "completed"
	SubAgentTaskFailed    SubAgentTaskStatus = "failed"
	SubAgentTaskAborted   SubAgentTaskStatus = "aborted"
)

// SubAgentTask 子智能体异步任务模型
type SubAgentTask struct {
	Base
	TaskID       string             `json:"taskId" gorm:"uniqueIndex;size:36"` // UUID
	SessionID    uint               `json:"sessionId" gorm:"index"`
	ParentTaskID string             `json:"parentTaskId" gorm:"index;size:36"` // 父任务 ID，根任务为空
	AgentName    string             `json:"agentName"`
	Query        string             `json:"query" gorm:"type:text"`
	Status       SubAgentTaskStatus `json:"status" gorm:"default:'pending'"`
	Depth        int                `json:"depth" gorm:"default:0"` // 当前递归深度
	Result       string             `json:"result" gorm:"type:text"`
	Error        string             `json:"error" gorm:"type:text"`
}
