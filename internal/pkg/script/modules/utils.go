package modules

import (
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

func init() {
	Register(ModuleDoc{
		Name: "utils",
		Desc: "General purpose utilities",
		Functions: []FunctionDoc{
			{
				Name: "uuid",
				Desc: "Generate a new UUID",
				Params: []Parameter{},
				Returns: "string",
			},
			{
				Name: "sleep",
				Desc: "Pause execution",
				Params: []Parameter{
					{Name: "ms", Type: "number", Desc: "Milliseconds to sleep"},
				},
				Returns: "void",
			},
			{
				Name: "now",
				Desc: "Get current timestamp (seconds)",
				Params: []Parameter{},
				Returns: "number",
			},
			{
				Name: "nowMS",
				Desc: "Get current timestamp (milliseconds)",
				Params: []Parameter{},
				Returns: "number",
			},
		},
		Register: registerUtils,
	})
}

func registerUtils(vm *goja.Runtime) {
	vm.Set("utils", map[string]interface{}{
		"uuid": func() string {
			return uuid.New().String()
		},
		"sleep": func(ms int) {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		},
		"now": func() int64 {
			return time.Now().Unix()
		},
		"nowMS": func() int64 {
			return time.Now().UnixMilli()
		},
	})
	
	// Global sleep for backward compatibility
	vm.Set("sleep", func(ms int) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	})
}
