package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"oiynlike/database"
	"oiynlike/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var gameCardCollection *mongo.Collection = database.OpenCollection("gamecards")

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
		firstName, _ := c.Get("first_name")
		lastName, _ := c.Get("last_name")

		hostUserID := fmt.Sprintf("%v", userID)
		hostUserFirstName := fmt.Sprintf("%v", firstName)
		hostUserLastName := fmt.Sprintf("%v", lastName)

		gameCard.HostUser = models.User{
			UserId:    &hostUserID,
			FirstName: &hostUserFirstName,
			LastName:  &hostUserLastName,
		}

		gameCard.CreatedAt = time.Now()
		gameCard.UpdatedAt = time.Now()
		gameCard.Status = "active"
		gameCard.CurrentPlayers = 1
		gameCard.MatchedPlayers = []*models.User{} // Пустой массив для начала

		// Вставляем созданную GameCard в базу данных
		result, err := gameCardCollection.InsertOne(ctx, gameCard)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отправляем успешный ответ с созданной GameCard
		c.JSON(http.StatusCreated, gin.H{"msg": "Game card successfully created.", "id": result.InsertedID})
	}
}

// Главная страница пользователя. Клиент запрашивает список карточек, и получает все активные кроме своих
func GetActiveGameCards() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Получаем UserID из JWT токена
		userID, _ := c.Get("uid")
		currentUserID := fmt.Sprintf("%v", userID)

		// Извлекаем параметры пагинации
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))
		if page <= 0 {
			page = 1
		}
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		// Рассчитываем смещение для запроса
		offset := (page - 1) * limit

		// Формируем фильтр для исключения карт текущего пользователя
		filter := bson.M{
			"status":          "active",
			"hostuser.userid": bson.M{"$ne": currentUserID},
		}

		// Запрашиваем активные игровые карты с учетом пагинации
		cursor, err := gameCardCollection.Find(ctx, filter, options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		// Проходим по результатам запроса и добавляем их в срез
		var gameCards []models.GameCard
		if err := cursor.All(ctx, &gameCards); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отправляем успешный ответ с активными игровыми картами
		c.JSON(http.StatusOK, gameCards)
	}
}

func GetUserGameCards() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Получаем UserID из JWT токена
		userID, _ := c.Get("uid")
		currentUserID := fmt.Sprintf("%v", userID)

		// Извлекаем параметры фильтрации и пагинации
		status := c.Query("status")
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))
		if page <= 0 {
			page = 1
		}
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		// Рассчитываем смещение для запроса
		offset := (page - 1) * limit

		// Формируем фильтр по UserID и статусу
		filter := bson.M{"hostuser.userid": currentUserID}
		if status != "" {
			filter["status"] = status
		}

		// Запрашиваем игровые карты пользователя с учетом фильтрации и пагинации
		cursor, err := gameCardCollection.Find(ctx, filter, options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		// Проходим по результатам запроса и добавляем их в срез
		var userGameCards []models.GameCard
		if err := cursor.All(ctx, &userGameCards); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отправляем успешный ответ с игровыми картами пользователя
		c.JSON(http.StatusOK, userGameCards)
	}
}
