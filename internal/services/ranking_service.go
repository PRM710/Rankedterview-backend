package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/redis/go-redis/v9"

	"github.com/PRM710/Rankedterview-backend/internal/database"
	"github.com/PRM710/Rankedterview-backend/internal/models"
	"github.com/PRM710/Rankedterview-backend/internal/repositories"
)

type RankingService struct {
	rankingRepo *repositories.RankingRepository
	userRepo    *repositories.UserRepository
	redis       *database.RedisClient
}

func NewRankingService(rankingRepo *repositories.RankingRepository, redis *database.RedisClient) *RankingService {
	return &RankingService{
		rankingRepo: rankingRepo,
		redis:       redis,
	}
}

// UpdateUserRanking updates a user's ranking after an interview
func (s *RankingService) UpdateUserRanking(ctx context.Context, userID string, scores models.Scores) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Update overall ranking
	err = s.updateRanking(ctx, userObjID, "overall", "all_time", scores.Overall)
	if err != nil {
		return err
	}

	// Update category rankings
	categories := map[string]float64{
		"communication": scores.Communication,
		"technical":     scores.Technical,
		"confidence":    scores.Confidence,
		"structure":     scores.Structure,
	}

	for category, score := range categories {
		err = s.updateRanking(ctx, userObjID, category, "all_time", score)
		if err != nil {
			// Continue even if one category fails
			continue
		}
	}

	// Recalculate ranks
	s.RecalculateRanks(ctx, "overall", "all_time")

	return nil
}

// updateRanking updates a single ranking entry
func (s *RankingService) updateRanking(ctx context.Context, userID primitive.ObjectID, category, period string, newScore float64) error {
	// Try to find existing ranking
	ranking, err := s.rankingRepo.FindByUserID(ctx, userID.Hex(), category, period)
	
	if err != nil {
		// Create new ranking
		ranking = &models.Ranking{
			UserID:   userID,
			Category: category,
			Period:   period,
			Score:    newScore,
			Elo:      1000 + int(newScore*10), // Simple ELO calculation
			Rank:     0,
		}
		return s.rankingRepo.Create(ctx, ranking)
	}

	// Update existing ranking
	oldScore := ranking.Score
	ranking.Score = (oldScore + newScore) / 2 // Average of old and new
	ranking.Elo = calculateNewElo(ranking.Elo, newScore)
	
	// Add to history
	history := models.RankingHistory{
		Date:  time.Now(),
		Rank:  ranking.Rank,
		Score: ranking.Score,
		Elo:   ranking.Elo,
	}
	
	ranking.History = append(ranking.History, history)

	return s.rankingRepo.Update(ctx, ranking)
}

// GetGlobalLeaderboard retrieves the global leaderboard
func (s *RankingService) GetGlobalLeaderboard(ctx context.Context, limit int64) ([]*models.Ranking, error) {
	// Try Redis cache first
	leaderboardKey := "leaderboard:global:overall:all_time"
	cached, err := s.getLeaderboardFromCache(ctx, leaderboardKey, limit)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	// Fetch from database
	rankings, err := s.rankingRepo.GetTopRankings(ctx, "overall", "all_time", limit)
	if err != nil {
		return nil, err
	}

	// Cache the results
	s.cacheLeaderboard(ctx, leaderboardKey, rankings)

	return rankings, nil
}

// GetCategoryLeaderboard retrieves a category-specific leaderboard
func (s *RankingService) GetCategoryLeaderboard(ctx context.Context, category string, limit int64) ([]*models.Ranking, error) {
	leaderboardKey := "leaderboard:" + category + ":all_time"
	
	// Try cache
	cached, err := s.getLeaderboardFromCache(ctx, leaderboardKey, limit)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	// Fetch from database
	rankings, err := s.rankingRepo.GetTopRankings(ctx, category, "all_time", limit)
	if err != nil {
		return nil, err
	}

	// Cache
	s.cacheLeaderboard(ctx, leaderboardKey, rankings)

	return rankings, nil
}

// GetUserRank retrieves a user's current rank
func (s *RankingService) GetUserRank(ctx context.Context, userID, category string) (int, error) {
	return s.rankingRepo.GetUserRank(ctx, userID, category, "all_time")
}

// GetRankHistory retrieves a user's ranking history
func (s *RankingService) GetRankHistory(ctx context.Context, userID string) (*models.Ranking, error) {
	return s.rankingRepo.FindByUserID(ctx, userID, "overall", "all_time")
}

// RecalculateRanks recalculates all ranks for a category
func (s *RankingService) RecalculateRanks(ctx context.Context, category, period string) error {
	err := s.rankingRepo.RecalculateRanks(ctx, category, period)
	if err != nil {
		return err
	}

	// Invalidate cache
	leaderboardKey := "leaderboard:" + category + ":" + period
	s.redis.Del(ctx, leaderboardKey)

	return nil
}

// Helper: Get leaderboard from Redis cache
func (s *RankingService) getLeaderboardFromCache(ctx context.Context, key string, limit int64) ([]*models.Ranking, error) {
	// This is a simplified version - in production you'd serialize/deserialize properly
	return nil, redis.Nil
}

// Helper: Cache leaderboard in Redis
func (s *RankingService) cacheLeaderboard(ctx context.Context, key string, rankings []*models.Ranking) {
	// Add each ranking to sorted set with rank as score
	for _, ranking := range rankings {
		s.redis.Client.ZAdd(ctx, key, database.Z{
			Score:  float64(ranking.Rank),
			Member: ranking.UserID.Hex(),
		})
	}
	
	// Set expiration (5 minutes)
	s.redis.Expire(ctx, key, 5*time.Minute)
}

// calculateNewElo calculates new ELO rating
func calculateNewElo(currentElo int, score float64) int {
	// Simplified ELO calculation
	// In production, use proper ELO algorithm with K-factor
	k := 32.0
	expectedScore := 1.0 / (1.0 + float64(1000-currentElo)/400.0)
	actualScore := score / 100.0 // Normalize to 0-1
	
	change := k * (actualScore - expectedScore)
	return currentElo + int(change)
}
