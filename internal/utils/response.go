package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// CreatedResponse sends a created response
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}

// BadRequestResponse sends a bad request error
func BadRequestResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, message)
}

// UnauthorizedResponse sends an unauthorized error
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message)
}

// NotFoundResponse sends a not found error
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message)
}

// InternalServerErrorResponse sends an internal server error
func InternalServerErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, message)
}

// ConflictResponse sends a conflict error
func ConflictResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusConflict, message)
}

// PaginatedResponse sends a paginated response
func PaginatedResponse(c *gin.Context, data interface{}, page, limit, total int64) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	})
}
