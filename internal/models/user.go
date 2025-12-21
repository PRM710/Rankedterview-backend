package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email          string             `bson:"email" json:"email"`
	Name           string             `bson:"name" json:"name"`
	Avatar         string             `bson:"avatar" json:"avatar"`
	OAuthProvider  string             `bson:"oauthProvider" json:"oauthProvider"` // "google", "github"
	OAuthID        string             `bson:"oauthId" json:"oauthId"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	LastLoginAt    time.Time          `bson:"lastLoginAt" json:"lastLoginAt"`
	Stats          UserStats          `bson:"stats" json:"stats"`
	Settings       UserSettings       `bson:"settings" json:"settings"`
}

// UserStats holds user statistics
type UserStats struct {
	TotalInterviews int     `bson:"totalInterviews" json:"totalInterviews"`
	TotalScore      float64 `bson:"totalScore" json:"totalScore"`
	AverageScore    float64 `bson:"averageScore" json:"averageScore"`
	CurrentRank     int     `bson:"currentRank" json:"currentRank"`
	CurrentElo      int     `bson:"currentElo" json:"currentElo"`
}

// UserSettings holds user preferences
type UserSettings struct {
	Notifications bool `bson:"notifications" json:"notifications"`
	EmailUpdates  bool `bson:"emailUpdates" json:"emailUpdates"`
}

// CreateUserInput is the input for creating a new user
type CreateUserInput struct {
	Email         string `json:"email" binding:"required,email"`
	Name          string `json:"name" binding:"required"`
	Avatar        string `json:"avatar"`
	OAuthProvider string `json:"oauthProvider" binding:"required"`
	OAuthID       string `json:"oauthId" binding:"required"`
}

// UpdateUserInput is the input for updating a user
type UpdateUserInput struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// UserResponse is the response format for user data
type UserResponse struct {
	ID            string       `json:"id"`
	Email         string       `json:"email"`
	Name          string       `json:"name"`
	Avatar        string       `json:"avatar"`
	OAuthProvider string       `json:"oauthProvider"`
	CreatedAt     time.Time    `json:"createdAt"`
	LastLoginAt   time.Time    `json:"lastLoginAt"`
	Stats         UserStats    `json:"stats"`
	Settings      UserSettings `json:"settings"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID.Hex(),
		Email:         u.Email,
		Name:          u.Name,
		Avatar:        u.Avatar,
		OAuthProvider: u.OAuthProvider,
		CreatedAt:     u.CreatedAt,
		LastLoginAt:   u.LastLoginAt,
		Stats:         u.Stats,
		Settings:      u.Settings,
	}
}
