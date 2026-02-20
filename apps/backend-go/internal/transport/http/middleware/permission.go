package middleware

import (
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

const (
	RowPermissionDatasetIDKey = "row_permission_dataset_id"
	RowPermissionFilterKey    = "row_permission_filter"
)

func RowPermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Warn("RowPermissionMiddleware applied - framework in place, needs real implementation",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		c.Next()
	}
}
