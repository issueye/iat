package tools

import (
	"context"
	"strings"
)

type Tool func(ctx context.Context, args map[string]any) (any, error)

func Builtins() map[string]Tool {
	return map[string]Tool{
		"uppercase": func(ctx context.Context, args map[string]any) (any, error) {
			s, _ := args["text"].(string)
			return strings.ToUpper(s), nil
		},
	}
}

