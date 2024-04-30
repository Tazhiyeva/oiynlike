package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetFilterValues() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Получаем уникальные значения city из коллекции gameCards
		cityCursor, err := gameCardCollection.Distinct(ctx, "city", bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var cities []string
		for _, city := range cityCursor {
			if cityString, ok := city.(string); ok {
				cities = append(cities, cityString)
			}
		}

		// Получаем уникальные значения category из коллекции gameCards
		categoryCursor, err := gameCardCollection.Distinct(ctx, "category", bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var categories []string
		for _, category := range categoryCursor {
			if categoryString, ok := category.(string); ok {
				categories = append(categories, categoryString)
			}
		}

		// Возвращаем успешный ответ с уникальными значениями city и category
		c.JSON(http.StatusOK, gin.H{"cities": cities, "categories": categories})
	}
}
