package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/PRM710/Rankedterview-backend/internal/database"
)

// Configuration for scalability
const (
	// Buffer sizes
	broadcastBufferSize = 1024
	clientBufferSize    = 256

	// Cache TTL
	roomCacheTTL = 30 * time.Second
)

// roomCache caches room participants to reduce Redis calls
type roomCache struct {
	participants map[string]string
	cachedAt     time.Time
}

// Hub maintains active WebSocket connections
type Hub struct {
	// Registered clients by userID
	clients map[string]*Client

	// Room membership (roomID -> set of userIDs) - local cache
	rooms map[string]*roomCache

	// User to room mapping for quick lookup
	userRooms map[string]string

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients
	broadcast chan *Message

	// Mutex for thread-safe access to clients
	clientsMu sync.RWMutex

	// Mutex for room cache
	roomsMu sync.RWMutex

	// Redis for persistence and pub/sub across instances
	redis *database.RedisClient

	// Shutdown channel
	shutdown chan struct{}
}

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"userId,omitempty"`
	RoomID    string                 `json:"roomId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Broadcast bool                   `json:"-"`
	Exclude   string                 `json:"-"` // UserID to exclude from broadcast
}

// NewHub creates a new Hub
func NewHub(redis *database.RedisClient) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]*roomCache),
		userRooms:  make(map[string]string),
		register:   make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		broadcast:  make(chan *Message, broadcastBufferSize),
		redis:      redis,
		shutdown:   make(chan struct{}),
	}
}

// Run starts the hub - now with multiple workers for scalability
func (h *Hub) Run() {
	// Start multiple broadcast workers
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		go h.broadcastWorker(i)
	}

	// Main loop for register/unregister
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case <-h.shutdown:
			return
		}
	}
}

// broadcastWorker handles broadcast messages
func (h *Hub) broadcastWorker(workerID int) {
	for {
		select {
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		case <-h.shutdown:
			return
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.clientsMu.Lock()

	// Check if user already has a connection
	if existingClient, ok := h.clients[client.UserID]; ok {
		// Just log it - DON'T close the old connection
		// The old connection will naturally close when ping/pong times out
		// Closing it causes a reconnect loop on the frontend
		log.Printf("User %s has existing connection, replacing with new one", client.UserID)

		// Close the old connection's underlying websocket to free resources
		// but don't close the send channel (that causes the reconnect loop)
		go func(oldClient *Client) {
			// Give the old client a moment to process, then close the conn
			time.Sleep(100 * time.Millisecond)
			if oldClient.conn != nil {
				oldClient.conn.Close()
			}
		}(existingClient)
	}

	h.clients[client.UserID] = client
	clientCount := len(h.clients)
	h.clientsMu.Unlock()

	log.Printf("Client registered: %s (Total: %d)", client.UserID, clientCount)

	// Set user online status in Redis (non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.redis.Set(ctx, "user:"+client.UserID+":online", "true", 30*time.Minute)
	}()

	// Send welcome message
	client.Send(map[string]interface{}{
		"type":    "connected",
		"message": "Connected to RANKEDterview",
	})
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.clientsMu.Lock()

	// Only unregister if this is the current client for this user
	if currentClient, ok := h.clients[client.UserID]; ok && currentClient == client {
		delete(h.clients, client.UserID)

		// Safely close the send channel
		select {
		case <-client.send:
			// Channel already closed
		default:
			close(client.send)
		}

		clientCount := len(h.clients)

		// Get the room ID before unlocking
		roomID := client.RoomID

		h.clientsMu.Unlock()

		log.Printf("Client unregistered: %s (Total: %d)", client.UserID, clientCount)

		// If the client was in a room, notify the partner
		if roomID != "" {
			log.Printf("Client %s was in room %s, notifying partner", client.UserID, roomID)
			h.BroadcastToRoomExcept(roomID, client.UserID, map[string]interface{}{
				"type":   "partner_disconnected",
				"from":   client.UserID,
				"roomId": roomID,
			})
		}

		// Remove online status in Redis (non-blocking)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			h.redis.Del(ctx, "user:"+client.UserID+":online")
		}()
	} else {
		h.clientsMu.Unlock()
	}
}

// handleBroadcast handles broadcast messages
func (h *Hub) handleBroadcast(message *Message) {
	payload, err := json.Marshal(message.Data)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	if message.Broadcast {
		h.broadcastToAllClients(payload, message.Exclude)
	} else if message.UserID != "" {
		h.sendToUser(message.UserID, payload)
	} else if message.RoomID != "" {
		h.broadcastToRoomInternal(message.RoomID, payload, message.Exclude)
	}
}

// broadcastToAllClients sends to all connected clients
func (h *Hub) broadcastToAllClients(payload []byte, exclude string) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for userID, client := range h.clients {
		if userID == exclude {
			continue
		}
		h.sendToClient(client, payload)
	}
}

// sendToUser sends to a specific user
func (h *Hub) sendToUser(userID string, payload []byte) {
	h.clientsMu.RLock()
	client, ok := h.clients[userID]
	h.clientsMu.RUnlock()

	if ok {
		h.sendToClient(client, payload)
	}
}

// sendToClient sends payload to a client with non-blocking write
func (h *Hub) sendToClient(client *Client, payload []byte) {
	select {
	case client.send <- payload:
		// Message sent
	default:
		// Buffer full - log but don't block
		log.Printf("Send buffer full for user %s, dropping message", client.UserID)
	}
}

// getRoomParticipants gets room participants with caching
func (h *Hub) getRoomParticipants(roomID string) (map[string]string, error) {
	h.roomsMu.RLock()
	if cached, ok := h.rooms[roomID]; ok {
		if time.Since(cached.cachedAt) < roomCacheTTL {
			h.roomsMu.RUnlock()
			return cached.participants, nil
		}
	}
	h.roomsMu.RUnlock()

	// Fetch from Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	participants, err := h.redis.HGetAll(ctx, "room:"+roomID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	h.roomsMu.Lock()
	h.rooms[roomID] = &roomCache{
		participants: participants,
		cachedAt:     time.Now(),
	}
	h.roomsMu.Unlock()

	return participants, nil
}

// invalidateRoomCache invalidates the room cache
func (h *Hub) invalidateRoomCache(roomID string) {
	h.roomsMu.Lock()
	delete(h.rooms, roomID)
	h.roomsMu.Unlock()
}

// broadcastToRoomInternal broadcasts to room participants
func (h *Hub) broadcastToRoomInternal(roomID string, payload []byte, exclude string) {
	log.Printf("broadcastToRoomInternal: roomID=%s, exclude=%s", roomID, exclude)

	participants, err := h.getRoomParticipants(roomID)
	if err != nil {
		log.Printf("Error getting room participants: %v", err)
		return
	}

	log.Printf("Room %s participants from Redis: %+v", roomID, participants)

	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	sentCount := 0
	for key, userID := range participants {
		log.Printf("Checking participant: key=%s, userID=%s, exclude=%s", key, userID, exclude)
		if (key == "user1" || key == "user2") && userID != exclude {
			if client, ok := h.clients[userID]; ok {
				log.Printf("Sending to user %s", userID)
				h.sendToClient(client, payload)
				sentCount++
			} else {
				log.Printf("User %s not connected (not in h.clients)", userID)
			}
		}
	}
	log.Printf("broadcastToRoomInternal: sent to %d clients", sentCount)
}

// Public methods

// BroadcastToRoomExcept broadcasts a message to all users in a room except the specified user
func (h *Hub) BroadcastToRoomExcept(roomID string, excludeUserID string, data map[string]interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.broadcastToRoomInternal(roomID, payload, excludeUserID)
}

// BroadcastToAll broadcasts a message to all connected clients
func (h *Hub) BroadcastToAll(data map[string]interface{}) {
	select {
	case h.broadcast <- &Message{Data: data, Broadcast: true}:
	default:
		log.Println("Broadcast buffer full, dropping message")
	}
}

// BroadcastToAllExcept broadcasts to all connected clients except the specified user
func (h *Hub) BroadcastToAllExcept(excludeUserID string, data map[string]interface{}) {
	select {
	case h.broadcast <- &Message{Data: data, Broadcast: true, Exclude: excludeUserID}:
	default:
		log.Println("Broadcast buffer full, dropping message")
	}
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID string, data map[string]interface{}) {
	select {
	case h.broadcast <- &Message{UserID: userID, Data: data}:
	default:
		log.Printf("Broadcast buffer full for user %s, dropping message", userID)
	}
}

// BroadcastToRoom broadcasts to all users in a room
func (h *Hub) BroadcastToRoom(roomID string, data map[string]interface{}) {
	select {
	case h.broadcast <- &Message{RoomID: roomID, Data: data}:
	default:
		log.Printf("Broadcast buffer full for room %s, dropping message", roomID)
	}
}

// GetOnlineUsers returns the number of online users
func (h *Hub) GetOnlineUsers() int {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	return len(h.clients)
}

// IsUserOnline checks if a user is connected
func (h *Hub) IsUserOnline(userID string) bool {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	_, exists := h.clients[userID]
	return exists
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	select {
	case h.register <- client:
	default:
		log.Printf("Register buffer full for user %s", client.UserID)
	}
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	select {
	case h.unregister <- client:
	default:
		log.Printf("Unregister buffer full for user %s", client.UserID)
	}
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	close(h.shutdown)
}
