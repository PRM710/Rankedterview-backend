package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/PRM710/Rankedterview-backend/internal/middleware"
	"github.com/PRM710/Rankedterview-backend/internal/services"
	"github.com/PRM710/Rankedterview-backend/internal/utils"
	"github.com/PRM710/Rankedterview-backend/internal/websocket"
)

type MatchmakingHandler struct {
	matchmakingService *services.MatchmakingService
	hub                *websocket.Hub
}

func NewMatchmakingHandler(matchmakingService *services.MatchmakingService, hub *websocket.Hub) *MatchmakingHandler {
	return &MatchmakingHandler{
		matchmakingService: matchmakingService,
		hub:                hub,
	}
}

// JoinQueue adds user to matchmaking queue
func (h *MatchmakingHandler) JoinQueue(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var input struct {
		SkillLevel int `json:"skillLevel"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		// Default skill level
		input.SkillLevel = 1000
	}

	err := h.matchmakingService.JoinQueue(c.Request.Context(), userID, input.SkillLevel)
	if err != nil {
		if err == services.ErrAlreadyInQueue {
			utils.ConflictResponse(c, "Already in queue")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to join queue: "+err.Error())
		return
	}

	// Notify via WebSocket
	h.hub.BroadcastToUser(userID, map[string]interface{}{
		"type":    "queue_joined",
		"message": "Successfully joined matchmaking queue",
	})

	// Try to find a match immediately
	go h.tryMatch(userID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Joined matchmaking queue",
	})
}

// LeaveQueue removes user from matchmaking queue
func (h *MatchmakingHandler) LeaveQueue(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	err := h.matchmakingService.LeaveQueue(c.Request.Context(), userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to leave queue")
		return
	}

	h.hub.BroadcastToUser(userID, map[string]interface{}{
		"type":    "queue_left",
		"message": "Left matchmaking queue",
	})

	utils.SuccessResponse(c, gin.H{"message": "Left queue"})
}

// GetQueueStatus returns the user's queue status
func (h *MatchmakingHandler) GetQueueStatus(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	position, estimatedWait, err := h.matchmakingService.GetQueueStatus(c.Request.Context(), userID)
	if err != nil {
		if err == services.ErrNotInQueue {
			utils.NotFoundResponse(c, "Not in queue")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get queue status")
		return
	}

	queueSize, _ := h.matchmakingService.GetQueueSize(c.Request.Context())

	// Also try to find a match on each poll (fallback if WebSocket fails)
	roomID, opponentID, matchErr := h.matchmakingService.FindMatch(c.Request.Context(), userID)
	if matchErr == nil {
		// Match found! Return match info instead of queue status
		matchData := map[string]interface{}{
			"type":   "match_found",
			"roomId": roomID,
		}

		// Notify opponent via WebSocket
		h.hub.BroadcastToUser(opponentID, matchData)

		utils.SuccessResponse(c, gin.H{
			"matchFound": true,
			"roomId":     roomID,
		})
		return
	}

	utils.SuccessResponse(c, gin.H{
		"position":      position,
		"estimatedWait": estimatedWait.Seconds(),
		"totalInQueue":  queueSize,
		"matchFound":    false,
	})
}

// tryMatch attempts to find a match for the user
func (h *MatchmakingHandler) tryMatch(userID string) {
	ctx := context.Background()

	roomID, opponentID, err := h.matchmakingService.FindMatch(ctx, userID)
	if err != nil {
		// No match found yet
		return
	}

	// Notify both users of the match
	matchData := map[string]interface{}{
		"type":   "match_found",
		"roomId": roomID,
	}

	h.hub.BroadcastToUser(userID, matchData)
	h.hub.BroadcastToUser(opponentID, matchData)
}
