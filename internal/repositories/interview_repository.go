package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/PRM710/Rankedterview-backend/internal/database"
	"github.com/PRM710/Rankedterview-backend/internal/models"
)

type InterviewRepository struct {
	collection *mongo.Collection
}

func NewInterviewRepository(db *database.MongoDB) *InterviewRepository {
	return &InterviewRepository{
		collection: db.Collection("interviews"),
	}
}

// Create creates a new interview
func (r *InterviewRepository) Create(ctx context.Context, interview *models.Interview) error {
	interview.ID = primitive.NewObjectID()
	interview.StartedAt = time.Now()
	interview.Status = "pending"

	_, err := r.collection.InsertOne(ctx, interview)
	return err
}

// FindByID finds an interview by ID
func (r *InterviewRepository) FindByID(ctx context.Context, id string) (*models.Interview, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var interview models.Interview
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&interview)
	if err != nil {
		return nil, err
	}

	return &interview, nil
}

// FindByRoomID finds an interview by room ID
func (r *InterviewRepository) FindByRoomID(ctx context.Context, roomID string) (*models.Interview, error) {
	var interview models.Interview
	err := r.collection.FindOne(ctx, bson.M{"roomId": roomID}).Decode(&interview)
	if err != nil {
		return nil, err
	}

	return &interview, nil
}

// FindByUserID finds all interviews for a user
func (r *InterviewRepository) FindByUserID(ctx context.Context, userID string, skip, limit int64) ([]*models.Interview, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "startedAt", Value: -1}}) // Most recent first

	cursor, err := r.collection.Find(
		ctx,
		bson.M{"participants.userId": objectID},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var interviews []*models.Interview
	if err = cursor.All(ctx, &interviews); err != nil {
		return nil, err
	}

	return interviews, nil
}

// Update updates an interview
func (r *InterviewRepository) Update(ctx context.Context, interview *models.Interview) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": interview.ID},
		bson.M{"$set": interview},
	)
	return err
}

// UpdateStatus updates the interview status
func (r *InterviewRepository) UpdateStatus(ctx context.Context, id, status string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{"$set": bson.M{"status": status}}
	
	// If completing, set endedAt and calculate duration
	if status == "completed" {
		update = bson.M{
			"$set": bson.M{
				"status":  status,
				"endedAt": time.Now(),
			},
		}
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// UpdateRecording updates the recording information
func (r *InterviewRepository) UpdateRecording(ctx context.Context, id string, recording models.Recording) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"recording": recording}},
	)
	return err
}

// UpdateTranscript updates the transcript
func (r *InterviewRepository) UpdateTranscript(ctx context.Context, id string, transcript models.Transcript) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"transcript": transcript}},
	)
	return err
}

// UpdateEvaluation updates the AI evaluation
func (r *InterviewRepository) UpdateEvaluation(ctx context.Context, id string, evaluation models.Evaluation) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"evaluation": evaluation}},
	)
	return err
}

// Delete deletes an interview
func (r *InterviewRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// CountByUserID counts interviews for a user
func (r *InterviewRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	return r.collection.CountDocuments(ctx, bson.M{"participants.userId": objectID})
}
