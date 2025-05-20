package config

import (
	"encoding/json"
	"io/ioutil"
	"log" // Standard log is used here as zerolog is not yet initialized
	"os"
)

// AppConfig holds the application configuration.
type AppConfig struct {
	GRPCServerAddress  string         `json:"grpc_server_address"`
	WebSocketAddress   string         `json:"websocket_address"`
	LogFile            string         `json:"log_file"`  // Deprecated by LoggingConfig
	LogLevel           string         `json:"log_level"` // Deprecated by LoggingConfig
	MinBSSCreationRSSI int            `json:"min_bss_creation_rssi"`
	Logging            *LoggingConfig `json:"logging,omitempty"`
}

// LoggingConfig holds the logging configuration.
type LoggingConfig struct {
	Level   string  `json:"level"`             // e.g., "debug", "info", "warn", "error"
	File    *string `json:"file,omitempty"`    // Optional: path to log file
	Console *bool   `json:"console,omitempty"` // Optional: enable/disable console logging
}

// DefaultConfig provides a default configuration.
var DefaultConfig = AppConfig{
	GRPCServerAddress:  "192.168.6.250:50051", // Default gRPC server address
	WebSocketAddress:   "0.0.0.0:8080",        // Default WebSocket server address
	MinBSSCreationRSSI: -84,                   // Default minimum RSSI for BSS creation
	Logging: &LoggingConfig{
		Level:   "info",
		Console: func(b bool) *bool { return &b }(true), // Default console to true
		File:    nil,                                    // Default no file logging
	},
}

// GlobalConfig holds the global application configuration.
// It's populated by LoadConfig at startup.
var GlobalConfig AppConfig

// LoadConfig loads configuration from a JSON file or returns default configuration.
// If the file path is empty, it returns the default config.
// If the file path is specified but the file doesn't exist or is invalid,
// it logs an error and returns the default config.
// It also sets the GlobalConfig variable.
func LoadConfig(filePath string) AppConfig {
	var cfg AppConfig
	if filePath == "" {
		// log.Println("No configuration file path provided, using default configuration.")
		cfg = DefaultConfig
	} else {
		log.Printf("Attempting to load configuration from: %s\n", filePath)
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("Configuration file '%s' not found. Using default configuration.\n", filePath)
			} else {
				log.Printf("Error reading configuration file '%s': %v. Using default configuration.\n", filePath, err)
			}
			cfg = DefaultConfig
		} else {
			err = json.Unmarshal(data, &cfg)
			if err != nil {
				log.Printf("Error unmarshalling configuration data from '%s': %v. Using default configuration.\n", filePath, err)
				cfg = DefaultConfig
			} else {
				log.Printf("Configuration loaded successfully from '%s'.\n", filePath)
			}
		}
	}

	// Fill in any missing fields with defaults if necessary
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
	// If MinBSSCreationRSSI is 0 (Go's zero value for int), it means it was either
	// not present in the JSON or explicitly set to 0. We apply the default in this case.
	if cfg.MinBSSCreationRSSI == 0 {
		cfg.MinBSSCreationRSSI = DefaultConfig.MinBSSCreationRSSI
		log.Printf("MinBSSCreationRSSI not found or set to 0 in config, using default value: %d\n", DefaultConfig.MinBSSCreationRSSI)
	}
	if cfg.Logging == nil {
		cfg.Logging = DefaultConfig.Logging
		// log.Println("Logging configuration not found, using default logging settings.")
	} else {
		if cfg.Logging.Level == "" {
			cfg.Logging.Level = DefaultConfig.Logging.Level
		}
		if cfg.Logging.Console == nil {
			cfg.Logging.Console = DefaultConfig.Logging.Console
		}
		// File can be nil by default, so no specific default fill needed if it's missing, unless we want to force a default file path.
	}
	// Deprecate old LogFile and LogLevel if new Logging is present
	if cfg.Logging != nil {
		if cfg.LogFile != "" {
			log.Printf("Warning: Deprecated 'log_file' field ('%s') found in config. Please use 'logging.file' instead.", cfg.LogFile)
			// Optionally, migrate if new file logging is not set:
			// if cfg.Logging.File == nil || *cfg.Logging.File == "" {
			//  cfg.Logging.File = &cfg.LogFile
			// }
		}
		if cfg.LogLevel != "" {
			log.Printf("Warning: Deprecated 'log_level' field ('%s') found in config. Please use 'logging.level' instead.", cfg.LogLevel)
			// Optionally, migrate if new level is not set:
			// if cfg.Logging.Level == "" {
			// 	cfg.Logging.Level = cfg.LogLevel
			// }
		}
	}

	GlobalConfig = cfg // Set the global config
	return cfg
}
