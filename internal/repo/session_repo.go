package repo

import (
	"iat/internal/model"
	"iat/internal/pkg/db"
)

type SessionRepo struct{}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{}
}

func (r *SessionRepo) Create(s *model.Session) error {
	return db.DB.Create(s).Error
}

func (r *SessionRepo) Update(s *model.Session) error {
	return db.DB.Save(s).Error
}

func (r *SessionRepo) Delete(id uint) error {
	return db.DB.Delete(&model.Session{}, id).Error
}

func (r *SessionRepo) ListByProjectID(projectID uint) ([]model.Session, error) {
	var sessions []model.Session
	err := db.DB.Where("project_id = ?", projectID).Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepo) GetByID(id uint) (*model.Session, error) {
	var s model.Session
	err := db.DB.First(&s, id).Error
	return &s, err
}
