package script

import (
	"iat/common/pkg/script"
	"iat/engine/pkg/tools/builtin"

	"github.com/dop251/goja"
)

type ScriptEngine struct {
	engine *script.ScriptEngine
}

func NewScriptEngine() *ScriptEngine {
	commonEngine := script.NewScriptEngine()
	vm := commonEngine.GetVM()

	// Register engine-specific builtin tools to JS environment
	for name, fn := range builtin.ToolFunctions {
		vm.Set(name, fn)
	}

	return &ScriptEngine{
		engine: commonEngine,
	}
}

func NewScriptEngineWithBaseDir(baseDir string) *ScriptEngine {
	engine := NewScriptEngine()
	engine.RegisterTool("baseDir", baseDir)
	return engine
}

// Run executes a JS script
func (e *ScriptEngine) Run(jsScript string) (interface{}, error) {
	return e.engine.Run(jsScript)
}

// RegisterTool registers a Go function as a tool in the JS VM
func (e *ScriptEngine) RegisterTool(name string, fn interface{}) {
	e.engine.RegisterGlobal(name, fn)
}

// GetGlobalObject returns the global object of the VM
func (e *ScriptEngine) GetGlobalObject() *goja.Object {
	return e.engine.GetVM().GlobalObject()
}
