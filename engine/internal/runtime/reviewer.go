package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/engine/pkg/ai"

	"github.com/cloudwego/eino/schema"
)

type ReviewResult struct {
	Approved bool   `json:"approved"`
	Feedback string `json:"feedback,omitempty"`
}

type Reviewer struct {
	client *ai.AIClient
}

func NewReviewer(client *ai.AIClient) *Reviewer {
	return &Reviewer{client: client}
}

const reviewerPrompt = `You are a task review agent. Your goal is to evaluate if a sub-task has been completed successfully based on the task description and the produced output.
Return the result as a JSON object:
{
  "approved": true/false,
  "feedback": "detailed feedback if not approved, or a brief summary if approved"
}
`

func (r *Reviewer) Review(ctx context.Context, task SubTask, output any) (*ReviewResult, error) {
	outputJSON, _ := json.Marshal(output)
	messages := []*schema.Message{
		schema.SystemMessage(reviewerPrompt),
		schema.UserMessage(fmt.Sprintf("Task: %s\nDescription: %s\nOutput: %s", task.Title, task.Description, string(outputJSON))),
	}

	resp, err := r.client.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	var result ReviewResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse reviewer output: %w", err)
	}

	return &result, nil
}
