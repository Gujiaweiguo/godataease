package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMsgCenterHandler_InvalidInputsAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMsgCenterHandler(nil)
	r := gin.New()
	RegisterMsgCenterRoutes(r.Group("/api"), h)

	t.Run("invalid request body", func(t *testing.T) {
		assertHandlerErrorResponse(t, r, "POST", "/api/msg-center/count", "{")
		assertHandlerErrorResponse(t, r, "POST", "/api/msg-center/list", "{")
		assertHandlerErrorResponse(t, r, "POST", "/api/msg-center/read", "{")
		assertHandlerErrorResponse(t, r, "POST", "/api/msg-center/read/batch", "{")
	})

	t.Run("registered routes reachable", func(t *testing.T) {
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/msg-center/count", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/msg-center/list", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/msg-center/read", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/msg-center/read/batch", "{}")
	})
}
