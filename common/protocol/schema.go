package protocol

// Predefined capability names
const (
	CapCodeAnalysis = "code_analysis"
	CapCodeWriting  = "code_writing"
	CapWebSearch    = "web_search"
	CapFileOperation = "file_operation"
	CapCommandExec  = "command_execution"
)

// TaskContract defines the input/output schema for a capability
type TaskContract struct {
	Capability string         `json:"capability"`
	Input      map[string]any `json:"input"`
	Output     map[string]any `json:"output"`
}

// Predefined Contracts (examples)
var Contracts = map[string]TaskContract{
	CapCodeAnalysis: {
		Capability: CapCodeAnalysis,
		Input: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
				"query": map[string]any{"type": "string"},
			},
			"required": []string{"path", "query"},
		},
	},
}
