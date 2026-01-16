package main

import (
	"context"
	"fmt"
	"iat/internal/model"
	"iat/internal/pkg/common"
	"iat/internal/service"
)

// App struct
type App struct {
	ctx            context.Context
	projectService *service.ProjectService
	modelService   *service.AIModelService
	sessionService *service.SessionService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		projectService: service.NewProjectService(),
		modelService:   service.NewAIModelService(),
		sessionService: service.NewSessionService(),
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

func (a *App) CreateSession(projectID uint, name string) *common.Result {
	err := a.sessionService.CreateSession(projectID, name)
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
