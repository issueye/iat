package main

import (
	"context"
	"fmt"
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
	projectService *service.ProjectService
	modelService   *service.AIModelService
	sessionService *service.SessionService
	indexService   *service.IndexService
	scriptService  *service.ScriptService
	agentService   *service.AgentService
	toolService    *service.ToolService
	chatService    *service.ChatService
	modeService    *service.ModeService
}

// NewApp creates a new App application struct
func NewApp(sseHandler *sse.SSEHandler) *App {
	return &App{
		projectService: service.NewProjectService(),
		modelService:   service.NewAIModelService(),
		sessionService: service.NewSessionService(),
		indexService:   service.NewIndexService(),
		scriptService:  service.NewScriptService(),
		agentService:   service.NewAgentService(),
		toolService:    service.NewToolService(),
		chatService:    service.NewChatService(sseHandler),
		modeService:    service.NewModeService(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
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
	if err := a.indexService.IndexProject(id); err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) IndexAllProjects() *common.Result {
	if err := a.indexService.IndexAllProjects(); err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
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

func (a *App) CreateAgent(name, description, systemPrompt string, modelID uint, toolIDs []uint, modeID uint) *common.Result {
	err := a.agentService.CreateAgent(name, description, systemPrompt, modelID, toolIDs, modeID)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) UpdateAgent(id uint, name, description, systemPrompt string, modelID uint, toolIDs []uint, modeID uint) *common.Result {
	err := a.agentService.UpdateAgent(id, name, description, systemPrompt, modelID, toolIDs, modeID)
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

func (a *App) SendMessage(sessionID uint, userMessage string, agentID uint) *common.Result {
	err := a.chatService.Chat(sessionID, userMessage, agentID)
	if err != nil {
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
