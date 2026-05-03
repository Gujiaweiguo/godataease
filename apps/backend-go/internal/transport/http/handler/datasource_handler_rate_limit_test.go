package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDatasourceRoutes_RateLimitsValidateRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(service.NewDatasourceService(nil))
	r := gin.New()
	RegisterDatasourceRoutes(r.Group("/api"), h, nil, nil)

	for i := 0; i < datasourceValidateRateLimitRequests; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/ds/validate", strings.NewReader("{"))
		req.RemoteAddr = "203.0.113.15:9999"
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), `"code":"500000"`)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/ds/validate/not-a-number", nil)
	req.RemoteAddr = "203.0.113.15:9999"
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	var body bridgeCodeResp
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	assert.Equal(t, "429001", body.Code)
}
