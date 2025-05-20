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

	"github.com/google/gopacket"        // For NewPacketSource
	"github.com/google/gopacket/pcapgo" // For NewReader to handle io.Reader stream
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
	isConnected         atomic.Bool
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
		logger.Log.Info().Msg("Wails pcapStreamHandler invoked, attempting to process with gopacket/pcapgo.")

		// Use pcapgo.NewReader to handle the io.Reader stream.
		// pcapgo.Reader implements gopacket.PacketDataSource.
		r, pcapErr := pcapgo.NewReader(pcapStream)
		if pcapErr != nil {
			logger.Log.Error().Err(pcapErr).Msg("Error creating pcapgo.Reader from stream")
			runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Error creating pcapgo.Reader: %v", pcapErr))
			return
		}

		// Create a PacketSource from the pcapgo.Reader.
		// The LinkType comes from the pcapgo.Reader itself after parsing the header.
		packetSource := gopacket.NewPacketSource(r, r.LinkType())

		// Call the modified ProcessPcapStream which now expects a *gopacket.PacketSource
		// The second argument (formerly tsharkPath) is now unused by ProcessPcapStream.
		processErr := frame_parser.ProcessPcapStream(packetSource, "", a.packetInfoHandler)
		if processErr != nil {
			logger.Log.Error().Err(processErr).Msg("Error processing pcap stream with gopacket")
			runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Error processing pcap stream: %v", processErr))
		}
	}

	// 不再自动连接gRPC服务器，而是通过前端调用ConnectToAgent函数连接
	// 设置连接状态为未连接
	a.isConnected.Store(false)

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

// ConnectToAgent connects to the gRPC server with the specified address.
// Exposed to the frontend.
func (a *App) ConnectToAgent(serverAddr string) error {
	logger.Log.Info().Str("address", serverAddr).Msg("Connecting to gRPC server...")

	// 如果已经连接，先关闭之前的连接
	if a.grpcClient != nil {
		a.grpcClient.Close()
		a.grpcClient = nil
		a.isConnected.Store(false)
	}

	// 连接到新的gRPC服务器
	var err error
	a.grpcClient, err = grpc_client.Connect(serverAddr)
	if err != nil {
		logger.Log.Error().Err(err).Str("address", serverAddr).Msg("Failed to connect to gRPC server")
		runtime.EventsEmit(a.ctx, "connection_status", "failed")
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// 更新连接状态
	a.isConnected.Store(true)
	logger.Log.Info().Str("address", serverAddr).Msg("gRPC client connected successfully.")
	runtime.EventsEmit(a.ctx, "connection_status", "connected")
	return nil
}

// IsConnected returns the current connection status.
// Exposed to the frontend.
func (a *App) IsConnected() bool {
	return a.isConnected.Load()
}

// DisconnectFromAgent disconnects from the gRPC server.
// Exposed to the frontend.
func (a *App) DisconnectFromAgent() error {
	logger.Log.Info().Msg("Disconnecting from gRPC server...")

	if a.isCaptureActive.Load() {
		logger.Log.Warn().Msg("Cannot disconnect while capture is active. Please stop capture first.")
		return fmt.Errorf("抓包过程中无法断开连接，请先停止抓包")
	}

	if a.grpcClient != nil {
		a.grpcClient.Close()
		a.grpcClient = nil
		logger.Log.Info().Msg("gRPC client closed.")
	}
	a.isConnected.Store(false)
	runtime.EventsEmit(a.ctx, "connection_status", "disconnected")
	logger.Log.Info().Msg("Successfully disconnected from gRPC server.")
	return nil
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

func (a *App) SelectPcapFileAndProcess() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Pcap File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Pcap Files (*.pcap, *.cap)", Pattern: "*.pcap;*.cap"},
		},
	})
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error selecting pcap file")
		return "", fmt.Errorf("file selection error: %w", err)
	}
	if filePath == "" {
		logger.Log.Info().Msg("No pcap file selected.")
		return "", nil // No file selected is not an error
	}

	logger.Log.Info().Str("filePath", filePath).Msg("Pcap file selected for processing.")

	// Clear existing state before processing a new file
	if a.stateMgr != nil {
		logger.Log.Info().Msg("Clearing previous BSS/STA state before processing new file.")
		a.stateMgr.ClearState()
	}
	a.isCaptureActive.Store(true) // Treat file processing like an active capture for UI
	runtime.EventsEmit(a.ctx, "capture_status", "processing_file")

	// Process the pcap file
	// Goroutine to prevent blocking the UI thread.
	go func() {
		defer func() {
			a.isCaptureActive.Store(false)
			runtime.EventsEmit(a.ctx, "capture_status", "file_processed")
			logger.Log.Info().Str("filePath", filePath).Msg("Finished processing pcap file.")
		}()
		// The second argument to ProcessPcapFile was a.appConfig.TsharkPath, now unused.
		err := frame_parser.ProcessPcapFile(filePath, "", a.packetInfoHandler)
		if err != nil {
			logger.Log.Error().Err(err).Str("filePath", filePath).Msg("Error processing pcap file")
			runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Error processing pcap file '%s': %v", filePath, err))
		}
	}()

	return fmt.Sprintf("Processing pcap file: %s", filePath), nil
}
