package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yourusername/rankedterview-backend/internal/middleware"
	"github.com/yourusername/rankedterview-backend/internal/models"
	"github.com/yourusername/rankedterview-backend/internal/services"
	"github.com/yourusername/rankedterview-backend/internal/utils"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetCurrentUser retrieves the authenticated user's profile
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	utils.SuccessResponse(c, user.ToResponse())
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	utils.SuccessResponse(c, user.ToResponse())
}

// UpdateProfile updates the authenticated user's profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var input models.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestResponse(c, "Invalid request: "+err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update profile")
		return
	}

	utils.SuccessResponse(c, user.ToResponse())
}

// GetUserStats retrieves user statistics
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID := c.Param("id")

	stats, err := h.userService.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	utils.SuccessResponse(c, stats)
}

// ListUsers lists all users with pagination
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	users, total, err := h.userService.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve users")
		return
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	utils.PaginatedResponse(c, userResponses, page, limit, total)
}
