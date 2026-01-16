package service

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/internal/model"
	"iat/internal/pkg/ai"
	"iat/internal/pkg/sse"
	"iat/internal/repo"

	"github.com/cloudwego/eino/schema"
)

type ChatService struct {
	sessionRepo *repo.SessionRepo
	agentRepo   *repo.AgentRepo
	modelRepo   *repo.AIModelRepo
	messageRepo *repo.MessageRepo
	sseHandler  *sse.SSEHandler
}

func NewChatService(sseHandler *sse.SSEHandler) *ChatService {
	return &ChatService{
		sessionRepo: repo.NewSessionRepo(),
		agentRepo:   repo.NewAgentRepo(),
		modelRepo:   repo.NewAIModelRepo(),
		messageRepo: repo.NewMessageRepo(),
		sseHandler:  sseHandler,
	}
}

// ListMessages returns history messages for a session
func (s *ChatService) ListMessages(sessionID uint) ([]model.Message, error) {
	return s.messageRepo.ListBySessionID(sessionID)
}

// Chat handles the main chat logic
func (s *ChatService) Chat(sessionID uint, userMessage string) error {
	// 1. Get Session
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %v", err)
	}

	// 2. Get Agent
	if session.AgentID == 0 {
		return fmt.Errorf("session has no agent assigned")
	}
	agent, err := s.agentRepo.GetByID(session.AgentID)
	if err != nil {
		return fmt.Errorf("agent not found: %v", err)
	}

	// 3. Get Model Config
	modelConfig, err := s.modelRepo.GetByID(agent.ModelID)
	if err != nil {
		return fmt.Errorf("model config not found: %v", err)
	}

	// 4. Init AI Client
	aiClient, err := ai.NewAIClient(modelConfig)
	if err != nil {
		return fmt.Errorf("failed to init ai client: %v", err)
	}

	// 5. Construct Messages (System + User)
	// Save User Message
	userMsg := &model.Message{
		SessionID: sessionID,
		Role:      "user",
		Content:   userMessage,
	}
	if err := s.messageRepo.Create(userMsg); err != nil {
		return fmt.Errorf("failed to save user message: %v", err)
	}

	// Load History (including the one just saved, but we need structure for AI client)
	history, err := s.messageRepo.ListBySessionID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to load history: %v", err)
	}

	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: agent.SystemPrompt,
		},
	}
	
	for _, msg := range history {
		role := schema.User
		if msg.Role == "assistant" {
			role = schema.Assistant
		}
		messages = append(messages, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// 6. Stream Chat
	// Run in goroutine to not block
	go func() {
		stream, err := aiClient.StreamChat(context.Background(), messages)
		if err != nil {
			errMsg, _ := json.Marshal(map[string]interface{}{
				"sessionId": sessionID,
				"error":     err.Error(),
			})
			s.sseHandler.Send(string(errMsg))
			return
		}
		defer stream.Close()

		fullResponse := ""
		for {
			chunk, err := stream.Recv()
			if err != nil {
				// EOF or Error
				break
			}
			// Send chunk to frontend via SSE
			// Format: JSON with sessionId and content delta
			if chunk.Content != "" {
				fullResponse += chunk.Content
				msg, _ := json.Marshal(map[string]interface{}{
					"sessionId": sessionID,
					"delta":     chunk.Content,
				})
				s.sseHandler.Send(string(msg))
			}
		}
		
		// Save Assistant Message
		aiMsg := &model.Message{
			SessionID: sessionID,
			Role:      "assistant",
			Content:   fullResponse,
		}
		s.messageRepo.Create(aiMsg)

		// Send done signal
		doneMsg, _ := json.Marshal(map[string]interface{}{
			"sessionId": sessionID,
			"done":      true,
		})
		s.sseHandler.Send(string(doneMsg))
	}()

	return nil
}
