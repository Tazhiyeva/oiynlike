package main

import (
	"os"

	"oiynlike/database"
	routes "oiynlike/routes"

	"github.com/gin-gonic/gin"
)

// @title Oiynlike
// @version 1.0
// @description Пока что пусто
// @termsOfService Your Terms of Service
// @contact.name Madina Tazhiyeva
// @contact.email madinatazhiyeva@gmail.com
// @host localhost:8000
// @BasePath /api

func main() {

	database.ConnectToMongoDB()

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.AdminRoutes(router)
	routes.GameCardRoutes(router)

	router.GET("api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api-1"})
	})

	router.GET("api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	router.Run(":" + port)
}
