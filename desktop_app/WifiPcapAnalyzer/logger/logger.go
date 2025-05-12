package logger

import (
	"WifiPcapAnalyzer/config"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var (
	// Log is the global logger instance
	Log zerolog.Logger
)

// InitLogger initializes the global logger based on the provided configuration.
func InitLogger(cfg *config.LoggingConfig) {
	var writers []io.Writer

	// Console writer
	if cfg.Console != nil && *cfg.Console {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			FormatLevel: func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("[%s]", i))
			},
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("%s", i)
			},
		}
		writers = append(writers, consoleWriter)
	}

	// File writer
	if cfg.File != nil && *cfg.File != "" {
		file, err := os.OpenFile(*cfg.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			// Fallback to console if file opening fails
			fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v. Logging to console only.\n", *cfg.File, err)
			if len(writers) == 0 { // If console wasn't already added
				consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
				writers = append(writers, consoleWriter)
			}
		} else {
			writers = append(writers, file)
		}
	}

	if len(writers) == 0 {
		// Default to console if no writers are configured (should not happen with default config)
		fmt.Fprintln(os.Stderr, "No log writers configured, defaulting to stderr console.")
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
		writers = append(writers, consoleWriter)
	}

	multiWriter := zerolog.MultiLevelWriter(writers...)
	Log = zerolog.New(multiWriter).With().Timestamp().Logger()

	// Set global log level
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level '%s': %v. Defaulting to 'info'.\n", cfg.Level, err)
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	Log.Info().Msgf("Logger initialized. Level: %s", level.String())
}
