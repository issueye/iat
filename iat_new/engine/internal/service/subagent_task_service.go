package service

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/common/model"
	"iat/common/pkg/chat"
	"iat/common/pkg/sse"
	"iat/engine/internal/repo"
	"sync"

	"github.com/google/uuid"
)

// SubAgentMaxDepth 子任务最大递归深度
const SubAgentMaxDepth = 3

// SubAgentTaskService 子智能体任务服务
type SubAgentTaskService struct {
	repo       *repo.SubAgentTaskRepo
	sseHandler *sse.SSEHandler
	mu         sync.RWMutex
	cancels    map[string]context.CancelFunc // taskID -> cancel
}

func NewSubAgentTaskService(sseHandler *sse.SSEHandler) *SubAgentTaskService {
	return &SubAgentTaskService{
		repo:       repo.NewSubAgentTaskRepo(),
		sseHandler: sseHandler,
		cancels:    make(map[string]context.CancelFunc),
	}
}

// CreateTask 创建子任务记录
func (s *SubAgentTaskService) CreateTask(sessionID uint, agentName, query, parentTaskID string, depth int) (*model.SubAgentTask, error) {
	if depth > SubAgentMaxDepth {
		return nil, fmt.Errorf("sub-agent recursion depth exceeded (max: %d)", SubAgentMaxDepth)
	}

	task := &model.SubAgentTask{
		TaskID:       uuid.New().String(),
		SessionID:    sessionID,
		ParentTaskID: parentTaskID,
		AgentName:    agentName,
		Query:        query,
		Status:       model.SubAgentTaskPending,
		Depth:        depth,
	}
	if err := s.repo.Create(task); err != nil {
		return nil, err
	}
	s.notifySubAgentEvent(sessionID, "subagent_start", map[string]interface{}{
		"taskId":       task.TaskID,
		"agentName":    agentName,
		"parentTaskId": parentTaskID,
		"depth":        depth,
	})
	return task, nil
}

// UpdateStatus 更新任务状态
func (s *SubAgentTaskService) UpdateStatus(taskID string, status model.SubAgentTaskStatus, result, errMsg string) error {
	task, err := s.repo.GetByTaskID(taskID)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateStatus(taskID, status, result, errMsg); err != nil {
		return err
	}
	eventType := "subagent_update"
	if status == model.SubAgentTaskCompleted || status == model.SubAgentTaskFailed || status == model.SubAgentTaskAborted {
		eventType = "subagent_done"
	}
	s.notifySubAgentEvent(task.SessionID, eventType, map[string]interface{}{
		"taskId": taskID,
		"status": status,
		"result": result,
		"error":  errMsg,
	})
	return nil
}

// GetTask 获取任务详情
func (s *SubAgentTaskService) GetTask(taskID string) (*model.SubAgentTask, error) {
	return s.repo.GetByTaskID(taskID)
}

// ListBySessionID 列出会话所有子任务
func (s *SubAgentTaskService) ListBySessionID(sessionID uint) ([]model.SubAgentTask, error) {
	return s.repo.ListBySessionID(sessionID)
}

// RegisterCancel 注册取消函数
func (s *SubAgentTaskService) RegisterCancel(taskID string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancels[taskID] = cancel
}

// UnregisterCancel 移除取消函数
func (s *SubAgentTaskService) UnregisterCancel(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cancels, taskID)
}

// AbortTask 中止任务
func (s *SubAgentTaskService) AbortTask(taskID string) error {
	s.mu.Lock()
	cancel, ok := s.cancels[taskID]
	if ok {
		delete(s.cancels, taskID)
	}
	s.mu.Unlock()

	if ok && cancel != nil {
		cancel()
	}
	return s.UpdateStatus(taskID, model.SubAgentTaskAborted, "", "aborted by user")
}

// AbortAllBySessionID 中止会话下所有运行中的子任务
func (s *SubAgentTaskService) AbortAllBySessionID(sessionID uint) error {
	tasks, err := s.repo.ListRunningBySessionID(sessionID)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		_ = s.AbortTask(t.TaskID)
	}
	return nil
}

// SendChunk 发送子任务中间输出
func (s *SubAgentTaskService) SendChunk(sessionID uint, taskID, content string, eventChan chan<- chat.ChatEvent) {
	if eventChan != nil {
		eventChan <- chat.ChatEvent{
			Type: chat.ChatEventToolCall,
			Extra: map[string]interface{}{
				"stage":   "subagent_chunk",
				"taskId":  taskID,
				"content": content,
			},
		}
	}
	s.notifySubAgentEvent(sessionID, "subagent_chunk", map[string]interface{}{
		"taskId":  taskID,
		"content": content,
	})
}

func (s *SubAgentTaskService) notifySubAgentEvent(sessionID uint, eventType string, data map[string]interface{}) {
	if s.sseHandler == nil {
		return
	}
	data["sessionId"] = sessionID
	data["eventType"] = eventType
	msg, _ := json.Marshal(data)
	s.sseHandler.Send(string(msg))
}
