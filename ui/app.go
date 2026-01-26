package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"iat/common/model"
	"iat/common/pkg/result"
	"iat/common/pkg/script/modules"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectDirectory opens a directory selection dialog
func (a *App) SelectDirectory() *result.Result {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	if err != nil {
		return result.Fail(err.Error())
	}
	return result.Success(path)
}

// GetScriptAPIDocs returns the documentation for the script engine modules
func (a *App) GetScriptAPIDocs() *result.Result {
	return result.Success(modules.GetModuleDocs())
}

// --- Proxy Methods to Engine ---

const EngineURL = "http://localhost:8080/api"

// Helper for HTTP requests
func (a *App) proxyRequest(method, endpoint string, payload interface{}) *result.Result {
	var body io.Reader
	if payload != nil {
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return result.Fail("Failed to marshal payload: " + err.Error())
		}
		body = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, EngineURL+endpoint, body)
	if err != nil {
		return result.Fail("Failed to create request: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result.Fail("Failed to connect to engine: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return result.Fail(fmt.Sprintf("Engine error (%d): %s", resp.StatusCode, string(respBody)))
	}

	var data interface{}
	// Only try to decode if content-length > 0 or not empty
	if resp.ContentLength != 0 {
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			if err != io.EOF {
				// Ignore
			}
		}
	}

	return result.Success(data)
}

// --- Project Methods ---

func (a *App) ListProjects() *result.Result {
	return a.proxyRequest(http.MethodGet, "/projects", nil)
}

func (a *App) CreateProject(name, description, path string) *result.Result {
	payload := map[string]string{
		"name":        name,
		"description": description,
		"path":        path,
	}
	return a.proxyRequest(http.MethodPost, "/projects", payload)
}

func (a *App) UpdateProject(id uint, name, description, path string) *result.Result {
	payload := map[string]string{
		"name":        name,
		"description": description,
		"path":        path,
	}
	return a.proxyRequest(http.MethodPut, fmt.Sprintf("/projects/%d", id), payload)
}

func (a *App) DeleteProject(id uint) *result.Result {
	return a.proxyRequest(http.MethodDelete, fmt.Sprintf("/projects/%d", id), nil)
}

func (a *App) IndexProject(id uint) *result.Result {
	return a.proxyRequest(http.MethodPost, fmt.Sprintf("/projects/%d/index", id), nil)
}

func (a *App) IndexAllProjects() *result.Result {
	return a.proxyRequest(http.MethodPost, "/projects/index-all", nil)
}

// --- AI Model Methods ---

func (a *App) ListAIModels() *result.Result {
	return a.proxyRequest(http.MethodGet, "/models", nil)
}

func (a *App) CreateAIModel(m model.AIModel) *result.Result {
	return a.proxyRequest(http.MethodPost, "/models", m)
}

func (a *App) DeleteAIModel(id uint) *result.Result {
	return a.proxyRequest(http.MethodDelete, fmt.Sprintf("/models/%d", id), nil)
}

func (a *App) TestAIModel(m model.AIModel) *result.Result {
	return a.proxyRequest(http.MethodPost, "/models/test", m)
}

// --- Agent Methods ---

func (a *App) ListAgents() *result.Result {
	return a.proxyRequest(http.MethodGet, "/agents", nil)
}

func (a *App) CreateAgent(name, description, systemPrompt, agentType, externalURL, externalType, externalParams string, modelID uint, toolIDs []uint, mcpServerIDs []uint, modeID uint, status string, capabilities string) *result.Result {
	payload := map[string]interface{}{
		"name":           name,
		"description":    description,
		"systemPrompt":   systemPrompt,
		"type":           agentType,
		"modelId":        modelID,
		"toolIds":        toolIDs,
		"mcpServerIds":   mcpServerIDs,
		"modeId":         modeID,
		"externalUrl":    externalURL,
		"externalType":   externalType,
		"externalParams": externalParams,
		"status":         status,
		"capabilities":   capabilities,
	}
	return a.proxyRequest(http.MethodPost, "/agents", payload)
}

func (a *App) UpdateAgent(id uint, name, description, systemPrompt, agentType, externalURL, externalType, externalParams string, modelID uint, toolIDs []uint, mcpServerIDs []uint, modeID uint, status string, capabilities string) *result.Result {
	payload := map[string]interface{}{
		"name":           name,
		"description":    description,
		"systemPrompt":   systemPrompt,
		"type":           agentType,
		"modelId":        modelID,
		"toolIds":        toolIDs,
		"mcpServerIds":   mcpServerIDs,
		"modeId":         modeID,
		"externalUrl":    externalURL,
		"externalType":   externalType,
		"externalParams": externalParams,
		"status":         status,
		"capabilities":   capabilities,
	}
	return a.proxyRequest(http.MethodPut, fmt.Sprintf("/agents/%d", id), payload)
}

func (a *App) DeleteAgent(id uint) *result.Result {
	return a.proxyRequest(http.MethodDelete, fmt.Sprintf("/agents/%d", id), nil)
}

// --- Tool Methods ---

func (a *App) ListTools() *result.Result {
	return a.proxyRequest(http.MethodGet, "/tools", nil)
}

func (a *App) CreateTool(tool model.Tool) *result.Result {
	return a.proxyRequest(http.MethodPost, "/tools", tool)
}

func (a *App) UpdateTool(tool model.Tool) *result.Result {
	return a.proxyRequest(http.MethodPut, "/tools", tool)
}

func (a *App) DeleteTool(id uint) *result.Result {
	return a.proxyRequest(http.MethodDelete, fmt.Sprintf("/tools/%d", id), nil)
}

// --- MCP Server Methods ---

func (a *App) ListMCPServers() *result.Result {
	return a.proxyRequest(http.MethodGet, "/mcp", nil)
}

func (a *App) CreateMCPServer(name, description, serverType, command, args, env, url string) *result.Result {
	payload := map[string]interface{}{
		"name":        name,
		"description": description,
		"serverType":  serverType,
		"command":     command,
		"args":        args,
		"env":         env,
		"url":         url,
	}
	return a.proxyRequest(http.MethodPost, "/mcp", payload)
}

func (a *App) UpdateMCPServer(id uint, name, description, serverType, command, args, env, url string, enabled bool) *result.Result {
	payload := map[string]interface{}{
		"name":        name,
		"description": description,
		"serverType":  serverType,
		"command":     command,
		"args":        args,
		"env":         env,
		"url":         url,
		"enabled":     enabled,
	}
	return a.proxyRequest(http.MethodPut, fmt.Sprintf("/mcp/%d", id), payload)
}

func (a *App) DeleteMCPServer(id uint) *result.Result {
	return a.proxyRequest(http.MethodDelete, fmt.Sprintf("/mcp/%d", id), nil)
}

func (a *App) ListMCPTools(id uint) *result.Result {
	return a.proxyRequest(http.MethodGet, fmt.Sprintf("/mcp/%d/tools", id), nil)
}

// --- Mode Methods ---

func (a *App) ListModes() *result.Result {
	return a.proxyRequest(http.MethodGet, "/modes", nil)
}

// --- Session Methods ---

func (a *App) ListSessions(projectId uint) *result.Result {
	return a.proxyRequest(http.MethodGet, fmt.Sprintf("/sessions?projectId=%d", projectId), nil)
}

func (a *App) CreateSession(name string, projectId uint, agentId uint) *result.Result {
	payload := map[string]interface{}{
		"name":      name,
		"projectId": projectId,
		"agentId":   agentId,
	}
	return a.proxyRequest(http.MethodPost, "/sessions", payload)
}

func (a *App) UpdateSession(id uint, name string) *result.Result {
	payload := map[string]interface{}{
		"name": name,
	}
	return a.proxyRequest(http.MethodPut, fmt.Sprintf("/sessions/%d", id), payload)
}

func (a *App) DeleteSession(id uint) *result.Result {
	return a.proxyRequest(http.MethodDelete, fmt.Sprintf("/sessions/%d", id), nil)
}

func (a *App) GetSessionMessages(id uint) *result.Result {
	return a.proxyRequest(http.MethodGet, fmt.Sprintf("/sessions/%d/messages", id), nil)
}
