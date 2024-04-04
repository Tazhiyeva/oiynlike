package middleware

import (
	"net/http"
	helper "oiynlike/helpers"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("Authorization")
		if clientToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "No Authorization header provided 2323"})
			c.Abort()
			return
		}

		if strings.HasPrefix(clientToken, "Bearer ") {
			clientToken = strings.TrimPrefix(clientToken, "Bearer ")
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("email", claims["email"])
		c.Set("first_name", claims["firstName"])
		c.Set("last_name", claims["lastName"])
		c.Set("uid", claims["uid"])
		c.Set("user_type", claims["userType"])
		c.Next()
	}
}
