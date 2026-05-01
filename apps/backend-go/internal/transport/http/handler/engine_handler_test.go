package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEngineHandler_InvalidInputsAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewEngineHandler(nil)
	r := gin.New()
	RegisterEngineRoutes(r.Group("/api"), h)

	t.Run("invalid request body and id", func(t *testing.T) {
		assertHandlerErrorResponse(t, r, "POST", "/api/engine/validate", "{")
		assertHandlerErrorResponse(t, r, "POST", "/api/engine/validate/not-a-number", "")
	})

	t.Run("registered routes reachable", func(t *testing.T) {
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/engine/getEngine", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/engine/validate", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/engine/validate/1", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/engine/supportSetKey", "")
	})
}
