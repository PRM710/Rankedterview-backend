package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/PRM710/Rankedterview-backend/internal/database"
	"github.com/PRM710/Rankedterview-backend/internal/models"
	"github.com/PRM710/Rankedterview-backend/internal/repositories"
)

var (
	ErrAlreadyInQueue = errors.New("user already in matchmaking queue")
	ErrNotInQueue     = errors.New("user not in matchmaking queue")
	ErrNoMatchFound   = errors.New("no suitable match found")
)

type MatchmakingService struct {
	redis    *database.RedisClient
	roomRepo *repositories.RoomRepository
}

func NewMatchmakingService(redis *database.RedisClient, roomRepo *repositories.RoomRepository) *MatchmakingService {
	return &MatchmakingService{
		redis:    redis,
		roomRepo: roomRepo,
	}
}

// JoinQueue adds a user to the matchmaking queue
func (s *MatchmakingService) JoinQueue(ctx context.Context, userID string, skillLevel int) error {
	// Check if user is already in queue
	inQueue, err := s.IsInQueue(ctx, userID)
	if err != nil {
		return err
	}
	if inQueue {
		return ErrAlreadyInQueue
	}

	// Add to queue with current timestamp as score (for FIFO matching)
	queueKey := "matchmaking:queue"
	score := float64(time.Now().Unix())
	
	err = s.redis.Client.ZAdd(ctx, queueKey, database.Z{
		Score:  score,
		Member: userID,
	}).Err()
	if err != nil {
		return err
	}

	// Store user metadata
	metaKey := fmt.Sprintf("matchmaking:user:%s", userID)
	err = s.redis.HSet(ctx, metaKey,
		"skillLevel", skillLevel,
		"joinedAt", time.Now().Unix(),
	)
	if err != nil {
		return err
	}

	// Set expiration (30 minutes)
	s.redis.Expire(ctx, metaKey, 30*time.Minute)

	return nil
}

// LeaveQueue removes a user from the matchmaking queue
func (s *MatchmakingService) LeaveQueue(ctx context.Context, userID string) error {
	queueKey := "matchmaking:queue"
	
	// Remove from queue
	err := s.redis.Client.ZRem(ctx, queueKey, userID).Err()
	if err != nil {
		return err
	}

	// Remove metadata
	metaKey := fmt.Sprintf("matchmaking:user:%s", userID)
	s.redis.Del(ctx, metaKey)

	return nil
}

// IsInQueue checks if a user is in the matchmaking queue
func (s *MatchmakingService) IsInQueue(ctx context.Context, userID string) (bool, error) {
	queueKey := "matchmaking:queue"
	score, err := s.redis.ZScore(ctx, queueKey, userID)
	if err != nil {
		// User not in queue
		return false, nil
	}
	return score > 0, nil
}

// GetQueueStatus returns the user's position and estimated wait time
func (s *MatchmakingService) GetQueueStatus(ctx context.Context, userID string) (int, time.Duration, error) {
	queueKey := "matchmaking:queue"
	
	// Get user's rank in queue
	rank, err := s.redis.ZRank(ctx, queueKey, userID)
	if err != nil {
		return 0, 0, ErrNotInQueue
	}

	// Estimate wait time (assume 30 seconds per match)
	position := int(rank) + 1
	estimatedWait := time.Duration(position/2) * 30 * time.Second

	return position, estimatedWait, nil
}

// FindMatch attempts to find a match for a user
func (s *MatchmakingService) FindMatch(ctx context.Context, userID string) (string, string, error) {
	queueKey := "matchmaking:queue"

	// Get all users in queue
	members, err := s.redis.Client.ZRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return "", "", err
	}

	if len(members) < 2 {
		return "", "", ErrNoMatchFound
	}

	// Find the first two users (FIFO)
	var user1, user2 string
	for _, member := range members {
		if member == userID {
			user1 = member
		} else if user1 != "" {
			user2 = member
			break
		} else {
			user1 = member
		}
	}

	if user1 == "" || user2 == "" {
		return "", "", ErrNoMatchFound
	}

	// Create a room for the matched users
	roomID, err := s.CreateRoomForMatch(ctx, user1, user2)
	if err != nil {
		return "", "", err
	}

	// Remove both users from queue
	s.LeaveQueue(ctx, user1)
	s.LeaveQueue(ctx, user2)

	return roomID, user2, nil
}

// CreateRoomForMatch creates a room for matched users
func (s *MatchmakingService) CreateRoomForMatch(ctx context.Context, user1ID, user2ID string) (string, error) {
	// Generate unique room ID
	roomID, err := generateRoomID()
	if err != nil {
		return "", err
	}

	// Convert user IDs to ObjectIDs
	userObjID1, err := primitive.ObjectIDFromHex(user1ID)
	if err != nil {
		return "", err
	}

	userObjID2, err := primitive.ObjectIDFromHex(user2ID)
	if err != nil {
		return "", err
	}

	// Create room
	room := &models.Room{
		RoomID:       roomID,
		Status:       "waiting",
		Participants: []primitive.ObjectID{userObjID1, userObjID2},
		Metadata: models.RoomMetadata{
			Topic:      "Technical Interview",
			Difficulty: "medium",
			Type:       "technical",
		},
	}

	err = s.roomRepo.Create(ctx, room)
	if err != nil {
		return "", err
	}

	// Store room state in Redis
	roomStateKey := fmt.Sprintf("room:%s", roomID)
	s.redis.HSet(ctx, roomStateKey,
		"status", "waiting",
		"user1", user1ID,
		"user2", user2ID,
		"createdAt", time.Now().Unix(),
	)
	s.redis.Expire(ctx, roomStateKey, 2*time.Hour)

	return roomID, nil
}

// GetQueueSize returns the number of users in queue
func (s *MatchmakingService) GetQueueSize(ctx context.Context) (int64, error) {
	queueKey := "matchmaking:queue"
	return s.redis.Client.ZCard(ctx, queueKey).Result()
}

// generateRoomID generates a unique room ID
func generateRoomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
