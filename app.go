package main

import (
	"context"
	"fmt"
	"iat/internal/model"
	"iat/internal/pkg/common"
	"iat/internal/pkg/sse"
	"iat/internal/service"
)

// App struct
type App struct {
	ctx            context.Context
	projectService *service.ProjectService
	modelService   *service.AIModelService
	sessionService *service.SessionService
	scriptService  *service.ScriptService
	agentService   *service.AgentService
	chatService    *service.ChatService
}

// NewApp creates a new App application struct
func NewApp(sseHandler *sse.SSEHandler) *App {
	return &App{
		projectService: service.NewProjectService(),
		modelService:   service.NewAIModelService(),
		sessionService: service.NewSessionService(),
		scriptService:  service.NewScriptService(),
		agentService:   service.NewAgentService(),
		chatService:    service.NewChatService(sseHandler),
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

// --- Project Methods ---

func (a *App) CreateProject(name, description string) *common.Result {
	err := a.projectService.CreateProject(name, description)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) UpdateProject(id uint, name, description string) *common.Result {
	err := a.projectService.UpdateProject(id, name, description)
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

// --- Agent Methods ---

func (a *App) CreateAgent(name, description, systemPrompt string, modelID uint) *common.Result {
	err := a.agentService.CreateAgent(name, description, systemPrompt, modelID)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}

func (a *App) UpdateAgent(id uint, name, description, systemPrompt string, modelID uint) *common.Result {
	err := a.agentService.UpdateAgent(id, name, description, systemPrompt, modelID)
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

// --- Chat Methods ---

func (a *App) SendMessage(sessionID uint, userMessage string) *common.Result {
	err := a.chatService.Chat(sessionID, userMessage)
	if err != nil {
		return common.Fail(err.Error())
	}
	return common.Success(nil)
}
