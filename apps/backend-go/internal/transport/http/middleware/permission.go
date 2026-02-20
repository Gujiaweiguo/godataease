package middleware

import (
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func Permission(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		if role == "" {
			response.Unauthorized(c, "authentication required")
			return
		}

		if role != "admin" && role != requiredRole {
			response.Forbidden(c, "insufficient permissions")
			return
		}

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return Permission("admin")
}
