package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubjectHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSubjectRoutes(r, NewSubjectHandler(nil))

	t.Run("invalid_body", func(t *testing.T) {
		assertSubjectHandlerErrorResponse(t, r, "POST", "/visualizationSubject/update", "", "500000", "Invalid request")
	})

	t.Run("valid_body_nil_repo_recovered", func(t *testing.T) {
		assertSubjectHandlerErrorResponse(t, r, "POST", "/visualizationSubject/update", `{"name":"test","details":"d","coverUrl":""}`, "500000", "Service unavailable")
	})
}

func TestSubjectHandler_DeleteAndQueryRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSubjectRoutes(r, NewSubjectHandler(nil))

	tests := []struct {
		name   string
		method string
		url    string
		body   string
	}{
		{name: "delete_nil_repo_recovered", method: "POST", url: "/visualizationSubject/delete/test-id"},
		{name: "query_nil_repo_recovered", method: "POST", url: "/visualizationSubject/query"},
		{name: "query_with_group_nil_repo_recovered", method: "POST", url: "/visualizationSubject/querySubjectWithGroup"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSubjectHandlerErrorResponse(t, r, tt.method, tt.url, tt.body, "500000", "Service unavailable")
		})
	}
}

func assertSubjectHandlerErrorResponse(t *testing.T, r *gin.Engine, method, url, body, expectedCode, expectedMessage string) {
	t.Helper()

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, expectedCode, resp["code"])
	assert.Contains(t, resp["msg"], expectedMessage)
}
