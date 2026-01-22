package modules

import (
	"os"

	"iat/internal/pkg/tools/builtin"

	"github.com/dop251/goja"
)

func init() {
	Register(ModuleDoc{
		Name: "fs",
		Desc: "File System operations",
		Functions: []FunctionDoc{
			{
				Name: "readFile",
				Desc: "Read file content",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "Path to file"},
				},
				Returns: "string",
			},
			{
				Name: "writeFile",
				Desc: "Write content to file",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "Path to file"},
					{Name: "content", Type: "string", Desc: "Content to write"},
				},
				Returns: "string (status)",
			},
			{
				Name: "listFiles",
				Desc: "List files in directory",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "Directory path"},
				},
				Returns: "string (formatted list)",
			},
			{
				Name: "remove",
				Desc: "Remove file or directory",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "Path to remove"},
				},
				Returns: "error",
			},
			{
				Name: "exists",
				Desc: "Check if path exists",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "Path to check"},
				},
				Returns: "boolean",
			},
			{
				Name: "mkdir",
				Desc: "Create directory (recursively)",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "Directory path"},
				},
				Returns: "error",
			},
		},
		Register: registerFS,
	})
}

func registerFS(vm *goja.Runtime) {
	vm.Set("fs", map[string]interface{}{
		"readFile":  builtin.ReadFile,
		"writeFile": builtin.WriteFile,
		"listFiles": builtin.ListFiles,
		"remove": func(path string) error {
			return os.RemoveAll(path)
		},
		"exists": func(path string) bool {
			_, err := os.Stat(path)
			return err == nil || !os.IsNotExist(err)
		},
		"mkdir": func(path string) error {
			return os.MkdirAll(path, 0755)
		},
	})
}
