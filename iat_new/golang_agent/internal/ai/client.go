package ai

import (
	"context"

	commonai "iat/common/pkg/ai"

	"github.com/cloudwego/eino/schema"
)

type Client struct {
	client *commonai.Client
}

func NewClientFromEnv() (*Client, error) {
	c, err := commonai.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &Client{client: c}, nil
}

func (c *Client) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return c.client.Stream(ctx, messages)
}
