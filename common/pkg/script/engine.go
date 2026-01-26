package script

import (
	"errors"
	"iat/common/pkg/script/modules"
	"time"

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

func (e *ScriptEngine) Run(script string) (interface{}, error) {
	return e.RunWithTimeout(script, 30*time.Second)
}

// Run executes a JS script and returns the exported value with a timeout
func (e *ScriptEngine) RunWithTimeout(script string, timeout time.Duration) (interface{}, error) {
	type result struct {
		val interface{}
		err error
	}

	resCh := make(chan result, 1)

	timer := time.AfterFunc(timeout, func() {
		e.vm.Interrupt("timeout")
	})
	defer timer.Stop()

	go func() {
		val, err := e.vm.RunString(script)
		if err != nil {
			resCh <- result{err: errors.New(err.Error())}
		} else {
			resCh <- result{val: val.Export()}
		}
	}()

	select {
	case res := <-resCh:
		return res.val, res.err
	}
}

// RegisterGlobal registers a Go value or function in the JS global scope
func (e *ScriptEngine) RegisterGlobal(name string, val interface{}) {
	e.vm.Set(name, val)
}

// GetVM returns the underlying goja Runtime
func (e *ScriptEngine) GetVM() *goja.Runtime {
	return e.vm
}
