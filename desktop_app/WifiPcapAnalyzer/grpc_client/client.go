package grpc_client

import (
	"WifiPcapAnalyzer/logger"
	"context"
	"io"
	"time"

	// "log" // Removed as it's no longer used

	router_agent_pb "WifiPcapAnalyzer/router_agent_pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// PcapStreamHandler is a function type that processes a stream of pcap data.
type PcapStreamHandler func(pcapStream io.Reader)

// CaptureAgentClient wraps the gRPC client
type CaptureAgentClient struct {
	client router_agent_pb.CaptureAgentClient
	conn   *grpc.ClientConn
}

// Connect establishes a connection to the gRPC server.
func Connect(serverAddr string) (*CaptureAgentClient, error) {
	logger.Log.Debug().Msgf("Attempting to connect to gRPC server at %s", serverAddr)
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithBlock(), // grpc.WithBlock() is deprecated, use a timeout context with DialContext instead if needed.
		// For simplicity, using Dial which is non-blocking by default. Connection state can be checked.
	)
	if err != nil {
		logger.Log.Error().Err(err).Msgf("Failed to dial gRPC server")
		return nil, err
	}
	// To ensure connection is up, you might want to use grpc.DialContext with a timeout
	// or perform a quick RPC call. For now, we assume connection will establish or fail on first RPC.
	logger.Log.Debug().Msgf("gRPC dial initiated to %s", serverAddr)

	client := router_agent_pb.NewCaptureAgentClient(conn)
	return &CaptureAgentClient{client: client, conn: conn}, nil
}

// Close closes the gRPC connection.
func (c *CaptureAgentClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		// log.Println("gRPC connection closed.")
	}
}

// SendControlCommand sends a control command to the router agent.
func (c *CaptureAgentClient) SendControlCommand(ctx context.Context, req *router_agent_pb.ControlRequest) (*router_agent_pb.ControlResponse, error) {
	logger.Log.Debug().Msgf("Sending control command: Type=%s, Interface=%s, Channel=%d, Bandwidth=%s", req.CommandType, req.InterfaceName, req.Channel, req.Bandwidth)

	// Use a timeout for the RPC call itself, separate from the main context if needed.
	callCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10-second timeout for this specific command
	defer cancel()

	res, err := c.client.SendControlCommand(callCtx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			logger.Log.Error().Msgf("Error sending control command: Code=%s, Message=%s", st.Code(), st.Message())
		} else {
			logger.Log.Error().Err(err).Msgf("Error sending control command (non-gRPC error)")
		}
		return nil, err
	}
	logger.Log.Debug().Msgf("Control command response: Success=%v, Message=%s", res.Success, res.Message)
	return res, nil
}

// StreamPackets starts streaming packets from the router agent and processes them using the packetHandler.
// This function will run the stream processing in a new goroutine.
// The provided context (ctx) is used for the lifetime of the stream.
func (c *CaptureAgentClient) StreamPackets(ctx context.Context, req *router_agent_pb.ControlRequest, pcapHandler PcapStreamHandler) error {
	logger.Log.Info().Msgf("Requesting to stream packets for interface: %s, Channel: %d, Bandwidth: %s", req.InterfaceName, req.Channel, req.Bandwidth)

	stream, err := c.client.StreamPackets(ctx, req) // Use the passed-in context for the stream
	if err != nil {
		logger.Log.Error().Err(err).Msgf("Error starting packet stream")
		return err
	}
	// log.Println("Packet stream started successfully. Waiting for packets...")

	pipeReader, pipeWriter := io.Pipe()

	go func() {
		defer pipeWriter.Close() // Ensure writer is closed when goroutine exits
		for {
			// Check if the context has been cancelled (e.g., by the caller)
			select {
			case <-ctx.Done():
				logger.Log.Info().Err(ctx.Err()).Msgf("Context cancelled, stopping packet stream for interface %s.", req.InterfaceName)
				return // Exit goroutine
			default:
				// Proceed with Recv()
			}

			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					logger.Log.Info().Msgf("Packet stream ended by server (EOF) for interface %s.", req.InterfaceName)
				} else if status.Code(err) == codes.Canceled {
					logger.Log.Info().Err(err).Msgf("Packet stream cancelled (client-side or server-side context cancellation) for interface %s.", req.InterfaceName)
				} else {
					logger.Log.Error().Err(err).Msgf("Error receiving packet from stream for interface %s", req.InterfaceName)
				}
				return // Exit goroutine on any error or EOF
			}
			if msg != nil {
				frameBytes := msg.GetFrame() // Use the getter method
				if len(frameBytes) > 0 {
					if _, err := pipeWriter.Write(frameBytes); err != nil {
						logger.Log.Error().Err(err).Msgf("Error writing to pipe for interface %s", req.InterfaceName)
						return // Exit goroutine
					}
				}
			}
		}
	}()

	// Start a new goroutine to handle the pcap stream processing
	go pcapHandler(pipeReader)

	return nil // Return immediately as the stream processing is in a goroutine
}
