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

type RankingRepository struct {
	collection *mongo.Collection
}

func NewRankingRepository(db *database.MongoDB) *RankingRepository {
	return &RankingRepository{
		collection: db.Collection("rankings"),
	}
}

// Create creates a new ranking
func (r *RankingRepository) Create(ctx context.Context, ranking *models.Ranking) error {
	ranking.ID = primitive.NewObjectID()
	ranking.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, ranking)
	return err
}

// FindByUserID finds rankings for a user
func (r *RankingRepository) FindByUserID(ctx context.Context, userID, category, period string) (*models.Ranking, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var ranking models.Ranking
	err = r.collection.FindOne(ctx, bson.M{
		"userId":   objectID,
		"category": category,
		"period":   period,
	}).Decode(&ranking)
	
	if err != nil {
		return nil, err
	}

	return &ranking, nil
}

// Update updates a ranking
func (r *RankingRepository) Update(ctx context.Context, ranking *models.Ranking) error {
	ranking.UpdatedAt = time.Now()
	
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": ranking.ID},
		bson.M{"$set": ranking},
	)
	return err
}

// Upsert creates or updates a ranking
func (r *RankingRepository) Upsert(ctx context.Context, ranking *models.Ranking) error {
	ranking.UpdatedAt = time.Now()
	
	filter := bson.M{
		"userId":   ranking.UserID,
		"category": ranking.Category,
		"period":   ranking.Period,
	}
	
	update := bson.M{"$set": ranking}
	opts := options.Update().SetUpsert(true)
	
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// AddHistory adds a history entry to a ranking
func (r *RankingRepository) AddHistory(ctx context.Context, rankingID string, history models.RankingHistory) error {
	objectID, err := primitive.ObjectIDFromHex(rankingID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$push": bson.M{"history": history},
			"$set":  bson.M{"updatedAt": time.Now()},
		},
	)
	return err
}

// GetTopRankings gets top N rankings for a category and period
func (r *RankingRepository) GetTopRankings(ctx context.Context, category, period string, limit int64) ([]*models.Ranking, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "rank", Value: 1}}). // Ascending (1 is best)
		SetLimit(limit)

	cursor, err := r.collection.Find(
		ctx,
		bson.M{
			"category": category,
			"period":   period,
		},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rankings []*models.Ranking
	if err = cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}

	return rankings, nil
}

// GetUserRank gets a user's rank in a specific category and period
func (r *RankingRepository) GetUserRank(ctx context.Context, userID, category, period string) (int, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	var ranking models.Ranking
	err = r.collection.FindOne(ctx, bson.M{
		"userId":   objectID,
		"category": category,
		"period":   period,
	}).Decode(&ranking)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, err
	}

	return ranking.Rank, nil
}

// RecalculateRanks recalculates ranks for all users in a category/period based on ELO
func (r *RankingRepository) RecalculateRanks(ctx context.Context, category, period string) error {
	// Find all rankings for this category/period, sorted by ELO descending
	opts := options.Find().SetSort(bson.D{{Key: "elo", Value: -1}})
	
	cursor, err := r.collection.Find(
		ctx,
		bson.M{
			"category": category,
			"period":   period,
		},
		opts,
	)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	// Update each ranking with its new rank
	rank := 1
	for cursor.Next(ctx) {
		var ranking models.Ranking
		if err := cursor.Decode(&ranking); err != nil {
			continue
		}

		_, err := r.collection.UpdateOne(
			ctx,
			bson.M{"_id": ranking.ID},
			bson.M{
				"$set": bson.M{
					"rank":      rank,
					"updatedAt": time.Now(),
				},
			},
		)
		if err != nil {
			continue
		}

		rank++
	}

	return nil
}

// Delete deletes a ranking
func (r *RankingRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
