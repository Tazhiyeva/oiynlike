// database/database.go

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
)

func ConnectToMongoDB() (*mongo.Client, error) {
	// Use sync.Once to ensure that the connection is established only once
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("Error loading .env file: %v\n", err)
			return
		}

		mongoURI := os.Getenv("MONGODB_URI")

		clientOptions := options.Client().ApplyURI(mongoURI)

		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			fmt.Printf("Error connecting to MongoDB: %v\n", err)
			return
		}

		err = client.Ping(context.Background(), nil)
		if err != nil {
			fmt.Printf("Error pinging MongoDB: %v\n", err)
			return
		}

		fmt.Println("Connected to MongoDB successfully!")
		mongoClient = client
	})

	return mongoClient, nil
}

func OpenCollection(collectionName string) *mongo.Collection {
	return mongoClient.Database("oiynlike").Collection(collectionName)
}
