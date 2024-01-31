package middleware

import (
	"net/http"
	helper "oiynlike/helpers"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access Forbidden"})
			c.Abort()
			return
		}

		c.Set("email", claims["email"])
		c.Set("fisrt_name", claims["firstName"])
		c.Set("last_name", claims["lastName"])
		c.Set("uid", claims["uid"])
		c.Set("user_type", claims["userType"])
		c.Next()

	}
}
