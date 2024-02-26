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
		incomingRoutes.GET("api/gamecards", controller.GetActiveGameCards())
		incomingRoutes.GET("api/user/gamecards", controller.GetUserGameCards())
		incomingRoutes.PUT("api/join", controller.JoinGameCard())
		incomingRoutes.PATCH("/api/gameCards/:gameCardID", controller.UpdateGameCard())
	}

	// incomingRoutes.GET("api/gamecards/:id", controller.GetGameCard())

}
