package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512 KB
)

// Client represents a WebSocket client
type Client struct {
	// The WebSocket connection
	conn *websocket.Conn

	// User ID
	UserID string

	// Current Room ID (if in interview)
	RoomID string

	// Buffered channel of outbound messages
	send chan []byte

	// Hub reference
	hub *Hub
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, userID string, hub *Hub) *Client {
	return &Client{
		conn:   conn,
		UserID: userID,
		RoomID: "",
		send:   make(chan []byte, 256),
		hub:    hub,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
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
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming WebSocket messages
func (c *Client) handleMessage(message []byte) {
	var msg Event
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	// Debug: log ALL incoming messages
	log.Printf("handleMessage: type=%s, from=%s, to=%s, roomId=%s", msg.Type, c.UserID, msg.To, msg.RoomID)

	// Handle different event types
	switch msg.Type {
	case "ping":
		// Heartbeat from client - respond with pong
		c.Send(map[string]interface{}{"type": "pong"})
		return

	case EventJoinQueue:
		// Client wants to join matchmaking queue
		// This is handled via HTTP API, so we just acknowledge
		c.Send(map[string]interface{}{
			"type":    "queue_ack",
			"message": "Queue request received",
		})

	case EventLeaveQueue:
		// Client wants to leave queue
		c.Send(map[string]interface{}{
			"type":    "queue_left_ack",
			"message": "Left queue",
		})

	case EventAcceptMatch:
		// User accepted the match
		c.handleAcceptMatch(msg)

	case EventWebRTCOffer, EventWebRTCAnswer, EventICECandidate:
		// Relay WebRTC signaling - handle all three the same way
		c.relayWebRTC(msg)

	case EventCallEnd:
		// User ended the call - notify room participants
		c.handleCallEnded(msg)

	case EventMediaStateChange:
		// User toggled mic/camera - notify room participants
		c.handleMediaStateChanged(msg)

	case EventMessage:
		// Relay chat message in room
		c.relayToRoom(msg)

	default:
		// Only log truly unknown events (not empty or common noise)
		if msg.Type != "" {
			log.Printf("Unknown event type: %s from user %s", msg.Type, c.UserID)
		}
	}
}

// handleAcceptMatch handles when a user accepts a match
func (c *Client) handleAcceptMatch(msg Event) {
	roomID := msg.RoomID
	if roomID == "" {
		log.Printf("No roomId in accept_match from %s", c.UserID)
		return
	}

	log.Printf("User %s accepted match for room %s", c.UserID, roomID)

	// Store acceptance in Redis
	ctx := context.Background()
	acceptKey := "room:" + roomID + ":accepted"

	// Add this user to the accepted set
	c.hub.redis.SAdd(ctx, acceptKey, c.UserID)
	c.hub.redis.Expire(ctx, acceptKey, 5*time.Minute)

	// Check how many users have accepted
	acceptedUsers, _ := c.hub.redis.SMembers(ctx, acceptKey)
	log.Printf("Accepted users for room %s: %v", roomID, acceptedUsers)

	if len(acceptedUsers) == 1 {
		// Only one user accepted so far - notify the other user
		// First, get the room participants
		roomKey := "room:" + roomID
		participants, _ := c.hub.redis.HGetAll(ctx, roomKey)

		// Find the other user and notify them
		for key, userID := range participants {
			if (key == "user1" || key == "user2") && userID != c.UserID {
				c.hub.BroadcastToUser(userID, map[string]interface{}{
					"type":   EventPartnerAccepted,
					"roomId": roomID,
				})
			}
		}
	} else if len(acceptedUsers) >= 2 {
		// Both users accepted! Determine roles and notify both
		// Sort users for deterministic role assignment (alphabetically)
		sort.Strings(acceptedUsers)

		log.Printf("Sorted accepted users for room %s: %v", roomID, acceptedUsers)

		for i, userID := range acceptedUsers {
			role := "caller"
			if i == 1 {
				role = "callee"
			}
			log.Printf("Assigning role %s to user %s", role, userID)
			c.hub.BroadcastToUser(userID, map[string]interface{}{
				"type":   EventBothReady,
				"roomId": roomID,
				"role":   role,
			})
		}

		// Clean up the acceptance key
		c.hub.redis.Del(ctx, acceptKey)
	}
}

// relayWebRTC relays WebRTC signaling messages to room participants only
func (c *Client) relayWebRTC(msg Event) {
	roomID := msg.To // The "to" field contains the roomId

	payload := map[string]interface{}{
		"type":      msg.Type,
		"from":      c.UserID,
		"roomId":    roomID,
		"sdp":       msg.SDP,
		"candidate": msg.Candidate,
	}

	if roomID == "" {
		log.Printf("No roomId in WebRTC message from %s, falling back to broadcast", c.UserID)
		c.hub.BroadcastToAllExcept(c.UserID, payload)
		return
	}

	// Track which room this client is in (for disconnect notification)
	if c.RoomID == "" {
		c.RoomID = roomID
		log.Printf("Client %s joined room %s", c.UserID, roomID)
	}

	log.Printf("relayWebRTC: %s from %s to room %s", msg.Type, c.UserID, roomID)

	// Try room-based delivery first, then fallback to broadcast all
	// This ensures WebRTC messages always get through
	c.hub.BroadcastToRoomExcept(roomID, c.UserID, payload)

	// Also broadcast to all as fallback (WebRTC is time-sensitive)
	// The receiver will filter by roomId anyway
	c.hub.BroadcastToAllExcept(c.UserID, payload)
}

// relayToRoom relays a message to all users in a room
func (c *Client) relayToRoom(msg Event) {
	if msg.RoomID == "" {
		return
	}

	c.hub.BroadcastToRoom(msg.RoomID, map[string]interface{}{
		"type":   msg.Type,
		"from":   c.UserID,
		"roomId": msg.RoomID,
		"data":   msg.Data,
	})
}

// handleCallEnded handles when a user ends the call and notifies room participants
func (c *Client) handleCallEnded(msg Event) {
	roomID := msg.RoomID
	if roomID == "" {
		log.Printf("No roomId in call_ended from %s", c.UserID)
		return
	}

	log.Printf("User %s ended call in room %s", c.UserID, roomID)

	// Notify other room participants (not the sender)
	c.hub.BroadcastToRoomExcept(roomID, c.UserID, map[string]interface{}{
		"type":   "call_ended",
		"from":   c.UserID,
		"roomId": roomID,
	})
}

// handleMediaStateChanged handles when a user toggles mic/camera
func (c *Client) handleMediaStateChanged(msg Event) {
	roomID := msg.RoomID
	if roomID == "" {
		log.Printf("No roomId in media_state_changed from %s", c.UserID)
		return
	}

	// Relay to other room participants
	c.hub.BroadcastToRoomExcept(roomID, c.UserID, map[string]interface{}{
		"type":       "media_state_changed",
		"from":       c.UserID,
		"roomId":     roomID,
		"isMuted":    msg.Data["isMuted"],
		"isVideoOff": msg.Data["isVideoOff"],
	})
}

// Send sends a message to the client
func (c *Client) Send(data map[string]interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		return
	}

	select {
	case c.send <- payload:
	default:
		// Send buffer is full
		log.Printf("Send buffer full for user %s", c.UserID)
	}
}
