package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/engine/pkg/ai"

	"github.com/cloudwego/eino/schema"
)

type SubTask struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DependsOn   []string `json:"dependsOn,omitempty"`
	Capability  string   `json:"capability"` // The required capability for this task
}

type TaskTree struct {
	Goal  string    `json:"goal"`
	Tasks []SubTask `json:"tasks"`
}

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

func (p *Planner) Plan(ctx context.Context, goal string) (*TaskTree, error) {
	messages := []*schema.Message{
		schema.SystemMessage(plannerPrompt),
		schema.UserMessage(fmt.Sprintf("Goal: %s", goal)),
	}

	resp, err := p.client.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	var tree TaskTree
	if err := json.Unmarshal([]byte(resp.Content), &tree); err != nil {
		// Attempt to extract JSON if there's surrounding text
		return nil, fmt.Errorf("failed to parse planner output: %w", err)
	}

	return &tree, nil
}
