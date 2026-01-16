package script

import (
	"errors"
	"iat/internal/pkg/script/modules"
	"iat/internal/pkg/tools/builtin"
	"os"

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

func NewScriptEngineWithBaseDir(baseDir string) *ScriptEngine {
	vm := goja.New()
	engine := &ScriptEngine{vm: vm}
	engine.registerGlobals()
	engine.registerProjectFS(baseDir)
	return engine
}

func (s *ScriptEngine) registerGlobals() {
	modules.RegisterAll(s.vm)
}

func (s *ScriptEngine) registerProjectFS(baseDir string) {
	if baseDir == "" {
		return
	}
	s.vm.Set("fs", map[string]interface{}{
		"readFile": func(path string) (string, error) {
			p, err := builtin.ResolvePathInBase(baseDir, path)
			if err != nil {
				return "", err
			}
			return builtin.ReadFile(p)
		},
		"writeFile": func(path string, content string) (string, error) {
			p, err := builtin.ResolvePathInBase(baseDir, path)
			if err != nil {
				return "", err
			}
			return builtin.WriteFile(p, content)
		},
		"listFiles": func(path string) (string, error) {
			p, err := builtin.ResolvePathInBase(baseDir, path)
			if err != nil {
				return "", err
			}
			return builtin.ListFiles(p)
		},
		"remove": func(path string) error {
			p, err := builtin.ResolvePathInBase(baseDir, path)
			if err != nil {
				return err
			}
			return os.RemoveAll(p)
		},
		"exists": func(path string) bool {
			p, err := builtin.ResolvePathInBase(baseDir, path)
			if err != nil {
				return false
			}
			_, statErr := os.Stat(p)
			return statErr == nil || !os.IsNotExist(statErr)
		},
		"mkdir": func(path string) error {
			p, err := builtin.ResolvePathInBase(baseDir, path)
			if err != nil {
				return err
			}
			return os.MkdirAll(p, 0755)
		},
	})
}

func (s *ScriptEngine) Run(script string) (goja.Value, error) {
	val, err := s.vm.RunString(script)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return val, nil
}
