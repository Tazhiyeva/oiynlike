package routes

import (
	"github.com/gin-gonic/gin"

	controller "oiynlike/controllers"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("api/users/signup", controller.Signup())
	incomingRoutes.POST("api/users/login", controller.Login())
	incomingRoutes.POST("api/upload_photo", controller.UploadPhoto())
	SetupStaticRoutes(incomingRoutes)
}

func SetupStaticRoutes(router *gin.Engine) {
	router.Static("/uploads", "./uploads")
}
