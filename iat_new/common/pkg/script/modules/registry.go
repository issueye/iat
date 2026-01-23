package modules

import "github.com/dop251/goja"

type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Desc string `json:"desc"`
}

type FunctionDoc struct {
	Name    string      `json:"name"`
	Desc    string      `json:"desc"`
	Params  []Parameter `json:"params"`
	Returns string      `json:"returns"`
}

type ModuleDoc struct {
	Name      string                 `json:"name"`
	Desc      string                 `json:"desc"`
	Functions []FunctionDoc          `json:"functions"`
	Register  func(vm *goja.Runtime) `json:"-"`
}

var registry []ModuleDoc

func Register(m ModuleDoc) {
	registry = append(registry, m)
}

func RegisterAll(vm *goja.Runtime) {
	for _, m := range registry {
		if m.Register != nil {
			m.Register(vm)
		}
	}
}

func GetModuleDocs() []ModuleDoc {
	return registry
}
