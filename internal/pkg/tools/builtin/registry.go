package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/internal/model"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

var ToolFunctions = map[string]interface{}{
	"ReadFile":    ReadFile,
	"WriteFile":   WriteFile,
	"ListFiles":   ListFiles,
	"RunCommand":  RunCommand,
	"HttpGet":     HttpGet,
	"HttpPost":    HttpPost,
}

func GetEinoTools(agentName string) []*schema.ToolInfo {
	var tools []*schema.ToolInfo
	for _, t := range BuiltinTools {
		// Permission Filter based on Agent Name
		if agentName == "Chat" {
			// Chat agent gets NO tools
			continue
		} else if agentName == "Plan" {
			// Plan agent only gets file operations (read/write/list)
			// We can filter by name or some property. 
			// Assuming file operations are: read_file, write_file, list_files
			if t.Name != "read_file" && t.Name != "write_file" && t.Name != "list_files" {
				continue
			}
		} else if agentName == "Build" {
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
			Name: t.Name,
			Desc: t.Description,
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
		Name:        "read_file",
		Description: "Read the contents of a file from the local filesystem",
		Type:        "builtin",
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
		Name:        "write_file",
		Description: "Write content to a file on the local filesystem (overwrites if exists)",
		Type:        "builtin",
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
		Name:        "list_files",
		Description: "List files and directories in a given directory path",
		Type:        "builtin",
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
		Name:        "run_command",
		Description: "Execute a shell command",
		Type:        "builtin",
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

	// Network Requests
	{
		Name:        "http_get",
		Description: "Perform an HTTP GET request",
		Type:        "builtin",
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
		Name:        "http_post",
		Description: "Perform an HTTP POST request",
		Type:        "builtin",
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
}
