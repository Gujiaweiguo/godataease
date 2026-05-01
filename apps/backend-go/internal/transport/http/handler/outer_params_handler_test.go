package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOuterParamsHandler_InvalidInputsAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOuterParamsHandler(nil)
	r := gin.New()
	RegisterOuterParamsRoutes(r.Group("/api"), h)

	t.Run("invalid request body", func(t *testing.T) {
		assertHandlerErrorResponse(t, r, "POST", "/api/outerParams/updateOuterParamsSet", "{")
	})

	t.Run("registered routes reachable", func(t *testing.T) {
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/outerParams/queryWithVisualizationId/1", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/outerParams/updateOuterParamsSet", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/outerParams/getOuterParamsInfo/1", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/outerParams/queryDsWithVisualizationId/1", "")
	})
}
