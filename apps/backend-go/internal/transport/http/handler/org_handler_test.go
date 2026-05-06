package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/governance"
	"dataease/backend/internal/domain/org"
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

func setupOrgHandlerTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&org.SysOrg{}, &audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, &role.SysRole{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))

	orgRepo := repository.NewOrgRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	loginFailureRepo := repository.NewLoginFailureRepository(db)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(db)
	auditSvc := service.NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)
	orgSvc := service.NewOrgService(orgRepo, auditSvc, repository.NewUserRepository(db), repository.NewRoleRepository(db), repository.NewUserRoleRepository(db), service.NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), auditSvc))

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("org_id", uint64(1))
		c.Next()
	})
	RegisterOrgRoutes(r.Group("/api"), NewOrgHandler(orgSvc))
	return r
}

func setupOrgHandlerTestRouterWithAudit(t *testing.T) (*gin.Engine, *repository.AuditLogRepository) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&org.SysOrg{}, &audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, &role.SysRole{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))

	orgRepo := repository.NewOrgRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	loginFailureRepo := repository.NewLoginFailureRepository(db)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(db)
	auditSvc := service.NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)
	orgSvc := service.NewOrgService(orgRepo, auditSvc, repository.NewUserRepository(db), repository.NewRoleRepository(db), repository.NewUserRoleRepository(db), service.NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), auditSvc))

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(77))
		c.Set("username", "tester")
		c.Set("org_id", uint64(1))
		c.Next()
	})
	RegisterOrgRoutes(r.Group("/api"), NewOrgHandler(orgSvc))
	return r, auditLogRepo
}

func TestOrgHandler_CreateOrg_RejectsMissingOrgContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&org.SysOrg{}, &audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, &role.SysRole{}, &user.SysUser{}, &user.SysUserRole{}, &governance.SysGovernancePolicy{}))
	auditSvc := service.NewAuditService(repository.NewAuditLogRepository(db), repository.NewLoginFailureRepository(db), repository.NewAuditLogDetailRepository(db))
	orgSvc := service.NewOrgService(repository.NewOrgRepository(db), auditSvc, repository.NewUserRepository(db), repository.NewRoleRepository(db), repository.NewUserRoleRepository(db), service.NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), auditSvc))
	r := gin.New()
	RegisterOrgRoutes(r.Group("/api"), NewOrgHandler(orgSvc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/organization/create", strings.NewReader(`{"orgName":"NoOrg"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
	assert.Equal(t, "Invalid org context", resp["msg"])
}

func TestOrgHandler_DeleteOrg_WithChildrenReturnsDeterministicErrorEnvelope(t *testing.T) {
	r := setupOrgHandlerTestRouter(t)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	repo := repository.NewOrgRepository(db)
	parent := &org.SysOrg{OrgName: "Parent", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	require.NoError(t, repo.Create(parent))
	child := &org.SysOrg{OrgName: "Child", ParentID: parent.OrgID, Level: 2, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	require.NoError(t, repo.Create(child))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/system/organization/delete/%d", parent.OrgID), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
	msg, _ := resp["msg"].(string)
	assert.Contains(t, msg, "child organizations")
}

func TestOrgHandler_DeleteOrg_UsesUserIDContextForAudit(t *testing.T) {
	r, auditLogRepo := setupOrgHandlerTestRouterWithAudit(t)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	repo := repository.NewOrgRepository(db)
	leaf := &org.SysOrg{OrgName: "Leaf", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	require.NoError(t, repo.Create(leaf))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/system/organization/delete/%d", leaf.OrgID), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])

	logs, total, err := auditLogRepo.Query(&audit.AuditLogQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, logs, 1)
	require.NotNil(t, logs[0].UserID)
	assert.Equal(t, int64(77), *logs[0].UserID)
	require.NotNil(t, logs[0].Username)
	assert.Equal(t, "tester", *logs[0].Username)
}

func TestOrgHandler_TransferUser_Success(t *testing.T) {
	r := setupOrgHandlerTestRouter(t)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Create(&org.SysOrg{OrgID: 1, OrgName: "Source", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}).Error)
	require.NoError(t, db.Create(&org.SysOrg{OrgID: 2, OrgName: "Target", ParentID: org.RootParentID, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}).Error)
	require.NoError(t, repository.NewGovernancePolicyRepository(db).SetLastRolePolicy(1, governance.LastRolePolicyWarnAllow, "tester"))
	require.NoError(t, db.Create(&role.SysRole{RoleID: 1, RoleName: "Source Role", RoleCode: "source-handler-role", Status: role.StatusEnabled}).Error)
	require.NoError(t, db.Create(&user.SysUser{UserID: 3, Username: "handler-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, db.Create(&user.SysUserRole{UserID: 3, RoleID: 1, OrgID: 1}).Error)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/organization/transfer-user", strings.NewReader(`{"sourceOrgId":1,"targetOrgId":2,"userId":3}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp["code"])

	var sourceCount int64
	require.NoError(t, db.Model(&user.SysUserRole{}).Where("user_id = ? AND org_id = ?", 3, 1).Count(&sourceCount).Error)
	assert.Equal(t, int64(0), sourceCount)

	defaultRole, err := repository.NewRoleRepository(db).GetByRoleCode(role.BuiltInOrgUserRoleCode)
	require.NoError(t, err)
	var targetCount int64
	require.NoError(t, db.Model(&user.SysUserRole{}).Where("user_id = ? AND org_id = ? AND role_id = ?", 3, 2, defaultRole.RoleID).Count(&targetCount).Error)
	assert.Equal(t, int64(1), targetCount)
}
