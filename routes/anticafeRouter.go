package routes

import (
	controller "oiynlike/controllers"
	middleware "oiynlike/middleware"

	"github.com/gin-gonic/gin"
)

func AnticafeRoutes(incomingRoutes *gin.Engine) {
	authMiddleware := middleware.AuthAdmin()

	incomingRoutes.Use(authMiddleware)
	{
		incomingRoutes.POST("api/anticafe", controller.CreateAnticafe())
		incomingRoutes.PATCH("api/anticafe", controller.UpdateAnticafe())
		incomingRoutes.GET("api/anticafe", controller.GetAllAnticafe())
		incomingRoutes.GET("api/anticafe/:anticafeID", controller.GetAnticafeByID())
	}
}
