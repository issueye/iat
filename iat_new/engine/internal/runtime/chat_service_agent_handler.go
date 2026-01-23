package runtime

import (
	"context"
	"fmt"
	"iat/common/protocol"
	"iat/engine/internal/service"
)

type ChatServiceAgentHandler struct {
	chatService *service.ChatService
}

func (h *ChatServiceAgentHandler) Handle(ctx context.Context, r *Runtime, inst *AgentInstance, msg protocol.Message) (protocol.Message, error) {
	if h.chatService == nil {
		return protocol.Message{}, fmt.Errorf("chat service not available")
	}

	content, err := getPayloadContent(msg.Payload)
	if err != nil {
		return protocol.Message{}, err
	}

	modeKey := inst.ModelAgent.Mode.Key
	if modeKey == "" {
		modeKey = "chat"
	}

	result, err := h.chatService.RunAgentInternal(
		0,
		inst.ModelAgent.Name,
		content,
		".",
		modeKey,
		nil,
	)
	if err != nil {
		return protocol.Message{}, err
	}

	return protocol.Message{
		Action:  "task_result",
		Payload: map[string]any{"result": result},
	}, nil
}

func getPayloadContent(payload any) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("missing payload")
	}
	if m, ok := payload.(map[string]any); ok {
		if v, ok := m["content"]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
		}
	}
	if m, ok := payload.(map[string]interface{}); ok {
		if v, ok := m["content"]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
		}
	}
	if m, ok := payload.(map[string]string); ok {
		if s, ok := m["content"]; ok {
			return s, nil
		}
	}
	return "", fmt.Errorf("invalid payload content")
}

