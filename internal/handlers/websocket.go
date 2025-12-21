package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	ws "github.com/PRM710/Rankedterview-backend/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate the origin properly
		return true
	},
}

type WebSocketHandler struct {
	hub *ws.Hub
}

func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Extract user ID from query parameter
	// In production, validate the user via JWT token
	userID := c.Query("userId")
	if userID == "" {
		// Try to get from Authorization header
		token := c.Query("token")
		if token != "" {
			// Validate token and extract userID
			// claims, err := utils.ValidateToken(token, cfg.JWTSecret)
			// if err == nil {
			//     userID = claims.UserID
			// }
		}
	}

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID or token required",
		})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create new client
	client := ws.NewClient(conn, userID, h.hub)

	// Register client with hub
	h.hub.Register(client)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}
