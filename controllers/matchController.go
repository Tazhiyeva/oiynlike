package controllers

import (
	"context"
	"fmt"
	"net/http"
	"oiynlike/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JoinRequest struct {
	GameCardID primitive.ObjectID `json:"gameCardID" binding:"required"`
}

func GetUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User

	// Преобразуем строку user_id в ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return user, fmt.Errorf("invalid user_id format")
	}

	// Формируем фильтр по _id
	filter := bson.M{"_id": objectID}

	// Ищем пользователя в коллекции
	err = userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, fmt.Errorf("user not found")
		}
		return user, fmt.Errorf("error querying user: %v", err)
	}

	return user, nil
}

func AddPlayerToGameCard(c *gin.Context, userID string, gameCardID primitive.ObjectID) error {
	// Получаем данные пользователя из коллекции users по user_id
	user, err := GetUserByID(c, userID)
	if err != nil {
		return fmt.Errorf("error retrieving user data: %v", err)
	}

	// Получаем gameCard по gameCardID
	gameCard, err := getGameCardByID(c, gameCardID)
	if err != nil {
		return fmt.Errorf("error retrieving gameCard data: %v", err)
	}

	// Проверяем, что пользователь еще не добавлен в matchedPlayers
	for _, player := range gameCard.MatchedPlayers {
		if player.UserID == userID {
			return fmt.Errorf("user already joined the gameCard")
		}
	}

	// Добавляем пользователя в matchedPlayers
	newMatchedPlayer := models.MatchedPlayer{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserID:    userID,
		PhotoURL:  user.PhotoURL,
		City:      user.City,
	}

	gameCard.MatchedPlayers = append(gameCard.MatchedPlayers, newMatchedPlayer)

	// Обновляем gameCard в базе данных
	err = updateGameCard(c, gameCardID, gameCard)
	if err != nil {
		return fmt.Errorf("error updating gameCard: %v", err)
	}

	return nil
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

		// Добавляем пользователя к игровой карте
		err := AddPlayerToGameCard(c, userIDString, joinRequest.GameCardID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Получаем данные игровой карты
		gameCard, err := getGameCardByID(c, joinRequest.GameCardID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving gameCard data"})
			return
		}

		// Создаем чат, если необходимо
		err = CreateChatIfNeeded(c, &gameCard)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = ChangeStatusIfNeeded(c, joinRequest.GameCardID, &gameCard)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отправляем успешный ответ пользователю
		c.JSON(http.StatusOK, gin.H{"msg": "User joined the gameCard successfully"})
	}
}

func ChangeStatusIfNeeded(ctx context.Context, gameCardID primitive.ObjectID, gameCard *models.GameCard) error {
	// Проверяем, равно ли количество присоединенных пользователей количеству нужных игроков
	if len(gameCard.MatchedPlayers) == gameCard.MaxPlayers {
		// Меняем статус карточки игры на "inactive"
		gameCard.Status = "inactive"

		// Обновляем карточку игры в базе данных
		err := updateGameCard(ctx, gameCardID, *gameCard)
		if err != nil {
			return fmt.Errorf("error updating gameCard status: %v", err)
		}
	}
	return nil
}

// updateGameCard обновляет gameCard с заданным gameCardID
func updateGameCard(ctx context.Context, gameCardID primitive.ObjectID, updatedGameCard models.GameCard) error {

	// Формируем фильтр по _id
	filter := bson.M{"_id": gameCardID}

	// Формируем динамический BSON-документ с учетом только указанных полей
	updateFields := bson.D{
		{Key: "$set", Value: bson.D{}},
	}

	if updatedGameCard.Title != "" {
		updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "title", Value: updatedGameCard.Title})
	}
	if updatedGameCard.Description != "" {
		updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "description", Value: updatedGameCard.Description})
	}
	if updatedGameCard.Status != "" {
		updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "status", Value: updatedGameCard.Status})
	}
	if updatedGameCard.City != "" {
		updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "city", Value: updatedGameCard.City})
	}
	if updatedGameCard.CoverURL != "" {
		updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "cover_url", Value: updatedGameCard.CoverURL})
	}
	if updatedGameCard.MaxPlayers != 0 {
		updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "max_players", Value: updatedGameCard.MaxPlayers})
	}
	if !updatedGameCard.ScheduledTime.IsZero() {
		updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "scheduled_time", Value: updatedGameCard.ScheduledTime})
	}

	// Обновляем matched_players
	if len(updatedGameCard.MatchedPlayers) > 0 {
		players := make([]interface{}, len(updatedGameCard.MatchedPlayers))
		for i, player := range updatedGameCard.MatchedPlayers {
			players[i] = bson.M{
				"first_name": player.FirstName,
				"last_name":  player.LastName,
				"user_id":    player.UserID,
				"photo_url":  player.PhotoURL,
				"city":       player.City,
			}
		}
		updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "matched_players", Value: players})
	}

	// Добавляем обновление поля "updated_at"
	updateFields[0].Value = append(updateFields[0].Value.(bson.D), bson.E{Key: "updated_at", Value: time.Now()})

	// Опции для FindOneAndUpdate
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	// Выполняем обновление в базе данных
	result := gameCardCollection.FindOneAndUpdate(ctx, filter, updateFields, options)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return fmt.Errorf("GameCard not found")
		}
		return fmt.Errorf("error updating gameCard: %v", result.Err())
	}

	// Декодируем обновленные данные
	var updatedCard models.GameCard
	if err := result.Decode(&updatedCard); err != nil {
		return fmt.Errorf("error decoding updated gameCard: %v", err)
	}

	return nil
}
