package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Ranking represents a user's ranking
type Ranking struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID   primitive.ObjectID `bson:"userId" json:"userId"`
	Category string             `bson:"category" json:"category"` // "overall", "communication", "technical"
	Period   string             `bson:"period" json:"period"`     // "all_time", "monthly", "weekly", "daily"
	Rank     int                `bson:"rank" json:"rank"`
	Score    float64            `bson:"score" json:"score"`
	Elo      int                `bson:"elo" json:"elo"`
	UpdatedAt time.Time         `bson:"updatedAt" json:"updatedAt"`
	History  []RankingHistory   `bson:"history" json:"history"`
}

// RankingHistory tracks ranking changes over time
type RankingHistory struct {
	Date  time.Time `bson:"date" json:"date"`
	Rank  int       `bson:"rank" json:"rank"`
	Score float64   `bson:"score" json:"score"`
	Elo   int       `bson:"elo" json:"elo"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	UserID   string  `json:"userId"`
	UserName string  `json:"userName"`
	Avatar   string  `json:"avatar"`
	Rank     int     `json:"rank"`
	Score    float64 `json:"score"`
	Elo      int     `json:"elo"`
}

// RankingResponse is the response format
type RankingResponse struct {
	ID       string           `json:"id"`
	UserID   string           `json:"userId"`
	Category string           `json:"category"`
	Period   string           `json:"period"`
	Rank     int              `json:"rank"`
	Score    float64          `json:"score"`
	Elo      int              `json:"elo"`
	UpdatedAt time.Time       `json:"updatedAt"`
	History  []RankingHistory `json:"history"`
}

// ToResponse converts Ranking to RankingResponse
func (r *Ranking) ToResponse() RankingResponse {
	return RankingResponse{
		ID:       r.ID.Hex(),
		UserID:   r.UserID.Hex(),
		Category: r.Category,
		Period:   r.Period,
		Rank:     r.Rank,
		Score:    r.Score,
		Elo:      r.Elo,
		UpdatedAt: r.UpdatedAt,
		History:  r.History,
	}
}
