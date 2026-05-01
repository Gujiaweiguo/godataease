package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRoleMenuHandler_InvalidInputsAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleMenuHandler(nil)
	r := gin.New()
	RegisterRoleMenuRoutes(r.Group("/api"), h)

	t.Run("invalid request body and id", func(t *testing.T) {
		assertHandlerErrorResponse(t, r, "GET", "/api/roleMenu/auth/not-a-number", "")
		assertHandlerErrorResponse(t, r, "POST", "/api/roleMenu/auth", "{")
	})

	t.Run("registered routes reachable", func(t *testing.T) {
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/roleMenu/auth/1", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/roleMenu/auth", "{}")
	})
}
