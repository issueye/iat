package repo

import (
	"iat/internal/model"
	"iat/internal/pkg/db"
)

type MessageRepo struct{}

func NewMessageRepo() *MessageRepo {
	return &MessageRepo{}
}

func (r *MessageRepo) Create(m *model.Message) error {
	return db.DB.Create(m).Error
}

func (r *MessageRepo) ListBySessionID(sessionID uint) ([]model.Message, error) {
	var messages []model.Message
	err := db.DB.Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

func (r *MessageRepo) DeleteBySessionID(sessionID uint) error {
	return db.DB.Where("session_id = ?", sessionID).Delete(&model.Message{}).Error
}
