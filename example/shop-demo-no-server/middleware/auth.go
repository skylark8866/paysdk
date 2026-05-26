package middleware

import (
	"net/http"
	"shop-demo/service"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			writeAuthError(c, "未登录")
			c.Abort()
			return
		}

		claims, err := userService.ParseToken(token)
		if err != nil {
			writeAuthError(c, "登录已过期")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	cookie, err := c.Cookie("token")
	if err != nil {
		return ""
	}
	return cookie
}

func writeAuthError(c *gin.Context, message string) {
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/event-stream") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    401,
		"message": message,
	})
}

func GetUserID(c *gin.Context) uint64 {
	if v, exists := c.Get("user_id"); exists {
		return v.(uint64)
	}
	return 0
}

func GetUsername(c *gin.Context) string {
	if v, exists := c.Get("username"); exists {
		return v.(string)
	}
	return ""
}
