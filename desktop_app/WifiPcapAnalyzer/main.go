package main

import (
	"WifiPcapAnalyzer/config"
	"WifiPcapAnalyzer/logger"
	"embed"
	"log" // Standard log for initial errors before zerolog is up

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load configuration first to initialize logger
	config.LoadConfig("config/config.json") // This sets config.GlobalConfig
	if config.GlobalConfig.Logging == nil {
		log.Println("Critical error: GlobalConfig.Logging is nil after loading. Using default logger setup.")
		// Fallback logic if needed, though LoadConfig should prevent this.
		trueVal := true
		logger.InitLogger(&config.LoggingConfig{Level: "info", Console: &trueVal})
	} else {
		logger.InitLogger(config.GlobalConfig.Logging) // Initialize zerolog using the global config
	}

	// Create an instance of the app structure
	app := NewApp() // NewApp might need to accept appConfig if it uses it directly at creation

	// Create application with options
	err := wails.Run(&options.App{
		Title: "WifiPcapAnalyzer",
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup, // app.startup will now use the initialized zerolog
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Error running Wails application: %v", err) // Use log.Fatalf for fatal errors
	}
}
