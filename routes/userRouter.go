package routes

import (
	controller "oiynlike/controllers"
	"oiynlike/middleware"

	"github.com/gin-gonic/gin"
)

// func AdminRoutes(incomingRoutes *gin.Engine) {
// 	adminGroup := incomingRoutes.Group("/admin")
// 	adminGroup.Use(middleware.Authenticate())

// 	//admin
// 	adminGroup.GET("api/admin/users", controller.GetUsers())
// 	adminGroup.GET("api/admin/users/:user_id", controller.GetUser())

// }

func UserRoutes(incomingRoutes *gin.Engine) {
	authMiddleware := middleware.Authenticate()
	incomingRoutes.Use(authMiddleware)

	incomingRoutes.PATCH("api/profile", controller.UpdateProfile())
	incomingRoutes.GET("api/profile", controller.GetProfile())

}
