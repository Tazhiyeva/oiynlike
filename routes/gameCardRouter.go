package routes

import (
	controller "oiynlike/controllers"
	middleware "oiynlike/middleware"

	"github.com/gin-gonic/gin"
)

func GameCardRoutes(incomingRoutes *gin.Engine) {
	authMiddleware := middleware.Authenticate()

	incomingRoutes.Use(authMiddleware)
	{
		// Все роуты, добавленные здесь, будут использовать authMiddleware
		incomingRoutes.POST("api/gamecards", controller.CreateGameCard())

	}

	// incomingRoutes.PUT("api/user/gamecards/:id", controller.UpdateGameCard())

	// incomingRoutes.GET("api/gamecards", controller.GetGameCards())
	// incomingRoutes.GET("api/gamecards/:id", controller.GetGameCard())

	//incomingRoutes.PUT("api/user/gamecards/:id/like", controller.CreateMatch())

}
