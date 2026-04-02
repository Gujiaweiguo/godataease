package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	domainauth "dataease/backend/internal/domain/auth"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserHandlerWithRepo(t *testing.T) (*UserHandler, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.SysUser{}, &user.SysUserRole{}))
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	userSvc := service.NewUserService(userRepo, userRoleRepo, nil)
	return NewUserHandler(userSvc, service.NewUserImportService(userSvc)), db
}

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
		buildBootstrap: func(userID int64, selectedOrgID int64, requestLanguage string) (*domainauth.IdentityBootstrap, error) {
			assert.Equal(t, int64(9), userID)
			assert.Equal(t, int64(0), selectedOrgID)
			assert.Empty(t, requestLanguage)
			return &domainauth.IdentityBootstrap{
				ID:            userID,
				Name:          "alice",
				Oid:           2,
				Language:      "tw",
				CurrentOrg:    &domainauth.OrgSummary{OrgID: 2, OrgName: "Org B"},
				AvailableOrgs: []domainauth.OrgSummary{{OrgID: 2, OrgName: "Org B"}},
			}, nil
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
	assert.Equal(t, float64(2), data["oid"])
	assert.Equal(t, "tw", data["language"])
	currentOrg, ok := data["currentOrg"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(2), currentOrg["orgId"])
}

func TestUserHandler_GetUserInfo_HeaderOverridesStoredLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &UserHandler{
		buildBootstrap: func(userID int64, selectedOrgID int64, requestLanguage string) (*domainauth.IdentityBootstrap, error) {
			assert.Equal(t, "en-US", requestLanguage)
			assert.Equal(t, int64(0), selectedOrgID)
			return &domainauth.IdentityBootstrap{ID: userID, Name: "bob", Oid: 0, Language: "en", AvailableOrgs: []domainauth.OrgSummary{}}, nil
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
		buildBootstrap: func(userID int64, selectedOrgID int64, requestLanguage string) (*domainauth.IdentityBootstrap, error) {
			assert.Equal(t, "fr-FR,en-US;q=0.8", requestLanguage)
			assert.Equal(t, int64(0), selectedOrgID)
			return &domainauth.IdentityBootstrap{ID: userID, Name: "dora", Oid: 0, Language: "en", AvailableOrgs: []domainauth.OrgSummary{}}, nil
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
		buildBootstrap: func(userID int64, selectedOrgID int64, requestLanguage string) (*domainauth.IdentityBootstrap, error) {
			assert.Equal(t, "fr-FR", requestLanguage)
			assert.Equal(t, int64(0), selectedOrgID)
			return &domainauth.IdentityBootstrap{ID: userID, Name: "carl", Oid: 0, Language: "zh-CN", AvailableOrgs: []domainauth.OrgSummary{}}, nil
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

func TestUserHandler_GetUserInfo_RejectsMissingAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &UserHandler{}
	r := gin.New()
	r.GET("/user/info", h.GetUserInfo)

	req := httptest.NewRequest("GET", "/user/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestUserHandler_SwitchOrg_ReturnsTokenVO(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &UserHandler{
		switchOrg: func(userID int64, targetOrgID int64, requestLanguage string) (*domainauth.TokenVO, error) {
			assert.Equal(t, int64(15), userID)
			assert.Equal(t, int64(3), targetOrgID)
			assert.Equal(t, "en-US", requestLanguage)
			return &domainauth.TokenVO{Token: "new-token", Exp: 123, Oid: 3, CurrentOrg: &domainauth.OrgSummary{OrgID: 3, OrgName: "Org Three"}}, nil
		},
	}
	r := gin.New()
	r.POST("/user/switch/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(15))
		h.SwitchOrg(c)
	})
	req := httptest.NewRequest("POST", "/user/switch/3", nil)
	req.Header.Set("Accept-Language", "en-US")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "new-token", data["token"])
	assert.Equal(t, float64(3), data["oid"])
}

func TestUserHandler_SwitchEnable_InvalidPayload(t *testing.T) {
	h, _ := setupUserHandlerWithRepo(t)
	r := gin.New()
	r.POST("/user/enable", h.SwitchEnable)

	req := httptest.NewRequest(http.MethodPost, "/user/enable", strings.NewReader(`{"id":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
	assert.Contains(t, resp["msg"], "id and status are required")
}

func TestUserHandler_SwitchEnable_Success(t *testing.T) {
	h, repoDB := setupUserHandlerWithRepo(t)
	r := gin.New()
	r.POST("/user/enable", func(c *gin.Context) {
		c.Set("org_id", int64(1))
		h.SwitchEnable(c)
	})

	userRepo := repository.NewUserRepository(repoDB)
	userRoleRepo := repository.NewUserRoleRepository(repoDB)
	require.NoError(t, userRepo.Create(&user.SysUser{UserID: 1, Username: "admin", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}))
	require.NoError(t, userRepo.Create(&user.SysUser{Username: "toggle-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}))

	var existing user.SysUser
	require.NoError(t, repoDB.Where("username = ?", "toggle-user").First(&existing).Error)
	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: existing.UserID, RoleID: 1, OrgID: 1}))
	req := httptest.NewRequest(http.MethodPost, "/user/enable", strings.NewReader(`{"id":`+strconv.FormatInt(existing.UserID, 10)+`,"status":0}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])

	var updated user.SysUser
	require.NoError(t, repoDB.First(&updated, "user_id = ?", existing.UserID).Error)
	assert.Equal(t, user.StatusDisabled, updated.Status)
}

func TestUserHandler_ListUsers_DefaultsToCurrentOrg(t *testing.T) {
	h, repoDB := setupUserHandlerWithRepo(t)
	r := gin.New()
	r.POST("/system/user/list", func(c *gin.Context) {
		c.Set("org_id", int64(7))
		h.ListUsers(c)
	})

	userRepo := repository.NewUserRepository(repoDB)
	userRoleRepo := repository.NewUserRoleRepository(repoDB)
	require.NoError(t, userRepo.Create(&user.SysUser{UserID: 11, Username: "org-seven-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}))
	require.NoError(t, userRepo.Create(&user.SysUser{UserID: 12, Username: "org-eight-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}))
	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 11, RoleID: 1, OrgID: 7}))
	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 12, RoleID: 1, OrgID: 8}))

	req := httptest.NewRequest(http.MethodPost, "/system/user/list", strings.NewReader(`{"current":1,"size":100}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
	data := resp["data"].(map[string]interface{})
	list := data["list"].([]interface{})
	require.Len(t, list, 1)
	item := list[0].(map[string]interface{})
	assert.Equal(t, "org-seven-user", item["username"])
}
