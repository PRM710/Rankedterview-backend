package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yourusername/rankedterview-backend/internal/services"
	"github.com/yourusername/rankedterview-backend/internal/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration (placeholder for future email/password auth)
func (h *AuthHandler) Register(c *gin.Context) {
	utils.ErrorResponse(c, http.StatusNotImplemented, "Email/password registration not implemented. Please use OAuth.")
}

// Login handles user login (placeholder for future email/password auth)
func (h *AuthHandler) Login(c *gin.Context) {
	utils.ErrorResponse(c, http.StatusNotImplemented, "Email/password login not implemented. Please use OAuth.")
}

// GoogleOAuth initiates Google OAuth flow
func (h *AuthHandler) GoogleOAuth(c *gin.Context) {
	url, err := h.authService.GetOAuthURL("google")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authUrl": url,
	})
}

// GitHubOAuth initiates GitHub OAuth flow
func (h *AuthHandler) GitHubOAuth(c *gin.Context) {
	url, err := h.authService.GetOAuthURL("github")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authUrl": url,
	})
}

// OAuthCallback handles OAuth callback
func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	var input struct {
		Provider string `json:"provider" binding:"required"`
		OAuthID  string `json:"oauthId" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Avatar   string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestResponse(c, "Invalid request: "+err.Error())
		return
	}

	// Register or login user
	user, token, err := h.authService.RegisterWithOAuth(
		c.Request.Context(),
		input.Provider,
		input.OAuthID,
		input.Email,
		input.Name,
		input.Avatar,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Authentication failed: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user":    user.ToResponse(),
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestResponse(c, "Invalid request: "+err.Error())
		return
	}

	// Validate old token
	claims, err := h.authService.ValidateToken(input.Token)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid token")
		return
	}

	// Generate new token
	newToken, err := h.authService.RefreshToken(c.Request.Context(), claims.UserID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to refresh token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   newToken,
	})
}
