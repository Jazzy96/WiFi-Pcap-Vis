package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"WifiPcapAnalyzer/config"
	"WifiPcapAnalyzer/frame_parser"
	"WifiPcapAnalyzer/grpc_client"
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
	log.Println("Wails App starting up...")

	// Load configuration
	// Assuming config.json is in a 'config' directory relative to the executable
	// or the Wails project root when developing.
	a.appConfig = config.LoadConfig("config/config.json")
	log.Printf("Configuration loaded: %+v", a.appConfig)
	config.GlobalConfig = a.appConfig // Make config globally accessible if needed by sub-packages

	// Initialize State Manager
	a.stateMgr = state_manager.NewStateManager()
	log.Println("State Manager initialized.")

	// Initialize Packet Info Handler
	a.packetInfoHandler = func(parsedInfo *frame_parser.ParsedFrameInfo) {
		if parsedInfo != nil {
			a.stateMgr.ProcessParsedFrame(parsedInfo)
			// Snapshot broadcasting will be handled by a ticker using Wails events
		}
	}

	// Initialize PCAP Stream Handler
	a.pcapStreamHandler = func(pcapStream io.Reader) {
		log.Println("Wails pcapStreamHandler invoked, starting ProcessPcapStream.")
		frame_parser.ProcessPcapStream(pcapStream, a.packetInfoHandler)
	}

	// Initialize gRPC Client
	var err error
	a.grpcClient, err = grpc_client.Connect(a.appConfig.GRPCServerAddress)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server at %s: %v", a.appConfig.GRPCServerAddress, err)
		// In a real app, might want to show an error to the user via Wails dialog
		runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Failed to connect to gRPC server: %v", err))
		return
	}
	log.Printf("gRPC client connected to %s.", a.appConfig.GRPCServerAddress)

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
				log.Println("Snapshot ticker stopping due to app context done.")
				return
			}
		}
	}()
	log.Println("Snapshot emission goroutine started.")

	// Goroutine to periodically prune old entries from state manager
	pruneTicker := time.NewTicker(30 * time.Second) // Prune every 30 seconds
	go func() {
		defer pruneTicker.Stop()
		for {
			select {
			case <-pruneTicker.C:
				a.stateMgr.PruneOldEntries(2 * time.Minute) // Timeout of 2 minutes
				log.Println("Pruned old entries from state manager.")
			case <-a.ctx.Done(): // App is shutting down
				log.Println("Pruning ticker stopping due to app context done.")
				return
			}
		}
	}()
	log.Println("Pruning goroutine started.")
	log.Println("Wails App startup complete.")
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(ctx context.Context) {
	log.Println("Wails App shutting down...")
	if a.grpcClient != nil {
		a.grpcClient.Close()
		log.Println("gRPC client closed.")
	}
	a.captureStreamMutex.Lock()
	if a.captureStreamCancel != nil {
		a.captureStreamCancel()
		log.Println("Capture stream cancelled.")
	}
	a.captureStreamMutex.Unlock()
	log.Println("Wails App shutdown complete.")
}

// StartCapture initiates packet capture via gRPC.
// Exposed to the frontend.
func (a *App) StartCapture(interfaceName string, channel int32, bandwidth string, bpfFilter string) error {
	log.Printf("StartCapture called: Interface=%s, Channel=%d, Bandwidth=%s, Filter=%s", interfaceName, channel, bandwidth, bpfFilter)
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
		log.Printf("Error sending START_CAPTURE gRPC command: %v", err)
		return fmt.Errorf("failed to send START_CAPTURE command: %w", err)
	}
	log.Println("Successfully sent START_CAPTURE gRPC command.")

	// Clear existing state before starting a new capture session
	if a.stateMgr != nil {
		log.Println("Clearing previous BSS/STA state before starting new capture.")
		a.stateMgr.ClearState()
	}

	a.captureStreamMutex.Lock()
	// Cancel any previous stream before starting a new one
	if a.captureStreamCancel != nil {
		log.Println("Cancelling previous capture stream before starting new one...")
		a.captureStreamCancel()
		a.captureStreamCancel = nil
	}

	// Create new context and cancel function for this stream
	streamCtx, streamCancel := context.WithCancel(context.Background())
	a.captureStreamCancel = streamCancel
	a.captureStreamMutex.Unlock()

	// Start the streaming in a new goroutine
	go func() {
		log.Printf("Starting new gRPC packet stream goroutine for interface %s.", interfaceName)
		err := a.grpcClient.StreamPackets(streamCtx, grpcReq, a.pcapStreamHandler)
		if err != nil && err != context.Canceled {
			log.Printf("Error during packet stream for %s: %v", interfaceName, err)
			runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Packet stream error: %v", err))
		} else if err == context.Canceled {
			log.Printf("Packet stream for %s cancelled successfully.", interfaceName)
		} else {
			log.Printf("Packet stream for %s finished without error.", interfaceName)
		}
		// Ensure capture active is false if stream ends not by explicit stop
		// This might need more robust handling if a new stream is started before this one ends.
		// For now, StopCapture is the primary way to set it false.
	}()

	log.Printf("Packet streaming goroutine initiated for interface %s.", interfaceName)
	a.isCaptureActive.Store(true)
	runtime.EventsEmit(a.ctx, "capture_status", "started")
	return nil
}

// StopCapture stops the packet capture via gRPC.
// Exposed to the frontend.
func (a *App) StopCapture() error {
	log.Println("StopCapture called.")
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
		log.Printf("Error sending STOP_CAPTURE gRPC command: %v", err)
		return fmt.Errorf("failed to send STOP_CAPTURE command: %w", err)
	}
	log.Println("Successfully sent STOP_CAPTURE gRPC command.")

	a.captureStreamMutex.Lock()
	if a.captureStreamCancel != nil {
		log.Println("Cancelling capture stream...")
		a.captureStreamCancel()
		a.captureStreamCancel = nil
	} else {
		log.Println("No active capture stream to stop.")
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
