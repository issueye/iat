package db

import (
	"iat/internal/model"
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
		db.Model(&model.Tool{}).Where("name = ? AND type = ?", tool.Name, "builtin").Count(&count)
		if count == 0 {
			db.Create(&tool)
			log.Printf("Seeded builtin tool: %s", tool.Name)
		}
	}
}

func seedBuiltinAgents(db *gorm.DB) {
	agents := []model.Agent{
		{
			Name:        "Chat",
			Description: "A helpful AI assistant for general conversation.",
			Type:        "builtin",
			SystemPrompt: `You are a helpful, intelligent assistant. 
Your goal is to provide accurate, concise, and useful information to the user.
You can help with a wide range of tasks, including answering questions, explaining concepts, and engaging in general conversation.`,
		},
		{
			Name:        "Plan",
			Description: "A planning expert that helps breakdown complex tasks.",
			Type:        "builtin",
			SystemPrompt: `You are an expert planner and project manager.
Your goal is to help the user break down complex problems or tasks into manageable steps.
When presented with a goal, you should:
1. Analyze the requirements.
2. Identify dependencies and potential challenges.
3. Create a structured, step-by-step plan.
4. Suggest tools or resources needed for each step.
Output the plan in a clear, Markdown-formatted list.`,
		},
		{
			Name:        "Build",
			Description: "A coding and build automation expert.",
			Type:        "builtin",
			SystemPrompt: `You are an expert software engineer and build automation specialist.
Your capabilities include writing code, debugging, and generating build scripts.
You should:
1. Write clean, efficient, and well-documented code.
2. Follow best practices for the language or framework being used.
3. Provide complete, runnable code snippets.
4. Explain your code logic clearly.
When asked to build something, consider the environment, dependencies, and execution steps.`,
		},
	}

	for _, agent := range agents {
		var count int64
		db.Model(&model.Agent{}).Where("name = ? AND type = ?", agent.Name, "builtin").Count(&count)
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
