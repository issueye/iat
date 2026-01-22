package main

import (
	"context"
	"embed"
	"log"
	"os/exec"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/dist
var assets embed.FS

type App struct {
	ctx       context.Context
	engineCmd *exec.Cmd
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Start Engine
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("./engine.exe")
	} else {
		cmd = exec.Command("./engine")
	}
	
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start engine: %v", err)
	} else {
		a.engineCmd = cmd
		log.Printf("Engine started with PID: %d", cmd.Process.Pid)
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.engineCmd != nil {
		if err := a.engineCmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill engine: %v", err)
		}
	}
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "IAT",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
