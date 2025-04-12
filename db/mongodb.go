package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance      *mongo.Client
	clientInstanceError error
	mongoOnce           sync.Once
)

// ConnectMongoDB establishes a connection to MongoDB (thread-safe singleton)
func ConnectMongoDB() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uri := "mongodb+srv://jefferyokesamuel1:YShPTjt5BETnP770@christville.rss18.mongodb.net/?retryWrites=true&w=majority&appName=Christville"

		clientOptions := options.Client().
			ApplyURI(uri).
			SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			clientInstanceError = fmt.Errorf("connection failed: %w", err)
			return
		}

		err = client.Ping(ctx, nil)
		if err != nil {
			clientInstanceError = fmt.Errorf("ping failed: %w", err)
			_ = client.Disconnect(ctx)
			return
		}

		clientInstance = client
		log.Println("✅ MongoDB connected successfully")
	})

	return clientInstance, clientInstanceError
}

// GetClient returns the singleton MongoDB client instance
// Note: Ensure ConnectMongoDB() is called first
func GetClient() *mongo.Client {
	if clientInstance == nil {
		log.Println("⚠️ Warning: MongoDB client accessed before initialization")
	}
	return clientInstance
}

// DisconnectMongoDB gracefully closes the MongoDB connection
func DisconnectMongoDB() error {
	if clientInstance == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := clientInstance.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect MongoDB: %w", err)
	}

	log.Println("🗑️ MongoDB connection closed")
	return nil
}
