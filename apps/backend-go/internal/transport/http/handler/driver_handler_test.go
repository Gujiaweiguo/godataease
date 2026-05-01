package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDriverHandler_InvalidInputsAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDriverHandler(nil)
	r := gin.New()
	RegisterDriverRoutes(r.Group("/api"), h)

	t.Run("invalid ids", func(t *testing.T) {
		assertHandlerErrorResponse(t, r, "GET", "/api/driver/get/not-a-number", "")
		assertHandlerErrorResponse(t, r, "GET", "/api/driver/listDriverJar/not-a-number", "")
	})

	t.Run("registered routes reachable", func(t *testing.T) {
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/driver/list", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/driver/list/mysql", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/driver/get/1", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/driver/listDriverJar/1", "")
	})
}
