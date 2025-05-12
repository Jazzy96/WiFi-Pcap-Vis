package main

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"WifiPcapAnalyzer/config"
	"WifiPcapAnalyzer/frame_parser"
	"WifiPcapAnalyzer/grpc_client"
	"WifiPcapAnalyzer/logger" // Import the new logger package
	router_agent_pb "WifiPcapAnalyzer/router_agent_pb"
	"WifiPcapAnalyzer/state_manager"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context

	grpcClient          *grpc_client.CaptureAgentClient
	stateMgr            *state_manager.StateManager
	appConfig           config.AppConfig
	packetInfoHandler   frame_parser.PacketInfoHandler
	pcapStreamHandler   grpc_client.PcapStreamHandler
	captureStreamCancel context.CancelFunc
	captureStreamMutex  sync.Mutex
	isCaptureActive     atomic.Bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.Log.Info().Msg("Wails App starting up...")

	// Configuration is already loaded in main.go before logger initialization
	// We can access it via config.GlobalConfig or pass it to NewApp if needed
	a.appConfig = config.GlobalConfig
	logger.Log.Info().Interface("config", a.appConfig).Msg("Configuration loaded")

	// Initialize State Manager with metrics calculation parameters
	metricsInterval := 1 * time.Second // Calculate metrics every second
	historyPoints := 60                // Keep 60 seconds of history
	a.stateMgr = state_manager.NewStateManager(metricsInterval, historyPoints)
	logger.Log.Info().
		Dur("metricsInterval", metricsInterval).
		Int("historyPoints", historyPoints).
		Msg("State Manager initialized.")

	// Initialize Packet Info Handler
	a.packetInfoHandler = func(frame *frame_parser.ParsedFrameInfo) {
		if frame != nil {
			// Log before calling StateManager updates
			// Note: ProcessParsedFrame handles both BSS and STA updates internally.
			a.stateMgr.ProcessParsedFrame(frame)
			// Snapshot broadcasting will be handled by a ticker using Wails events
		}
	}

	// Initialize PCAP Stream Handler
	a.pcapStreamHandler = func(pcapStream io.Reader) {
		logger.Log.Info().Msg("Wails pcapStreamHandler invoked, starting ProcessPcapStream.")
		// The pcapStream is an io.Reader directly from the gRPC client.
		// It will be piped to tshark's stdin.
		err := frame_parser.ProcessPcapStream(pcapStream, a.appConfig.TsharkPath, a.packetInfoHandler)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Error processing pcap stream with tshark")
			runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Error processing pcap stream with tshark: %v", err))
		}
	}

	// Initialize gRPC Client
	var err error
	a.grpcClient, err = grpc_client.Connect(a.appConfig.GRPCServerAddress)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("address", a.appConfig.GRPCServerAddress).Msg("Failed to connect to gRPC server")
		// In a real app, might want to show an error to the user via Wails dialog
		runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Failed to connect to gRPC server: %v", err))
		return
	}
	logger.Log.Info().Str("address", a.appConfig.GRPCServerAddress).Msg("gRPC client connected.")

	// Goroutine to periodically send state snapshot to WebSocket clients via Wails events
	snapshotTicker := time.NewTicker(500 * time.Millisecond) // Send updates every 500 milliseconds
	go func() {
		defer snapshotTicker.Stop()
		for {
			select {
			case <-snapshotTicker.C:
				if a.isCaptureActive.Load() {
					snapshot := a.stateMgr.GetSnapshot()
					runtime.EventsEmit(a.ctx, "state_snapshot", snapshot)
				}
			case <-a.ctx.Done(): // App is shutting down
				logger.Log.Info().Msg("Snapshot ticker stopping due to app context done.")
				return
			}
		}
	}()
	logger.Log.Info().Msg("Snapshot emission goroutine started.")

	// Goroutine to periodically prune old entries from state manager
	pruneTicker := time.NewTicker(30 * time.Second) // Prune every 30 seconds
	go func() {
		defer pruneTicker.Stop()
		for {
			select {
			case <-pruneTicker.C:
				a.stateMgr.PruneOldEntries(2 * time.Minute) // Timeout of 2 minutes
				logger.Log.Info().Msg("Pruned old entries from state manager.")
			case <-a.ctx.Done(): // App is shutting down
				logger.Log.Info().Msg("Pruning ticker stopping due to app context done.")
				return
			}
		}
	}()
	logger.Log.Info().Msg("Pruning goroutine started.")

	// Goroutine to periodically calculate metrics
	metricsTicker := time.NewTicker(metricsInterval)
	go func() {
		defer metricsTicker.Stop()
		for {
			select {
			case <-metricsTicker.C:
				if a.isCaptureActive.Load() { // Only calculate if capture is active
					a.stateMgr.PeriodicallyCalculateMetrics()
				}
			case <-a.ctx.Done(): // App is shutting down
				logger.Log.Info().Msg("Metrics calculation ticker stopping due to app context done.")
				return
			}
		}
	}()
	logger.Log.Info().Msg("Metrics calculation goroutine started.")

	logger.Log.Info().Msg("Wails App startup complete.")
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(ctx context.Context) {
	logger.Log.Info().Msg("Wails App shutting down...")
	if a.grpcClient != nil {
		a.grpcClient.Close()
		logger.Log.Info().Msg("gRPC client closed.")
	}
	a.captureStreamMutex.Lock()
	if a.captureStreamCancel != nil {
		a.captureStreamCancel()
		logger.Log.Info().Msg("Capture stream cancelled.")
	}
	a.captureStreamMutex.Unlock()
	logger.Log.Info().Msg("Wails App shutdown complete.")
}

// StartCapture initiates packet capture via gRPC.
// Exposed to the frontend.
func (a *App) StartCapture(interfaceName string, channel int32, bandwidth string, bpfFilter string) error {
	logger.Log.Info().
		Str("interface", interfaceName).
		Int32("channel", channel).
		Str("bandwidth", bandwidth).
		Str("filter", bpfFilter).
		Msg("StartCapture called")
	if a.grpcClient == nil {
		return fmt.Errorf("gRPC client not initialized")
	}

	if interfaceName == "" {
		return fmt.Errorf("interface name cannot be empty")
	}

	grpcReq := &router_agent_pb.ControlRequest{
		CommandType:   router_agent_pb.ControlCommandType_START_CAPTURE,
		InterfaceName: interfaceName,
		Channel:       channel,
		Bandwidth:     bandwidth,
		BpfFilter:     bpfFilter,
	}

	// Send START_CAPTURE command
	cmdCtx, cmdCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cmdCancel()
	_, err := a.grpcClient.SendControlCommand(cmdCtx, grpcReq)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error sending START_CAPTURE gRPC command")
		return fmt.Errorf("failed to send START_CAPTURE command: %w", err)
	}
	logger.Log.Info().Msg("Successfully sent START_CAPTURE gRPC command.")

	// Clear existing state before starting a new capture session
	if a.stateMgr != nil {
		logger.Log.Info().Msg("Clearing previous BSS/STA state before starting new capture.")
		a.stateMgr.ClearState()
	}

	a.captureStreamMutex.Lock()
	// Cancel any previous stream before starting a new one
	if a.captureStreamCancel != nil {
		logger.Log.Info().Msg("Cancelling previous capture stream before starting new one...")
		a.captureStreamCancel()
		a.captureStreamCancel = nil
	}

	// Create new context and cancel function for this stream
	streamCtx, streamCancel := context.WithCancel(context.Background())
	a.captureStreamCancel = streamCancel
	a.captureStreamMutex.Unlock()

	// Start the streaming in a new goroutine
	go func() {
		logger.Log.Info().Str("interface", interfaceName).Msg("Starting new gRPC packet stream goroutine.")
		err := a.grpcClient.StreamPackets(streamCtx, grpcReq, a.pcapStreamHandler)
		if err != nil && err != context.Canceled {
			logger.Log.Error().Err(err).Str("interface", interfaceName).Msg("Error during packet stream")
			runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Packet stream error: %v", err))
		} else if err == context.Canceled {
			logger.Log.Info().Str("interface", interfaceName).Msg("Packet stream cancelled successfully.")
		} else {
			logger.Log.Info().Str("interface", interfaceName).Msg("Packet stream finished without error.")
		}
		// Ensure capture active is false if stream ends not by explicit stop
		// This might need more robust handling if a new stream is started before this one ends.
		// For now, StopCapture is the primary way to set it false.
	}()

	logger.Log.Info().Str("interface", interfaceName).Msg("Packet streaming goroutine initiated.")
	a.isCaptureActive.Store(true)
	runtime.EventsEmit(a.ctx, "capture_status", "started")
	return nil
}

// StopCapture stops the packet capture via gRPC.
// Exposed to the frontend.
func (a *App) StopCapture() error {
	logger.Log.Info().Msg("StopCapture called.")
	if a.grpcClient == nil {
		return fmt.Errorf("gRPC client not initialized")
	}

	grpcReq := &router_agent_pb.ControlRequest{
		CommandType: router_agent_pb.ControlCommandType_STOP_CAPTURE,
		// InterfaceName might not be strictly needed for STOP by some agents,
		// but can be included if the agent uses it.
	}

	cmdCtx, cmdCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cmdCancel()
	_, err := a.grpcClient.SendControlCommand(cmdCtx, grpcReq)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error sending STOP_CAPTURE gRPC command")
		return fmt.Errorf("failed to send STOP_CAPTURE command: %w", err)
	}
	logger.Log.Info().Msg("Successfully sent STOP_CAPTURE gRPC command.")

	a.captureStreamMutex.Lock()
	if a.captureStreamCancel != nil {
		logger.Log.Info().Msg("Cancelling capture stream...")
		a.captureStreamCancel()
		a.captureStreamCancel = nil
	} else {
		logger.Log.Info().Msg("No active capture stream to stop.")
	}
	a.captureStreamMutex.Unlock()
	a.isCaptureActive.Store(false)
	runtime.EventsEmit(a.ctx, "capture_status", "stopped")
	return nil
}

// GetAppConfig returns the current application configuration.
// Exposed to the frontend.
func (a *App) GetAppConfig() config.AppConfig {
	return a.appConfig
}

// GetCurrentSnapshot returns the current BSS/STA snapshot.
// Exposed to the frontend.
func (a *App) GetCurrentSnapshot() state_manager.Snapshot {
	if a.stateMgr == nil {
		return state_manager.Snapshot{} // Return empty if not initialized
	}
	return a.stateMgr.GetSnapshot()
}
