package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/PRM710/Rankedterview-backend/internal/middleware"
	"github.com/PRM710/Rankedterview-backend/internal/services"
	"github.com/PRM710/Rankedterview-backend/internal/utils"
)

type RoomHandler struct {
	roomService *services.RoomService
}

func NewRoomHandler(roomService *services.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

// GetRoom retrieves room details
func (h *RoomHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	room, err := h.roomService.GetRoom(c.Request.Context(), roomID)
	if err != nil {
		utils.NotFoundResponse(c, "Room not found")
		return
	}

	utils.SuccessResponse(c, room.ToResponse())
}

// JoinRoom adds user to a room
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	err := h.roomService.JoinRoom(c.Request.Context(), roomID, userID)
	if err != nil {
		if err == services.ErrRoomFull {
			utils.ConflictResponse(c, "Room is full")
			return
		}
		if err == services.ErrRoomNotFound {
			utils.NotFoundResponse(c, "Room not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to join room: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Joined room successfully"})
}

// LeaveRoom removes user from a room
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	err := h.roomService.LeaveRoom(c.Request.Context(), roomID, userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to leave room")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Left room successfully"})
}

// GetRoomState retrieves current room state from Redis
func (h *RoomHandler) GetRoomState(c *gin.Context) {
	roomID := c.Param("roomId")

	state, err := h.roomService.GetRoomState(c.Request.Context(), roomID)
	if err != nil {
		utils.NotFoundResponse(c, "Room state not found")
		return
	}

	utils.SuccessResponse(c, state)
}
