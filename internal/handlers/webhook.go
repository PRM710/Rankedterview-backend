package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/PRM710/Rankedterview-backend/internal/config"
	"github.com/PRM710/Rankedterview-backend/internal/services"
	"github.com/PRM710/Rankedterview-backend/internal/utils"
)

type WebhookHandler struct {
	interviewService  *services.InterviewService
	evaluationService *services.EvaluationService
	rankingService    *services.RankingService
	config            *config.Config
}

func NewWebhookHandler(
	interviewService *services.InterviewService,
	rankingService *services.RankingService,
	cfg *config.Config,
) *WebhookHandler {
	return &WebhookHandler{
		interviewService:  interviewService,
		evaluationService: services.NewEvaluationService(cfg),
		rankingService:    rankingService,
		config:            cfg,
	}
}

// RecallWebhook handles webhooks from Recall.ai
func (h *WebhookHandler) RecallWebhook(c *gin.Context) {
	// Verify webhook secret
	secret := c.GetHeader("X-Recall-Secret")
	if secret != h.config.RecallWebhookSecret {
		utils.UnauthorizedResponse(c, "Invalid webhook secret")
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.BadRequestResponse(c, "Invalid payload")
		return
	}

	// Extract interview ID from payload
	interviewID, ok := payload["interview_id"].(string)
	if !ok {
		utils.BadRequestResponse(c, "Missing interview_id")
		return
	}

	// Process the webhook asynchronously
	go h.processRecallWebhook(interviewID, payload)

	// Return success immediately
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook received",
	})
}

// processRecallWebhook processes the webhook data
func (h *WebhookHandler) processRecallWebhook(interviewID string, payload map[string]interface{}) {
	ctx := context.Background()

	// Step 1: Update recording information
	err := h.interviewService.ProcessWebhook(ctx, interviewID, payload)
	if err != nil {
		// Log error but don't fail
		return
	}

	// Step 2: Get transcript
	transcript, err := h.interviewService.GetTranscript(ctx, interviewID)
	if err != nil || transcript.Raw == "" {
		return
	}

	// Step 3: Evaluate interview with AI
	evaluation, err := h.evaluationService.EvaluateInterview(ctx, transcript.Raw)
	if err != nil {
		// Log error
		return
	}

	// Step 4: Save evaluation
	err = h.interviewService.UpdateEvaluation(ctx, interviewID, *evaluation)
	if err != nil {
		return
	}

	// Step 5: Get interview to find participants
	interview, err := h.interviewService.GetInterview(ctx, interviewID)
	if err != nil {
		return
	}

	// Step 6: Update rankings for participants
	for _, participant := range interview.Participants {
		h.rankingService.UpdateUserRanking(ctx, participant.UserID.Hex(), evaluation.Scores)
	}
}
