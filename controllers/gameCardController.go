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

func createHostUserFromJWT(c *gin.Context) models.HostUser {
	userID, _ := c.Get("uid")
	firstName, _ := c.Get("first_name")
	lastName, _ := c.Get("last_name")

	hostUserID := fmt.Sprintf("%v", userID)
	hostUserFirstName := fmt.Sprintf("%v", firstName)
	hostUserLastName := fmt.Sprintf("%v", lastName)

	return models.HostUser{
		FirstName: hostUserFirstName,
		LastName:  hostUserLastName,
		UserID:    hostUserID,
	}
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

		// Получаем данные пользователя из JWT токена
		gameCard.HostUser = createHostUserFromJWT(c)

		// Устанавливаем временные метки и статус
		gameCard.CreatedAt = time.Now()
		gameCard.UpdatedAt = time.Now()
		gameCard.Status = "active"
		gameCard.CurrentPlayers = 1
		gameCard.MatchedPlayers = []*models.User{} // Пустой массив для начала

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
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))
		sort := c.Query("sort")
		if page <= 0 {
			page = 1
		}
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		// Рассчитываем смещение для запроса
		offset := (page - 1) * limit

		userID, _ := c.Get("uid")
		currentUserID := fmt.Sprintf("%v", userID)

		// Формируем фильтр для исключения карт текущего пользователя
		filter := bson.M{
			"status":            "active",
			"host_user.user_id": bson.M{"$ne": currentUserID},
		}

		sortOption := buildSortOption(sort, "created_at")

		// Запрашиваем активные игровые карты с учетом пагинации
		cursor, err := gameCardCollection.Find(ctx, filter, options.Find().SetSort(sortOption).SetLimit(int64(limit)).SetSkip(int64(offset)))
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
