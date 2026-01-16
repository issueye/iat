package script

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

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
	// Console
	s.vm.Set("console", map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value {
			args := make([]interface{}, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			fmt.Println(args...)
			return goja.Undefined()
		},
	})

	// HTTP
	s.vm.Set("http", map[string]interface{}{
		"get": func(url string) (string, error) {
			resp, err := http.Get(url)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			return string(body), nil
		},
	})

	// Sleep
	s.vm.Set("sleep", func(ms int) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	})
}

func (s *ScriptEngine) Run(script string) (goja.Value, error) {
	val, err := s.vm.RunString(script)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return val, nil
}
