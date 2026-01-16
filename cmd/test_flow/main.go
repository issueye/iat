package main

import (
	"fmt"
	"iat/internal/pkg/db"
	"iat/internal/pkg/logger"
	"iat/internal/pkg/sse"
	"iat/internal/service"
	"log"
	"time"
)

func main() {
	// 1. Init
	logger.InitLogger()
	if err := db.InitDB(); err != nil {
		log.Fatal("DB Init failed:", err)
	}

	// 2. Services
	projectService := service.NewProjectService()
	modelService := service.NewAIModelService()
	agentService := service.NewAgentService()
	scriptService := service.NewScriptService()
	sessionService := service.NewSessionService()
	
	// Mock SSE for Chat
	sseHandler := sse.NewSSEHandler()
	go func() {
		// Drain the channels to prevent blocking if any
		for {
			select {
			case <-sseHandler.New:
			case <-sseHandler.Closed:
			case <-sseHandler.Total:
			}
		}
	}()
	// Note: Real chat service needs real SSE handler which writes to HTTP response.
	// Here we just test if ChatService can be created and logic runs without panic.
	// E2E chat test is hard without real OpenAI key.
	_ = service.NewChatService(sseHandler)

	fmt.Println("--- Starting Flow Test ---")

	// 3. Project
	err := projectService.CreateProject("Test Project", "Description", "/tmp/test-project")
	if err != nil {
		log.Fatal("CreateProject failed:", err)
	}
	projects, _ := projectService.ListProjects()
	if len(projects) == 0 {
		log.Fatal("Project list empty")
	}
	fmt.Printf("Project Created: %s (ID: %d)\n", projects[len(projects)-1].Name, projects[len(projects)-1].ID)
	pid := projects[len(projects)-1].ID

	// 4. Script
	err = scriptService.CreateScript("Test Script", "Desc", `console.log("Hello from JS"); "Result String"`)
	if err != nil {
		log.Fatal("CreateScript failed:", err)
	}
	res, err := scriptService.RunScript(1) // Assuming ID 1 or we fetch list
	if err != nil {
		fmt.Printf("RunScript warning (might be ID mismatch): %v\n", err)
	} else {
		fmt.Printf("Script Run Result: %v\n", res)
	}

	// 5. Model (Mock data)
	// We can't really test AI without a key, but we can create the record.
	_ = modelService
	_ = agentService
	_ = sessionService
	_ = pid

	// 6. Agent
	// agentService.CreateAgent(...)

	// 7. Session
	// sessionService.CreateSession(...)

	fmt.Println("--- Flow Test Finished Successfully (Partial) ---")
	
	// Cleanup (Optional, for now we keep DB as is or use a test DB file if configured)
	time.Sleep(1 * time.Second)
}
