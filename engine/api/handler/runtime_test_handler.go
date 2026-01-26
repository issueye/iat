package handler

import (
	"context"
	"encoding/json"
	"iat/common/model"
	"iat/engine/internal/runtime"
	"net/http"
	"strings"
)

type RuntimeTestHandler struct{}

func NewRuntimeTestHandler() *RuntimeTestHandler {
	return &RuntimeTestHandler{}
}

func (h *RuntimeTestHandler) Run(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rt := runtime.NewRuntime(nil, nil, nil)
	rt.RegisterGlobalTool("uppercase", func(ctx context.Context, args map[string]any) (any, error) {
		s, _ := args["text"].(string)
		return strings.ToUpper(s), nil
	})

	a := rt.RegisterDetachedAgent(model.Agent{
		Base: model.Base{ID: 1},
		Name: "test_orchestrator",
		Modes: []model.Mode{{Key: "chat"}},
	}, &runtime.TestSeparatedAgentHandler{})
	b := rt.RegisterDetachedAgent(model.Agent{
		Base: model.Base{ID: 2},
		Name: "test_worker",
		Modes: []model.Mode{{Key: "chat"}},
	}, &runtime.TestSeparatedAgentHandler{})

	_ = rt.GrantTool(a.ID, b.ID, "uppercase")

	ctx := r.Context()

	_, _ = rt.Call(ctx, a.ID, a.ID, "execute_task", map[string]any{"content": "remember apple"})
	_, _ = rt.Call(ctx, a.ID, b.ID, "execute_task", map[string]any{"content": "remember banana"})

	recallA, _ := rt.Call(ctx, a.ID, a.ID, "execute_task", map[string]any{"content": "recall"})
	recallB, _ := rt.Call(ctx, a.ID, b.ID, "execute_task", map[string]any{"content": "recall"})

	toolOk, _ := rt.Call(ctx, a.ID, b.ID, "execute_task", map[string]any{"content": "use uppercase hello"})
	toolDenied, _ := rt.Call(ctx, a.ID, a.ID, "execute_task", map[string]any{"content": "use uppercase hello"})

	dispatch, _ := rt.Call(ctx, a.ID, a.ID, "execute_task", map[string]any{"content": "dispatch " + b.ID + " recall"})

	resp := map[string]any{
		"agents": map[string]any{
			"orchestrator": a.ID,
			"worker":       b.ID,
		},
		"recall": map[string]any{
			"orchestrator": recallA.Payload,
			"worker":       recallB.Payload,
		},
		"tools": map[string]any{
			"worker_uppercase":       toolOk.Payload,
			"orchestrator_uppercase": toolDenied.Payload,
		},
		"dispatch": dispatch.Payload,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

