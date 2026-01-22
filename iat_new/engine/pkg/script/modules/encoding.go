package modules

import (
	"encoding/base64"
	"encoding/json"

	"github.com/dop251/goja"
)

func init() {
	Register(ModuleDoc{
		Name: "json",
		Desc: "JSON encoding/decoding",
		Functions: []FunctionDoc{
			{
				Name: "stringify",
				Desc: "Convert object to JSON string",
				Params: []Parameter{
					{Name: "value", Type: "any", Desc: "Value to stringify"},
				},
				Returns: "string",
			},
			{
				Name: "parse",
				Desc: "Parse JSON string",
				Params: []Parameter{
					{Name: "jsonString", Type: "string", Desc: "JSON string to parse"},
				},
				Returns: "any",
			},
		},
		Register: registerJSON,
	})

	Register(ModuleDoc{
		Name: "base64",
		Desc: "Base64 encoding/decoding",
		Functions: []FunctionDoc{
			{
				Name: "encode",
				Desc: "Encode string to Base64",
				Params: []Parameter{
					{Name: "data", Type: "string", Desc: "String to encode"},
				},
				Returns: "string",
			},
			{
				Name: "decode",
				Desc: "Decode Base64 string",
				Params: []Parameter{
					{Name: "data", Type: "string", Desc: "Base64 string to decode"},
				},
				Returns: "string",
			},
		},
		Register: registerBase64,
	})
}

func registerEncoding(vm *goja.Runtime) {
	// Deprecated: used by individual registers now
}

func registerJSON(vm *goja.Runtime) {
	vm.Set("json", map[string]interface{}{
		"stringify": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
		"parse": func(str string) (interface{}, error) {
			var v interface{}
			err := json.Unmarshal([]byte(str), &v)
			return v, err
		},
	})
}

func registerBase64(vm *goja.Runtime) {
	vm.Set("base64", map[string]interface{}{
		"encode": func(str string) string {
			return base64.StdEncoding.EncodeToString([]byte(str))
		},
		"decode": func(str string) (string, error) {
			b, err := base64.StdEncoding.DecodeString(str)
			return string(b), err
		},
	})
}
