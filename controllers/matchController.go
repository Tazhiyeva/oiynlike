package controllers

import (
	"context"
	"fmt"
	"net/http"
	"oiynlike/database"
	"oiynlike/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JoinRequest struct {
	GameCardID string `json:"gameCardID" binding:"required"`
}

func GetUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User

	// Открываем коллекцию "users"
	usersCollection := database.OpenCollection("users")
	if usersCollection == nil {
		return user, fmt.Errorf("Failed to open users collection")
	}

	// Преобразуем строку user_id в ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return user, fmt.Errorf("Invalid user_id format")
	}

	// Формируем фильтр по _id
	filter := bson.M{"_id": objectID}

	// Ищем пользователя в коллекции
	err = usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, fmt.Errorf("User not found")
		}
		return user, fmt.Errorf("Error querying user: %v", err)
	}

	return user, nil
}

func JoinGameCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем user_id из JWT
		userID, _ := c.Get("uid")
		userIDString := fmt.Sprintf("%v", userID)

		// Извлекаем данные из JSON-тела запроса
		var joinRequest JoinRequest
		if err := c.ShouldBindJSON(&joinRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request"})
			return
		}

		// Получаем данные пользователя из коллекции users по user_id
		user, err := GetUserByID(c, userIDString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user data"})
			return
		}

		// Преобразуем строку в ObjectID
		gameCardID, err := primitive.ObjectIDFromHex(joinRequest.GameCardID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gameCardID"})
			return
		}

		// Получаем gameCard по gameCardID
		gameCard, err := getGameCardByID(c, gameCardID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving gameCard data"})
			return
		}

		// Проверяем, что пользователь еще не добавлен в matchedPlayers
		for _, player := range gameCard.MatchedPlayers {
			if player.UserID == userIDString {
				c.JSON(http.StatusBadRequest, gin.H{"error": "User already joined the gameCard"})
				return
			}
		}

		// Добавляем пользователя в matchedPlayers
		newMatchedPlayer := models.MatchedPlayer{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			UserID:    userIDString,
			PhotoURL:  user.PhotoURL,
			City:      user.City,
		}

		gameCard.MatchedPlayers = append(gameCard.MatchedPlayers, newMatchedPlayer)

		// Обновляем gameCard в базе данных
		err = updateGameCard(c, joinRequest.GameCardID, gameCard)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating gameCard"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "User joined the gameCard successfully"})
	}
}

// updateGameCard обновляет gameCard с заданным gameCardID
func updateGameCard(ctx context.Context, gameCardID string, updatedGameCard models.GameCard) error {
	objectID, err := primitive.ObjectIDFromHex(gameCardID)
	if err != nil {
		return fmt.Errorf("Invalid gameCardID format")
	}

	// Формируем фильтр по _id
	filter := bson.M{"_id": objectID}

	// Формируем обновленные данные
	update := bson.D{{Key: "$set", Value: updatedGameCard}}

	// Опции для FindOneAndUpdate
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	// Выполняем обновление в базе данных
	result := gameCardCollection.FindOneAndUpdate(ctx, filter, update, options)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return fmt.Errorf("GameCard not found")
		}
		return fmt.Errorf("Error updating gameCard: %v", result.Err())
	}

	// Декодируем обновленные данные
	var updatedCard models.GameCard
	if err := result.Decode(&updatedCard); err != nil {
		return fmt.Errorf("Error decoding updated gameCard: %v", err)
	}

	return nil
}
