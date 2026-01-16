package db

import (
	"iat/internal/model"
	"iat/internal/pkg/consts"
	"iat/internal/pkg/tools/builtin"
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
		&model.AIModel{},
		&model.Script{},
		&model.Agent{},
		&model.Tool{},
	)
	if err != nil {
		return err
	}

	DB = db

	// Seed Builtin Tools
	seedBuiltinTools(db)

	// Seed Builtin Agents
	seedBuiltinAgents(db)

	log.Println("Database initialized at", dbPath)
	return nil
}

func seedBuiltinTools(db *gorm.DB) {
	for _, tool := range builtin.BuiltinTools {
		var count int64
		db.Model(&model.Tool{}).Where("name = ? AND type = ?", tool.Name, consts.ToolTypeBuiltin).Count(&count)
		if count == 0 {
			db.Create(&tool)
			log.Printf("Seeded builtin tool: %s", tool.Name)
		}
	}
}

func seedBuiltinAgents(db *gorm.DB) {
	agents := []model.Agent{
		{
			Name:        consts.AgentNameChat,
			Description: "A helpful AI assistant for general conversation.",
			Type:        consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptChat,
		},
		{
			Name:        consts.AgentNamePlan,
			Description: "A planning expert that helps breakdown complex tasks.",
			Type:        consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptPlan,
		},
		{
			Name:        consts.AgentNameBuild,
			Description: "A coding and build automation expert.",
			Type:        consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptBuild,
		},
	}

	for _, agent := range agents {
		var count int64
		db.Model(&model.Agent{}).Where("name = ? AND type = ?", agent.Name, consts.AgentTypeBuiltin).Count(&count)
		if count == 0 {
			// Try to assign a default model if exists
			var firstModel model.AIModel
			if err := db.First(&firstModel).Error; err == nil {
				agent.ModelID = firstModel.ID
			}
			db.Create(&agent)
			log.Printf("Seeded builtin agent: %s", agent.Name)
		}
	}
}
