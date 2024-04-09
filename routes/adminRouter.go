package routes

import (
	controller "oiynlike/controllers"
	middleware "oiynlike/middleware"

	"github.com/gin-gonic/gin"
)

func AdminRoutes(incomingRoutes *gin.Engine) {
	authMiddleware := middleware.AuthAdmin()

	incomingRoutes.Use(authMiddleware)

	incomingRoutes.GET("api/admin/gamecards", controller.GetAllGameCards())
	incomingRoutes.GET("api/admin/gamecards/:gameCardID", controller.GetGameCardByID())
	incomingRoutes.POST("api/admin/gamecards/:gameCardID", controller.UpdateStatus())
}
