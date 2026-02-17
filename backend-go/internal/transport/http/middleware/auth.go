package middleware

import (
	"strings"

	"dataease/backend/internal/pkg/auth"
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func Auth(jwtInstance *auth.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			response.Unauthorized(c, "missing authorization header")
			return
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		claims, err := jwtInstance.ParseToken(token)
		if err != nil {
			if err == auth.ErrTokenExpired {
				response.Unauthorized(c, "token has expired")
				return
			}
			response.Unauthorized(c, "invalid token")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func GetUserID(c *gin.Context) uint64 {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(uint64)
	}
	return 0
}

func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		return username.(string)
	}
	return ""
}

func GetRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return ""
}
