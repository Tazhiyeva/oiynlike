package controllers

import (
	"context"
	"net/http"
	"time"

	"oiynlike/database"
	"oiynlike/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var gameCardCollection *mongo.Collection = database.OpenCollection(database.ConnectToMongoDB(), "gamecards")

func CreateGameCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Получаем данные из запроса
		var gameCard models.GameCard
		if err := c.BindJSON(&gameCard); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Получаем данные пользователя из JWT токена
		userID, _ := c.Get("uid")
		firstName, _ := c.Get("firstName")
		lastName, _ := c.Get("lastName")

		hostUserID := userID.(string)
		hostUserFirstName := firstName.(string)
		hostUserLastName := lastName.(string)

		gameCard.HostUser = models.User{
			UserId:    &hostUserID,
			FirstName: &hostUserFirstName,
			LastName:  &hostUserLastName,
		}

		gameCard.CreatedAt = time.Now()
		gameCard.UpdatedAt = time.Now()
		gameCard.Status = "active"
		gameCard.CurrentPlayers = 1
		gameCard.MatchedPlayers = []models.User{} // Пустой массив для начала

		// Вставляем созданную GameCard в базу данных
		result, err := gameCardCollection.InsertOne(ctx, gameCard)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отправляем успешный ответ с созданной GameCard
		c.JSON(http.StatusCreated, gin.H{"message": "GameCard created successfully", "id": result.InsertedID})
	}
}
