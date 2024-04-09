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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var gameCardCollection *mongo.Collection = database.OpenCollection("gamecards")

func validateGameCard(gameCard *models.GameCard) error {
	if err := validate.Struct(gameCard); err != nil {
		return err
	}
	return nil
}

func insertGameCard(ctx context.Context, gameCard models.GameCard) (primitive.ObjectID, error) {
	result, err := gameCardCollection.InsertOne(ctx, gameCard)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func getGameCardByID(ctx context.Context, id primitive.ObjectID) (models.GameCard, error) {
	gameCard := models.GameCard{}
	err := gameCardCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&gameCard)
	return gameCard, err
}

func CreateGameCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID, _ := c.Get("uid")
		userIDString := fmt.Sprintf("%v", userID)

		// Получаем данные из запроса
		var gameCard models.GameCard
		if err := c.BindJSON(&gameCard); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Валидация данных
		if err := validateGameCard(&gameCard); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		user, err := GetUserByID(c, userIDString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user data"})
			return
		}

		log.Print(user)

		hostUser := models.HostUser{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			UserID:    userIDString,
			PhotoURL:  user.PhotoURL,
			City:      user.City,
		}

		gameCard.HostUser = hostUser

		log.Printf("UserID: %s", userIDString)
		log.Printf("User: %+v", user)
		log.Printf("HostUser: %+v", hostUser)

		gameCard.CreatedAt = time.Now()
		gameCard.UpdatedAt = time.Now()
		gameCard.Status = "active"
		gameCard.MatchedPlayers = []models.MatchedPlayer{} // Пустой массив для начала

		// Вставляем созданную GameCard в базу данных
		insertedID, err := insertGameCard(ctx, gameCard)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Получаем созданный объект из базы данных по его ID
		createdGameCard, err := getGameCardByID(ctx, insertedID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отправляем успешный ответ с созданной GameCard
		c.JSON(http.StatusCreated, gin.H{"data": createdGameCard})
	}
}

func buildSortOption(sort, key string) bson.D {
	switch sort {
	case "desc":
		return bson.D{{Key: key, Value: -1}}
	case "asc":
		return bson.D{{Key: key, Value: 1}}
	default:
		return bson.D{{Key: key, Value: 1}} // По умолчанию сортируем по возрастанию
	}
}

func GetActiveGameCards() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Извлекаем параметры пагинации

		sort := c.Query("sort")
		city := c.Query("city")
		from := c.Query("from")
		to := c.Query("to")

		// Рассчитываем смещение для запроса

		userID, _ := c.Get("uid")
		currentUserID := fmt.Sprintf("%v", userID)

		// Формируем фильтр для исключения карт текущего пользователя
		filter := bson.M{
			"status":            "active",
			"host_user.user_id": bson.M{"$ne": currentUserID},
		}

		if city != "" {
			filter["city"] = city
		}

		if from != "" && to != "" {
			fromTime, err := time.Parse(time.RFC3339, from)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date format"})
				return
			}
			toTime, err := time.Parse(time.RFC3339, to)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date format"})
				return
			}
			filter["scheduled_time"] = bson.M{"$gte": fromTime, "$lte": toTime}
		}

		sortOption := buildSortOption(sort, "created_at")

		// Запрашиваем активные игровые карты с учетом пагинации
		cursor, err := gameCardCollection.Find(ctx, filter, options.Find().SetSort(sortOption))
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

func UpdateGameCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		gameCardID := c.Param("gameCardID")

		// Получаем UserID из JWT токена
		userID, _ := c.Get("uid")
		currentUserID := fmt.Sprintf("%v", userID)

		isHost, err := isUserHostOfGameCard(context.Background(), currentUserID, gameCardID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking host status"})
			return
		}

		if !isHost {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the host can update the gameCard"})
			return
		}

		var updateData models.GameCard
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Вызовите функцию обновления gameCard
		err = updateGameCard(context.Background(), gameCardID, updateData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating gameCard"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "GameCard updated successfully"})
	}
}

func isUserHostOfGameCard(ctx context.Context, userID interface{}, gameCardID string) (bool, error) {
	// Преобразуем gameCardID в ObjectID
	objectID, err := primitive.ObjectIDFromHex(gameCardID)
	if err != nil {
		return false, fmt.Errorf("Invalid gameCardID format")
	}

	// Найдем gameCard в базе данных по ID и проверим, является ли пользователь хостом
	filter := bson.M{"_id": objectID, "host_user.user_id": userID}
	count, err := gameCardCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("Error checking host status: %v", err)
	}

	return count > 0, nil
}
