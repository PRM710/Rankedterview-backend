package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB wraps the MongoDB client
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(uri, database string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return &MongoDB{
		Client:   client,
		Database: client.Database(database),
	}, nil
}

// Disconnect closes the MongoDB connection
func (m *MongoDB) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return m.Client.Disconnect(ctx)
}

// Ping checks if the database is reachable
func (m *MongoDB) Ping(ctx context.Context) error {
	return m.Client.Ping(ctx, readpref.Primary())
}

// Collection returns a collection from the database
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
