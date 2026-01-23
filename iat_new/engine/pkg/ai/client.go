package ai

import (
	"context"

	commonai "iat/common/pkg/ai"

	"github.com/cloudwego/eino/schema"
	"iat/common/model"
)

type AIClient struct {
	client *commonai.Client
}

func NewAIClient(config *model.AIModel, tools []*schema.ToolInfo) (*AIClient, error) {
	c, err := commonai.NewClientFromModel(config, tools)
	if err != nil {
		return nil, err
	}
	return &AIClient{client: c}, nil
}

func (c *AIClient) StreamChat(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return c.client.Stream(ctx, messages)
}

func (c *AIClient) Chat(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return c.client.Chat(ctx, messages)
}
