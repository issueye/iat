package script

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"iat/engine/pkg/tools/builtin"

	"github.com/dop251/goja"
	"github.com/google/uuid"
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
	s.registerConsole()
	s.registerHTTP()
	s.registerFS()
	s.registerPath()
	s.registerOS()
	s.registerUtils()
	s.registerEncoding()

	// Global sleep for backward compatibility
	s.vm.Set("sleep", func(ms int) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	})
}

func (s *ScriptEngine) registerConsole() {
	s.vm.Set("console", map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value {
			args := make([]interface{}, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			fmt.Println(args...)
			return goja.Undefined()
		},
		"error": func(call goja.FunctionCall) goja.Value {
			args := make([]interface{}, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			fmt.Println("ERROR:", fmt.Sprint(args...))
			return goja.Undefined()
		},
		"warn": func(call goja.FunctionCall) goja.Value {
			args := make([]interface{}, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}
			fmt.Println("WARN:", fmt.Sprint(args...))
			return goja.Undefined()
		},
	})
}

func (s *ScriptEngine) registerHTTP() {
	s.vm.Set("http", map[string]interface{}{
		"get":     builtin.HttpGet,
		"post":    builtin.HttpPost,
		"request": func(method, url string, headers map[string]string, body string) (map[string]interface{}, error) {
			client := &http.Client{Timeout: 30 * time.Second}
			req, err := http.NewRequest(method, url, strings.NewReader(body))
			if err != nil {
				return nil, err
			}

			for k, v := range headers {
				req.Header.Set(k, v)
			}

			resp, err := client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			respHeaders := make(map[string]string)
			for k, v := range resp.Header {
				respHeaders[k] = strings.Join(v, ", ")
			}

			return map[string]interface{}{
				"status":  resp.StatusCode,
				"body":    string(respBody),
				"headers": respHeaders,
			}, nil
		},
	})
}

func (s *ScriptEngine) registerFS() {
	s.vm.Set("fs", map[string]interface{}{
		"readFile":  builtin.ReadFile,
		"writeFile": builtin.WriteFile,
		"listFiles": builtin.ListFiles,
		"remove": func(path string) error {
			return os.RemoveAll(path)
		},
		"exists": func(path string) bool {
			_, err := os.Stat(path)
			return err == nil || !os.IsNotExist(err)
		},
		"mkdir": func(path string) error {
			return os.MkdirAll(path, 0755)
		},
	})
}

func (s *ScriptEngine) registerPath() {
	s.vm.Set("path", map[string]interface{}{
		"join": func(call goja.FunctionCall) goja.Value {
			var parts []string
			for _, arg := range call.Arguments {
				parts = append(parts, arg.String())
			}
			return s.vm.ToValue(filepath.Join(parts...))
		},
		"base": filepath.Base,
		"dir":  filepath.Dir,
		"ext":  filepath.Ext,
		"abs":  filepath.Abs,
	})
}

func (s *ScriptEngine) registerOS() {
	s.vm.Set("os", map[string]interface{}{
		"getenv": os.Getenv,
		"setenv": os.Setenv,
		"exec":   builtin.RunCommand,
	})
}

func (s *ScriptEngine) registerUtils() {
	s.vm.Set("utils", map[string]interface{}{
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
}

func (s *ScriptEngine) registerEncoding() {
	s.vm.Set("json", map[string]interface{}{
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

	s.vm.Set("base64", map[string]interface{}{
		"encode": func(str string) string {
			return base64.StdEncoding.EncodeToString([]byte(str))
		},
		"decode": func(str string) (string, error) {
			b, err := base64.StdEncoding.DecodeString(str)
			return string(b), err
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
