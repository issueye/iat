package repo

import (
	"iat/common/model"
	"iat/common/pkg/db"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() {
	d, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	d.AutoMigrate(&model.Message{})
	db.DB = d
}

func TestMessageRepo_DeleteByID(t *testing.T) {
	setupTestDB()
	repo := NewMessageRepo()

	msg := &model.Message{
		SessionID: 1,
		Content:   "test message",
	}
	repo.Create(msg)

	if msg.ID == 0 {
		t.Fatalf("expected msg ID to be set")
	}

	err := repo.DeleteByID(msg.ID)
	if err != nil {
		t.Fatalf("failed to delete message: %v", err)
	}

	var count int64
	db.DB.Model(&model.Message{}).Where("id = ?", msg.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected message to be deleted, found %d", count)
	}
}
