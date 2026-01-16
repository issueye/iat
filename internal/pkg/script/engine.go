package script

import (
	"errors"
	"iat/internal/pkg/script/modules"

	"github.com/dop251/goja"
)

type ScriptEngine struct {
	vm *goja.Runtime
}

func NewScriptEngine() *ScriptEngine {
	vm := goja.New()
	engine := &ScriptEngine{vm: vm}
	engine.registerGlobals()
	return engine
}

func (s *ScriptEngine) registerGlobals() {
	modules.RegisterAll(s.vm)
}

func (s *ScriptEngine) Run(script string) (goja.Value, error) {
	val, err := s.vm.RunString(script)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return val, nil
}
