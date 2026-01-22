package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/common/model"

	"iat/common/pkg/consts"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

var ToolFunctions = map[string]interface{}{
	"ReadFile":   ReadFile,
	"WriteFile":  WriteFile,
	"ListFiles":  ListFiles,
	"RunCommand": RunCommand,
	"RunScript":  RunScript,
	"HttpGet":    HttpGet,
	"HttpPost":   HttpPost,
}

func GetEinoTools(modeKey string) []*schema.ToolInfo {
	var tools []*schema.ToolInfo
	for _, t := range BuiltinTools {
		// Permission Filter based on Agent Name
		if modeKey == consts.ChatMode {
			// Chat agent gets NO tools
			continue
		} else if modeKey == consts.PlanMode {
			// Plan agent only gets file operations (read/write/list)
			// We can filter by name or some property.
			// Assuming file operations are: read_file, write_file, list_files, read_file_range, diff_file
			if t.Name != consts.ToolReadFile.ToString() &&
				t.Name != consts.ToolWriteFile.ToString() &&
				t.Name != consts.ToolListFiles.ToString() &&
				t.Name != consts.ToolReadFileRange.ToString() &&
				t.Name != consts.ToolDiffFile.ToString() {
				continue
			}
		} else if modeKey == consts.BuildMode {
			// Build agent gets ALL tools
		} else {
			// Custom agents or unknown builtins: default to ALL (or based on binding if we implement binding check here)
			// If we are strictly following "Chat/Plan/Build" logic, we might restrict others too.
			// But for custom agents, they usually have explicit tool bindings in DB.
			// This function currently returns ALL builtin tools filtered by hardcoded logic.
			// Ideally, we should pass the list of allowed tool names.
		}

		var s jsonschema.Schema
		if err := json.Unmarshal([]byte(t.Parameters), &s); err != nil {
			// Skip tools with invalid schema or log error
			// For now, we print error and continue, but in production we should handle better
			fmt.Printf("Failed to parse schema for tool %s: %v\n", t.Name, err)
			continue
		}
		tools = append(tools, &schema.ToolInfo{
			Name:        t.Name,
			Desc:        t.Description,
			ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&s),
		})
	}
	return tools
}

// WrapToolFunction wraps the tool function to match Eino's Tool implementation
func WrapToolFunction(name string, fn interface{}) func(ctx context.Context, input map[string]interface{}) (string, error) {
	// TODO: Implement reflection based wrapper or specific wrappers
	// Since Eino tools usually take (ctx, input) and return (output, error)
	// We might need to adapt our simple functions to this signature.
	// For now, let's keep it simple and just return a placeholder or specific wrappers.

	// Real implementation would use reflection to map input map to function arguments.
	return func(ctx context.Context, input map[string]interface{}) (string, error) {
		return "", fmt.Errorf("WrapToolFunction not implemented for %s", name)
	}
}

var BuiltinTools = []model.Tool{
	// File Operations
	{
		Name:        consts.ToolReadFile.ToString(),
		Description: "Read the contents of a file from the local filesystem",
		Type:        consts.ToolTypeBuiltin,
		Content:     "ReadFile",
		Parameters: `{
			"type": "object",
			"properties": {
				"path": {
					"type": "string",
					"description": "The absolute path to the file to read"
				}
			},
			"required": ["path"]
		}`,
	},
	{
		Name:        consts.ToolReadFileRange.ToString(),
		Description: "Read a specific range of lines from a file (useful for large files)",
		Type:        consts.ToolTypeBuiltin,
		Content:     "ReadFileRange",
		Parameters: `{
			"type": "object",
			"properties": {
				"path": {
					"type": "string",
					"description": "The absolute path to the file to read"
				},
				"startLine": {
					"type": "integer",
					"description": "The line number to start reading from (1-based)"
				},
				"limit": {
					"type": "integer",
					"description": "The number of lines to read"
				}
			},
			"required": ["path", "startLine", "limit"]
		}`,
	},
	{
		Name:        consts.ToolDiffFile.ToString(),
		Description: "Show the differences between two files",
		Type:        consts.ToolTypeBuiltin,
		Content:     "DiffFile",
		Parameters: `{
			"type": "object",
			"properties": {
				"path1": {
					"type": "string",
					"description": "The absolute path to the first file"
				},
				"path2": {
					"type": "string",
					"description": "The absolute path to the second file"
				}
			},
			"required": ["path1", "path2"]
		}`,
	},
	{
		Name:        consts.ToolWriteFile.ToString(),
		Description: "Write content to a file on the local filesystem (overwrites if exists)",
		Type:        consts.ToolTypeBuiltin,
		Content:     "WriteFile",
		Parameters: `{
			"type": "object",
			"properties": {
				"path": {
					"type": "string",
					"description": "The absolute path to the file to write"
				},
				"content": {
					"type": "string",
					"description": "The content to write to the file"
				}
			},
			"required": ["path", "content"]
		}`,
	},
	{
		Name:        consts.ToolListFiles.ToString(),
		Description: "List files and directories in a given directory path",
		Type:        consts.ToolTypeBuiltin,
		Content:     "ListFiles",
		Parameters: `{
			"type": "object",
			"properties": {
				"path": {
					"type": "string",
					"description": "The absolute path to the directory to list"
				}
			},
			"required": ["path"]
		}`,
	},

	// Command Execution
	{
		Name:        consts.ToolRunCommand.ToString(),
		Description: "Execute a shell command",
		Type:        consts.ToolTypeBuiltin,
		Content:     "RunCommand",
		Parameters: `{
			"type": "object",
			"properties": {
				"command": {
					"type": "string",
					"description": "The command to execute (e.g., 'ls', 'git')"
				},
				"args": {
					"type": "array",
					"items": {
						"type": "string"
					},
					"description": "List of arguments for the command"
				}
			},
			"required": ["command"]
		}`,
	},
	{
		Name:        consts.ToolRunScript.ToString(),
		Description: "Execute a script file (python, js, sh, go)",
		Type:        consts.ToolTypeBuiltin,
		Content:     "RunScript",
		Parameters: `{
			"type": "object",
			"properties": {
				"scriptPath": {
					"type": "string",
					"description": "The absolute path to the script file"
				},
				"args": {
					"type": "array",
					"items": {
						"type": "string"
					},
					"description": "List of arguments for the script"
				}
			},
			"required": ["scriptPath"]
		}`,
	},

	// Network Requests
	{
		Name:        consts.ToolHttpGet.ToString(),
		Description: "Perform an HTTP GET request",
		Type:        consts.ToolTypeBuiltin,
		Content:     "HttpGet",
		Parameters: `{
			"type": "object",
			"properties": {
				"url": {
					"type": "string",
					"description": "The URL to send the GET request to"
				}
			},
			"required": ["url"]
		}`,
	},
	{
		Name:        consts.ToolHttpPost.ToString(),
		Description: "Perform an HTTP POST request",
		Type:        consts.ToolTypeBuiltin,
		Content:     "HttpPost",
		Parameters: `{
			"type": "object",
			"properties": {
				"url": {
					"type": "string",
					"description": "The URL to send the POST request to"
				},
				"contentType": {
					"type": "string",
					"description": "Content-Type header (default: application/json)"
				},
				"body": {
					"type": "string",
					"description": "The request body"
				}
			},
			"required": ["url", "body"]
		}`,
	},
	{
		Name:        consts.ToolIndexProject.ToString(),
		Description: "Index projects for searching sessions by project name",
		Type:        consts.ToolTypeBuiltin,
		Content:     "IndexProject",
		Parameters: `{
			"type": "object",
			"properties": {
				"projectId": {
					"type": "number",
					"description": "Project ID to index (default: current session project)"
				},
				"all": {
					"type": "boolean",
					"description": "Index all projects"
				}
			}
		}`,
	},
	{
		Name:        "call_subagent",
		Description: "Call another agent (SubAgent) to perform a specific task",
		Type:        consts.ToolTypeBuiltin,
		Content:     "CallSubAgent",
		Parameters: `{
			"type": "object",
			"properties": {
				"agentName": {
					"type": "string",
					"description": "The name of the agent to call"
				},
				"query": {
					"type": "string",
					"description": "The task or query for the sub-agent"
				}
			},
			"required": ["agentName", "query"]
		}`,
	},
	{
		Name:        "manage_tasks",
		Description: "Manage a task list for the current session (add, update, delete, list)",
		Type:        consts.ToolTypeBuiltin,
		Content:     "ManageTasks",
		Parameters: `{
			"type": "object",
			"properties": {
				"action": {
					"type": "string",
					"enum": ["add", "update", "delete", "list"],
					"description": "The action to perform"
				},
				"content": {
					"type": "string",
					"description": "Task content/description (for add/update)"
				},
				"id": {
					"type": "number",
					"description": "Task ID (for update/delete)"
				},
				"status": {
					"type": "string",
					"enum": ["pending", "in_progress", "completed", "failed"],
					"description": "Task status (for update)"
				},
				"priority": {
					"type": "string",
					"enum": ["high", "medium", "low"],
					"description": "Task priority (for add/update)"
				}
			},
			"required": ["action"]
		}`,
	},
}
