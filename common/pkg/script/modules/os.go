package modules

import (
	"os"

	"iat/common/pkg/tools"

	"github.com/dop251/goja"
)

func init() {
	Register(ModuleDoc{
		Name: "os",
		Desc: "Operating System utilities",
		Functions: []FunctionDoc{
			{
				Name: "getenv",
				Desc: "Get environment variable",
				Params: []Parameter{
					{Name: "key", Type: "string", Desc: "Environment variable name"},
				},
				Returns: "string",
			},
			{
				Name: "setenv",
				Desc: "Set environment variable",
				Params: []Parameter{
					{Name: "key", Type: "string", Desc: "Variable name"},
					{Name: "value", Type: "string", Desc: "Variable value"},
				},
				Returns: "error",
			},
			{
				Name: "exec",
				Desc: "Execute shell command",
				Params: []Parameter{
					{Name: "command", Type: "string", Desc: "Command to run"},
					{Name: "args", Type: "[]string", Desc: "Command arguments"},
				},
				Returns: "string (output)",
			},
		},
		Register: registerOS,
	})
}

func registerOS(vm *goja.Runtime) {
	vm.Set("os", map[string]interface{}{
		"getenv": os.Getenv,
		"setenv": os.Setenv,
		"exec":   tools.RunCommand,
	})
}
