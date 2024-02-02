package routes

import (
	controller "oiynlike/controllers"
	"oiynlike/middleware"

	"github.com/gin-gonic/gin"
)

func AdminRoutes(incomingRoutes *gin.Engine) {
	userGroup := incomingRoutes.Group("/admin")
	userGroup.Use(middleware.Authenticate())

	//admin
	userGroup.GET("api/admin/users", controller.GetUsers())
	userGroup.GET("api/admin/users/:user_id", controller.GetUser())

}
