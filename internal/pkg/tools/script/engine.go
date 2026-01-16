package script

import (
	"fmt"
	"iat/internal/pkg/tools/builtin"

	"github.com/dop251/goja"
)

type ScriptEngine struct {
	vm *goja.Runtime
}

func NewScriptEngine() *ScriptEngine {
	vm := goja.New()
	
	// Register builtin tools to JS environment
	// This allows scripts to use built-in capabilities like file operations
	for name, fn := range builtin.ToolFunctions {
		vm.Set(name, fn)
	}
	
	// Add console.log support
	vm.Set("console", map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value {
			var args []interface{}
			for _, arg := range call.Arguments {
				args = append(args, arg.Export())
			}
			fmt.Println(args...)
			return goja.Undefined()
		},
	})

	return &ScriptEngine{
		vm: vm,
	}
}

// Run executes a JS script
func (e *ScriptEngine) Run(script string) (interface{}, error) {
	val, err := e.vm.RunString(script)
	if err != nil {
		return nil, err
	}
	return val.Export(), nil
}

// RegisterTool registers a Go function as a tool in the JS VM
func (e *ScriptEngine) RegisterTool(name string, fn interface{}) {
	e.vm.Set(name, fn)
}

// GetGlobalObject returns the global object of the VM
func (e *ScriptEngine) GetGlobalObject() *goja.Object {
	return e.vm.GlobalObject()
}
