package services

import (
	"context"
	"errors"
	"time"

	"github.com/PRM710/Rankedterview-backend/internal/models"
	"github.com/PRM710/Rankedterview-backend/internal/repositories"
)

var (
	ErrInterviewNotFound = errors.New("interview not found")
)

type InterviewService struct {
	interviewRepo *repositories.InterviewRepository
	roomRepo      *repositories.RoomRepository
}

func NewInterviewService(interviewRepo *repositories.InterviewRepository, roomRepo *repositories.RoomRepository) *InterviewService {
	return &InterviewService{
		interviewRepo: interviewRepo,
		roomRepo:      roomRepo,
	}
}

// CreateInterview creates a new interview for a room
func (s *InterviewService) CreateInterview(ctx context.Context, roomID string, participants []models.Participant) (*models.Interview, error) {
	interview := &models.Interview{
		RoomID:       roomID,
		Participants: participants,
		Status:       "in_progress",
		StartedAt:    time.Now(),
	}

	err := s.interviewRepo.Create(ctx, interview)
	if err != nil {
		return nil, err
	}

	// Link interview to room
	s.roomRepo.SetInterviewID(ctx, roomID, interview.ID)

	return interview, nil
}

// GetInterview retrieves an interview by ID
func (s *InterviewService) GetInterview(ctx context.Context, interviewID string) (*models.Interview, error) {
	return s.interviewRepo.FindByID(ctx, interviewID)
}

// GetInterviewByRoomID retrieves an interview by room ID
func (s *InterviewService) GetInterviewByRoomID(ctx context.Context, roomID string) (*models.Interview, error) {
	return s.interviewRepo.FindByRoomID(ctx, roomID)
}

// ListUserInterviews retrieves all interviews for a user
func (s *InterviewService) ListUserInterviews(ctx context.Context, userID string, page, limit int64) ([]*models.Interview, error) {
	skip := (page - 1) * limit
	return s.interviewRepo.FindByUserID(ctx, userID, skip, limit)
}

// CompleteInterview marks an interview as completed
func (s *InterviewService) CompleteInterview(ctx context.Context, interviewID string) error {
	// Get interview
	interview, err := s.interviewRepo.FindByID(ctx, interviewID)
	if err != nil {
		return err
	}

	// Update status and calculate duration
	interview.Status = "completed"
	interview.EndedAt = time.Now()
	interview.Duration = int(interview.EndedAt.Sub(interview.StartedAt).Seconds())

	return s.interviewRepo.Update(ctx, interview)
}

// UpdateRecording updates the recording information
func (s *InterviewService) UpdateRecording(ctx context.Context, interviewID string, recording models.Recording) error {
	return s.interviewRepo.UpdateRecording(ctx, interviewID, recording)
}

// UpdateTranscript updates the interview transcript
func (s *InterviewService) UpdateTranscript(ctx context.Context, interviewID string, transcript models.Transcript) error {
	return s.interviewRepo.UpdateTranscript(ctx, interviewID, transcript)
}

// UpdateEvaluation updates the AI evaluation results
func (s *InterviewService) UpdateEvaluation(ctx context.Context, interviewID string, evaluation models.Evaluation) error {
	return s.interviewRepo.UpdateEvaluation(ctx, interviewID, evaluation)
}

// GetTranscript retrieves the interview transcript
func (s *InterviewService) GetTranscript(ctx context.Context, interviewID string) (*models.Transcript, error) {
	interview, err := s.interviewRepo.FindByID(ctx, interviewID)
	if err != nil {
		return nil, err
	}
	return &interview.Transcript, nil
}

// GetRecording retrieves the recording URLs
func (s *InterviewService) GetRecording(ctx context.Context, interviewID string) (*models.Recording, error) {
	interview, err := s.interviewRepo.FindByID(ctx, interviewID)
	if err != nil {
		return nil, err
	}
	return &interview.Recording, nil
}

// GetFeedback retrieves the AI feedback
func (s *InterviewService) GetFeedback(ctx context.Context, interviewID string) (*models.Feedback, error) {
	interview, err := s.interviewRepo.FindByID(ctx, interviewID)
	if err != nil {
		return nil, err
	}
	return &interview.Evaluation.Feedback, nil
}

// DeleteInterview deletes an interview
func (s *InterviewService) DeleteInterview(ctx context.Context, interviewID string) error {
	return s.interviewRepo.Delete(ctx, interviewID)
}

// CountUserInterviews counts the total interviews for a user
func (s *InterviewService) CountUserInterviews(ctx context.Context, userID string) (int64, error) {
	return s.interviewRepo.CountByUserID(ctx, userID)
}

// ProcessWebhook processes a webhook from Recall.ai
func (s *InterviewService) ProcessWebhook(ctx context.Context, interviewID string, webhookData map[string]interface{}) error {
	// Extract recording information from webhook
	recording := models.Recording{
		Status:        "completed",
		VideoURL:      getStringOrEmpty(webhookData, "video_url"),
		AudioURL:      getStringOrEmpty(webhookData, "audio_url"),
		TranscriptURL: getStringOrEmpty(webhookData, "transcript_url"),
	}

	return s.UpdateRecording(ctx, interviewID, recording)
}

// Helper function
func getStringOrEmpty(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
