package modules

import (
	"path/filepath"

	"github.com/dop251/goja"
)

func init() {
	Register(ModuleDoc{
		Name: "path",
		Desc: "Path manipulation utilities",
		Functions: []FunctionDoc{
			{
				Name: "join",
				Desc: "Join path elements",
				Params: []Parameter{
					{Name: "...parts", Type: "string", Desc: "Path parts"},
				},
				Returns: "string",
			},
			{
				Name: "base",
				Desc: "Return the last element of path",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "The path"},
				},
				Returns: "string",
			},
			{
				Name: "dir",
				Desc: "Return all but the last element of path",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "The path"},
				},
				Returns: "string",
			},
			{
				Name: "ext",
				Desc: "Return the file extension",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "The path"},
				},
				Returns: "string",
			},
			{
				Name: "abs",
				Desc: "Return an absolute representation of path",
				Params: []Parameter{
					{Name: "path", Type: "string", Desc: "The path"},
				},
				Returns: "string",
			},
		},
		Register: registerPath,
	})
}

func registerPath(vm *goja.Runtime) {
	vm.Set("path", map[string]interface{}{
		"join": func(call goja.FunctionCall) goja.Value {
			var parts []string
			for _, arg := range call.Arguments {
				parts = append(parts, arg.String())
			}
			return vm.ToValue(filepath.Join(parts...))
		},
		"base": filepath.Base,
		"dir":  filepath.Dir,
		"ext":  filepath.Ext,
		"abs":  filepath.Abs,
	})
}
