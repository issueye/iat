package agent

import (
	"context"
	"iat/common/protocol"
	"iat/golang_agent/internal/tools"
	"testing"
	"time"
)

func TestAgentMemoryIsolation(t *testing.T) {
	a := New("a", "A")
	b := New("b", "B")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = a.Handle(ctx, protocol.Message{
		ID:     "1",
		From:   "tester",
		To:     "a",
		Type:   protocol.MsgRequest,
		Action: "execute_task",
		Payload: map[string]any{
			"content": "remember apple",
		},
	})
	_ = b.Handle(ctx, protocol.Message{
		ID:     "2",
		From:   "tester",
		To:     "b",
		Type:   protocol.MsgRequest,
		Action: "execute_task",
		Payload: map[string]any{
			"content": "remember banana",
		},
	})

	recallA := a.Handle(ctx, protocol.Message{
		ID:     "3",
		From:   "tester",
		To:     "a",
		Type:   protocol.MsgRequest,
		Action: "execute_task",
		Payload: map[string]any{
			"content": "recall",
		},
	})
	if got := payloadString(recallA.Payload, "result"); got != "apple" {
		t.Fatalf("recall A mismatch: got %q", got)
	}

	recallB := b.Handle(ctx, protocol.Message{
		ID:     "4",
		From:   "tester",
		To:     "b",
		Type:   protocol.MsgRequest,
		Action: "execute_task",
		Payload: map[string]any{
			"content": "recall",
		},
	})
	if got := payloadString(recallB.Payload, "result"); got != "banana" {
		t.Fatalf("recall B mismatch: got %q", got)
	}
}

func TestAgentToolGranting(t *testing.T) {
	a := New("a", "A")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	denied := a.Handle(ctx, protocol.Message{
		ID:     "1",
		From:   "tester",
		To:     "a",
		Type:   protocol.MsgRequest,
		Action: "execute_task",
		Payload: map[string]any{
			"content": "use uppercase hello",
		},
	})
	if payloadString(denied.Payload, "error") == "" {
		t.Fatalf("expected error when tool not granted")
	}

	a.GrantTool("uppercase", tools.Builtins()["uppercase"])

	ok := a.Handle(ctx, protocol.Message{
		ID:     "2",
		From:   "tester",
		To:     "a",
		Type:   protocol.MsgRequest,
		Action: "execute_task",
		Payload: map[string]any{
			"content": "use uppercase hello",
		},
	})
	if got := payloadString(ok.Payload, "result"); got != "HELLO" {
		t.Fatalf("tool result mismatch: got %q", got)
	}
}

func payloadString(payload any, key string) string {
	if m, ok := payload.(map[string]any); ok {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	if m, ok := payload.(map[string]interface{}); ok {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

