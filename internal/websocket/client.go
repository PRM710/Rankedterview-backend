package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket client connection
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	UserID string
	RoomID string
}

// NewClient creates a new client
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		UserID: userID,
	}
}

// Send sends a message to the client
func (c *Client) Send(data map[string]interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	select {
	case c.send <- payload:
		return nil
	default:
		return nil // Drop if buffer full
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
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

			// Send each message as a separate WebSocket frame
			// Do NOT batch messages - frontend expects each JSON separately
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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
		log.Printf("Unknown event type: %s", msg.Type)
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

	ctx := context.Background()

	// Track which users have accepted this match
	acceptKey := "match:" + roomID + ":accepted"

	// Add this user to the set of users who have accepted
	c.hub.redis.SAdd(ctx, acceptKey, c.UserID)
	c.hub.redis.Expire(ctx, acceptKey, 2*time.Minute)

	// Check how many users have accepted
	acceptedUsers, err := c.hub.redis.SMembers(ctx, acceptKey)
	if err != nil {
		log.Printf("Error getting accepted users: %v", err)
		return
	}

	log.Printf("Room %s accepted users: %v (count: %d)", roomID, acceptedUsers, len(acceptedUsers))

	if len(acceptedUsers) == 1 {
		// First user accepted - notify them to wait
		c.Send(map[string]interface{}{
			"type":    EventPartnerAccepted,
			"message": "Waiting for partner to accept...",
		})
	}

	if len(acceptedUsers) >= 2 {
		// Both users accepted - notify everyone to start the call
		log.Printf("Both users accepted for room %s, notifying with roles", roomID)

		// Get room participants from Redis
		roomKey := "room:" + roomID
		participants, err := c.hub.redis.HGetAll(ctx, roomKey)
		if err != nil {
			log.Printf("Error getting room participants: %v", err)
		}

		// Determine caller/callee - first to accept is caller
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

		// Also notify room participants who might have different IDs
		for key, userID := range participants {
			if (key == "user1" || key == "user2") && userID != c.UserID {
				c.hub.BroadcastToUser(userID, map[string]interface{}{
					"type":   EventPartnerAccepted,
					"roomId": roomID,
				})
			}
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

	c.hub.BroadcastToRoomExcept(msg.RoomID, c.UserID, map[string]interface{}{
		"type":    msg.Type,
		"from":    c.UserID,
		"roomId":  msg.RoomID,
		"message": msg.Data["message"],
	})
}

// handleCallEnded handles when a user ends a call
func (c *Client) handleCallEnded(msg Event) {
	roomID := msg.RoomID
	if roomID == "" {
		roomID = c.RoomID
	}

	if roomID == "" {
		return
	}

	log.Printf("User %s ended call in room %s", c.UserID, roomID)

	c.hub.BroadcastToRoomExcept(roomID, c.UserID, map[string]interface{}{
		"type":   EventCallEnd,
		"from":   c.UserID,
		"roomId": roomID,
	})
}

// handleMediaStateChanged handles when a user changes their media state (mute/video toggle)
func (c *Client) handleMediaStateChanged(msg Event) {
	roomID := msg.RoomID
	if roomID == "" {
		roomID = c.RoomID
	}

	if roomID == "" {
		return
	}

	c.hub.BroadcastToRoomExcept(roomID, c.UserID, map[string]interface{}{
		"type":       EventMediaStateChange,
		"from":       c.UserID,
		"roomId":     roomID,
		"isMuted":    msg.Data["isMuted"],
		"isVideoOff": msg.Data["isVideoOff"],
	})
}
