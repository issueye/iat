package service

import (
	"encoding/json"
	"iat/internal/model"
	"iat/internal/pkg/sse"
	"iat/internal/repo"
)

type TaskService struct {
	repo       *repo.TaskRepo
	sseHandler *sse.SSEHandler
}

func NewTaskService(sseHandler *sse.SSEHandler) *TaskService {
	return &TaskService{
		repo:       repo.NewTaskRepo(),
		sseHandler: sseHandler,
	}
}

func (s *TaskService) ListTasks(sessionID uint) ([]model.Task, error) {
	return s.repo.ListBySessionID(sessionID)
}

func (s *TaskService) CreateTask(sessionID uint, content, priority string) (*model.Task, error) {
	task := &model.Task{
		SessionID: sessionID,
		Content:   content,
		Priority:  priority,
		Status:    model.TaskStatusPending,
	}
	if err := s.repo.Create(task); err != nil {
		return nil, err
	}
	s.notifyTasksUpdate(sessionID)
	return task, nil
}

func (s *TaskService) UpdateTask(id uint, status string) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	task.Status = model.TaskStatus(status)
	if err := s.repo.Update(task); err != nil {
		return err
	}
	s.notifyTasksUpdate(task.SessionID)
	return nil
}

func (s *TaskService) DeleteTask(id uint) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.notifyTasksUpdate(task.SessionID)
	return nil
}

func (s *TaskService) notifyTasksUpdate(sessionID uint) {
	tasks, _ := s.repo.ListBySessionID(sessionID)
	msg, _ := json.Marshal(map[string]interface{}{
		"sessionId": sessionID,
		"tasks":     tasks,
	})
	s.sseHandler.Send(string(msg))
}
