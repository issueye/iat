package modules

import (
	"fmt"

	"github.com/dop251/goja"
)

func init() {
	Register(ModuleDoc{
		Name: "console",
		Desc: "Logging utilities",
		Functions: []FunctionDoc{
			{
				Name: "log",
				Desc: "Log a message to stdout",
				Params: []Parameter{
					{Name: "...args", Type: "any", Desc: "Values to log"},
				},
				Returns: "void",
			},
			{
				Name: "error",
				Desc: "Log an error message to stdout",
				Params: []Parameter{
					{Name: "...args", Type: "any", Desc: "Values to log"},
				},
				Returns: "void",
			},
			{
				Name: "warn",
				Desc: "Log a warning message to stdout",
				Params: []Parameter{
					{Name: "...args", Type: "any", Desc: "Values to log"},
				},
				Returns: "void",
			},
		},
		Register: registerConsole,
	})
}

func registerConsole(vm *goja.Runtime) {
	vm.Set("console", map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value {
			args := make([]interface{}, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			fmt.Println(args...)
			return goja.Undefined()
		},
		"error": func(call goja.FunctionCall) goja.Value {
			args := make([]interface{}, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			fmt.Println("ERROR:", fmt.Sprint(args...))
			return goja.Undefined()
		},
		"warn": func(call goja.FunctionCall) goja.Value {
			args := make([]interface{}, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			fmt.Println("WARN:", fmt.Sprint(args...))
			return goja.Undefined()
		},
	})
}
