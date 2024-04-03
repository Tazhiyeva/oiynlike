package routes

import (
	controller "oiynlike/controllers"
	middleware "oiynlike/middleware"

	"github.com/gin-gonic/gin"
)

func PhotoUploadRoutes(incomingRoutes *gin.Engine) {
	authMiddleware := middleware.Authenticate()

	incomingRoutes.POST("api/upload_photo", authMiddleware, controller.UploadPhoto())
	incomingRoutes.Static("/uploads", "./uploads")
}
