package websocket_server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time" // Added for potential future use (e.g. ping/pong timeouts)

	"wifi-pcap-demo/pc_analyzer/state_manager" // For BSSInfo, STAInfo

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte // Channel for broadcasting messages to clients
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex

	// Callback to get data snapshot from StateManager
	GetDataSnapshot func() ([]*state_manager.BSSInfo, []*state_manager.STAInfo)

	// Callback to handle control messages from WebSocket clients
	HandleControlMessage func(messageType int, message []byte) error
}

// Client is a middleman between the WebSocket connection and the hub.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte // Buffered channel of outbound messages.
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections for development
	},
}

// NewHub creates a new Hub.
func NewHub(snapshotGetter func() ([]*state_manager.BSSInfo, []*state_manager.STAInfo), controlHandler func(messageType int, message []byte) error) *Hub {
	return &Hub{
		broadcast:            make(chan []byte),
		register:             make(chan *Client),
		unregister:           make(chan *Client),
		clients:              make(map[*Client]bool),
		GetDataSnapshot:      snapshotGetter,
		HandleControlMessage: controlHandler,
	}
}

// Run starts the hub's event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Println("Client registered to WebSocket hub")
			// Optionally send initial snapshot upon registration
			if h.GetDataSnapshot != nil {
				bssInfos, staInfos := h.GetDataSnapshot()
				snapshot := struct {
					BSSs []*state_manager.BSSInfo `json:"bsss"`
					STAs []*state_manager.STAInfo `json:"stas"`
				}{BSSs: bssInfos, STAs: staInfos}
				jsonData, err := json.Marshal(snapshot)
				if err == nil {
					client.send <- jsonData
				} else {
					log.Printf("Error marshalling initial snapshot for client: %v", err)
				}
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("Client unregistered from WebSocket hub")
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default: // Don't block if client's send buffer is full
					log.Printf("Client send buffer full or closed, unregistering client: %v", client.conn.RemoteAddr())
					// Ensure unregister is handled properly without deadlocking
					// This direct delete and close might be problematic if Run loop is also trying to modify h.clients
					// It's better to send to unregister channel if possible, but that might deadlock here.
					// For simplicity now, direct removal, but review for concurrent safety.
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastSnapshot fetches the current state and broadcasts it as JSON.
func (h *Hub) BroadcastSnapshot() {
	if h.GetDataSnapshot == nil {
		log.Println("GetDataSnapshot callback is not set in Hub")
		return
	}
	bssInfos, staInfos := h.GetDataSnapshot()
	snapshot := struct {
		Type string `json:"type"`
		Data struct {
			BSSs []*state_manager.BSSInfo `json:"bsss"`
			STAs []*state_manager.STAInfo `json:"stas"`
		} `json:"data"`
	}{
		Type: "snapshot",
		Data: struct {
			BSSs []*state_manager.BSSInfo `json:"bsss"`
			STAs []*state_manager.STAInfo `json:"stas"`
		}{BSSs: bssInfos, STAs: staInfos},
	}

	jsonData, err := json.Marshal(snapshot)
	if err != nil {
		log.Printf("Error marshalling data for broadcast: %v", err)
		return
	}

	// DEBUG_WEBSOCKET_SEND log
	log.Printf("DEBUG_WEBSOCKET_SEND: Sending data: %s", string(jsonData))

	h.broadcast <- jsonData
}

// readPump pumps messages from the WebSocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket client read error: %v", err)
			}
			break
		}
		// Process message (e.g., control command from client)
		if c.hub.HandleControlMessage != nil {
			if err := c.hub.HandleControlMessage(messageType, message); err != nil {
				log.Printf("Error handling control message from client %v: %v", c.conn.RemoteAddr(), err)
				// Optionally send an error message back to the client
			}
		} else {
			log.Printf("Received message from client %v: Type=%d, Message=%s (No handler set)", c.conn.RemoteAddr(), messageType, message)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error getting next writer for client %v: %v", c.conn.RemoteAddr(), err)
				return
			}
			w.Write(message)

			// Add queued messages. Not strictly necessary if messages are self-contained JSON.
			// n := len(c.send)
			// for i := 0; i < n; i++ {
			// 	w.Write(<-c.send) // Consider message boundaries
			// }

			if err := w.Close(); err != nil {
				log.Printf("Error closing writer for client %v: %v", c.conn.RemoteAddr(), err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping to client %v: %v", c.conn.RemoteAddr(), err)
				return
			}
		}
	}
}

// ServeWs handles WebSocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()

	log.Printf("WebSocket client connected: %v", conn.RemoteAddr())
}

// StartServer starts the WebSocket server.
// It takes the hub and the server address (e.g., "0.0.0.0:8080") as arguments.
func StartServer(hub *Hub, addr string) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	})
	log.Printf("WebSocket server starting on %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("WebSocket ListenAndServe failed: %v", err)
	}
}
