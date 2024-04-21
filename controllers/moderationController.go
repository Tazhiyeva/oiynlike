package controllers

import (
	"context"
	"net/http"
	"oiynlike/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetGameCardByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gamecard ID format"})
			return
		}

		// Формируем фильтр для поиска пользователя по ID
		filter := bson.M{"_id": objectID}

		// Выполняем запрос к базе данных
		var gameCard models.GameCard
		err = gameCardCollection.FindOne(context.Background(), filter).Decode(&gameCard)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": gameCard})

	}

}

func GetAllGameCards() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Запрашиваем активные игровые карты с учетом пагинации
		cursor, err := gameCardCollection.Find(ctx, options.Find())
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

		response := struct {
			Items []models.GameCard `json:"items"`
			Meta  struct {
				Current  int `json:"current"`
				Total    int `json:"total"`
				PageSize int `json:"page_size"`
			} `json:"meta"`
		}{
			Items: gameCards,
			Meta: struct {
				Current  int `json:"current"`
				Total    int `json:"total"`
				PageSize int `json:"page_size"`
			}{
				Current:  1,              // текущая страница (ваше значение)
				Total:    len(gameCards), // общее количество элементов
				PageSize: 10,             // размер страницы (ваше значение)
			},
		}

		// Отправляем успешный ответ с активными gamecards
		c.JSON(http.StatusOK, response)

	}
}

func UpdateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		gameCardID := c.Param("gameCardID")

		objectID, err := primitive.ObjectIDFromHex(gameCardID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "id is incorrect"})
			return
		}

		// Извлечение значения параметра status из form-data
		status := c.PostForm("status")

		// Проверка, было ли передано значение status
		if status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status parameter is required"})
			return
		}

		// Создание объекта GameCard для обновления
		updateData := models.GameCard{
			Status: status,
		}

		// Вызовите функцию обновления gameCard
		err = updateGameCard(context.Background(), objectID, updateData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating gameCard"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "GameCard status updated successfully"})
	}
}
