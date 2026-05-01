package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTicketHandler_InvalidInputsAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTicketHandler(nil)
	r := gin.New()
	RegisterTicketRoutes(r.Group("/api"), h)

	t.Run("invalid request body", func(t *testing.T) {
		assertHandlerErrorResponse(t, r, "POST", "/api/ticket/create", "{")
	})

	t.Run("registered routes reachable", func(t *testing.T) {
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/ticket/create", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/ticket/validate/test-ticket", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "DELETE", "/api/ticket/delete/test-ticket", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/ticket/list/test-uuid?page=bad&pageSize=bad", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/ticket/temp", "")
	})
}
