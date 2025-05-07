package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"        // Import sync package for Mutex
	"sync/atomic" // Import atomic package
	"syscall"
	"time"

	"wifi-pcap-demo/pc_analyzer/config"
	"wifi-pcap-demo/pc_analyzer/frame_parser"
	"wifi-pcap-demo/pc_analyzer/grpc_client"
	router_agent_pb "wifi-pcap-demo/pc_analyzer/router_agent_pb"
	"wifi-pcap-demo/pc_analyzer/state_manager"
	"wifi-pcap-demo/pc_analyzer/websocket_server"
)

var (
	grpcClient        *grpc_client.CaptureAgentClient
	stateMgr          *state_manager.StateManager
	wsHub             *websocket_server.Hub
	appConfig         config.AppConfig
	packetInfoHandler frame_parser.PacketInfoHandler // Renamed and updated
	pcapStreamHandler grpc_client.PcapStreamHandler  // New handler for pcap stream

	// Global variables to manage the WebSocket-controlled gRPC stream
	wsStreamCancel    context.CancelFunc // Corrected: Removed 'var'
	wsStreamMutex     sync.Mutex         // Corrected: Removed 'var'
	isWsCaptureActive atomic.Bool        // Tracks if WebSocket initiated capture is active
)

// webSocketControlMessageHandler handles messages received from WebSocket clients.
func webSocketControlMessageHandler(messageType int, message []byte) error {
	log.Printf("Received control message from WebSocket: Type=%d, Message=%s", messageType, string(message))

	// Adjusted to handle both "action" and "command" keys, and nested payload.
	// Also matches frontend's data.ts field name "interface" for InterfaceName.
	type CommandPayload struct {
		InterfaceName string `json:"interface,omitempty"`
		Channel       int32  `json:"channel,omitempty"`
		Bandwidth     string `json:"bandwidth,omitempty"`
		BpfFilter     string `json:"bpf_filter,omitempty"`
	}

	type ControlCommandMsg struct {
		Action  string         `json:"action,omitempty"`
		Command string         `json:"command,omitempty"`
		Payload CommandPayload `json:"payload,omitempty"`
	}

	var cmdMsg ControlCommandMsg
	if err := json.Unmarshal(message, &cmdMsg); err != nil {
		log.Printf("Error unmarshalling WebSocket control message: %v. Raw message: %s", err, string(message))
		return fmt.Errorf("invalid control message format: %w", err)
	}

	// DEBUG: Print the parsed command message
	log.Printf("DEBUG: Parsed command message: %+v", cmdMsg)

	actualCommand := cmdMsg.Command
	if actualCommand == "" {
		actualCommand = cmdMsg.Action // Fallback to action if command is empty
	}

	// DEBUG: Print the command to be dispatched
	log.Printf("DEBUG: Command to dispatch: [%s]", actualCommand)

	if actualCommand == "" {
		// This log might be redundant now with the one above, but good for explicit check
		log.Printf("No command or action specified in WebSocket control message: %s", string(message))
		return fmt.Errorf("command or action field is missing or empty")
	}

	if grpcClient == nil {
		log.Println("gRPC client not initialized, cannot send control command.")
		return fmt.Errorf("gRPC client not available")
	}

	var grpcCmdType router_agent_pb.ControlCommandType
	var requiresPayload bool // Flag to indicate if payload is generally expected for the command

	switch strings.ToLower(actualCommand) {
	case "start_capture":
		grpcCmdType = router_agent_pb.ControlCommandType_START_CAPTURE
		requiresPayload = true
	case "stop_capture":
		grpcCmdType = router_agent_pb.ControlCommandType_STOP_CAPTURE
		requiresPayload = false
	case "set_channel":
		log.Printf("Received 'set_channel' command. Payload: %+v. This command is logged but not yet fully mapped to a gRPC action.", cmdMsg.Payload)
		// TODO: Implement gRPC call for set_channel if/when backend supports it
		return nil
	case "set_bandwidth":
		log.Printf("Received 'set_bandwidth' command. Payload: %+v. This command is logged but not yet fully mapped to a gRPC action.", cmdMsg.Payload)
		// TODO: Implement gRPC call for set_bandwidth if/when backend supports it
		return nil
	default:
		log.Printf("Unknown WebSocket control command: '%s' in message: %s", actualCommand, string(message))
		return fmt.Errorf("unknown command: %s", actualCommand)
	}

	if requiresPayload && cmdMsg.Payload.InterfaceName == "" && grpcCmdType == router_agent_pb.ControlCommandType_START_CAPTURE {
		log.Printf("START_CAPTURE command received without 'interface' in payload: %+v", cmdMsg)
		// Depending on requirements, either return error or allow gRPC to handle default/error
		return fmt.Errorf("START_CAPTURE command requires 'interface' in payload")
	}

	grpcReq := &router_agent_pb.ControlRequest{
		CommandType:   grpcCmdType,
		InterfaceName: cmdMsg.Payload.InterfaceName,
		Channel:       cmdMsg.Payload.Channel,
		Bandwidth:     cmdMsg.Payload.Bandwidth,
		BpfFilter:     cmdMsg.Payload.BpfFilter,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := grpcClient.SendControlCommand(ctx, grpcReq)
	if err != nil {
		log.Printf("Error sending gRPC control command for '%s': %v", actualCommand, err)
		return fmt.Errorf("failed to send gRPC command '%s': %w", actualCommand, err)
	}

	// Handle stream cancellation for STOP command
	if grpcCmdType == router_agent_pb.ControlCommandType_STOP_CAPTURE {
		wsStreamMutex.Lock()
		if wsStreamCancel != nil {
			log.Println("Cancelling existing WebSocket-controlled gRPC stream...")
			wsStreamCancel()
			wsStreamCancel = nil // Clear the cancel function
			// isWsCaptureActive.Store(false) // Moved setting inactive state after successful gRPC command
		} else {
			log.Println("No active WebSocket-controlled gRPC stream to stop.")
		}
		wsStreamMutex.Unlock()
		// Set inactive state AFTER successfully sending the STOP command via gRPC
		// (or if there was no stream to cancel)
		isWsCaptureActive.Store(false)
	}

	// Send the gRPC command (START or STOP)
	// Use existing 'err' variable, so use '=' instead of ':='
	var grpcErr error
	_, grpcErr = grpcClient.SendControlCommand(ctx, grpcReq) // Assign to a new variable or use existing 'err' with '='
	if grpcErr != nil {
		log.Printf("Error sending gRPC control command for '%s': %v", actualCommand, grpcErr)
		return fmt.Errorf("failed to send gRPC command '%s': %w", actualCommand, grpcErr)
	}
	log.Printf("Successfully sent gRPC command '%s' for interface '%s'", actualCommand, cmdMsg.Payload.InterfaceName)

	// Handle stream starting for START command
	if grpcCmdType == router_agent_pb.ControlCommandType_START_CAPTURE {
		log.Printf("Attempting to start packet streaming for interface %s via WebSocket command.", cmdMsg.Payload.InterfaceName)

		// Clear existing state before starting a new capture session
		if stateMgr != nil {
			log.Println("Clearing previous BSS/STA state before starting new capture.")
			stateMgr.ClearState()
		}

		wsStreamMutex.Lock()
		// Cancel any previous stream before starting a new one
		if wsStreamCancel != nil {
			log.Println("Cancelling previous WebSocket-controlled gRPC stream before starting new one...")
			wsStreamCancel()
			wsStreamCancel = nil
		}

		// Create new context and cancel function for this stream
		streamCtx, streamCancel := context.WithCancel(context.Background())
		wsStreamCancel = streamCancel // Store the new cancel function
		wsStreamMutex.Unlock()

		// Start the streaming in a new goroutine so this handler doesn't block
		go func() {
			log.Printf("Starting new gRPC packet stream goroutine for interface %s.", cmdMsg.Payload.InterfaceName)
			err := grpcClient.StreamPackets(streamCtx, grpcReq, pcapStreamHandler)
			if err != nil && err != context.Canceled {
				// Log error only if it's not a context cancellation
				log.Printf("Error during WebSocket-controlled packet stream for %s: %v", cmdMsg.Payload.InterfaceName, err)
				// Optionally notify WebSocket clients of the error?
			} else if err == context.Canceled {
				log.Printf("WebSocket-controlled packet stream for %s cancelled successfully.", cmdMsg.Payload.InterfaceName)
			} else {
				log.Printf("WebSocket-controlled packet stream for %s finished without error.", cmdMsg.Payload.InterfaceName)
			}

			// The cancel function (wsStreamCancel) is managed by START_CAPTURE and STOP_CAPTURE commands.
			// This goroutine should not clear wsStreamCancel itself upon natural completion,
			// as a new stream might have already been started and wsStreamCancel updated.
			// The streamCtx (local to this goroutine) will be cancelled by a new START or a STOP.
		}()

		log.Printf("Packet streaming goroutine initiated for interface %s.", cmdMsg.Payload.InterfaceName)
		// Set capture state to active only AFTER successfully sending the START command via gRPC
		// and initiating the goroutine. The SendControlCommand was already checked for errors.
		isWsCaptureActive.Store(true)
	}

	return nil
}

func main() {
	log.Println("Starting PC-Side Real-time Analysis Engine...")

	configFile := flag.String("config", "config/config.json", "Path to configuration file")
	defaultCaptureInterface := flag.String("iface", "ath1", "Default wireless interface for capture (if auto-start)")
	autoStartCapture := flag.Bool("autostart", false, "Automatically start capture on the default interface on startup")
	defaultChannel := flag.Int("channel", 1, "Default channel for auto-start capture")
	defaultBandwidth := flag.String("bw", "HT20", "Default bandwidth for auto-start capture (e.g., HT20, HT40, VHT80)")

	flag.Parse()

	var err error
	appConfig = config.LoadConfig(*configFile)
	log.Printf("Configuration loaded: %+v", appConfig)

	stateMgr = state_manager.NewStateManager()
	log.Println("State Manager initialized.")

	// Initialize packetInfoHandler
	packetInfoHandler = func(parsedInfo *frame_parser.ParsedFrameInfo) {
		if parsedInfo != nil {
			stateMgr.ProcessParsedFrame(parsedInfo)
			// Note: Broadcasting snapshot is handled by a ticker now,
			// so no direct wsHub.BroadcastSnapshot() here per packet.
		}
	}

	// Initialize pcapStreamHandler
	pcapStreamHandler = func(pcapStream io.Reader) {
		log.Println("Global pcapStreamHandler invoked, starting ProcessPcapStream.")
		frame_parser.ProcessPcapStream(pcapStream, packetInfoHandler)
	}

	wsHub = websocket_server.NewHub(stateMgr.GetSnapshot, webSocketControlMessageHandler)
	go wsHub.Run()
	go websocket_server.StartServer(wsHub, appConfig.WebSocketAddress)
	log.Printf("WebSocket Hub initialized and server starting on %s.", appConfig.WebSocketAddress)

	grpcClient, err = grpc_client.Connect(appConfig.GRPCServerAddress)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server at %s: %v", appConfig.GRPCServerAddress, err)
	}
	defer grpcClient.Close()
	log.Printf("gRPC client connected to %s.", appConfig.GRPCServerAddress)

	// Context for managing the lifetime of gRPC streams
	// This main context can be cancelled on shutdown to stop all streams.
	mainStreamCtx, mainStreamCancel := context.WithCancel(context.Background())
	defer mainStreamCancel() // Ensure all streams are signalled to stop on main exit

	if *autoStartCapture {
		log.Printf("Auto-starting capture on interface %s, channel %d, bandwidth %s", *defaultCaptureInterface, *defaultChannel, *defaultBandwidth)
		startReq := &router_agent_pb.ControlRequest{
			CommandType:   router_agent_pb.ControlCommandType_START_CAPTURE,
			InterfaceName: *defaultCaptureInterface,
			Channel:       int32(*defaultChannel),
			Bandwidth:     *defaultBandwidth,
		}
		_, err := grpcClient.SendControlCommand(mainStreamCtx, startReq)
		if err != nil {
			log.Printf("Error sending initial START_CAPTURE command: %v", err)
		} else {
			err = grpcClient.StreamPackets(mainStreamCtx, startReq, pcapStreamHandler)
			if err != nil {
				log.Fatalf("Failed to start initial packet stream for %s: %v", *defaultCaptureInterface, err)
			}
			log.Printf("Initial packet streaming started for %s.", *defaultCaptureInterface)
		}
	}

	// Goroutine to periodically send state snapshot to WebSocket clients
	snapshotTicker := time.NewTicker(2 * time.Second) // Send updates every 2 seconds
	defer snapshotTicker.Stop()
	go func() {
		for range snapshotTicker.C {
			// Only broadcast if WebSocket-initiated capture is active
			if isWsCaptureActive.Load() {
				wsHub.BroadcastSnapshot()
			}
		}
	}()

	// Goroutine to periodically prune old entries from state manager
	pruneTicker := time.NewTicker(30 * time.Second) // Prune every 30 seconds
	defer pruneTicker.Stop()
	go func() {
		for range pruneTicker.C {
			stateMgr.PruneOldEntries(2 * time.Minute) // Timeout of 2 minutes
			log.Println("Pruned old entries from state manager.")
		}
	}()

	log.Println("PC-Side Analysis Engine is running. Press Ctrl+C to exit.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down PC-Side Analysis Engine...")
	mainStreamCancel() // Signal gRPC streams to stop
	// Add any other cleanup, e.g., closing WebSocket hub if it has a dedicated close method.
	log.Println("Shutdown complete.")
}
