package service

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/common/model"
	"iat/common/pkg/script"
	"iat/engine/internal/repo"
	"net/http"
	"strings"
	"time"
)

type HookService struct {
	repo *repo.HookRepo
}

func NewHookService() *HookService {
	return &HookService{
		repo: repo.NewHookRepo(),
	}
}

func (s *HookService) CreateHook(hook *model.Hook) error {
	return s.repo.Create(hook)
}

func (s *HookService) ListHooks() ([]model.Hook, error) {
	return s.repo.List()
}

func (s *HookService) GetHook(id uint) (*model.Hook, error) {
	return s.repo.Get(id)
}

func (s *HookService) UpdateHook(hook *model.Hook) error {
	return s.repo.Update(hook)
}

func (s *HookService) DeleteHook(id uint) error {
	return s.repo.Delete(id)
}

func (s *HookService) GetHooksByType(hookType string) ([]model.Hook, error) {
	return s.repo.ListByType(hookType)
}

func (s *HookService) ExecuteHooks(ctx context.Context, hookType string, targetType string, targetID uint, contextData map[string]any) error {
	// 1. Get Global Hooks for this type
	globalHooks, err := s.repo.ListByTarget("global", 0)
	if err != nil {
		return err
	}

	// 2. Get Target Hooks for this type
	var targetHooks []model.Hook
	if targetType != "" && targetID != 0 {
		targetHooks, err = s.repo.ListByTarget(targetType, targetID)
		if err != nil {
			return err
		}
	}

	// 3. Filter and Execute
	allHooks := append(globalHooks, targetHooks...)
	for _, hook := range allHooks {
		if !hook.Enabled {
			continue
		}
		if hook.Type != hookType {
			continue
		}

		// Execute asynchronously to avoid blocking main flow?
		// For pre-hooks, we might want to block (validation).
		// For post-hooks, maybe async.
		// For now, let's run synchronously but safely.
		if err := s.runSingleHook(ctx, &hook, contextData); err != nil {
			fmt.Printf("[HookService] Hook %s (ID: %d) failed: %v\n", hook.Name, hook.ID, err)
			// Continue executing other hooks even if one fails
		}
	}
	return nil
}

func (s *HookService) runSingleHook(ctx context.Context, hook *model.Hook, data map[string]any) error {
	switch hook.Action {
	case "script":
		engine := script.NewScriptEngine()
		engine.RegisterGlobal("context", data)
		_, err := engine.RunWithTimeout(hook.Content, 10*time.Second)
		return err
	case "http":
		client := &http.Client{Timeout: 10 * time.Second}
		body, _ := json.Marshal(data)
		req, err := http.NewRequestWithContext(ctx, "POST", hook.Content, strings.NewReader(string(body)))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hook-Event", hook.Type)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("http hook returned status %d", resp.StatusCode)
		}
		return nil
	default:
		return fmt.Errorf("unknown hook action: %s", hook.Action)
	}
}
