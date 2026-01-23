package agent

import (
	"context"
	"fmt"
	"iat/common/protocol"
	"iat/golang_agent/internal/tools"
	"strings"
	"sync"
	"time"
)

type Agent struct {
	ID    string
	Name  string
	mu    sync.Mutex
	mem   []string
	tools map[string]tools.Tool
	clock func() int64
}

func New(id, name string) *Agent {
	return &Agent{
		ID:    id,
		Name:  name,
		mem:   nil,
		tools: make(map[string]tools.Tool),
		clock: func() int64 { return time.Now().UnixMilli() },
	}
}

func (a *Agent) GrantTool(name string, tool tools.Tool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tools[name] = tool
}

func (a *Agent) Handle(ctx context.Context, msg protocol.Message) protocol.Message {
	if msg.Type == protocol.MsgNotification && msg.Action == "grant_tool" {
		if payload, ok := msg.Payload.(map[string]any); ok {
			if toolName, ok := payload["tool"].(string); ok {
				a.GrantTool(toolName, nil)
			}
		}
		return protocol.Message{
			ID:        msg.ID,
			From:      a.ID,
			To:        msg.From,
			Type:      protocol.MsgResponse,
			Action:    "grant_tool_ack",
			Timestamp: a.clock(),
			Payload:   map[string]any{"ok": true},
		}
	}

	if msg.Type != protocol.MsgRequest {
		return protocol.Message{
			ID:        msg.ID,
			From:      a.ID,
			To:        msg.From,
			Type:      protocol.MsgResponse,
			Action:    "ignored",
			Timestamp: a.clock(),
			Payload:   map[string]any{"error": "unsupported message type"},
		}
	}

	switch msg.Action {
	case "execute_task":
		return a.handleExecuteTask(ctx, msg)
	default:
		return protocol.Message{
			ID:        msg.ID,
			From:      a.ID,
			To:        msg.From,
			Type:      protocol.MsgResponse,
			Action:    "error",
			Timestamp: a.clock(),
			Payload:   map[string]any{"error": "unknown action"},
		}
	}
}

func (a *Agent) handleExecuteTask(ctx context.Context, msg protocol.Message) protocol.Message {
	content, err := payloadContent(msg.Payload)
	if err != nil {
		return protocol.Message{
			ID:        msg.ID,
			From:      a.ID,
			To:        msg.From,
			Type:      protocol.MsgResponse,
			Action:    "task_result",
			Timestamp: a.clock(),
			Payload:   map[string]any{"error": err.Error()},
		}
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return protocol.Message{
			ID:        msg.ID,
			From:      a.ID,
			To:        msg.From,
			Type:      protocol.MsgResponse,
			Action:    "task_result",
			Timestamp: a.clock(),
			Payload:   map[string]any{"result": ""},
		}
	}

	if strings.HasPrefix(content, "remember ") {
		value := strings.TrimSpace(strings.TrimPrefix(content, "remember "))
		a.mu.Lock()
		a.mem = append(a.mem, value)
		a.mu.Unlock()
		return protocol.Message{
			ID:        msg.ID,
			From:      a.ID,
			To:        msg.From,
			Type:      protocol.MsgResponse,
			Action:    "task_result",
			Timestamp: a.clock(),
			Payload:   map[string]any{"result": "ok"},
		}
	}

	if content == "recall" {
		a.mu.Lock()
		out := strings.Join(a.mem, "\n")
		a.mu.Unlock()
		return protocol.Message{
			ID:        msg.ID,
			From:      a.ID,
			To:        msg.From,
			Type:      protocol.MsgResponse,
			Action:    "task_result",
			Timestamp: a.clock(),
			Payload:   map[string]any{"result": out},
		}
	}

	if strings.HasPrefix(content, "use ") {
		rest := strings.TrimSpace(strings.TrimPrefix(content, "use "))
		toolName, argsText, _ := strings.Cut(rest, " ")
		toolName = strings.TrimSpace(toolName)
		argsText = strings.TrimSpace(argsText)

		a.mu.Lock()
		tool, ok := a.tools[toolName]
		a.mu.Unlock()

		if !ok || tool == nil {
			return protocol.Message{
				ID:        msg.ID,
				From:      a.ID,
				To:        msg.From,
				Type:      protocol.MsgResponse,
				Action:    "task_result",
				Timestamp: a.clock(),
				Payload:   map[string]any{"error": fmt.Sprintf("tool %s not granted", toolName)},
			}
		}

		out, err := tool(ctx, map[string]any{"text": argsText})
		if err != nil {
			return protocol.Message{
				ID:        msg.ID,
				From:      a.ID,
				To:        msg.From,
				Type:      protocol.MsgResponse,
				Action:    "task_result",
				Timestamp: a.clock(),
				Payload:   map[string]any{"error": err.Error()},
			}
		}

		return protocol.Message{
			ID:        msg.ID,
			From:      a.ID,
			To:        msg.From,
			Type:      protocol.MsgResponse,
			Action:    "task_result",
			Timestamp: a.clock(),
			Payload:   map[string]any{"result": fmt.Sprintf("%v", out)},
		}
	}

	return protocol.Message{
		ID:        msg.ID,
		From:      a.ID,
		To:        msg.From,
		Type:      protocol.MsgResponse,
		Action:    "task_result",
		Timestamp: a.clock(),
		Payload:   map[string]any{"result": "echo: " + content},
	}
}

func payloadContent(payload any) (string, error) {
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
