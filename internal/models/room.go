package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Room represents an interview room
type Room struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID       string             `bson:"roomId" json:"roomId"` // unique identifier
	Status       string             `bson:"status" json:"status"` // "waiting", "active", "ended"
	Participants []primitive.ObjectID `bson:"participants" json:"participants"` // user IDs
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	StartedAt    time.Time          `bson:"startedAt" json:"startedAt"`
	EndedAt      time.Time          `bson:"endedAt" json:"endedAt"`
	InterviewID  primitive.ObjectID `bson:"interviewId,omitempty" json:"interviewId,omitempty"`
	Metadata     RoomMetadata       `bson:"metadata" json:"metadata"`
}

// RoomMetadata holds room configuration
type RoomMetadata struct {
	Topic      string `bson:"topic" json:"topic"`
	Difficulty string `bson:"difficulty" json:"difficulty"` // "easy", "medium", "hard"
	Type       string `bson:"type" json:"type"`             // "technical", "behavioral"
}

// RoomResponse is the response format
type RoomResponse struct {
	ID           string           `json:"id"`
	RoomID       string           `json:"roomId"`
	Status       string           `json:"status"`
	Participants []string         `json:"participants"`
	CreatedAt    time.Time        `json:"createdAt"`
	StartedAt    time.Time        `json:"startedAt"`
	EndedAt      time.Time        `json:"endedAt"`
	InterviewID  string           `json:"interviewId,omitempty"`
	Metadata     RoomMetadata     `json:"metadata"`
}

// ToResponse converts Room to RoomResponse
func (r *Room) ToResponse() RoomResponse {
	participantIDs := make([]string, len(r.Participants))
	for i, p := range r.Participants {
		participantIDs[i] = p.Hex()
	}

	interviewID := ""
	if !r.InterviewID.IsZero() {
		interviewID = r.InterviewID.Hex()
	}

	return RoomResponse{
		ID:           r.ID.Hex(),
		RoomID:       r.RoomID,
		Status:       r.Status,
		Participants: participantIDs,
		CreatedAt:    r.CreatedAt,
		StartedAt:    r.StartedAt,
		EndedAt:      r.EndedAt,
		InterviewID:  interviewID,
		Metadata:     r.Metadata,
	}
}
