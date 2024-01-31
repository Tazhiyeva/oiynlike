package database

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	once        sync.Once
	initErr     error // Variable to store the initialization error
)

func ConnectToMongoDB() *mongo.Client {
	// Use sync.Once to ensure that the connection is established only once
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			initErr = fmt.Errorf("Error loading .env file: %v", err)
			return
		}

		mongoURI := os.Getenv("MONGODB_URI")

		clientOptions := options.Client().ApplyURI(mongoURI)

		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			initErr = fmt.Errorf("Error connecting to MongoDB: %v", err)
			return
		}

		err = client.Ping(context.Background(), nil)
		if err != nil {
			initErr = fmt.Errorf("Error pinging MongoDB: %v", err)
			return
		}

		fmt.Println("Connected to MongoDB successfully!")
		mongoClient = client
	})

	return mongoClient
}

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("oiynlike").Collection(collectionName)
	return collection
}
