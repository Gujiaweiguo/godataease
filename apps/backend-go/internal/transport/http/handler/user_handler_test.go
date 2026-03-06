package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_GetDefaultPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv(service.DefaultPasswordEnvName, "custom-default-pwd")

	h := NewUserHandler(&service.UserService{}, service.NewUserImportService(&service.UserService{}))
	r := gin.New()
	r.GET("/user/defaultPwd", h.GetDefaultPassword)

	req := httptest.NewRequest("GET", "/user/defaultPwd", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "custom-default-pwd", data["defaultPwd"])
}

func TestUserHandler_ResetPasswordCompat_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserHandler(&service.UserService{}, service.NewUserImportService(&service.UserService{}))
	r := gin.New()
	r.POST("/user/resetPwd/:uid", h.ResetPasswordCompat)

	req := httptest.NewRequest("POST", "/user/resetPwd/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestUserHandler_DownloadExcelTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserHandler(&service.UserService{}, service.NewUserImportService(&service.UserService{}))
	r := gin.New()
	r.POST("/user/excelTemplate", h.DownloadExcelTemplate)

	req := httptest.NewRequest("POST", "/user/excelTemplate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Body.Bytes())
	assert.Contains(t, w.Header().Get("Content-Type"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
}
