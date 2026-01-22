package service

import (
	"iat/common/model"
	"iat/engine/internal/repo"
)

type SessionService struct {
	repo *repo.SessionRepo
}

func NewSessionService() *SessionService {
	return &SessionService{
		repo: repo.NewSessionRepo(),
	}
}

func (s *SessionService) CreateSession(projectID uint, name string, agentID uint) error {
	session := &model.Session{
		ProjectID: projectID,
		Name:      name,
		AgentID:   agentID,
	}
	return s.repo.Create(session)
}

func (s *SessionService) ListSessions(projectID uint) ([]model.Session, error) {
	return s.repo.ListByProjectID(projectID)
}

func (s *SessionService) DeleteSession(id uint) error {
	return s.repo.Delete(id)
}
