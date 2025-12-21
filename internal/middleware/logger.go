package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/PRM710/Rankedterview-backend/pkg/logger"
)

// Logger middleware logs all HTTP requests
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Log request
		log.Info(
			"%s %s %d %s %s",
			c.Request.Method,
			c.Request.URL.Path,
			statusCode,
			latency,
			clientIP,
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Error("Request error: %v", err.Err)
			}
		}
	}
}
