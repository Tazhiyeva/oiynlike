package routes

import (
	controller "oiynlike/controllers"
	"oiynlike/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	authMiddleware := middleware.Authenticate()
	incomingRoutes.Use(authMiddleware)

	incomingRoutes.PATCH("api/user/profile", controller.UpdateProfile())
	incomingRoutes.GET("api/user/profile", controller.GetProfile())
	incomingRoutes.GET("api/user/chats", controller.GetUserChatsHandler())
	incomingRoutes.DELETE("api/user/leave_chat/:chat_id", controller.LeaveChatHandler())
}
