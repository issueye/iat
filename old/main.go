package main

import (
	"embed"
	"iat/internal/pkg/db"
	"iat/internal/pkg/logger"
	"iat/internal/pkg/sse"
	"log"
	"net/http"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Init Logger
	logger.InitLogger()

	// Init DB
	if err := db.InitDB(); err != nil {
		log.Fatal("Failed to init DB:", err)
	}

	// Init SSE
	sseHandler := sse.NewSSEHandler()
	go func() {
		http.Handle("/events", sseHandler)
		log.Println("SSE Server started on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("SSE Server failed:", err)
		}
	}()

	// Create an instance of the app structure
	app := NewApp(sseHandler)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "iat",
		Width:  1500,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
