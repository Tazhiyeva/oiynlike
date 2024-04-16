package controllers

import (
	"context"
	"net/http"
	"oiynlike/database"
	"oiynlike/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var anticafeCollection *mongo.Collection = database.OpenCollection("anticafe")

// admin
func CreateAnticafe() gin.HandlerFunc {
	return func(c *gin.Context) {
		var anticafe models.AnticafeModel
		if err := c.ShouldBindJSON(&anticafe); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		anticafe.CreatedAt = time.Now()
		anticafe.UpdatedAt = time.Now()

		result, err := anticafeCollection.InsertOne(context.Background(), anticafe)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating anticafe"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"msg": "Anticafe has created succesfully", "data": result.InsertedID})
	}
}

// admin

func UpdateAnticafe() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получение ID антикафе из параметра запроса
		id := c.Param("id")

		// Проверка наличия ID
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Anticafe ID is required"})
			return
		}

		// Поиск антикафе по ID в базе данных
		var existingAnticafe models.AnticafeModel
		err := anticafeCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&existingAnticafe)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Anticafe not found"})
			return
		}

		// Привязка данных запроса к структуре антикафе
		var updatedAnticafe models.AnticafeModel
		if err := c.ShouldBindJSON(&updatedAnticafe); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Обновление только тех полей, которые были переданы в запросе
		updateFields := bson.M{}
		if updatedAnticafe.Title != "" {
			updateFields["name"] = updatedAnticafe.Title
		}
		if updatedAnticafe.Rating != "" {
			updateFields["rating"] = updatedAnticafe.Rating
		}
		if updatedAnticafe.Address != "" {
			updateFields["address"] = updatedAnticafe.Address
		}
		if !updatedAnticafe.OpeningTime.IsZero() {
			updateFields["openingTime"] = updatedAnticafe.OpeningTime
		}
		if !updatedAnticafe.ClosingTime.IsZero() {
			updateFields["closingTime"] = updatedAnticafe.ClosingTime
		}
		if updatedAnticafe.PhoneNumber != "" {
			updateFields["phoneNumber"] = updatedAnticafe.PhoneNumber
		}
		if updatedAnticafe.Description != "" {
			updateFields["description"] = updatedAnticafe.Description
		}
		if len(updatedAnticafe.Photos) > 0 {
			updateFields["photos"] = updatedAnticafe.Photos
		}
		if updatedAnticafe.Latitude != "" {
			updateFields["latitude"] = updatedAnticafe.Latitude
		}
		if updatedAnticafe.Longitude != "" {
			updateFields["longitude"] = updatedAnticafe.Longitude
		}

		// Установка времени обновления
		updateFields["updatedAt"] = time.Now()

		// Выполнение частичного обновления антикафе в базе данных
		_, err = anticafeCollection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": updateFields})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating anticafe"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Anticafe updated successfully"})
	}
}

// admin

func GetAllAnticafe() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получение всех антикафе из базы данных
		cursor, err := anticafeCollection.Find(context.Background(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving anticafe"})
			return
		}
		defer cursor.Close(context.Background())

		// Преобразование результатов запроса в массив антикафе
		var anticafeList []models.AnticafeModel
		if err := cursor.All(context.Background(), &anticafeList); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding anticafe list"})
			return
		}

		// Формирование ответа в заданном формате
		response := struct {
			Items []models.AnticafeModel `json:"items"`
			Meta  struct {
				Current  int `json:"current"`
				Total    int `json:"total"`
				PageSize int `json:"page_size"`
			} `json:"meta"`
		}{
			Items: anticafeList,
			Meta: struct {
				Current  int `json:"current"`
				Total    int `json:"total"`
				PageSize int `json:"page_size"`
			}{
				Current:  1,                 // текущая страница (ваше значение)
				Total:    len(anticafeList), // общее количество элементов
				PageSize: 10,                // размер страницы (ваше значение)
			},
		}

		c.JSON(http.StatusOK, response)
	}
}

// admin

func GetAnticafeByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получение ID антикафе из параметра запроса
		id := c.Param("id")

		// Проверка наличия ID
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Anticafe ID is required"})
			return
		}

		// Поиск антикафе по ID в базе данных
		var anticafe models.AnticafeModel
		err := anticafeCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&anticafe)
		if err != nil {
			// Если антикафе с указанным ID не найдено, возвращаем ошибку 404
			c.JSON(http.StatusNotFound, gin.H{"error": "Anticafe not found"})
			return
		}

		// Отправка найденного антикафе в качестве ответа
		c.JSON(http.StatusOK, anticafe)
	}
}

// user
func GetListActiveAnticafe() {

}

// user
func GetActiveAnticafe() {

}
