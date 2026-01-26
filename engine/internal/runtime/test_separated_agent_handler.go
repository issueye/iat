package runtime

import (
	"context"
	"fmt"
	"iat/common/protocol"
	"strings"
)

type TestSeparatedAgentHandler struct{}

func (h *TestSeparatedAgentHandler) Handle(ctx context.Context, r *Runtime, inst *AgentInstance, msg protocol.Message) (protocol.Message, error) {
	content, err := getPayloadContent(msg.Payload)
	if err != nil {
		return protocol.Message{}, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return protocol.Message{Action: "task_result", Payload: map[string]any{"result": ""}}, nil
	}

	if strings.HasPrefix(content, "remember ") {
		value := strings.TrimSpace(strings.TrimPrefix(content, "remember "))
		inst.Memory = append(inst.Memory, value)
		return protocol.Message{Action: "task_result", Payload: map[string]any{"result": "ok"}}, nil
	}

	if content == "recall" {
		return protocol.Message{
			Action:  "task_result",
			Payload: map[string]any{"result": strings.Join(inst.Memory, "\n")},
		}, nil
	}

	if strings.HasPrefix(content, "use ") {
		rest := strings.TrimSpace(strings.TrimPrefix(content, "use "))
		toolName, argsText, _ := strings.Cut(rest, " ")
		toolName = strings.TrimSpace(toolName)
		argsText = strings.TrimSpace(argsText)
		tool, ok := inst.Tools[toolName]
		if !ok {
			return protocol.Message{}, fmt.Errorf("tool %s not granted", toolName)
		}
		out, err := tool(ctx, map[string]any{"text": argsText})
		if err != nil {
			return protocol.Message{}, err
		}
		return protocol.Message{Action: "task_result", Payload: map[string]any{"result": fmt.Sprintf("%v", out)}}, nil
	}

	if strings.HasPrefix(content, "dispatch ") {
		rest := strings.TrimSpace(strings.TrimPrefix(content, "dispatch "))
		agentID, task, _ := strings.Cut(rest, " ")
		agentID = strings.TrimSpace(agentID)
		task = strings.TrimSpace(task)
		if agentID == "" {
			return protocol.Message{}, fmt.Errorf("missing agent id")
		}
		resp, err := r.Call(ctx, inst.ID, agentID, "execute_task", map[string]any{"content": task})
		if err != nil {
			return protocol.Message{}, err
		}
		return protocol.Message{Action: "task_result", Payload: resp.Payload}, nil
	}

	return protocol.Message{
		Action: "task_result",
		Payload: map[string]any{
			"result": fmt.Sprintf("echo: %s", content),
		},
	}, nil
}

