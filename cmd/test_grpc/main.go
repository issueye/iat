package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	pb "iat/pkg/pb/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)

	// Create a test request
	// Note: You need a valid SessionID, AgentID usually. 
	// For this test, we might expect an error if IDs are invalid, but that confirms connectivity.
	req := &pb.ChatRequest{
		SessionId:   1, // Assuming session 1 exists or will be handled gracefully
		UserMessage: "Hello from gRPC test client",
		AgentId:     1, // Assuming agent 1 exists
		Mode:        "chat",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Sending Chat Request...")
	stream, err := client.Chat(ctx, req)
	if err != nil {
		log.Fatalf("Error calling Chat: %v", err)
	}

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error receiving stream: %v", err)
		}

		fmt.Printf("Received Event: Type=%s Content=%s Extra=%s\n", event.Type, event.Content, event.Extra)
	}

	fmt.Println("Stream finished.")
}
