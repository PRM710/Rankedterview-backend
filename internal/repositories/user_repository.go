package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/yourusername/rankedterview-backend/internal/database"
	"github.com/yourusername/rankedterview-backend/internal/models"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *database.MongoDB) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.LastLoginAt = time.Now()
	
	// Initialize stats
	user.Stats = models.UserStats{
		TotalInterviews: 0,
		TotalScore:      0,
		AverageScore:    0,
		CurrentRank:     0,
		CurrentElo:      1000, // Starting ELO
	}
	
	// Initialize settings
	user.Settings = models.UserSettings{
		Notifications: true,
		EmailUpdates:  true,
	}

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByOAuthID finds a user by OAuth provider and ID
func (r *UserRepository) FindByOAuthID(ctx context.Context, provider, oauthID string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{
		"oauthProvider": provider,
		"oauthId":       oauthID,
	}).Decode(&user)
	
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": user},
	)
	return err
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"lastLoginAt": time.Now()}},
	)
	return err
}

// UpdateStats updates user statistics
func (r *UserRepository) UpdateStats(ctx context.Context, userID string, stats models.UserStats) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"stats": stats}},
	)
	return err
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// List lists users with pagination
func (r *UserRepository) List(ctx context.Context, skip, limit int64) ([]*models.User, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{},
		// options.Find().SetSkip(skip).SetLimit(limit),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
