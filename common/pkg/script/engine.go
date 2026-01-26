package script

import (
	"errors"
	"iat/common/pkg/script/modules"

	"github.com/dop251/goja"
)

type ScriptEngine struct {
	vm *goja.Runtime
}

func NewScriptEngine() *ScriptEngine {
	vm := goja.New()
	modules.RegisterAll(vm)
	return &ScriptEngine{vm: vm}
}

// Run executes a JS script and returns the exported value
func (e *ScriptEngine) Run(script string) (interface{}, error) {
	val, err := e.vm.RunString(script)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return val.Export(), nil
}

// RegisterGlobal registers a Go value or function in the JS global scope
func (e *ScriptEngine) RegisterGlobal(name string, val interface{}) {
	e.vm.Set(name, val)
}

// GetVM returns the underlying goja Runtime
func (e *ScriptEngine) GetVM() *goja.Runtime {
	return e.vm
}
