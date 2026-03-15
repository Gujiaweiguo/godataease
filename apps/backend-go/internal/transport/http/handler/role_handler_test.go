package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleHandlerTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}))

	repo := repository.NewRoleRepository(db)
	systemType := "system"
	now := time.Unix(300, 0)
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "Admin", RoleCode: "admin", RoleType: &systemType, Status: role.StatusEnabled, CreateTime: &now}))

	svc := service.NewRoleService(repo, nil, nil)
	r := gin.New()
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
