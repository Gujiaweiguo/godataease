package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/user"
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

func TestUserHandler_GetUserInfo_UsesNormalizedLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &UserHandler{
		loadUserByID: func(userID int64) (*user.SysUser, error) {
			lang := "zh-TW"
			return &user.SysUser{UserID: userID, Username: "alice", Language: &lang}, nil
		},
	}
	r := gin.New()
	r.GET("/user/info", func(c *gin.Context) {
		c.Set("user_id", uint64(9))
		c.Set("username", "context-name")
		h.GetUserInfo(c)
	})

	req := httptest.NewRequest("GET", "/user/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(9), data["id"])
	assert.Equal(t, "alice", data["name"])
	assert.Equal(t, "tw", data["language"])
}

func TestUserHandler_GetUserInfo_HeaderOverridesStoredLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &UserHandler{
		loadUserByID: func(userID int64) (*user.SysUser, error) {
			lang := "zh-CN"
			return &user.SysUser{UserID: userID, Username: "bob", Language: &lang}, nil
		},
	}
	r := gin.New()
	r.GET("/user/info", func(c *gin.Context) {
		c.Set("user_id", uint64(10))
		h.GetUserInfo(c)
	})

	req := httptest.NewRequest("GET", "/user/info", nil)
	req.Header.Set("Accept-Language", "en-US")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "en", data["language"])
}

func TestUserHandler_GetUserInfo_UsesFirstSupportedLocaleFromHeaderList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &UserHandler{
		loadUserByID: func(userID int64) (*user.SysUser, error) {
			lang := "zh-CN"
			return &user.SysUser{UserID: userID, Username: "dora", Language: &lang}, nil
		},
	}
	r := gin.New()
	r.GET("/user/info", func(c *gin.Context) {
		c.Set("user_id", uint64(12))
		h.GetUserInfo(c)
	})

	req := httptest.NewRequest("GET", "/user/info", nil)
	req.Header.Set("Accept-Language", "fr-FR,en-US;q=0.8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "en", data["language"])
}

func TestUserHandler_GetUserInfo_UnsupportedInputFallsBackToDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &UserHandler{
		loadUserByID: func(userID int64) (*user.SysUser, error) {
			lang := "de-DE"
			return &user.SysUser{UserID: userID, Username: "carl", Language: &lang}, nil
		},
	}
	r := gin.New()
	r.GET("/user/info", func(c *gin.Context) {
		c.Set("user_id", uint64(11))
		h.GetUserInfo(c)
	})

	req := httptest.NewRequest("GET", "/user/info", nil)
	req.Header.Set("Accept-Language", "fr-FR")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "zh-CN", data["language"])
}
