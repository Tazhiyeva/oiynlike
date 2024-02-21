package controllers

import (
	"context"
	"fmt"
	"log"
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

// @Summary Create a new game card
// @Description Creates a new game card with the provided details and associates it with the requesting user
// @ID create-game-card
// @Tags GameCards
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}" default(api_key) "JWT token for user authentication"
// @Param gameCard body models.GameCard true "GameCard object to be created"
// @Success 201 {object} SuccessResponse "Created"
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /api/gamecards [post]
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

		gameCard.HostUser = models.HostUser{
			FirstName: hostUserFirstName,
			LastName:  hostUserLastName,
			UserID:    hostUserID,
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

// @Summary Get active game cards excluding the ones of the requesting user
// @Description Returns a list of active game cards, excluding the ones owned by the requesting user
// @ID get-active-game-cards
// @Tags GameCards
// @Produce json
// @Param page query int false "Page number for pagination (default is 1)"
// @Param limit query int false "Number of items to return per page (default is 10, maximum is 100)"
// @Security ApiKeyAuth
// @Success 200 {array} models.GameCard "OK"
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /api/gamecards [get]

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
		cursor, err := gameCardCollection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetLimit(int64(limit)).SetSkip(int64(offset)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		// Проходим по результатам запроса и добавляем их в слайс
		var gameCards []models.GameCard
		if err := cursor.All(ctx, &gameCards); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отправляем успешный ответ с активными gamecards
		c.JSON(http.StatusOK, gameCards)
	}
}

// @Summary Get user game cards
// @Description Get a list of game cards for the current user
// @Produce json
// @Param status query string false "Filter by game card status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} []GameCard
// @Router /gamecards [get]
func GetUserGameCards() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Handler reached!")

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
		filter := bson.M{"host_user.user_id": currentUserID}
		if status != "" {
			filter["status"] = status
		}

		// Запрашиваем игровые карты пользователя с учетом фильтрации и пагинации
		cursor, err := gameCardCollection.Find(ctx, filter, options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)))
		if err != nil {
			log.Println("Error querying database:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		// Проходим по результатам запроса и добавляем их в срез
		var userGameCards []models.GameCard
		if err := cursor.All(ctx, &userGameCards); err != nil {
			log.Println("Error decoding results:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отправляем успешный ответ с игровыми картами пользователя
		c.JSON(http.StatusOK, userGameCards)
		log.Println("Response sent:", userGameCards)
	}
}
