package ai

import (
	"context"
	"fmt"
	"iat/common/model"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

type AIClient struct {
	chatModel *openai.ChatModel
}

func NewAIClient(config *model.AIModel, tools []*schema.ToolInfo) (*AIClient, error) {
	// Eino currently supports OpenAI compatible interfaces
	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL: config.BaseURL,
		APIKey:  config.APIKey,
		Model:   config.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %v", err)
	}

	if len(tools) > 0 {
		// Bind tools to the model
		if err := chatModel.BindTools(tools); err != nil {
			return nil, fmt.Errorf("failed to bind tools: %v", err)
		}
	}

	return &AIClient{
		chatModel: chatModel,
	}, nil
}

func (c *AIClient) StreamChat(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return c.chatModel.Stream(ctx, messages)
}

func (c *AIClient) Chat(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return c.chatModel.Generate(ctx, messages)
}
