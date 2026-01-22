package main

import (
	"log"
	"iat/engine/api"
	"iat/engine/pkg/seeder"
	"iat/common/pkg/db"
)

func main() {
	// Initialize Database
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	
	// Seed Data
	seeder.Seed(db.DB)

	server := api.NewServer(":8080")
	log.Println("Starting Engine on :8080")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
