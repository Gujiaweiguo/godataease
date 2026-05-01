package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPermHandler_InvalidInputsAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermHandler(nil)
	r := gin.New()
	RegisterPermRoutes(r.Group("/api"), h)

	t.Run("invalid request body and id", func(t *testing.T) {
		assertHandlerErrorResponse(t, r, "POST", "/api/system/permission/create", "{")
		assertHandlerErrorResponse(t, r, "POST", "/api/system/permission/update", "{")
		assertHandlerErrorResponse(t, r, "POST", "/api/system/permission/delete/not-a-number", "")
	})

	t.Run("registered routes reachable", func(t *testing.T) {
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/system/permission/list", "{")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/system/permission/create", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/system/permission/update", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/system/permission/delete/1", "")
	})
}
