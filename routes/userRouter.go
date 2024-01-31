package routes

import (
	controller "oiynlike/controllers"
	"oiynlike/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	userGroup := incomingRoutes.Group("/users")
	userGroup.Use(middleware.Authenticate())

	userGroup.GET("/", controller.GetUsers())
	userGroup.GET("/:user_id", controller.GetUser())
}
