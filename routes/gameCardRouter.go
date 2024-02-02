package routes

import (
	controller "oiynlike/controllers"
	middleware "oiynlike/middleware"

	"github.com/gin-gonic/gin"
)

func GameCardRoutes(incomingRoutes *gin.Engine) {
	userGroup := incomingRoutes.Group("/user")
	userGroup.Use(middleware.Authenticate())

	incomingRoutes.POST("api/user/gamecards", controller.CreateGameCard())
	// incomingRoutes.PUT("api/user/gamecards/:id", controller.UpdateGameCard())

	// incomingRoutes.GET("api/gamecards", controller.GetGameCards())
	// incomingRoutes.GET("api/gamecards/:id", controller.GetGameCard())

	//incomingRoutes.PUT("api/user/gamecards/:id/like", controller.CreateMatch())

}
