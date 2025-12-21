package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PRM710/Rankedterview-backend/internal/database"
	"github.com/PRM710/Rankedterview-backend/internal/models"
)

type RoomRepository struct {
	collection *mongo.Collection
}

func NewRoomRepository(db *database.MongoDB) *RoomRepository {
	return &RoomRepository{
		collection: db.Collection("rooms"),
	}
}

// Create creates a new room
func (r *RoomRepository) Create(ctx context.Context, room *models.Room) error {
	room.ID = primitive.NewObjectID()
	room.CreatedAt = time.Now()
	room.Status = "waiting"

	_, err := r.collection.InsertOne(ctx, room)
	return err
}

// FindByID finds a room by ID
func (r *RoomRepository) FindByID(ctx context.Context, id string) (*models.Room, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var room models.Room
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&room)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

// FindByRoomID finds a room by its unique room ID
func (r *RoomRepository) FindByRoomID(ctx context.Context, roomID string) (*models.Room, error) {
	var room models.Room
	err := r.collection.FindOne(ctx, bson.M{"roomId": roomID}).Decode(&room)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

// Update updates a room
func (r *RoomRepository) Update(ctx context.Context, room *models.Room) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": room.ID},
		bson.M{"$set": room},
	)
	return err
}

// UpdateStatus updates the room status
func (r *RoomRepository) UpdateStatus(ctx context.Context, roomID, status string) error {
	update := bson.M{"$set": bson.M{"status": status}}

	// Set timestamps based on status
	switch status {
	case "active":
		update = bson.M{"$set": bson.M{
			"status":    status,
			"startedAt": time.Now(),
		}}
	case "ended":
		update = bson.M{"$set": bson.M{
			"status":  status,
			"endedAt": time.Now(),
		}}
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"roomId": roomID},
		update,
	)
	return err
}

// AddParticipant adds a participant to the room
func (r *RoomRepository) AddParticipant(ctx context.Context, roomID string, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"roomId": roomID},
		bson.M{"$addToSet": bson.M{"participants": userID}},
	)
	return err
}

// RemoveParticipant removes a participant from the room
func (r *RoomRepository) RemoveParticipant(ctx context.Context, roomID string, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"roomId": roomID},
		bson.M{"$pull": bson.M{"participants": userID}},
	)
	return err
}

// SetInterviewID sets the interview ID for the room
func (r *RoomRepository) SetInterviewID(ctx context.Context, roomID string, interviewID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"roomId": roomID},
		bson.M{"$set": bson.M{"interviewId": interviewID}},
	)
	return err
}

// Delete deletes a room
func (r *RoomRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// FindActiveRooms finds all active rooms
func (r *RoomRepository) FindActiveRooms(ctx context.Context) ([]*models.Room, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": "active"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []*models.Room
	if err = cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}

	return rooms, nil
}

// CleanupOldRooms deletes rooms older than the specified duration
func (r *RoomRepository) CleanupOldRooms(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	
	_, err := r.collection.DeleteMany(ctx, bson.M{
		"status":    "ended",
		"endedAt":   bson.M{"$lt": cutoff},
	})
	
	return err
}
