package middleware

import (
	"strings"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/pkg/auth"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

var auditService *service.AuditService

func SetAuditService(svc *service.AuditService) {
	auditService = svc
}

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

type AuditConfig struct {
	ActionType   audit.ActionType
	ActionName   string
	ResourceType audit.ResourceType
}

func AuditLog(cfg AuditConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		if auditService == nil {
			return
		}

		status := audit.StatusSuccess
		var failureReason *string
		if len(c.Errors) > 0 {
			status = audit.StatusFailed
			errMsg := c.Errors.String()
			failureReason = &errMsg
		}

		var userID *int64
		if uid := GetUserID(c); uid > 0 {
			id := int64(uid)
			userID = &id
		}

		var username *string
		if uname := GetUsername(c); uname != "" {
			username = &uname
		}

		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		operation := inferOperation(c.Request.Method)
		resourceName := c.Request.URL.Path

		req := &audit.AuditLogCreateRequest{
			UserID:        userID,
			Username:      username,
			ActionType:    cfg.ActionType,
			ActionName:    cfg.ActionName,
			ResourceType:  ptrString(string(cfg.ResourceType)),
			Operation:     operation,
			Status:        &status,
			FailureReason: failureReason,
			IPAddress:     &ipAddress,
			UserAgent:     &userAgent,
			ResourceName:  &resourceName,
		}

		go func() {
			_, _ = auditService.CreateAuditLog(req)
		}()

		_ = start
	}
}

func inferOperation(method string) audit.Operation {
	switch method {
	case "POST":
		return audit.OperationCreate
	case "PUT", "PATCH":
		return audit.OperationUpdate
	case "DELETE":
		return audit.OperationDelete
	case "GET":
		return audit.OperationExport
	default:
		return audit.OperationCreate
	}
}

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
