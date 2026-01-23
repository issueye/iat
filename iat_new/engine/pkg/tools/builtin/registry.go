package builtin

import (
	"encoding/json"
	"iat/common/model"
	"iat/common/pkg/consts"
	"iat/common/pkg/tools"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

var ToolFunctions = map[string]interface{}{
	"ReadFile":      tools.ReadFile,
	"WriteFile":     tools.WriteFile,
	"ListFiles":     tools.ListFiles,
	"RunCommand":    tools.RunCommand,
	"RunScript":     tools.RunScript,
	"ReadFileRange": tools.ReadFileRange,
	"DiffFile":      tools.DiffFile,
}

var BuiltinTools = []model.Tool{
	{
		Name:        "read_file",
		Description: "Read the content of a file (limit 100KB)",
		Type:        consts.ToolTypeBuiltin,
		Parameters: `{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "Path to the file"}
			},
			"required": ["path"]
		}`,
	},
	{
		Name:        "write_file",
		Description: "Write content to a file",
		Type:        consts.ToolTypeBuiltin,
		Parameters: `{
			"type": "object",
			"properties": {
				"path":    {"type": "string", "description": "Path to the file"},
				"content": {"type": "string", "description": "Content to write"}
			},
			"required": ["path", "content"]
		}`,
	},
	{
		Name:        "list_files",
		Description: "List files in a directory",
		Type:        consts.ToolTypeBuiltin,
		Parameters: `{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "Directory path"}
			},
			"required": ["path"]
		}`,
	},
	{
		Name:        "run_command",
		Description: "Execute a shell command",
		Type:        consts.ToolTypeBuiltin,
		Parameters: `{
			"type": "object",
			"properties": {
				"command": {"type": "string", "description": "Command to execute"},
				"args": {
					"type": "array",
					"items": {"type": "string"},
					"description": "Command arguments"
				}
			},
			"required": ["command"]
		}`,
	},
	{
		Name:        "run_script",
		Description: "Execute a python/js/sh/go script",
		Type:        consts.ToolTypeBuiltin,
		Parameters: `{
			"type": "object",
			"properties": {
				"scriptPath": {"type": "string", "description": "Path to script file"},
				"args": {
					"type": "array",
					"items": {"type": "string"},
					"description": "Script arguments"
				}
			},
			"required": ["scriptPath"]
		}`,
	},
	{
		Name:        "read_file_range",
		Description: "Read a range of lines from a file",
		Type:        consts.ToolTypeBuiltin,
		Parameters: `{
			"type": "object",
			"properties": {
				"path":      {"type": "string", "description": "Path to the file"},
				"startLine": {"type": "integer", "description": "Starting line number (1-based)"},
				"limit":     {"type": "integer", "description": "Number of lines to read"}
			},
			"required": ["path", "startLine", "limit"]
		}`,
	},
	{
		Name:        "diff_file",
		Description: "Compare two files and return pretty diff",
		Type:        consts.ToolTypeBuiltin,
		Parameters: `{
			"type": "object",
			"properties": {
				"path1": {"type": "string", "description": "Path to first file"},
				"path2": {"type": "string", "description": "Path to second file"}
			},
			"required": ["path1", "path2"]
		}`,
	},
	{
		Name:        "manage_tasks",
		Description: "Create, update, delete or list tasks in the current session",
		Type:        consts.ToolTypeBuiltin,
		Parameters: `{
			"type": "object",
			"properties": {
				"action":   {"type": "string", "enum": ["add", "update", "delete", "list"], "description": "Action to perform"},
				"id":       {"type": "integer", "description": "Task ID (required for update/delete)"},
				"content":  {"type": "string", "description": "Task content (required for add)"},
				"status":   {"type": "string", "enum": ["pending", "in_progress", "completed"], "description": "Task status (for update)"},
				"priority": {"type": "string", "enum": ["low", "medium", "high"], "description": "Task priority (for add)"},
				"parentId": {"type": "integer", "description": "Parent task ID (for sub-tasks)"}
			},
			"required": ["action"]
		}`,
	},
}

func mustParseSchema(jsonStr string) *schema.ParamsOneOf {
	var s jsonschema.Schema
	if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
		panic("invalid builtin tool schema: " + err.Error())
	}
	return schema.NewParamsOneOfByJSONSchema(&s)
}

// Wrapper for Eino
func GetEinoTools(mode string) []*schema.ToolInfo {
	var infos []*schema.ToolInfo

	// Read File
	infos = append(infos, &schema.ToolInfo{
		Name: "read_file",
		Desc: "Read the content of a file (limit 100KB)",
		ParamsOneOf: mustParseSchema(`{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "Path to the file"}
			},
			"required": ["path"]
		}`),
	})

	// Read File Range
	infos = append(infos, &schema.ToolInfo{
		Name: "read_file_range",
		Desc: "Read a range of lines from a file",
		ParamsOneOf: mustParseSchema(`{
			"type": "object",
			"properties": {
				"path":      {"type": "string", "description": "Path to the file"},
				"startLine": {"type": "integer", "description": "Starting line number (1-based)"},
				"limit":     {"type": "integer", "description": "Number of lines to read"}
			},
			"required": ["path", "startLine", "limit"]
		}`),
	})

	// List Files
	infos = append(infos, &schema.ToolInfo{
		Name: "list_files",
		Desc: "List files in a directory",
		ParamsOneOf: mustParseSchema(`{
			"type": "object",
			"properties": {
				"path": {"type": "string", "description": "Directory path"}
			},
			"required": ["path"]
		}`),
	})

	// Diff File
	infos = append(infos, &schema.ToolInfo{
		Name: "diff_file",
		Desc: "Compare two files and return pretty diff",
		ParamsOneOf: mustParseSchema(`{
			"type": "object",
			"properties": {
				"path1": {"type": "string", "description": "Path to first file"},
				"path2": {"type": "string", "description": "Path to second file"}
			},
			"required": ["path1", "path2"]
		}`),
	})

	// Manage Tasks
	infos = append(infos, &schema.ToolInfo{
		Name: "manage_tasks",
		Desc: "Create, update, delete or list tasks in the current session",
		ParamsOneOf: mustParseSchema(`{
			"type": "object",
			"properties": {
				"action":   {"type": "string", "enum": ["add", "update", "delete", "list"], "description": "Action to perform"},
				"id":       {"type": "integer", "description": "Task ID (required for update/delete)"},
				"content":  {"type": "string", "description": "Task content (required for add)"},
				"status":   {"type": "string", "enum": ["pending", "in_progress", "completed"], "description": "Task status (for update)"},
				"priority": {"type": "string", "enum": ["low", "medium", "high"], "description": "Task priority (for add)"},
				"parentId": {"type": "integer", "description": "Parent task ID (for sub-tasks)"}
			},
			"required": ["action"]
		}`),
	})

	if mode == "build" {
		// Write File
		infos = append(infos, &schema.ToolInfo{
			Name: "write_file",
			Desc: "Write content to a file",
			ParamsOneOf: mustParseSchema(`{
				"type": "object",
				"properties": {
					"path":    {"type": "string", "description": "Path to the file"},
					"content": {"type": "string", "description": "Content to write"}
				},
				"required": ["path", "content"]
			}`),
		})

		// Run Command
		infos = append(infos, &schema.ToolInfo{
			Name: "run_command",
			Desc: "Execute a shell command",
			ParamsOneOf: mustParseSchema(`{
				"type": "object",
				"properties": {
					"command": {"type": "string", "description": "Command to execute"},
					"args": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Command arguments"
					}
				},
				"required": ["command"]
			}`),
		})

		// Run Script
		infos = append(infos, &schema.ToolInfo{
			Name: "run_script",
			Desc: "Execute a python/js/sh/go script",
			ParamsOneOf: mustParseSchema(`{
				"type": "object",
				"properties": {
					"scriptPath": {"type": "string", "description": "Path to script file"},
					"args": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Script arguments"
					}
				},
				"required": ["scriptPath"]
			}`),
		})
	}

	// Call Sub-agent
	infos = append(infos, &schema.ToolInfo{
		Name: "call_subagent",
		Desc: "Call another specialized agent to perform a task",
		ParamsOneOf: mustParseSchema(`{
			"type": "object",
			"properties": {
				"agentName": {"type": "string", "description": "Name of the agent to call"},
				"query":     {"type": "string", "description": "The task or question for the sub-agent"}
			},
			"required": ["agentName", "query"]
		}`),
	})

	return infos
}

// Helper to resolve paths safely
func ResolvePathInBase(base, user string) (string, error) {
	return tools.ResolvePathInBase(base, user)
}

// Re-export implementations for direct call if needed
func ReadFile(path string) (string, error)           { return tools.ReadFile(path) }
func WriteFile(path, content string) (string, error) { return tools.WriteFile(path, content) }
func ListFiles(path string) (string, error)          { return tools.ListFiles(path) }
func ReadFileRange(path string, start, limit int) (string, error) {
	return tools.ReadFileRange(path, start, limit)
}
func DiffFile(path1, path2 string) (string, error) { return tools.DiffFile(path1, path2) }
func RunCommand(command string, args []string) (string, error) {
	return tools.RunCommand(command, args)
}
func RunScript(path string, args []string) (string, error) { return tools.RunScript(path, args) }

// Http helpers for script modules
func HttpGet(url string) (string, error) { return tools.HttpGet(url) }
func HttpPost(url, contentType, body string) (string, error) {
	return tools.HttpPost(url, contentType, body)
}
