package consts

const (
	// Agent Types
	AgentTypeBuiltin = "builtin"
	AgentTypeCustom  = "custom"
	AgentTypeDynamic = "dynamic"

	// Agent Names
	ChatMode  = "CHAT"
	PlanMode  = "PLAN"
	BuildMode = "BUILD"

	// 产品经理Agent
	AgentNameProductManager = "product manager"
	// 项目经理Agent
	AgentNameProjectManager = "project manager"
	// UI/UX Agent
	AgentNameUxUi = "ux/ui developer"
	// golang开发人员Agent
	AgentNameGolang = "golang developer"
	// python开发人员Agent
	AgentNamePython = "python developer"
	// javascript开发人员Agent
	AgentNameJavascript = "javascript developer"
	// 测试开发人员Agent
	AgentNameTest = "test developer"

	// Tool Types
	ToolTypeBuiltin = "builtin"
	ToolTypeCustom  = "custom"
	ToolTypeScript  = "script"

	// Message Roles
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"

	// Message Categories
	MessageCategoryTool = "tool"

	// Tool stages
	ToolStageCall   = "call"
	ToolStageResult = "result"
)

/*
"ReadFile":   ReadFile,
	"WriteFile":  WriteFile,
	"ListFiles":  ListFiles,
	"RunCommand": RunCommand,
	"RunScript":  RunScript,
	"HttpGet":    HttpGet,
	"HttpPost":   HttpPost,
*/

type IATTool struct {
	Name        string
	Content     string
	Description string
	Type        string
}

var (
	ToolReadFile      IATTool = IATTool{Name: "read_file", Content: "ReadFile", Description: "Read a file", Type: ToolTypeBuiltin}
	ToolWriteFile     IATTool = IATTool{Name: "write_file", Content: "WriteFile", Description: "Write to a file", Type: ToolTypeBuiltin}
	ToolListFiles     IATTool = IATTool{Name: "list_files", Content: "ListFiles", Description: "List files in a directory", Type: ToolTypeBuiltin}
	ToolReadFileRange IATTool = IATTool{Name: "read_file_range", Content: "ReadFileRange", Description: "Read a range of lines from a file", Type: ToolTypeBuiltin}
	ToolDiffFile      IATTool = IATTool{Name: "diff_file", Content: "DiffFile", Description: "Compare two files", Type: ToolTypeBuiltin}
	ToolRunCommand    IATTool = IATTool{Name: "run_command", Content: "RunCommand", Description: "Run a shell command", Type: ToolTypeBuiltin}
	ToolRunScript     IATTool = IATTool{Name: "run_script", Content: "RunScript", Description: "Run a script", Type: ToolTypeBuiltin}
	ToolHttpGet       IATTool = IATTool{Name: "http_get", Content: "HttpGet", Description: "Perform an HTTP GET request", Type: ToolTypeBuiltin}
	ToolHttpPost      IATTool = IATTool{Name: "http_post", Content: "HttpPost", Description: "Perform an HTTP POST request", Type: ToolTypeBuiltin}
	ToolIndexProject  IATTool = IATTool{Name: "index_project", Content: "IndexProject", Description: "Index projects for searching sessions by project name", Type: ToolTypeBuiltin}
)

func (t IATTool) ToString() string {
	return t.Name
}
