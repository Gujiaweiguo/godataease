package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLicenseHandler_InvalidInputsAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLicenseHandler(nil)
	r := gin.New()
	RegisterLicenseRoutes(r.Group("/api"), h)

	t.Run("invalid request body", func(t *testing.T) {
		assertHandlerErrorResponse(t, r, "POST", "/api/license/validate", "{")
		assertHandlerErrorResponse(t, r, "POST", "/api/license/update", "{")
	})

	t.Run("registered routes reachable", func(t *testing.T) {
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/license/validate", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/license/update", "{}")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/license/version", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/license/revert", "")
		assertHandlerRouteReachableWithRecoveredPanic(t, r, "GET", "/api/license/expiryWarning", "")
	})
}
