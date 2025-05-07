package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

const (
	defaultPort = ":50051"
	bufferSize  = 65536 // 64KB buffer for reading tcpdump output
)

// server is used to implement CaptureAgentServer.
type server struct {
	UnimplementedCaptureAgentServer
	mu               sync.Mutex
	tcpdumpCmd       *exec.Cmd
	tcpdumpPipe      io.ReadCloser
	isCapturing      bool
	currentInterface string
	currentChannel   int32
	currentBandwidth string
	currentBpfFilter string
}

// SendControlCommand implements CaptureAgentServer
func (s *server) SendControlCommand(ctx context.Context, req *ControlRequest) (*ControlResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Received control command: %v for interface %s", req.CommandType, req.InterfaceName)

	switch req.CommandType {
	case ControlCommandType_START_CAPTURE:
		if s.isCapturing {
			log.Printf("Capture already in progress on %s", s.currentInterface)
			return &ControlResponse{Success: false, Message: "Capture already in progress"}, nil
		}
		if req.InterfaceName == "" {
			return &ControlResponse{Success: false, Message: "Interface name cannot be empty for START_CAPTURE"}, nil
		}
		s.currentInterface = req.InterfaceName
		s.currentBpfFilter = req.BpfFilter
		// Note: Channel and Bandwidth setting via iw command is not implemented here yet,
		// as the primary focus is tcpdump. Assumes interface is pre-configured.
		// If needed, add logic here:
		// err := s.setInterfaceParams(req.InterfaceName, req.Channel, req.Bandwidth)
		// if err != nil { ... }

		log.Printf("Starting capture on interface %s with filter '%s'", s.currentInterface, s.currentBpfFilter)
		// tcpdump command: -i <interface> -U (buffer per packet) -w - (write to stdout)
		// -s 0 can be added to capture full packets if needed, but default is usually sufficient.
		args := []string{"-i", s.currentInterface, "-U", "-w", "-"}
		if s.currentBpfFilter != "" {
			// Split filter string by space, as exec.Command expects separate arguments
			filterParts := strings.Fields(s.currentBpfFilter)
			args = append(args, filterParts...)
		}
		s.tcpdumpCmd = exec.Command("tcpdump", args...)

		var err error
		s.tcpdumpPipe, err = s.tcpdumpCmd.StdoutPipe()
		if err != nil {
			log.Printf("Error creating StdoutPipe for tcpdump: %v", err)
			return &ControlResponse{Success: false, Message: "Failed to create tcpdump pipe"}, err
		}

		if err := s.tcpdumpCmd.Start(); err != nil {
			log.Printf("Error starting tcpdump: %v", err)
			s.tcpdumpPipe.Close() // Clean up pipe
			s.tcpdumpPipe = nil
			s.tcpdumpCmd = nil
			return &ControlResponse{Success: false, Message: "Failed to start tcpdump"}, err
		}
		s.isCapturing = true
		log.Printf("tcpdump process started (PID: %d) on interface %s", s.tcpdumpCmd.Process.Pid, s.currentInterface)
		return &ControlResponse{Success: true, Message: "Capture started successfully"}, nil

	case ControlCommandType_STOP_CAPTURE:
		if !s.isCapturing || s.tcpdumpCmd == nil || s.tcpdumpCmd.Process == nil {
			log.Println("No capture in progress to stop.")
			return &ControlResponse{Success: false, Message: "No capture in progress"}, nil
		}
		log.Printf("Stopping capture on interface %s (PID: %d)", s.currentInterface, s.tcpdumpCmd.Process.Pid)

		// Send SIGINT to tcpdump for graceful shutdown
		if err := s.tcpdumpCmd.Process.Signal(syscall.SIGINT); err != nil {
			log.Printf("Error sending SIGINT to tcpdump: %v. Attempting SIGKILL.", err)
			// If SIGINT fails, try SIGKILL
			if killErr := s.tcpdumpCmd.Process.Kill(); killErr != nil {
				log.Printf("Error sending SIGKILL to tcpdump: %v", killErr)
				return &ControlResponse{Success: false, Message: "Failed to stop tcpdump (SIGKILL failed)"}, killErr
			}
		}

		// Wait for the process to exit
		go func(cmd *exec.Cmd, pipe io.ReadCloser) {
			cmd.Wait() // Wait for the command to finish
			if pipe != nil {
				pipe.Close()
			}
			log.Printf("tcpdump process (PID: %d) stopped.", cmd.ProcessState.Pid())
		}(s.tcpdumpCmd, s.tcpdumpPipe)

		s.isCapturing = false
		s.tcpdumpCmd = nil
		s.tcpdumpPipe = nil // Nullify the pipe
		log.Println("Capture stopped successfully.")
		return &ControlResponse{Success: true, Message: "Capture stopped successfully"}, nil

	case ControlCommandType_SET_CHANNEL:
		// Placeholder for future implementation
		// s.currentChannel = req.Channel
		// err := s.setInterfaceParams(req.InterfaceName, req.Channel, s.currentBandwidth)
		// if err != nil { ... }
		log.Printf("SET_CHANNEL command received for channel %d (not fully implemented)", req.Channel)
		return &ControlResponse{Success: true, Message: "SET_CHANNEL command received (not fully implemented)"}, nil

	case ControlCommandType_SET_BANDWIDTH:
		// Placeholder for future implementation
		// s.currentBandwidth = req.Bandwidth
		// err := s.setInterfaceParams(req.InterfaceName, s.currentChannel, req.Bandwidth)
		// if err != nil { ... }
		log.Printf("SET_BANDWIDTH command received for bandwidth %s (not fully implemented)", req.Bandwidth)
		return &ControlResponse{Success: true, Message: "SET_BANDWIDTH command received (not fully implemented)"}, nil

	default:
		log.Printf("Unknown command type: %v", req.CommandType)
		return &ControlResponse{Success: false, Message: "Unknown command type"}, nil
	}
}

// StreamPackets implements CaptureAgentServer
func (s *server) StreamPackets(req *ControlRequest, stream CaptureAgent_StreamPacketsServer) error {
	log.Printf("StreamPackets called. Waiting for capture to be active.")
	// This stream will only send data if capture is active.
	// The client calls this, and then separately calls SendControlCommand to start.
	// Alternatively, this call itself could trigger a capture if req contains START_CAPTURE,
	// but current design separates control (SendControlCommand) and data (StreamPackets).

	// Continuously check if capture is active and pipe is available
	var localPipe io.ReadCloser
	for {
		s.mu.Lock()
		if s.isCapturing && s.tcpdumpPipe != nil {
			localPipe = s.tcpdumpPipe
			s.mu.Unlock()
			break
		}
		s.mu.Unlock()
		// Wait a bit before checking again to avoid busy-looping if capture isn't started immediately.
		time.Sleep(100 * time.Millisecond)

		// Check if context is cancelled (client disconnected)
		if stream.Context().Err() != nil {
			log.Printf("Client disconnected while waiting for capture to start: %v", stream.Context().Err())
			return stream.Context().Err()
		}
	}

	log.Printf("Capture active. Streaming packets from interface %s", s.currentInterface)
	reader := bufio.NewReaderSize(localPipe, bufferSize)
	buffer := make([]byte, bufferSize)

	for {
		// Check if context is cancelled (client disconnected or STOP_CAPTURE called)
		select {
		case <-stream.Context().Done():
			log.Printf("Stream context done (client disconnected or stream cancelled): %v", stream.Context().Err())
			return stream.Context().Err()
		default:
			// Non-blocking check for capture status
			s.mu.Lock()
			isStillCapturing := s.isCapturing
			s.mu.Unlock()
			if !isStillCapturing {
				log.Println("Capture stopped, ending packet stream.")
				return nil // Gracefully end stream
			}

			n, err := reader.Read(buffer)
			if err != nil {
				if err == io.EOF {
					log.Println("tcpdump pipe EOF reached.")
					// This might happen if tcpdump exits unexpectedly or finishes.
					// Or if STOP_CAPTURE was called and pipe was closed.
					s.mu.Lock()
					if s.isCapturing { // If still marked as capturing, it's an unexpected EOF
						log.Println("EOF reached but still marked as capturing. This might be an error or tcpdump exited.")
						s.isCapturing = false // Mark as not capturing
						s.tcpdumpCmd = nil
						s.tcpdumpPipe = nil
					}
					s.mu.Unlock()
					return nil // End stream
				}
				log.Printf("Error reading from tcpdump pipe: %v", err)
				// Decide if this error is fatal for the stream
				return err
			}

			if n > 0 {
				if err := stream.Send(&CaptureData{Frame: buffer[:n]}); err != nil {
					log.Printf("Error sending packet data to client: %v", err)
					return err
				}
			}
		}
	}
}

// setInterfaceParams (Optional, for future use with SET_CHANNEL/SET_BANDWIDTH)
// This function would use 'iw' command to set channel and bandwidth.
// Example: iw dev ath1 set channel <channel> [HT20|HT40|VHT20|VHT40|VHT80|VHT160]
func (s *server) setInterfaceParams(iface string, channel int32, bandwidth string) error {
	if channel <= 0 && bandwidth == "" {
		return nil // Nothing to set
	}
	// Ensure interface is up first (optional, depends on system state)
	// exec.Command("ip", "link", "set", iface, "up").Run()

	if channel > 0 {
		var cmdArgs []string
		if bandwidth != "" {
			// Note: 'iw' syntax for bandwidth might vary (e.g. 'HT20', 'VHT80', '80MHz')
			// This needs to be aligned with what the specific 'iw' version on the router expects.
			cmdArgs = []string{"dev", iface, "set", "channel", fmt.Sprintf("%d", channel), bandwidth}
		} else {
			cmdArgs = []string{"dev", iface, "set", "channel", fmt.Sprintf("%d", channel)}
		}
		log.Printf("Executing: iw %s", strings.Join(cmdArgs, " "))
		cmd := exec.Command("iw", cmdArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error setting channel/bandwidth for %s to ch %d %s: %v. Output: %s", iface, channel, bandwidth, err, string(output))
			return fmt.Errorf("failed to set channel/bandwidth: %s, %v", string(output), err)
		}
		log.Printf("Successfully set channel/bandwidth for %s. Output: %s", iface, string(output))
		s.currentChannel = channel
		s.currentBandwidth = bandwidth
	}
	return nil
}

func main() {
	port := os.Getenv("CAPTURE_AGENT_PORT")
	if port == "" {
		port = defaultPort
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("gRPC server listening on %s", port)

	s_grpc := grpc.NewServer() // Renamed to s_grpc to avoid conflict if 's' is used above
	RegisterCaptureAgentServer(s_grpc, &server{})

	if err := s_grpc.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
