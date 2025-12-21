package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/PRM710/Rankedterview-backend/internal/middleware"
	"github.com/PRM710/Rankedterview-backend/internal/services"
	"github.com/PRM710/Rankedterview-backend/internal/utils"
)

type InterviewHandler struct {
	interviewService *services.InterviewService
}

func NewInterviewHandler(interviewService *services.InterviewService) *InterviewHandler {
	return &InterviewHandler{
		interviewService: interviewService,
	}
}

// ListInterviews retrieves all interviews for the authenticated user
func (h *InterviewHandler) ListInterviews(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	interviews, err := h.interviewService.ListUserInterviews(c.Request.Context(), userID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve interviews")
		return
	}

	// Convert to response format
	responses := make([]interface{}, len(interviews))
	for i, interview := range interviews {
		responses[i] = interview.ToResponse()
	}

	// Get total count
	total, _ := h.interviewService.CountUserInterviews(c.Request.Context(), userID)

	utils.PaginatedResponse(c, responses, page, limit, total)
}

// GetInterview retrieves a specific interview
func (h *InterviewHandler) GetInterview(c *gin.Context) {
	interviewID := c.Param("id")

	interview, err := h.interviewService.GetInterview(c.Request.Context(), interviewID)
	if err != nil {
		utils.NotFoundResponse(c, "Interview not found")
		return
	}

	utils.SuccessResponse(c, interview.ToResponse())
}

// GetTranscript retrieves the interview transcript
func (h *InterviewHandler) GetTranscript(c *gin.Context) {
	interviewID := c.Param("id")

	transcript, err := h.interviewService.GetTranscript(c.Request.Context(), interviewID)
	if err != nil {
		utils.NotFoundResponse(c, "Interview not found")
		return
	}

	utils.SuccessResponse(c, transcript)
}

// GetRecordingURLs retrieves the recording URLs
func (h *InterviewHandler) GetRecordingURLs(c *gin.Context) {
	interviewID := c.Param("id")

	recording, err := h.interviewService.GetRecording(c.Request.Context(), interviewID)
	if err != nil {
		utils.NotFoundResponse(c, "Interview not found")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"videoUrl":      recording.VideoURL,
		"audioUrl":      recording.AudioURL,
		"transcriptUrl": recording.TranscriptURL,
		"status":        recording.Status,
	})
}

// GetFeedback retrieves the AI-generated feedback
func (h *InterviewHandler) GetFeedback(c *gin.Context) {
	interviewID := c.Param("id")

	feedback, err := h.interviewService.GetFeedback(c.Request.Context(), interviewID)
	if err != nil {
		utils.NotFoundResponse(c, "Interview feedback not found")
		return
	}

	utils.SuccessResponse(c, feedback)
}
