package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Interview represents an interview session
type Interview struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID         string             `bson:"roomId" json:"roomId"`
	Participants   []Participant      `bson:"participants" json:"participants"`
	Status         string             `bson:"status" json:"status"` // "pending", "in_progress", "completed", "failed"
	StartedAt      time.Time          `bson:"startedAt" json:"startedAt"`
	EndedAt        time.Time          `bson:"endedAt" json:"endedAt"`
	Duration       int                `bson:"duration" json:"duration"` // seconds
	Recording      Recording          `bson:"recording" json:"recording"`
	Transcript     Transcript         `bson:"transcript" json:"transcript"`
	Evaluation     Evaluation         `bson:"evaluation" json:"evaluation"`
	RankingImpact  RankingImpact      `bson:"rankingImpact" json:"rankingImpact"`
}

// Participant represents a participant in an interview
type Participant struct {
	UserID   primitive.ObjectID `bson:"userId" json:"userId"`
	Role     string             `bson:"role" json:"role"` // "interviewer", "interviewee"
	JoinedAt time.Time          `bson:"joinedAt" json:"joinedAt"`
	LeftAt   time.Time          `bson:"leftAt" json:"leftAt"`
}

// Recording holds recording information
type Recording struct {
	RecallBotID   string    `bson:"recallBotId" json:"recallBotId"`
	Status        string    `bson:"status" json:"status"` // "recording", "processing", "completed", "failed"
	VideoURL      string    `bson:"videoUrl" json:"videoUrl"`
	AudioURL      string    `bson:"audioUrl" json:"audioUrl"`
	TranscriptURL string    `bson:"transcriptUrl" json:"transcriptUrl"`
	Metadata      string    `bson:"metadata" json:"metadata"`
}

// Transcript holds the interview transcript
type Transcript struct {
	Raw      string            `bson:"raw" json:"raw"`
	Segments []TranscriptSegment `bson:"segments" json:"segments"`
}

// TranscriptSegment represents a segment of the transcript
type TranscriptSegment struct {
	Speaker    string  `bson:"speaker" json:"speaker"`
	Text       string  `bson:"text" json:"text"`
	StartTime  float64 `bson:"startTime" json:"startTime"`
	EndTime    float64 `bson:"endTime" json:"endTime"`
	Confidence float64 `bson:"confidence" json:"confidence"`
}

// Evaluation holds AI evaluation results
type Evaluation struct {
	ProcessedAt time.Time  `bson:"processedAt" json:"processedAt"`
	Scores      Scores     `bson:"scores" json:"scores"`
	Feedback    Feedback   `bson:"feedback" json:"feedback"`
	AIModel     string     `bson:"aiModel" json:"aiModel"`
	TokensUsed  int        `bson:"tokensUsed" json:"tokensUsed"`
}

// Scores holds evaluation scores
type Scores struct {
	Communication float64 `bson:"communication" json:"communication"` // 0-100
	Technical     float64 `bson:"technical" json:"technical"`
	Confidence    float64 `bson:"confidence" json:"confidence"`
	Structure     float64 `bson:"structure" json:"structure"`
	Overall       float64 `bson:"overall" json:"overall"`
}

// Feedback holds AI-generated feedback
type Feedback struct {
	Strengths    []string    `bson:"strengths" json:"strengths"`
	Improvements []string    `bson:"improvements" json:"improvements"`
	Summary      string      `bson:"summary" json:"summary"`
	Highlights   []Highlight `bson:"highlights" json:"highlights"`
}

// Highlight represents a timestamped highlight
type Highlight struct {
	Timestamp float64 `bson:"timestamp" json:"timestamp"`
	Type      string  `bson:"type" json:"type"` // "good", "improve"
	Comment   string  `bson:"comment" json:"comment"`
}

// RankingImpact holds ranking changes
type RankingImpact struct {
	EloChange  int `bson:"eloChange" json:"eloChange"`
	RankChange int `bson:"rankChange" json:"rankChange"`
}

// InterviewResponse is the response format
type InterviewResponse struct {
	ID            string        `json:"id"`
	RoomID        string        `json:"roomId"`
	Participants  []Participant `json:"participants"`
	Status        string        `json:"status"`
	StartedAt     time.Time     `json:"startedAt"`
	EndedAt       time.Time     `json:"endedAt"`
	Duration      int           `json:"duration"`
	Recording     Recording     `json:"recording"`
	Transcript    Transcript    `json:"transcript"`
	Evaluation    Evaluation    `json:"evaluation"`
	RankingImpact RankingImpact `json:"rankingImpact"`
}

// ToResponse converts Interview to InterviewResponse
func (i *Interview) ToResponse() InterviewResponse {
	return InterviewResponse{
		ID:            i.ID.Hex(),
		RoomID:        i.RoomID,
		Participants:  i.Participants,
		Status:        i.Status,
		StartedAt:     i.StartedAt,
		EndedAt:       i.EndedAt,
		Duration:      i.Duration,
		Recording:     i.Recording,
		Transcript:    i.Transcript,
		Evaluation:    i.Evaluation,
		RankingImpact: i.RankingImpact,
	}
}
