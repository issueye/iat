package main

import (
	"iat/internal/grpc"
	"iat/internal/pkg/db"
	"iat/internal/service"
	"log"
)

func main() {
	// 1. Init DB
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// 2. Init Services
	mcpService := service.NewMCPService()
	// TaskService with nil SSE handler (headless)
	taskService := service.NewTaskService(nil)
	
	chatService := service.NewChatService(mcpService, taskService)

	// 3. Init gRPC Server
	server := grpc.NewServer(50051, chatService)

	// 4. Start
	log.Println("Starting Engine on port 50051...")
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
