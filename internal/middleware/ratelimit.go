package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/rankedterview-backend/internal/database"
)

// RateLimiter middleware implements rate limiting using Redis
func RateLimiter(redis *database.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Create rate limit key
		key := fmt.Sprintf("ratelimit:%s", clientIP)

		// Get current count
		ctx := c.Request.Context()
		countStr, err := redis.Get(ctx, key)

		var count int
		if err == nil {
			fmt.Sscanf(countStr, "%d", &count)
		}

		// Check if limit exceeded (1000 requests per minute - increased for development)
		if count >= 1000 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		// Increment counter
		if count == 0 {
			// First request, set with expiration
			redis.Set(ctx, key, 1, time.Minute)
		} else {
			// Increment existing counter
			redis.Client.Incr(ctx, key)
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", "1000")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", 1000-count-1))

		c.Next()
	}
}
