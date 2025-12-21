package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header missing",
			})
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is not valid",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Set user ID in context
		userID, ok := claims["userId"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in token",
			})
			c.Abort()
			return
		}

		c.Set("userId", userID)
		c.Set("userEmail", claims["email"])

		c.Next()
	}
}

// GetUserID extracts the user ID from context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userId")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetUserEmail extracts the user email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("userEmail")
	if !exists {
		return "", false
	}
	return email.(string), true
}
