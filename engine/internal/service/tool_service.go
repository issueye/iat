package service

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/common/model"
	"iat/common/pkg/consts"
	"iat/common/pkg/script"
	"iat/engine/internal/repo"
	"iat/engine/pkg/tools/builtin"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

type ToolService struct {
	repo       *repo.ToolRepo
	mcpService *MCPService
}

func NewToolService(mcpService *MCPService) *ToolService {
	return &ToolService{
		repo:       repo.NewToolRepo(),
		mcpService: mcpService,
	}
}

// ... existing CRUD methods ...

func (s *ToolService) ListTools() ([]model.Tool, error) {
	return s.repo.List()
}

func (s *ToolService) CreateTool(tool *model.Tool) error {
	return s.repo.Create(tool)
}

func (s *ToolService) UpdateTool(tool *model.Tool) error {
	return s.repo.Update(tool)
}

func (s *ToolService) DeleteTool(id uint) error {
	tool, err := s.repo.Get(id)
	if err != nil {
		return err
	}
	if tool.Type == "builtin" {
		return fmt.Errorf("cannot delete builtin tool")
	}
	return s.repo.Delete(id)
}

// New Execution Logic
func (s *ToolService) Call(ctx context.Context, name string, args map[string]any, agent *model.Agent, projectRoot string) (string, error) {
	// 1. Try Builtin
	switch name {
	case "read_file", "write_file", "list_files", "run_command", "run_script", "read_file_range", "diff_file", "manage_tasks":
		// Handle via existing builtin logic (needs slight refactor to be more modular)
		return s.executeBuiltin(name, args, projectRoot)
	}

	// 2. Try Agent-attached Script Tools
	for _, t := range agent.Tools {
		if t.Name == name && (t.Type == consts.ToolTypeCustom || t.Type == consts.ToolTypeScript) {
			engine := script.NewScriptEngine()
			engine.RegisterGlobal("args", args)
			res, err := engine.Run(t.Content)
			if err != nil {
				return "", fmt.Errorf("script error: %w", err)
			}
			return fmt.Sprintf("%v", res), nil
		}
	}

	// 3. Try MCP
	if s.mcpService != nil {
		res, err := s.mcpService.CallTool(ctx, name, args)
		if err == nil {
			return res, nil
		}
	}

	return "", fmt.Errorf("tool %s not found", name)
}

func (s *ToolService) executeBuiltin(name string, args map[string]any, projectRoot string) (string, error) {
	// Implementation similar to chat_service.go's switch but using tools pkg directly
	switch name {
	case "read_file":
		path, _ := args["path"].(string)
		p, _ := builtin.ResolvePathInBase(projectRoot, path)
		return builtin.ReadFile(p)
	case "write_file":
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		p, _ := builtin.ResolvePathInBase(projectRoot, path)
		return builtin.WriteFile(p, content)
	case "list_files":
		path, _ := args["path"].(string)
		p, _ := builtin.ResolvePathInBase(projectRoot, path)
		return builtin.ListFiles(p)
	case "run_command":
		cmd, _ := args["command"].(string)
		cmdArgsRaw, _ := args["args"].([]any)
		var cmdArgs []string
		for _, a := range cmdArgsRaw {
			cmdArgs = append(cmdArgs, fmt.Sprintf("%v", a))
		}
		return builtin.RunCommand(cmd, cmdArgs)
	case "run_script":
		path, _ := args["scriptPath"].(string)
		p, _ := builtin.ResolvePathInBase(projectRoot, path)
		scriptArgsRaw, _ := args["args"].([]any)
		var scriptArgs []string
		for _, a := range scriptArgsRaw {
			scriptArgs = append(scriptArgs, fmt.Sprintf("%v", a))
		}
		return builtin.RunScript(p, scriptArgs)
	case "read_file_range":
		path, _ := args["path"].(string)
		p, _ := builtin.ResolvePathInBase(projectRoot, path)
		start, _ := args["startLine"].(float64) // JSON numbers are float64
		limit, _ := args["limit"].(float64)
		return builtin.ReadFileRange(p, int(start), int(limit))
	case "diff_file":
		path1, _ := args["path1"].(string)
		path2, _ := args["path2"].(string)
		p1, _ := builtin.ResolvePathInBase(projectRoot, path1)
		p2, _ := builtin.ResolvePathInBase(projectRoot, path2)
		return builtin.DiffFile(p1, p2)
	}
	return "", fmt.Errorf("builtin %s not implemented in ToolService or handled by Orchestrator", name)
}

func (s *ToolService) GetEinoTools(agent *model.Agent) ([]*schema.ToolInfo, error) {
	var infos []*schema.ToolInfo

	// Add Builtins based on mode
	infos = append(infos, builtin.GetEinoTools(agent.Mode.Key)...)

	// Add MCP Tools
	if s.mcpService != nil {
		mcpTools, err := s.mcpService.GetToolsForServers(agent.MCPServers)
		if err == nil {
			infos = append(infos, mcpTools...)
		}
	}

	// Add Custom Tools
	for _, t := range agent.Tools {
		if t.Type == consts.ToolTypeCustom || t.Type == consts.ToolTypeScript {
			var js jsonschema.Schema
			if err := json.Unmarshal([]byte(t.Parameters), &js); err == nil {
				infos = append(infos, &schema.ToolInfo{
					Name:        t.Name,
					Desc:        t.Description,
					ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&js),
				})
			}
		}
	}

	return infos, nil
}
