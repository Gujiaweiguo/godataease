package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================================
// SystemParamHandler tests (10 functions < 70%)
// =====================================================================

// --- QueryBasic ---

func TestRound9B_SystemParam_QueryBasic_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.QueryBasic(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveBasic ---

func TestRound9B_SystemParam_SaveBasic_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveBasic(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_SystemParam_SaveBasic_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `[]`)
	h.SaveBasic(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- QueryOnlineMap ---

func TestRound9B_SystemParam_QueryOnlineMap_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.QueryOnlineMap(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- QueryOnlineMapByType ---

func TestRound9B_SystemParam_QueryOnlineMapByType_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "type", Value: "gaode"}}
	h.QueryOnlineMapByType(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveOnlineMap ---

func TestRound9B_SystemParam_SaveOnlineMap_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveOnlineMap(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_SystemParam_SaveOnlineMap_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"key":"test"}`)
	h.SaveOnlineMap(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- QuerySQLBot ---

func TestRound9B_SystemParam_QuerySQLBot_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.QuerySQLBot(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveSQLBot ---

func TestRound9B_SystemParam_SaveSQLBot_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveSQLBot(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_SystemParam_SaveSQLBot_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"enabled":true}`)
	h.SaveSQLBot(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ShareBase ---

func TestRound9B_SystemParam_ShareBase_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ShareBase(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- RequestTimeOut ---

func TestRound9B_SystemParam_RequestTimeOut_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.RequestTimeOut(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DefaultSettings ---

func TestRound9B_SystemParam_DefaultSettings_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.DefaultSettings(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- UI ---

func TestRound9B_SystemParam_UI_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.UI(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DefaultLogin ---

func TestRound9B_SystemParam_DefaultLogin_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.DefaultLogin(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- I18nOptions ---

func TestRound9B_SystemParam_I18nOptions_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.I18nOptions(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- RegisterRoutes ---

func TestRound9B_SystemParam_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSystemParamHandler(nil)
	RegisterSystemParamRoutes(r, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sysParameter/basic/query", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Constructor ---

func TestRound9B_SystemParam_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

// =====================================================================
// PermissionCompatHandler tests (9 functions < 70%)
// =====================================================================

// --- MenuPermission ---

func TestRound9B_PermCompat_MenuPermission_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.MenuPermission(c)
	// nil menuService → panic → recovered → CodeInternalError
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveMenuPer ---

func TestRound9B_PermCompat_SaveMenuPer_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveMenuPer(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_PermCompat_SaveMenuPer_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"roleId":1,"menuIds":[1,2]}`)
	h.SaveMenuPer(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- BusiPermission ---

func TestRound9B_PermCompat_BusiPermission_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.BusiPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveBusiPer ---

func TestRound9B_PermCompat_SaveBusiPer_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.SaveBusiPer(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_PermCompat_SaveBusiPer_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"roleId":1,"permIds":[1,2]}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.SaveBusiPer(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- BusiResource ---

func TestRound9B_PermCompat_BusiResource_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	c.Params = gin.Params{{Key: "flag", Value: "all"}}
	h.BusiResource(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- MenuTargetPermission ---

func TestRound9B_PermCompat_MenuTargetPermission_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.MenuTargetPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_PermCompat_MenuTargetPermission_ZeroRoleID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"roleId":0}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.MenuTargetPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- BusiTargetPermission ---

func TestRound9B_PermCompat_BusiTargetPermission_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.BusiTargetPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_PermCompat_BusiTargetPermission_MissingFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"type":1,"flag":"","roleId":0}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.BusiTargetPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- UserPerspective ---

func TestRound9B_PermCompat_UserPerspective_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.UserPerspective(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_PermCompat_UserPerspective_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"userId":1}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.UserPerspective(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveMenuTargetPer ---

func TestRound9B_PermCompat_SaveMenuTargetPer_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveMenuTargetPer(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_PermCompat_SaveMenuTargetPer_NilRoleMenuService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"roleId":1}`)
	h.SaveMenuTargetPer(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveBusiTargetPer ---

func TestRound9B_PermCompat_SaveBusiTargetPer_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.SaveBusiTargetPer(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_PermCompat_SaveBusiTargetPer_MissingFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"flag":"","roleId":0}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.SaveBusiTargetPer(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Constructor & route registration ---

func TestRound9B_PermCompat_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	assert.NotNil(t, h)
}

func TestRound9B_PermCompat_RegisterRoutes_NilHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api")
	RegisterPermissionCompatRoutes(group, nil)
	// Should not panic
}

func TestRound9B_PermCompat_SaveRolePermission_Delegates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.SaveRolePermission(c)
	// Delegates to SaveBusiPer which fails on bad JSON
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// PermissionCompatHandler target helpers tests (6 functions < 70%)
// =====================================================================

// --- extractTargetPermIDs ---

func TestRound9B_PermCompat_ExtractTargetPermIDs_WithPermIDs(t *testing.T) {
	target := targetPermissionTarget{
		PermIDs: []int64{1, 2, 3},
	}
	result := extractTargetPermIDs(target)
	assert.Equal(t, []int64{1, 2, 3}, result)
}

func TestRound9B_PermCompat_ExtractTargetPermIDs_FallbackPermissions(t *testing.T) {
	target := targetPermissionTarget{
		Permissions: []targetPermissionEntry{
			{ID: 10},
			{ID: 20},
		},
	}
	result := extractTargetPermIDs(target)
	assert.Equal(t, []int64{10, 20}, result)
}

func TestRound9B_PermCompat_ExtractTargetPermIDs_Empty(t *testing.T) {
	target := targetPermissionTarget{}
	result := extractTargetPermIDs(target)
	assert.Equal(t, []int64(nil), result)
}

// --- uniqueInt64 ---

func TestRound9B_PermCompat_UniqueInt64_Dedup(t *testing.T) {
	result := uniqueInt64([]int64{3, 1, 2, 3, 1, 4})
	assert.Equal(t, []int64{3, 1, 2, 4}, result)
}

func TestRound9B_PermCompat_UniqueInt64_IgnoresZeroAndNegative(t *testing.T) {
	result := uniqueInt64([]int64{0, -1, 2, 2, 0})
	assert.Equal(t, []int64{2}, result)
}

func TestRound9B_PermCompat_UniqueInt64_Empty(t *testing.T) {
	result := uniqueInt64([]int64{})
	assert.Equal(t, []int64{}, result)
}

// --- normalizeTargetType ---

func TestRound9B_PermCompat_NormalizeTargetType_PreferTarget(t *testing.T) {
	assert.Equal(t, "role", normalizeTargetType("role", "user"))
}

func TestRound9B_PermCompat_NormalizeTargetType_FallbackSource(t *testing.T) {
	assert.Equal(t, "user", normalizeTargetType("", "user"))
}

// --- normalizeTargetID ---

func TestRound9B_PermCompat_NormalizeTargetID_PreferTarget(t *testing.T) {
	assert.Equal(t, int64(5), normalizeTargetID(5, 10))
}

func TestRound9B_PermCompat_NormalizeTargetID_FallbackSource(t *testing.T) {
	assert.Equal(t, int64(10), normalizeTargetID(0, 10))
}

// --- permissionMatchesResourceType (nil service) ---

func TestRound9B_PermCompat_PermissionMatchesResourceType_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	_, err := h.permissionMatchesResourceType(1, "chart")
	assert.Error(t, err)
}

// =====================================================================
// AuditHandler tests (8 functions < 70%)
// =====================================================================

// --- GetAuditAlertSettings ---

func TestRound9B_Audit_GetAuditAlertSettings_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.GetAuditAlertSettings(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveAuditAlertSettings ---

func TestRound9B_Audit_SaveAuditAlertSettings_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPut, "/", "not-json")
	h.SaveAuditAlertSettings(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- CleanupNow ---

func TestRound9B_Audit_CleanupNow_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.CleanupNow(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- TestNotification ---

func TestRound9B_Audit_TestNotification_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.TestNotification(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- CreateAuditLog ---

func TestRound9B_Audit_CreateAuditLog_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.CreateAuditLog(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Audit_CreateAuditLog_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"actionType":"login","resourceType":"user"}`)
	h.CreateAuditLog(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- GetAuditLogByID ---

func TestRound9B_Audit_GetAuditLogByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.GetAuditLogByID(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Audit_GetAuditLogByID_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetAuditLogByID(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- GetAuditLogsByUserID ---

func TestRound9B_Audit_GetAuditLogsByUserID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "userId", Value: "bad"}}
	h.GetAuditLogsByUserID(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Audit_GetAuditLogsByUserID_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "userId", Value: "1"}}
	h.GetAuditLogsByUserID(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- RecordLoginFailure ---

func TestRound9B_Audit_RecordLoginFailure_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.RecordLoginFailure(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Audit_RecordLoginFailure_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"username":"test","loginIp":"127.0.0.1"}`)
	h.RecordLoginFailure(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ExportAuditLogs ---

func TestRound9B_Audit_ExportAuditLogs_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.ExportAuditLogs(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Audit_ExportAuditLogs_EmptyIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"ids":[]}`)
	h.ExportAuditLogs(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Audit_ExportAuditLogs_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"ids":[1,2],"format":"csv"}`)
	h.ExportAuditLogs(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DeleteAuditLogsRetention ---

func TestRound9B_Audit_DeleteAuditLogsRetention_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodDelete, "/", `{"days":30}`)
	h.DeleteAuditLogsRetention(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DownloadExportFile ---

func TestRound9B_Audit_DownloadExportFile_MissingPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.DownloadExportFile(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Audit_DownloadExportFile_InvalidAbsPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Request.URL.RawQuery = "path=/etc/passwd"
	c.Request.URL.Query()
	// Recreate request with query
	req := httptest.NewRequest(http.MethodGet, "/?path=/etc/passwd", nil)
	c.Request = req
	h.DownloadExportFile(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- QueryAuditLogs ---

func TestRound9B_Audit_QueryAuditLogs_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.QueryAuditLogs(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Constructor ---

func TestRound9B_Audit_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.auditService)
	assert.Nil(t, h.systemParamService)
}

// =====================================================================
// ThresholdHandler tests (7 functions < 70%)
// =====================================================================

// --- Save ---

func TestRound9B_Threshold_Save_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Save(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_Save_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"test"}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.Save(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Edit ---

func TestRound9B_Threshold_Edit_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Edit(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_Edit_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"test"}`)
	h.Edit(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Pager ---

func TestRound9B_Threshold_Pager_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "goPage", Value: "0"}, {Key: "pageSize", Value: "10"}}
	h.Pager(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_Pager_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.Pager(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- FormInfo ---

func TestRound9B_Threshold_FormInfo_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}, {Key: "resourceTable", Value: "chart"}}
	h.FormInfo(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_FormInfo_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "resourceTable", Value: "chart"}}
	h.FormInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SwitchEnable ---

func TestRound9B_Threshold_SwitchEnable_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SwitchEnable(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_SwitchEnable_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"enable":true}`)
	h.SwitchEnable(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Delete ---

func TestRound9B_Threshold_Delete_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "resourceTable", Value: "chart"}}
	h.Delete(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_Delete_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `[1,2,3]`)
	c.Params = gin.Params{{Key: "resourceTable", Value: "chart"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- BatchReci ---

func TestRound9B_Threshold_BatchReci_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.BatchReci(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_BatchReci_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"ids":[1]}`)
	h.BatchReci(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- InstancePager ---

func TestRound9B_Threshold_InstancePager_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "goPage", Value: "0"}, {Key: "pageSize", Value: "10"}}
	h.InstancePager(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_InstancePager_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"thresholdId":1}`)
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.InstancePager(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Preview ---

func TestRound9B_Threshold_Preview_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Preview(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_Preview_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"thresholdId":1}`)
	h.Preview(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- AnyThreshold ---

func TestRound9B_Threshold_AnyThreshold_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "chartId", Value: "bad"}, {Key: "resourceTable", Value: "chart"}}
	h.AnyThreshold(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_AnyThreshold_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "chartId", Value: "1"}, {Key: "resourceTable", Value: "chart"}}
	h.AnyThreshold(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DeleteWithChart ---

func TestRound9B_Threshold_DeleteWithChart_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "chartId", Value: "bad"}, {Key: "resourceTable", Value: "chart"}}
	h.DeleteWithChart(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Threshold_DeleteWithChart_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "chartId", Value: "1"}, {Key: "resourceTable", Value: "chart"}}
	h.DeleteWithChart(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Constructor ---

func TestRound9B_Threshold_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.thresholdService)
}

// =====================================================================
// LinkJumpHandler tests (7 functions < 70%)
// =====================================================================

// --- GetTableFieldWithViewID ---

func TestRound9B_LinkJump_GetTableFieldWithViewID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "viewId", Value: "bad"}}
	h.GetTableFieldWithViewID(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_LinkJump_GetTableFieldWithViewID_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "viewId", Value: "1"}}
	h.GetTableFieldWithViewID(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- QueryWithViewId ---

func TestRound9B_LinkJump_QueryWithViewId_InvalidDvID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "bad"}, {Key: "viewId", Value: "1"}}
	h.QueryWithViewId(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_LinkJump_QueryWithViewId_InvalidViewID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "1"}, {Key: "viewId", Value: "bad"}}
	h.QueryWithViewId(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_LinkJump_QueryWithViewId_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "1"}, {Key: "viewId", Value: "1"}}
	h.QueryWithViewId(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- UpdateJumpSet ---

func TestRound9B_LinkJump_UpdateJumpSet_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UpdateJumpSet(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_LinkJump_UpdateJumpSet_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"sourceViewId":"1"}`)
	h.UpdateJumpSet(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- QueryTargetVisualizationJumpInfo ---

func TestRound9B_LinkJump_QueryTargetVisualizationJumpInfo_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.QueryTargetVisualizationJumpInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_LinkJump_QueryTargetVisualizationJumpInfo_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"sourceDvId":"1","sourceViewId":"1"}`)
	h.QueryTargetVisualizationJumpInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- QueryVisualizationJumpInfo ---

func TestRound9B_LinkJump_QueryVisualizationJumpInfo_InvalidDvID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "bad"}, {Key: "resourceTable", Value: "chart"}}
	h.QueryVisualizationJumpInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_LinkJump_QueryVisualizationJumpInfo_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "1"}, {Key: "resourceTable", Value: "chart"}}
	h.QueryVisualizationJumpInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ViewTableDetailList ---

func TestRound9B_LinkJump_ViewTableDetailList_InvalidDvID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "bad"}}
	h.ViewTableDetailList(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_LinkJump_ViewTableDetailList_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "1"}}
	h.ViewTableDetailList(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- UpdateJumpSetActive ---

func TestRound9B_LinkJump_UpdateJumpSetActive_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UpdateJumpSetActive(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_LinkJump_UpdateJumpSetActive_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"sourceDvId":"1","sourceViewId":"1"}`)
	h.UpdateJumpSetActive(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- RemoveJumpSet ---

func TestRound9B_LinkJump_RemoveJumpSet_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.RemoveJumpSet(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_LinkJump_RemoveJumpSet_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"sourceViewId":"1"}`)
	h.RemoveJumpSet(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Constructor ---

func TestRound9B_LinkJump_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkJumpHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

// =====================================================================
// CustomGeoHandler tests (7 functions < 70%)
// =====================================================================

// --- ListGeoAreas ---

func TestRound9B_CustomGeo_ListGeoAreas_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListGeoAreas(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetGeoArea ---

func TestRound9B_CustomGeo_GetGeoArea_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "area1"}}
	h.GetGeoArea(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DeleteGeoArea ---

func TestRound9B_CustomGeo_DeleteGeoArea_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodDelete, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "area1"}}
	h.DeleteGeoArea(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveGeoArea ---

func TestRound9B_CustomGeo_SaveGeoArea_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveGeoArea(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_CustomGeo_SaveGeoArea_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":"area1","name":"Test Area"}`)
	h.SaveGeoArea(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DeleteGeoSubArea ---

func TestRound9B_CustomGeo_DeleteGeoSubArea_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodDelete, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.DeleteGeoSubArea(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_CustomGeo_DeleteGeoSubArea_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodDelete, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.DeleteGeoSubArea(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveGeoSubArea ---

func TestRound9B_CustomGeo_SaveGeoSubArea_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveGeoSubArea(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_CustomGeo_SaveGeoSubArea_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"sub","scope":"scope","geoAreaId":"area1"}`)
	h.SaveGeoSubArea(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListSubAreaOptions ---

func TestRound9B_CustomGeo_ListSubAreaOptions_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListSubAreaOptions(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Constructor ---

func TestRound9B_CustomGeo_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.repo)
}

// =====================================================================
// StaticHandler tests (6 functions < 70%)
// =====================================================================

// --- ListResources ---

func TestRound9B_Static_ListResources_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListResources(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetResource ---

func TestRound9B_Static_GetResource_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "res1"}}
	h.GetResource(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListStores ---

func TestRound9B_Static_ListStores_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListStores(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListTypefaces ---

func TestRound9B_Static_ListTypefaces_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListTypefaces(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListFont ---

func TestRound9B_Static_ListFont_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListFont(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DefaultFont ---

func TestRound9B_Static_DefaultFont_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.DefaultFont(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- XpackModel ---

func TestRound9B_Static_XpackModel_ReturnsFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.XpackModel(c)
	assertCode(t, w, response.CodeSuccess)
	resp := parseRound9Resp(t, w)
	assert.Equal(t, false, resp["data"])
}

// --- Upload ---

func TestRound9B_Static_Upload_EmptyFileID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "fileId", Value: ""}}
	h.Upload(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_Static_Upload_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "fileId", Value: "test-file"}}
	c.Request.Header.Del("Content-Type")
	h.Upload(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- FindResourceAsBase64 ---

func TestRound9B_Static_FindResourceAsBase64_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.FindResourceAsBase64(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_Static_FindResourceAsBase64_TooManyFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	// Build a request with > 200 paths
	paths := make([]string, 201)
	for i := range paths {
		paths[i] = "file.txt"
	}
	body, _ := json.Marshal(map[string][]string{"resourcePathList": paths})
	w, c := newRound9Ctx(t, http.MethodPost, "/", string(body))
	h.FindResourceAsBase64(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_Static_FindResourceAsBase64_NonexistentFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"resourcePathList":["nonexistent.png"]}`)
	h.FindResourceAsBase64(c)
	assertCode(t, w, response.CodeSuccess)
	resp := parseRound9Resp(t, w)
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.Equal(t, "", data["nonexistent.png"])
}

// --- resolveSafeStaticUploadPath ---

func TestRound9B_Static_ResolveSafePath_EmptyID(t *testing.T) {
	path, ok := resolveSafeStaticUploadPath("/tmp/static", "", ".png")
	assert.False(t, ok)
	assert.Empty(t, path)
}

func TestRound9B_Static_ResolveSafePath_PathTraversal(t *testing.T) {
	path, ok := resolveSafeStaticUploadPath("/tmp/static", "../etc/passwd", ".png")
	assert.False(t, ok)
	assert.Empty(t, path)
}

func TestRound9B_Static_ResolveSafePath_Valid(t *testing.T) {
	path, ok := resolveSafeStaticUploadPath("/tmp/static", "abc123", ".png")
	assert.True(t, ok)
	assert.Equal(t, "/tmp/static/abc123.png", path)
}

// --- Constructor ---

func TestRound9B_Static_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

// =====================================================================
// FontHandler tests (additional low-coverage: Download at 11.6%)
// =====================================================================

// --- Download ---

func TestRound9B_Font_Download_EmptyFileName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "file", Value: ""}}
	h.Download(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_Font_Download_PathTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "file", Value: "../etc/passwd"}}
	h.Download(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_Font_Download_DisallowedExt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "file", Value: "test.exe"}}
	h.Download(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_Font_Download_FileNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "file", Value: "nonexistent.ttf"}}
	h.Download(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- List ---

func TestRound9B_Font_List_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.List(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Create ---

func TestRound9B_Font_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Create(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_Font_Create_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"TestFont","fileName":"test.ttf"}`)
	h.Create(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Edit ---

func TestRound9B_Font_Edit_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Edit(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_Font_Edit_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"TestFont"}`)
	h.Edit(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Delete ---

func TestRound9B_Font_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9B_Font_Delete_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- UploadFile ---

func TestRound9B_Font_UploadFile_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Request.Header.Del("Content-Type")
	h.UploadFile(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DefaultFont ---

func TestRound9B_Font_DefaultFont_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.DefaultFont(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Constructor ---

func TestRound9B_Font_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.repo)
}

// --- resolveSafeFontFilePath ---

func TestRound9B_Font_ResolveSafePath_EmptyName(t *testing.T) {
	path, ok := resolveSafeFontFilePath("/tmp/font", "")
	assert.False(t, ok)
	assert.Empty(t, path)
}

func TestRound9B_Font_ResolveSafePath_PathTraversal(t *testing.T) {
	path, ok := resolveSafeFontFilePath("/tmp/font", "../etc/passwd")
	assert.False(t, ok)
	assert.Empty(t, path)
}

func TestRound9B_Font_ResolveSafePath_DisallowedExt(t *testing.T) {
	path, ok := resolveSafeFontFilePath("/tmp/font", "test.exe")
	assert.False(t, ok)
	assert.Empty(t, path)
}

func TestRound9B_Font_ResolveSafePath_Valid(t *testing.T) {
	path, ok := resolveSafeFontFilePath("/tmp/font", "test.ttf")
	assert.True(t, ok)
	assert.Equal(t, "/tmp/font/test.ttf", path)
}

// --- isAllowedFontDownloadExtension ---

func TestRound9B_Font_IsAllowedExt(t *testing.T) {
	assert.True(t, isAllowedFontDownloadExtension("test.ttf"))
	assert.True(t, isAllowedFontDownloadExtension("test.otf"))
	assert.True(t, isAllowedFontDownloadExtension("test.woff"))
	assert.True(t, isAllowedFontDownloadExtension("test.woff2"))
	assert.False(t, isAllowedFontDownloadExtension("test.exe"))
	assert.False(t, isAllowedFontDownloadExtension("test.png"))
}

// --- fontToDTO ---

func TestRound9B_Font_FontToDTO(t *testing.T) {
	f := &auto.CoreFont{
		ID:            1,
		Name:          "Test",
		FileName:      "test.ttf",
		FileTransName: "uuid.ttf",
		IsDefault:     true,
		IsBuiltIn:     false,
		Size:          12.5,
		SizeType:      "KB",
	}
	dto := fontToDTO(f)
	assert.Equal(t, int64(1), dto.ID)
	assert.Equal(t, "Test", dto.Name)
	assert.Equal(t, "test.ttf", dto.FileName)
	assert.Equal(t, "uuid.ttf", dto.FileTransName)
	assert.True(t, dto.IsDefault)
	assert.False(t, dto.IsBuiltIn)
	assert.Equal(t, 12.5, dto.Size)
	assert.Equal(t, "KB", dto.SizeType)
}

// =====================================================================
// PermissionCompat parseRoleIDQuery tests
// =====================================================================

func TestRound9B_PermCompat_ParseRoleIDQuery_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodPost, "/", "")
	roleID, err := parseRoleIDQuery(c)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), roleID)
}

func TestRound9B_PermCompat_ParseRoleIDQuery_WithRoleID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodPost, "/", `{"roleId":42}`)
	roleID, err := parseRoleIDQuery(c)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), roleID)
}

func TestRound9B_PermCompat_ParseRoleIDQuery_WithQueryParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/?roleId=99", "")
	c.Request.URL.RawQuery = "roleId=99"
	c.Request.URL.Query()
	// Re-create request with query
	req := httptest.NewRequest(http.MethodGet, "/?roleId=99", nil)
	c.Request = req
	roleID, err := parseRoleIDQuery(c)
	assert.NoError(t, err)
	assert.Equal(t, int64(99), roleID)
}

func TestRound9B_PermCompat_ParseRoleIDQuery_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	_, err := parseRoleIDQuery(c)
	assert.Error(t, err)
}

// =====================================================================
// Audit DownloadExportFile additional edge case tests
// =====================================================================

func TestRound9B_Audit_DownloadExportFile_ValidPathNonExistentFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	// Build path to a file in temp dir with correct naming convention
	tmpDir := t.TempDir()
	filePath := tmpDir + "/audit_logs_test.csv"
	req := httptest.NewRequest(http.MethodGet, "/?path="+filePath, nil)
	c.Request = req
	h.DownloadExportFile(c)
	// File doesn't exist but passes path validation → NotFound
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9B_Audit_DownloadExportFile_InvalidFileName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	tmpDir := t.TempDir()
	filePath := tmpDir + "/invalid_name.csv"
	req := httptest.NewRequest(http.MethodGet, "/?path="+filePath, nil)
	c.Request = req
	h.DownloadExportFile(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9B_Audit_DownloadExportFile_InvalidExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	tmpDir := t.TempDir()
	filePath := tmpDir + "/audit_logs_test.txt"
	req := httptest.NewRequest(http.MethodGet, "/?path="+filePath, nil)
	c.Request = req
	h.DownloadExportFile(c)
	assertCode(t, w, response.CodeBadRequest)
}
