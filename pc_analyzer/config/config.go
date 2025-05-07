package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// AppConfig holds the application configuration.
type AppConfig struct {
	GRPCServerAddress string `json:"grpc_server_address"`
	WebSocketAddress  string `json:"websocket_address"`
	LogFile           string `json:"log_file"`
	LogLevel          string `json:"log_level"`
}

// DefaultConfig provides a default configuration.
var DefaultConfig = AppConfig{
	GRPCServerAddress: "192.168.110.1:50051", // Default gRPC server address
	WebSocketAddress:  "0.0.0.0:8080",        // Default WebSocket server address
	LogFile:           "pc_analyzer.log",     // Default log file
	LogLevel:          "info",                // Default log level (e.g., debug, info, warn, error)
}

// LoadConfig loads configuration from a JSON file or returns default configuration.
// If the file path is empty, it returns the default config.
// If the file path is specified but the file doesn't exist or is invalid,
// it logs an error and returns the default config.
func LoadConfig(filePath string) AppConfig {
	if filePath == "" {
		log.Println("No configuration file path provided, using default configuration.")
		return DefaultConfig
	}

	log.Printf("Attempting to load configuration from: %s\n", filePath)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Configuration file '%s' not found. Using default configuration.\n", filePath)
		} else {
			log.Printf("Error reading configuration file '%s': %v. Using default configuration.\n", filePath, err)
		}
		return DefaultConfig
	}

	var cfg AppConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		log.Printf("Error unmarshalling configuration data from '%s': %v. Using default configuration.\n", filePath, err)
		return DefaultConfig
	}

	log.Printf("Configuration loaded successfully from '%s'.\n", filePath)
	// Fill in any missing fields with defaults if necessary (optional, depends on desired behavior)
	if cfg.GRPCServerAddress == "" {
		cfg.GRPCServerAddress = DefaultConfig.GRPCServerAddress
	}
	if cfg.WebSocketAddress == "" {
		cfg.WebSocketAddress = DefaultConfig.WebSocketAddress
	}
	if cfg.LogFile == "" {
		cfg.LogFile = DefaultConfig.LogFile
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = DefaultConfig.LogLevel
	}

	return cfg
}
