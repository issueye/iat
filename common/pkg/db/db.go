package db

import (
	"iat/common/model"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	// Ensure data directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(homeDir, ".iat")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return err
	}
	dbPath := filepath.Join(appDir, "iat.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	// Auto Migrate
	err = db.AutoMigrate(
		&model.Project{},
		&model.Session{},
		&model.Message{},
		&model.ToolInvocation{},
		&model.AIModel{},
		&model.Script{},
		&model.Agent{},
		&model.Tool{},
		&model.Mode{},
		&model.MCPServer{},
		&model.Task{},
		&model.SubAgentTask{},
		&model.Workflow{},
		&model.WorkflowTask{},
		&model.Hook{},
	)
	if err != nil {
		return err
	}

	DB = db
	log.Println("Database initialized at", dbPath)
	return nil
}
