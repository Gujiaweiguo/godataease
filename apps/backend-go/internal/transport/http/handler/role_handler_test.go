package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleHandlerTestRouter(t *testing.T) *gin.Engine {
	return setupRoleHandlerTestRouterWithOrg(t, 0)
}

func setupRoleHandlerTestRouterWithOrg(t *testing.T, orgID uint64) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &user.SysUser{}, &user.SysUserRole{}))

	repo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	systemType := "system"
	now := time.Unix(300, 0)
	adminRole := &role.SysRole{RoleName: "Admin", RoleCode: "admin", RoleType: &systemType, Status: role.StatusEnabled, CreateTime: &now}
	require.NoError(t, repo.Create(adminRole))
	require.NoError(t, db.Create(&user.SysUserRole{UserID: 1, RoleID: adminRole.RoleID, OrgID: 1}).Error)

	svc := service.NewRoleService(repo, userRepo, userRoleRepo)
	r := gin.New()
	if orgID > 0 {
		r.Use(func(c *gin.Context) {
			c.Set("org_id", orgID)
			c.Next()
		})
	}
	RegisterRoleRoutes(r.Group("/api"), NewRoleHandler(svc))
	return r
}

func TestRoleHandler_Page_Success(t *testing.T) {
	r := setupRoleHandlerTestRouter(t)
	body := map[string]interface{}{"current": 1, "size": 10}
	buf, err := json.Marshal(body)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/page", bytes.NewBuffer(buf))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string `json:"code"`
		Data struct {
			Total int64 `json:"total"`
			List  []struct {
				Name     string  `json:"roleName"`
				RoleType *string `json:"roleType"`
			} `json:"list"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, int64(1), resp.Data.Total)
	assert.Len(t, resp.Data.List, 1)
	assert.Equal(t, "Admin", resp.Data.List[0].Name)
	require.NotNil(t, resp.Data.List[0].RoleType)
	assert.Equal(t, "system", *resp.Data.List[0].RoleType)
}

func TestRoleHandler_Page_InvalidRequest(t *testing.T) {
	r := setupRoleHandlerTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/page", bytes.NewBuffer([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRoleHandler_QueryWithOrgID_Success(t *testing.T) {
	r := setupRoleHandlerTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/role/queryWithOid/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string `json:"code"`
		Data []struct {
			Name string `json:"roleName"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "Admin", resp.Data[0].Name)
}

func TestRoleHandler_QueryWithOrgID_InvalidOID(t *testing.T) {
	r := setupRoleHandlerTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/role/queryWithOid/bad", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func TestRoleHandler_QueryByCurrentOrg_UsesContextOrg(t *testing.T) {
	r := setupRoleHandlerTestRouterWithOrg(t, 1)
	body := []byte(`{"keyword":"Admin"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/byCurOrg", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string `json:"code"`
		Data struct {
			List []struct {
				Name string `json:"roleName"`
			} `json:"list"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Len(t, resp.Data.List, 1)
	assert.Equal(t, "Admin", resp.Data.List[0].Name)
}

func TestRoleHandler_MountUser_UsesContextOrgWhenMissingInRequest(t *testing.T) {
	r := setupRoleHandlerTestRouterWithOrg(t, 3)
	body := []byte(`{"rid":1,"uids":[99]}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/mountUser", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])
}

func TestRoleHandler_UnmountUser_LastRoleReturnsDeterministicErrorEnvelope(t *testing.T) {
	r := setupRoleHandlerTestRouter(t)
	body := []byte(`{"rid":1,"uid":1}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/unMountUser", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
	assert.Equal(t, service.ErrLastRoleRemovalBlocked.Error(), resp["msg"])
}
