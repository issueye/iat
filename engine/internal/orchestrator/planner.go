package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/common/model"
	"iat/engine/pkg/ai"

	"github.com/cloudwego/eino/schema"
)

type Planner struct {
	client *ai.AIClient
}

func NewPlanner(client *ai.AIClient) *Planner {
	return &Planner{client: client}
}

const plannerPrompt = `You are a task planning agent. Your goal is to break down a complex user goal into a set of discrete, actionable sub-tasks.
Each sub-task should require a specific capability.
Return the result as a JSON object with the following structure:
{
  "goal": "the original goal",
  "tasks": [
    {
      "id": "unique_id",
      "title": "short title",
      "description": "detailed description",
      "dependsOn": ["list of task ids this task depends on"],
      "capability": "the name of the capability required (e.g., code_analysis, web_search, file_write)"
    }
  ]
}
`

func (p *Planner) Plan(ctx context.Context, goal string) (*model.TaskTree, error) {
	messages := []*schema.Message{
		schema.SystemMessage(plannerPrompt),
		schema.UserMessage(fmt.Sprintf("Goal: %s", goal)),
	}

	resp, err := p.client.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	var tree model.TaskTree
	if err := json.Unmarshal([]byte(resp.Content), &tree); err != nil {
		return nil, fmt.Errorf("failed to parse planner output: %w", err)
	}

	return &tree, nil
}
