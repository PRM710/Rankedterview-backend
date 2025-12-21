package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/PRM710/Rankedterview-backend/internal/database"
	"github.com/PRM710/Rankedterview-backend/internal/models"
	"github.com/PRM710/Rankedterview-backend/internal/repositories"
)

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrRoomFull        = errors.New("room is full")
	ErrNotParticipant  = errors.New("user is not a participant")
	ErrRoomNotActive   = errors.New("room is not active")
)

type RoomService struct {
	roomRepo *repositories.RoomRepository
	redis    *database.RedisClient
}

func NewRoomService(roomRepo *repositories.RoomRepository, redis *database.RedisClient) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
		redis:    redis,
	}
}

// GetRoom retrieves a room by its room ID
func (s *RoomService) GetRoom(ctx context.Context, roomID string) (*models.Room, error) {
	return s.roomRepo.FindByRoomID(ctx, roomID)
}

// JoinRoom adds a user to a room
func (s *RoomService) JoinRoom(ctx context.Context, roomID, userID string) error {
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return ErrRoomNotFound
	}

	// Check if room is full (max 2 participants)
	if len(room.Participants) >= 2 {
		return ErrRoomFull
	}

	// Convert userID to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Add participant
	err = s.roomRepo.AddParticipant(ctx, roomID, userObjID)
	if err != nil {
		return err
	}

	// Update Redis room state
	roomStateKey := fmt.Sprintf("room:%s", roomID)
	s.redis.Client.HSet(ctx, roomStateKey, fmt.Sprintf("user_%d", len(room.Participants)+1), userID)

	// If room now has 2 participants, mark as active
	if len(room.Participants)+1 >= 2 {
		s.roomRepo.UpdateStatus(ctx, roomID, "active")
		s.redis.Client.HSet(ctx, roomStateKey, "status", "active")
	}

	return nil
}

// LeaveRoom removes a user from a room
func (s *RoomService) LeaveRoom(ctx context.Context, roomID, userID string) error {
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return ErrRoomNotFound
	}

	// Convert userID to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Remove participant
	err = s.roomRepo.RemoveParticipant(ctx, roomID, userObjID)
	if err != nil {
		return err
	}

	// If room becomes empty, mark as ended
	if len(room.Participants) <= 1 {
		s.EndRoom(ctx, roomID)
	}

	return nil
}

// StartRoom marks a room as active
func (s *RoomService) StartRoom(ctx context.Context, roomID string) error {
	err := s.roomRepo.UpdateStatus(ctx, roomID, "active")
	if err != nil {
		return err
	}

	// Update Redis
	roomStateKey := fmt.Sprintf("room:%s", roomID)
	s.redis.Client.HSet(ctx, roomStateKey,
		"status", "active",
		"startedAt", time.Now().Unix(),
	)

	return nil
}

// EndRoom marks a room as ended
func (s *RoomService) EndRoom(ctx context.Context, roomID string) error {
	err := s.roomRepo.UpdateStatus(ctx, roomID, "ended")
	if err != nil {
		return err
	}

	// Update Redis
	roomStateKey := fmt.Sprintf("room:%s", roomID)
	s.redis.Client.HSet(ctx, roomStateKey,
		"status", "ended",
		"endedAt", time.Now().Unix(),
	)

	// Set TTL for cleanup
	s.redis.Expire(ctx, roomStateKey, 24*time.Hour)

	return nil
}

// GetRoomState retrieves the current room state from Redis
func (s *RoomService) GetRoomState(ctx context.Context, roomID string) (map[string]string, error) {
	roomStateKey := fmt.Sprintf("room:%s", roomID)
	return s.redis.HGetAll(ctx, roomStateKey)
}

// LinkInterview links an interview to a room
func (s *RoomService) LinkInterview(ctx context.Context, roomID string, interviewID primitive.ObjectID) error {
	return s.roomRepo.SetInterviewID(ctx, roomID, interviewID)
}

// GetActiveRooms retrieves all currently active rooms
func (s *RoomService) GetActiveRooms(ctx context.Context) ([]*models.Room, error) {
	return s.roomRepo.FindActiveRooms(ctx)
}

// CleanupOldRooms removes ended rooms older than the specified duration
func (s *RoomService) CleanupOldRooms(ctx context.Context, olderThan time.Duration) error {
	return s.roomRepo.CleanupOldRooms(ctx, olderThan)
}

// IsParticipant checks if a user is a participant in the room
func (s *RoomService) IsParticipant(ctx context.Context, roomID, userID string) (bool, error) {
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return false, err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	for _, participantID := range room.Participants {
		if participantID == userObjID {
			return true, nil
		}
	}

	return false, nil
}
