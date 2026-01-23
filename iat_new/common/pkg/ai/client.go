package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"iat/common/model"
)

type Client struct {
	chatModel *openai.ChatModel
}

func NewClientFromModel(config *model.AIModel, tools []*schema.ToolInfo) (*Client, error) {
	cfg := &openai.ChatModelConfig{
		BaseURL: config.BaseURL,
		APIKey:  config.APIKey,
		Model:   config.Name,
	}
	m, err := openai.NewChatModel(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %v", err)
	}
	if len(tools) > 0 {
		if err := m.BindTools(tools); err != nil {
			return nil, fmt.Errorf("failed to bind tools: %v", err)
		}
	}
	return &Client{chatModel: m}, nil
}

func NewClientFromEnv() (*Client, error) {
	baseURL := strings.TrimSpace(os.Getenv("AI_BASE_URL"))
	apiKey := strings.TrimSpace(os.Getenv("AI_API_KEY"))
	modelName := strings.TrimSpace(os.Getenv("AI_MODEL_NAME"))

	if apiKey == "" {
		return nil, fmt.Errorf("AI_API_KEY is required")
	}
	if modelName == "" {
		modelName = "gpt-4.1-mini"
	}

	cfg := &openai.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
	}

	m, err := openai.NewChatModel(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %v", err)
	}

	return &Client{chatModel: m}, nil
}

func (c *Client) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return c.chatModel.Stream(ctx, messages)
}

func (c *Client) Chat(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return c.chatModel.Generate(ctx, messages)
}

