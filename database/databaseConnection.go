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

// InitializeMongoDB создает и возвращает новый клиент MongoDB
func InitializeMongoDB() (*mongo.Client, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file: %v", err)
	}

	mongoURI := os.Getenv("MONGODB_URI")

	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to MongoDB: %v", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error pinging MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB successfully!")
	return client, nil
}

// ConnectToMongoDB возвращает существующий клиент MongoDB, или создает новый при первом вызове
func ConnectToMongoDB() (*mongo.Client, error) {
	once.Do(func() {
		client, err := InitializeMongoDB()
		if err != nil {
			initErr = err
			return
		}

		mongoClient = client
	})

	if initErr != nil {
		return nil, initErr
	}

	if mongoClient == nil {
		return nil, fmt.Errorf("MongoDB client is nil")
	}

	return mongoClient, nil
}

// OpenCollection открывает коллекцию MongoDB
func OpenCollection(collectionName string) *mongo.Collection {
	client, err := ConnectToMongoDB()
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		return nil
	}

	var collection *mongo.Collection = client.Database("oiynlike").Collection(collectionName)
	return collection
}
