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

func (s *SessionService) ListSessions(projectID uint) ([]model.Session, error) {
	return s.repo.ListByProjectID(projectID)
}

func (s *SessionService) CreateSession(name string, projectID uint, agentID uint) (*model.Session, error) {
	session := &model.Session{
		Name:      name,
		ProjectID: projectID,
		AgentID:   agentID,
	}
	if err := s.repo.Create(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionService) UpdateSession(id uint, name string) error {
	session, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	session.Name = name
	return s.repo.Update(session)
}

func (s *SessionService) DeleteSession(id uint) error {
	return s.repo.Delete(id)
}

func (s *SessionService) GetSession(id uint) (*model.Session, error) {
	return s.repo.GetByID(id)
}
