package main

import (
	"context"
	"encoding/json"
	"iat/internal/model"
	"iat/internal/pkg/common"
	"iat/internal/pkg/script/modules"
	"iat/internal/pkg/sse"
	"iat/internal/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	sseHandler     *sse.SSEHandler
	projectService *service.ProjectService
	modelService   *service.AIModelService
	sessionService *service.SessionService
	indexService   *service.IndexService
	scriptService  *service.ScriptService
	agentService   *service.AgentService
	toolService    *service.ToolService
	chatService    *service.ChatService
	modeService    *service.ModeService
	mcpService     *service.MCPService
	taskService    *service.TaskService
}

// NewApp creates a new App application struct
func NewApp(sseHandler *sse.SSEHandler) *App {
	mcpService := service.NewMCPService()
	taskService := service.NewTaskService(sseHandler)
	return &App{
		sseHandler:     sseHandler,
		projectService: service.NewProjectService(),
		modelService:   service.NewAIModelService(),
		sessionService: service.NewSessionService(),
		indexService:   service.NewIndexService(),
		scriptService:  service.NewScriptService(),
		agentService:   service.NewAgentService(),
		toolService:    service.NewToolService(),
		chatService:    service.NewChatService(mcpService, taskService),
		modeService:    service.NewModeService(),
		mcpService:     mcpService,
		taskService:    taskService,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectDirectory opens a directory selection dialog
func (a *App) SelectDirectory() *common.Result {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(path)
}

// --- Project Methods ---

func (a *App) CreateProject(name, description, path string) *common.Result {
	err := a.projectService.CreateProject(name, description, path)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) UpdateProject(id uint, name, description, path string) *common.Result {
	err := a.projectService.UpdateProject(id, name, description, path)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) ListProjects() *common.Result {
	projects, err := a.projectService.ListProjects()
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(projects)
}

func (a *App) DeleteProject(id uint) *common.Result {
	err := a.projectService.DeleteProject(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) IndexProject(id uint) *common.Result {
	data, err := a.indexService.IndexProject(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(data)
}

func (a *App) IndexAllProjects() *common.Result {
	data, err := a.indexService.IndexAllProjects()
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(data)
}

// --- AI Model Methods ---

func (a *App) CreateAIModel(m model.AIModel) *common.Result {
	err := a.modelService.CreateModel(&m)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) ListAIModels() *common.Result {
	models, err := a.modelService.ListModels()
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(models)
}

func (a *App) DeleteAIModel(id uint) *common.Result {
	err := a.modelService.DeleteModel(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) TestAIModel(m model.AIModel) *common.Result {
	err := a.modelService.TestConnection(&m)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

// --- Session Methods ---

func (a *App) CreateSession(projectID uint, name string, agentID uint) *common.Result {
	err := a.sessionService.CreateSession(projectID, name, agentID)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) ListSessions(projectID uint) *common.Result {
	sessions, err := a.sessionService.ListSessions(projectID)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(sessions)
}

func (a *App) DeleteSession(id uint) *common.Result {
	err := a.sessionService.DeleteSession(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) SearchSessionsByProjectName(query string) *common.Result {
	items, err := a.indexService.SearchSessionsByProjectName(query)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(items)
}

// --- MCP Server Methods ---

func (a *App) CreateMCPServer(name, description, serverType, command, args, env, url string) *common.Result {
	err := a.mcpService.CreateMCPServer(name, description, serverType, command, args, env, url)
	if err != nil {
		return common.Fail(err.Error())
	}

	return common.Success(nil)
}

func (a *App) UpdateMCPServer(id uint, name, description, serverType, command, args, env, url string, enabled bool) *common.Result {
	err := a.mcpService.UpdateMCPServer(id, name, description, serverType, command, args, env, url, enabled)
	if err != nil {
		return common.Fail(err.Error())
	}

	return common.Success(nil)
}

func (a *App) ListMCPServers() *common.Result {
	servers, err := a.mcpService.ListMCPServers()
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(servers)
}

func (a *App) ListMCPTools(id uint) *common.Result {
	tools, err := a.mcpService.ListToolsForServer(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(tools)
}

func (a *App) DeleteMCPServer(id uint) *common.Result {
	err := a.mcpService.DeleteMCPServer(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

// --- Script Methods ---

func (a *App) CreateScript(name, description, content string) *common.Result {
	err := a.scriptService.CreateScript(name, description, content)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) UpdateScript(id uint, name, description, content string) *common.Result {
	err := a.scriptService.UpdateScript(id, name, description, content)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) ListScripts() *common.Result {
	scripts, err := a.scriptService.ListScripts()
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(scripts)
}

func (a *App) DeleteScript(id uint) *common.Result {
	err := a.scriptService.DeleteScript(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) RunScript(id uint) *common.Result {
	res, err := a.scriptService.RunScript(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(res)
}

// GetScriptAPIDocs returns the documentation for the script engine modules
func (a *App) GetScriptAPIDocs() *common.Result {
	return common.Success(modules.GetModuleDocs())
}

// --- Agent Methods ---

func (a *App) CreateAgent(name, description, systemPrompt string, modelID uint, toolIDs []uint, mcpServerIDs []uint, modeID uint) *common.Result {
	err := a.agentService.CreateAgent(name, description, systemPrompt, modelID, toolIDs, mcpServerIDs, modeID)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) UpdateAgent(id uint, name, description, systemPrompt string, modelID uint, toolIDs []uint, mcpServerIDs []uint, modeID uint) *common.Result {
	err := a.agentService.UpdateAgent(id, name, description, systemPrompt, modelID, toolIDs, mcpServerIDs, modeID)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) ListAgents() *common.Result {
	agents, err := a.agentService.ListAgents()
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(agents)
}

func (a *App) DeleteAgent(id uint) *common.Result {
	err := a.agentService.DeleteAgent(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

// --- Mode Methods ---

func (a *App) ListModes() *common.Result {
	modes, err := a.modeService.ListModes()
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(modes)
}

// --- Tool Methods ---

func (a *App) CreateTool(tool model.Tool) *common.Result {
	err := a.toolService.CreateTool(&tool)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) UpdateTool(tool model.Tool) *common.Result {
	err := a.toolService.UpdateTool(&tool)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) ListTools() *common.Result {
	tools, err := a.toolService.ListTools()
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(tools)
}

func (a *App) DeleteTool(id uint) *common.Result {
	err := a.toolService.DeleteTool(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

// --- Chat Methods ---

func (a *App) SendMessage(sessionID uint, userMessage string, agentID uint, mode string) *common.Result {
	eventChan := make(chan service.ChatEvent, 100) // Buffer to decouple service from SSE

	go func() {
		for evt := range eventChan {
			var msg map[string]interface{}
			switch evt.Type {
			case service.ChatEventChunk:
				msg = map[string]interface{}{
					"sessionId": sessionID,
					"delta":     evt.Content,
				}
			case service.ChatEventError:
				msg = map[string]interface{}{
					"sessionId": sessionID,
					"error":     evt.Content,
				}
			case service.ChatEventTerminated:
				msg = map[string]interface{}{
					"sessionId":  sessionID,
					"terminated": true,
					"done":       true,
					"error":      "terminated",
				}
			case service.ChatEventDone:
				msg = evt.Extra
				if msg == nil {
					msg = make(map[string]interface{})
				}
				msg["sessionId"] = sessionID
				msg["done"] = true
			case service.ChatEventUsage:
				msg = evt.Extra
				if msg == nil {
					msg = make(map[string]interface{})
				}
				msg["sessionId"] = sessionID
			default:
				// Tool events and others
				msg = evt.Extra
				if msg == nil {
					msg = make(map[string]interface{})
				}
				msg["sessionId"] = sessionID
			}

			if msg != nil {
				jsonBytes, _ := json.Marshal(msg)
				a.sseHandler.Send(string(jsonBytes))
			}
		}
	}()

	err := a.chatService.Chat(sessionID, userMessage, agentID, mode, eventChan)
	if err != nil {
		close(eventChan)
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) ListMessages(sessionID uint) *common.Result {
	messages, err := a.chatService.ListMessages(sessionID)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(messages)
}

func (a *App) ClearSessionMessages(sessionID uint) *common.Result {
	if err := a.chatService.ClearMessages(sessionID); err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) TerminateSession(sessionID uint) *common.Result {
	a.chatService.AbortSession(sessionID)
	return common.Success(nil)
}

func (a *App) CompressSession(sessionID uint) *common.Result {
	if err := a.chatService.CompressSession(sessionID); err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) ListToolInvocations(sessionID uint) *common.Result {
	items, err := a.chatService.ListToolInvocations(sessionID)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(items)
}

// --- Task Methods ---

func (a *App) ListTasks(sessionID uint) *common.Result {
	tasks, err := a.taskService.ListTasks(sessionID)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(tasks)
}

func (a *App) CreateTask(sessionID uint, content, priority string) *common.Result {
	task, err := a.taskService.CreateTask(sessionID, content, priority, nil)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(task)
}

func (a *App) UpdateTask(id uint, status string) *common.Result {
	err := a.taskService.UpdateTask(id, status)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) DeleteTask(id uint) *common.Result {
	err := a.taskService.DeleteTask(id)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}
