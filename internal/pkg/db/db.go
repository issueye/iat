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
		&model.ToolInvocation{},
		&model.AIModel{},
		&model.Script{},
		&model.Agent{},
		&model.Tool{},
		&model.Mode{},
		&model.MCPServer{},
	)
	if err != nil {
		return err
	}

	DB = db

	// Seed Builtin Tools
	seedBuiltinTools(db)

	// Seed Modes
	seedModes(db)

	// Seed Builtin Agents
	seedBuiltinAgents(db)

	log.Println("Database initialized at", dbPath)
	return nil
}

func seedModes(db *gorm.DB) {
	modes := []model.Mode{
		{
			Key:          "chat",
			Name:         "Chat",
			Description:  "General conversational mode",
			SystemPrompt: consts.SystemPromptChat,
		},
		{
			Key:          "plan",
			Name:         "Plan",
			Description:  "Planning mode with restricted file access",
			SystemPrompt: consts.SystemPromptPlan,
		},
		{
			Key:          "build",
			Name:         "Build",
			Description:  "Build mode with full project access",
			SystemPrompt: consts.SystemPromptBuild,
		},
	}

	for _, mode := range modes {
		var count int64
		db.Model(&model.Mode{}).Where("key = ?", mode.Key).Count(&count)
		if count == 0 {
			db.Create(&mode)
			log.Printf("Seeded mode: %s", mode.Name)
		}
	}
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

	// Seed a sample script tool
	sampleScript := model.Tool{
		Name:        "calculate_sum",
		Description: "Calculates the sum of two numbers using a JS script",
		Type:        consts.ToolTypeScript,
		Content:     `
// args is injected by the engine
var a = args.a;
var b = args.b;
// Return result
a + b;
`,
		Parameters: `{
			"type": "object",
			"properties": {
				"a": { "type": "number", "description": "First number" },
				"b": { "type": "number", "description": "Second number" }
			},
			"required": ["a", "b"]
		}`,
	}
	var scriptCount int64
	db.Model(&model.Tool{}).Where("name = ? AND type = ?", sampleScript.Name, consts.ToolTypeScript).Count(&scriptCount)
	if scriptCount == 0 {
		db.Create(&sampleScript)
		log.Printf("Seeded sample script tool: %s", sampleScript.Name)
	}
}

func seedBuiltinAgents(db *gorm.DB) {
	// Pre-fetch modes
	var chatMode, planMode, buildMode model.Mode
	db.Where("key = ?", "chat").First(&chatMode)
	db.Where("key = ?", "plan").First(&planMode)
	db.Where("key = ?", "build").First(&buildMode)

	agents := []model.Agent{
		{
			Name:         consts.AgentNameChat,
			Description:  "A helpful AI assistant for general conversation.",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptChat,
			ModeID:       chatMode.ID,
		},
		{
			Name:         consts.AgentNamePlan,
			Description:  "A planning expert that helps breakdown complex tasks.",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptPlan,
			ModeID:       planMode.ID,
		},
		{
			Name:         consts.AgentNameBuild,
			Description:  "A coding and build automation expert.",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptBuild,
			ModeID:       buildMode.ID,
		},
		{
			Name:         consts.AgentNameProductManager,
			Description:  "Analyzes requirements and defines product features (PRD).",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptProductManager,
			ModeID:       planMode.ID, // PM fits well in Plan mode
		},
		{
			Name:         consts.AgentNameProjectManager,
			Description:  "Plans execution, manages timelines, and coordinates tasks.",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptProjectManager,
			ModeID:       planMode.ID, // Project Manager fits well in Plan mode
		},
		{
			Name:         consts.AgentNameUxUi,
			Description:  "Designs intuitive and accessible user interfaces.",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptUxUi,
			ModeID:       buildMode.ID, // Designers might need to see code/assets (Build mode) or just Plan
		},
		{
			Name:         consts.AgentNameGolang,
			Description:  "Expert in Golang backend development.",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptGolang,
			ModeID:       buildMode.ID,
		},
		{
			Name:         consts.AgentNamePython,
			Description:  "Expert in Python scripting and backend development.",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptPython,
			ModeID:       buildMode.ID,
		},
		{
			Name:         consts.AgentNameJavascript,
			Description:  "Expert in JavaScript and Frontend development.",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptJavascript,
			ModeID:       buildMode.ID,
		},
		{
			Name:         consts.AgentNameTest,
			Description:  "Ensures quality through automated and manual testing.",
			Type:         consts.AgentTypeBuiltin,
			SystemPrompt: consts.SystemPromptTest,
			ModeID:       buildMode.ID,
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
		} else {
			// Update ModeID and SystemPrompt for existing agents if they changed
			var existingAgent model.Agent
			db.Where("name = ? AND type = ?", agent.Name, consts.AgentTypeBuiltin).First(&existingAgent)
			
			needsUpdate := false
			if existingAgent.ModeID == 0 {
				existingAgent.ModeID = agent.ModeID
				needsUpdate = true
			}
			if existingAgent.SystemPrompt != agent.SystemPrompt {
				existingAgent.SystemPrompt = agent.SystemPrompt
				needsUpdate = true
			}
			
			if needsUpdate {
				db.Save(&existingAgent)
				log.Printf("Updated builtin agent: %s", agent.Name)
			}
		}
	}
}
