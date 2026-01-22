package modules

import (
	"io"
	"net/http"
	"strings"
	"time"

	"iat/engine/pkg/tools/builtin"

	"github.com/dop251/goja"
)

func init() {
	Register(ModuleDoc{
		Name: "http",
		Desc: "HTTP client utilities",
		Functions: []FunctionDoc{
			{
				Name: "get",
				Desc: "Perform a GET request",
				Params: []Parameter{
					{Name: "url", Type: "string", Desc: "The URL to fetch"},
				},
				Returns: "string (body)",
			},
			{
				Name: "post",
				Desc: "Perform a POST request",
				Params: []Parameter{
					{Name: "url", Type: "string", Desc: "The URL to post to"},
					{Name: "contentType", Type: "string", Desc: "MIME type of body"},
					{Name: "body", Type: "string", Desc: "Request body"},
				},
				Returns: "string (body)",
			},
			{
				Name: "request",
				Desc: "Perform a custom HTTP request",
				Params: []Parameter{
					{Name: "method", Type: "string", Desc: "HTTP method (GET, POST, etc.)"},
					{Name: "url", Type: "string", Desc: "The URL"},
					{Name: "headers", Type: "object", Desc: "Map of header key-values"},
					{Name: "body", Type: "string", Desc: "Request body"},
				},
				Returns: "object {status, body, headers}",
			},
		},
		Register: registerHTTP,
	})
}

func registerHTTP(vm *goja.Runtime) {
	vm.Set("http", map[string]interface{}{
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
