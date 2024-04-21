package routes

import (
	"github.com/gin-gonic/gin"

	controller "oiynlike/controllers"
)

func AuthRoutes(r *gin.Engine) {
	r.POST("api/users/signup", controller.Signup())
	r.POST("api/users/login", controller.Login())
	r.POST("api/upload_photo", controller.UploadPhoto())
	SetupStaticRoutes(r)
}

func SetupStaticRoutes(router *gin.Engine) {
	router.Static("/uploads", "./uploads")
}
