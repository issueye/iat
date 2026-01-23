package runtime

import (
	"context"
	"iat/common/model"
	"testing"
	"time"
)

func TestSeparatedAgentIsolationAndTools(t *testing.T) {
	rt := NewRuntime(nil)
	rt.RegisterGlobalTool("uppercase", func(ctx context.Context, args map[string]any) (any, error) {
		s, _ := args["text"].(string)
		out := ""
		for _, r := range s {
			if r >= 'a' && r <= 'z' {
				out += string(r - 32)
			} else {
				out += string(r)
			}
		}
		return out, nil
	})

	a := rt.RegisterDetachedAgent(model.Agent{
		Base: model.Base{ID: 1},
		Name: "A",
		Mode: model.Mode{Key: "chat"},
	}, &TestSeparatedAgentHandler{})
	b := rt.RegisterDetachedAgent(model.Agent{
		Base: model.Base{ID: 2},
		Name: "B",
		Mode: model.Mode{Key: "chat"},
	}, &TestSeparatedAgentHandler{})

	if err := rt.GrantTool(a.ID, b.ID, "uppercase"); err != nil {
		t.Fatalf("GrantTool failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := rt.Call(ctx, a.ID, a.ID, "execute_task", map[string]any{"content": "remember apple"})
	if err != nil {
		t.Fatalf("remember A failed: %v", err)
	}
	_, err = rt.Call(ctx, a.ID, b.ID, "execute_task", map[string]any{"content": "remember banana"})
	if err != nil {
		t.Fatalf("remember B failed: %v", err)
	}

	recallA, err := rt.Call(ctx, a.ID, a.ID, "execute_task", map[string]any{"content": "recall"})
	if err != nil {
		t.Fatalf("recall A failed: %v", err)
	}
	if got := payloadString(recallA.Payload, "result"); got != "apple" {
		t.Fatalf("recall A mismatch: got %q", got)
	}

	recallB, err := rt.Call(ctx, a.ID, b.ID, "execute_task", map[string]any{"content": "recall"})
	if err != nil {
		t.Fatalf("recall B failed: %v", err)
	}
	if got := payloadString(recallB.Payload, "result"); got != "banana" {
		t.Fatalf("recall B mismatch: got %q", got)
	}

	toolOk, err := rt.Call(ctx, a.ID, b.ID, "execute_task", map[string]any{"content": "use uppercase hello"})
	if err != nil {
		t.Fatalf("tool use failed: %v", err)
	}
	if got := payloadString(toolOk.Payload, "result"); got != "HELLO" {
		t.Fatalf("tool result mismatch: got %q", got)
	}

	toolDenied, err := rt.Call(ctx, a.ID, a.ID, "execute_task", map[string]any{"content": "use uppercase hello"})
	if err != nil {
		t.Fatalf("tool denied call failed: %v", err)
	}
	if got := payloadString(toolDenied.Payload, "error"); got == "" {
		t.Fatalf("expected error payload for denied tool")
	}

	dispatch, err := rt.Call(ctx, a.ID, a.ID, "execute_task", map[string]any{"content": "dispatch " + b.ID + " recall"})
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if got := payloadString(dispatch.Payload, "result"); got != "banana" {
		t.Fatalf("dispatch result mismatch: got %q", got)
	}
}

func payloadString(payload any, key string) string {
	m, ok := payload.(map[string]any)
	if !ok {
		if m2, ok := payload.(map[string]interface{}); ok {
			if v, ok := m2[key]; ok {
				if s, ok := v.(string); ok {
					return s
				}
			}
		}
		return ""
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

