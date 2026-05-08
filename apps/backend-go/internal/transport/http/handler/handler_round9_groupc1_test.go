package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =====================================================================
// SystemParamHandler remaining branches
// =====================================================================

func TestRound9C1_SystemParam_QueryBasic_NilHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var h *SystemParamHandler
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.QueryBasic(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_SystemParam_SaveBasic_ValidArray(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `[{"key":"k1","val":"v1"}]`)
	h.SaveBasic(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_SystemParam_SaveOnlineMap_EmptyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	h.SaveOnlineMap(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_SystemParam_SaveSQLBot_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	h.SaveSQLBot(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_SystemParam_QueryOnlineMapByType_EmptyType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemParamHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "type", Value: ""}}
	h.QueryOnlineMapByType(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// SyncHandler remaining branches
// =====================================================================

func TestRound9C1_Sync_SourceDatasourcePager_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "page", Value: "1"}, {Key: "limit", Value: "10"}}
	h.SourceDatasourcePager(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Sync_TargetDatasourcePager_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "page", Value: "1"}, {Key: "limit", Value: "10"}}
	h.TargetDatasourcePager(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Sync_ListDatasourceByType_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "type", Value: ""}}
	h.ListDatasourceByType(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Sync_TaskPager_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "current", Value: "1"}, {Key: "size", Value: "10"}}
	h.TaskPager(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Sync_TaskLogPager_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "current", Value: "1"}, {Key: "size", Value: "10"}}
	h.TaskLogPager(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Sync_ExecuteTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.ExecuteTask(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Sync_ResourceCount_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ResourceCount(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Sync_ListDatasourceTables_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dsId", Value: "1"}}
	h.ListDatasourceTables(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Sync_ValidateDatasourceByID_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.ValidateDatasourceByID(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// CustomGeoHandler remaining branches
// =====================================================================

func TestRound9C1_CustomGeo_SaveGeoArea_DuplicateName_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":"area1","name":"Test Area"}`)
	h.SaveGeoArea(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_CustomGeo_SaveGeoSubArea_DuplicateName_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"sub","scope":"scope","geoAreaId":"area1"}`)
	h.SaveGeoSubArea(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_CustomGeo_DeleteGeoSubArea_ValidID_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodDelete, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	h.DeleteGeoSubArea(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_CustomGeo_ListGeoAreas_NilRepo_CoverSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListGeoAreas(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_CustomGeo_GetGeoArea_NilRepo_CoverSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "area1"}}
	h.GetGeoArea(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_CustomGeo_DeleteGeoArea_NilRepo_CoverSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCustomGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodDelete, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "area1"}}
	h.DeleteGeoArea(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_CustomGeo_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api")
	h := NewCustomGeoHandler(nil)
	RegisterCustomGeoRoutes(group, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/customGeo/geoArea/list", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// DatasourceHandler remaining branches
// =====================================================================

func TestRound9C1_Datasource_UploadFile_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Request.Header.Del("Content-Type")
	h.UploadFile(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Datasource_Save_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"test","type":"mysql","configuration":"{}"}`)
	h.Save(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Datasource_Tables_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"datasourceId":1}`)
	h.Tables(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Datasource_TableStatus_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"datasourceId":1}`)
	h.TableStatus(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Datasource_CheckRepeat_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"test","type":"mysql","configuration":"{}"}`)
	h.CheckRepeat(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Datasource_Rename_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"new-name","type":"mysql"}`)
	h.Rename(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Datasource_LoadRemoteFile_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.LoadRemoteFile(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Datasource_LoadRemoteFile_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"url":"http://example.com/file.csv"}`)
	h.LoadRemoteFile(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// VisualizationHandler remaining branches
// =====================================================================

func TestRound9C1_Visualization_FindByID_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVisualizationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.FindByID(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Visualization_FindByID_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVisualizationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1}`)
	h.FindByID(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Visualization_UpdateCheckVersion_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVisualizationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.UpdateCheckVersion(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Visualization_UpdateCheckVersion_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVisualizationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.UpdateCheckVersion(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Visualization_RecoverToPublished_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVisualizationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.RecoverToPublished(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Visualization_RecoverToPublished_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVisualizationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1}`)
	h.RecoverToPublished(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Visualization_ViewDetailList_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVisualizationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "bad"}}
	h.ViewDetailList(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Visualization_ViewDetailList_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVisualizationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "1"}}
	h.ViewDetailList(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Visualization_getUpdateBy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h := &VisualizationHandler{}
	result := h.getUpdateBy(c)
	assert.Equal(t, "system", result)
}

func TestRound9C1_Visualization_getUpdateBy_Int64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", int64(42))
	h := &VisualizationHandler{}
	result := h.getUpdateBy(c)
	assert.Equal(t, "42", result)
}

func TestRound9C1_Visualization_getUpdateBy_Int(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", 99)
	h := &VisualizationHandler{}
	result := h.getUpdateBy(c)
	assert.Equal(t, "99", result)
}

func TestRound9C1_Visualization_getUpdateBy_String(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", "admin")
	h := &VisualizationHandler{}
	result := h.getUpdateBy(c)
	assert.Equal(t, "admin", result)
}

// =====================================================================
// RoleHandler remaining branches
// =====================================================================

func TestRound9C1_Role_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Role_Delete_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Role_getCreateBy_String(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Set("userId", "admin")
	h := &RoleHandler{}
	result := h.getCreateBy(c)
	assert.Equal(t, "admin", result)
}

func TestRound9C1_Role_getCreateBy_Int64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Set("userId", int64(42))
	h := &RoleHandler{}
	result := h.getCreateBy(c)
	assert.Equal(t, "42", result)
}

func TestRound9C1_Role_getCreateBy_Int(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Set("userId", 99)
	h := &RoleHandler{}
	result := h.getCreateBy(c)
	assert.Equal(t, "99", result)
}

func TestRound9C1_Role_getCreateBy_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h := &RoleHandler{}
	result := h.getCreateBy(c)
	assert.Equal(t, embeddedDefaultUpdateBy, result)
}

func TestRound9C1_Role_BeforeUnmountInfo_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.BeforeUnmountInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Role_BeforeUnmountInfo_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"orgId":1,"roleId":1,"userId":1}`)
	h.BeforeUnmountInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Role_UpdateLastRolePolicy_NilGovernanceSvc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPut, "/", `{"policy":"all"}`)
	h.UpdateLastRolePolicy(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Role_SelectedForUser_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SelectedForUser(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Role_SelectedForUser_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"userId":1}`)
	h.SelectedForUser(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Role_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api")
	h := NewRoleHandler(nil)
	RegisterRoleRoutes(group, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/role/query", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// OrgHandler remaining branches
// =====================================================================

func TestRound9C1_Org_CreateOrg_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrgHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.CreateOrg(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Org_CreateOrg_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrgHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"org1"}`)
	round9SetOrg(c, 1)
	h.CreateOrg(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Org_UpdateOrg_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrgHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UpdateOrg(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Org_UpdateOrg_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrgHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"org-updated"}`)
	round9SetOrg(c, 1)
	h.UpdateOrg(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Org_UpdateOrgStatus_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrgHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UpdateOrgStatus(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Org_UpdateOrgStatus_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrgHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"orgId":1,"status":1}`)
	round9SetOrg(c, 1)
	h.UpdateOrgStatus(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Org_ListOrgs_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrgHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListOrgs(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Org_GetOrgTree_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrgHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.GetOrgTree(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Org_DeleteOrg_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrgHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "orgId", Value: "bad"}}
	h.DeleteOrg(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Org_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api")
	h := NewOrgHandler(nil)
	RegisterOrgRoutes(group, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/system/organization/list", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// AuditHandler remaining branches
// =====================================================================

func TestRound9C1_Audit_CreateAuditLog_NilAuditService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"actionType":"login","resourceType":"user"}`)
	h.CreateAuditLog(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C1_Audit_GetAuditLogByID_ValidID_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetAuditLogByID(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C1_Audit_CleanupNow_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.CleanupNow(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C1_Audit_GetAuditAlertSettings_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.GetAuditAlertSettings(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C1_Audit_DeleteAuditLogsRetention_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodDelete, "/", "")
	h.DeleteAuditLogsRetention(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C1_Audit_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api")
	h := NewAuditHandler(nil, nil)
	RegisterAuditRoutes(group, h, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/audit/settings", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// UserHandler remaining branches
// =====================================================================

func TestRound9C1_User_ClearErrorRecord_NilImportService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "key", Value: "test-key"}}
	h.ClearErrorRecord(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_User_ResetPasswordCompat_EmptyUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	round9SetOrg(c, 1)
	h.ResetPasswordCompat(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_User_ResetPasswordCompat_InvalidUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "uid", Value: "bad"}}
	round9SetOrg(c, 1)
	h.ResetPasswordCompat(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_User_SwitchEnable_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"status":1}`)
	round9SetOrg(c, 1)
	h.SwitchEnable(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_User_SetAuthService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	h.SetAuthService(nil)
	assert.Nil(t, h.switchOrg)
}

// =====================================================================
// TemplateHandler remaining branches
// =====================================================================

func TestRound9C1_Template_SearchTemplates_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTemplateHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.SearchTemplates(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C1_Template_SearchTemplateMarket_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTemplateHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.SearchTemplateMarket(c)
	assertCode(t, w, response.CodeSuccess)
}

func TestRound9C1_Template_SearchTemplateMarketRecommend_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTemplateHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.SearchTemplateMarketRecommend(c)
	assertCode(t, w, response.CodeSuccess)
}

func TestRound9C1_Template_SearchTemplateMarketPreview_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTemplateHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.SearchTemplateMarketPreview(c)
	assertCode(t, w, response.CodeSuccess)
}

// =====================================================================
// SubjectHandler remaining branches
// =====================================================================

func TestRound9C1_Subject_Query_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubjectHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.Query(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Subject_QueryWithGroup_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubjectHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.QueryWithGroup(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Subject_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubjectHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Update(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Subject_Update_Create_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubjectHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"NewSubject","details":"{}","coverUrl":""}`)
	h.Update(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Subject_Update_Update_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubjectHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":"abc123","name":"Updated","details":"{}","coverUrl":""}`)
	h.Update(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Subject_Delete_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubjectHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: ""}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Subject_Delete_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubjectHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "abc123"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Subject_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubjectHandler(nil)
	RegisterSubjectRoutes(r, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/visualizationSubject/query", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// PermissionCompatHandler remaining branches
// =====================================================================

func TestRound9C1_PermCompat_BusiPermission_NilPermService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.BusiPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_PermCompat_SaveBusiPer_NilRoleService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"roleId":1,"permIds":[1,2]}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.SaveBusiPer(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_PermCompat_MenuTargetPermission_NilServices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"roleId":1}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.MenuTargetPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_PermCompat_MenuPermission_WithRoleID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"roleId":5}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.MenuPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_PermCompat_BusiResource_NilPermService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermissionCompatHandler(nil, nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	c.Params = gin.Params{{Key: "flag", Value: "dashboard"}}
	h.BusiResource(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// PermHandler remaining branches
// =====================================================================

func TestRound9C1_Perm_ListPerms_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.ListPerms(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Perm_ListPerms_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	round9SetUser(c, 1)
	round9SetOrg(c, 1)
	h.ListPerms(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Perm_CreatePerm_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.CreatePerm(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Perm_CreatePerm_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"permKey":"test:read","permName":"Test Read"}`)
	h.CreatePerm(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Perm_UpdatePerm_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UpdatePerm(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Perm_UpdatePerm_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"permKey":"test:write"}`)
	h.UpdatePerm(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Perm_DeletePerm_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.DeletePerm(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Perm_DeletePerm_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPermHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.DeletePerm(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// ExportHandler remaining branches
// =====================================================================

func TestRound9C1_Export_Delete_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewExportHandler(nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: ""}}
	h.Delete(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9C1_Export_Delete_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewExportHandler(nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Export_DeleteBatch_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewExportHandler(nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.DeleteBatch(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9C1_Export_DeleteBatch_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewExportHandler(nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"ids":["id1","id2"]}`)
	h.DeleteBatch(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Export_DeleteAll_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewExportHandler(nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "type", Value: "chart"}}
	h.DeleteAll(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Export_Retry_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewExportHandler(nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: ""}}
	h.Retry(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9C1_Export_Retry_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewExportHandler(nil, nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}
	h.Retry(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Export_normalizeExportResourceType(t *testing.T) {
	assert.Equal(t, "dataset", normalizeExportResourceType("dataset"))
	assert.Equal(t, "dashboard", normalizeExportResourceType("dashboard"))
	assert.Equal(t, "screen", normalizeExportResourceType("screen"))
	assert.Equal(t, "datasource", normalizeExportResourceType("datasource"))
	assert.Equal(t, "", normalizeExportResourceType("unknown"))
}

func TestRound9C1_Export_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewExportHandler(nil, nil, nil)
	RegisterExportRoutes(r, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/exportTasks/records", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// DriverHandler remaining branches
// =====================================================================

func TestRound9C1_Driver_List_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDriverHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.List(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Driver_ListByType_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDriverHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dsType", Value: "mysql"}}
	h.ListByType(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Driver_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDriverHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.GetByID(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Driver_GetByID_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDriverHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetByID(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Driver_ListDriverJars_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDriverHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "driverId", Value: "bad"}}
	h.ListDriverJars(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Driver_ListDriverJars_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDriverHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "driverId", Value: "1"}}
	h.ListDriverJars(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C1_Driver_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api")
	h := NewDriverHandler(nil)
	RegisterDriverRoutes(group, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/driver/list", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// Datasource request parsers (remaining branches)
// =====================================================================

func TestRound9C1_Datasource_toInt64ID(t *testing.T) {
	assert.Equal(t, int64(42), toInt64ID(int64(42)))
	assert.Equal(t, int64(42), toInt64ID(42))
	assert.Equal(t, int64(42), toInt64ID(uint64(42)))
	assert.Equal(t, int64(42), toInt64ID(float64(42)))
	assert.Equal(t, int64(42), toInt64ID("42"))
	assert.Equal(t, int64(0), toInt64ID("bad"))
	assert.Equal(t, int64(0), toInt64ID(nil))
	assert.Equal(t, int64(0), toInt64ID(true))
}

func TestRound9C1_Datasource_parseEditType_Nil(t *testing.T) {
	body := map[string]interface{}{}
	result := parseEditType(body)
	assert.Nil(t, result)
}

func TestRound9C1_Datasource_parseEditType_Number(t *testing.T) {
	body := map[string]interface{}{"editType": float64(1)}
	result := parseEditType(body)
	assert.NotNil(t, result)
	assert.Equal(t, "1", *result)
}

func TestRound9C1_Datasource_parseEditType_String(t *testing.T) {
	body := map[string]interface{}{"editType": "0"}
	result := parseEditType(body)
	assert.NotNil(t, result)
	assert.Equal(t, "0", *result)
}

func TestRound9C1_Datasource_parseRequestBody_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	c.Request = req
	body, ok := parseRequestBody(c)
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{}, body)
}

func TestRound9C1_Datasource_parseDatasourceConfiguration_Nil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := map[string]interface{}{"configuration": "already-string"}
	result, ok := parseDatasourceConfiguration(c, body)
	assert.True(t, ok)
	assert.NotNil(t, result)
	assert.Equal(t, "already-string", *result)
}

func TestRound9C1_Datasource_parseDatasourceConfiguration_Map(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := map[string]interface{}{"configuration": map[string]interface{}{"host": "localhost"}}
	result, ok := parseDatasourceConfiguration(c, body)
	assert.True(t, ok)
	assert.NotNil(t, result)
	assert.Contains(t, *result, "host")
}

func TestRound9C1_Datasource_parseDatasourceConfiguration_Array(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := map[string]interface{}{"configuration": []interface{}{map[string]interface{}{"host": "localhost"}}}
	result, ok := parseDatasourceConfiguration(c, body)
	assert.True(t, ok)
	assert.NotNil(t, result)
}

func TestRound9C1_Datasource_parseDatasourceConfiguration_Other(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := map[string]interface{}{"configuration": 123}
	result, ok := parseDatasourceConfiguration(c, body)
	assert.True(t, ok)
	assert.Nil(t, result)
}

func TestRound9C1_Datasource_buildDatasourceTreeResponse_Nil(t *testing.T) {
	result := buildDatasourceTreeResponse(nil)
	assert.Len(t, result, 1)
	assert.Equal(t, "0", result[0].ID)
}

// =====================================================================
// Dataset request parsers (remaining branches)
// =====================================================================

func TestRound9C1_Dataset_parseEnumFilters_Nil(t *testing.T) {
	result := parseEnumFilters(nil)
	assert.Equal(t, []dataset.EnumFilter{}, result)
}

func TestRound9C1_Dataset_parseEnumFilters_Empty(t *testing.T) {
	result := parseEnumFilters([]interface{}{})
	assert.Equal(t, []dataset.EnumFilter{}, result)
}

func TestRound9C1_Dataset_parseEnumFilters_InvalidItem(t *testing.T) {
	items := []interface{}{"not-a-map"}
	result := parseEnumFilters(items)
	assert.Equal(t, []dataset.EnumFilter{}, result)
}

func TestRound9C1_Dataset_parseEnumFilters_WithFieldIDString(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"fieldId": "abc", "operator": "eq", "value": []interface{}{"v1"}},
	}
	result := parseEnumFilters(items)
	assert.Len(t, result, 1)
	assert.Equal(t, "abc", result[0].FieldID)
	assert.Equal(t, "eq", result[0].Operator)
}

func TestRound9C1_Dataset_parseEnumFilters_WithFieldIDNumber(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"fieldId": float64(42), "operator": "in", "value": []interface{}{1, 2}},
	}
	result := parseEnumFilters(items)
	assert.Len(t, result, 1)
	assert.Equal(t, "42", result[0].FieldID)
}

func TestRound9C1_Dataset_parseEnumFilters_NoFieldID(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"operator": "eq"},
	}
	result := parseEnumFilters(items)
	assert.Len(t, result, 1)
	assert.Equal(t, "", result[0].FieldID)
}

func TestRound9C1_Dataset_parseEnumFilters_NoValues(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"fieldId": "f1"},
	}
	result := parseEnumFilters(items)
	assert.Len(t, result, 1)
	assert.Equal(t, []interface{}{}, result[0].Value)
}

func TestRound9C1_Dataset_parseEnumValueRequest_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	req, ok := parseEnumValueRequest(c)
	assert.False(t, ok)
	assert.Nil(t, req)
}

func TestRound9C1_Dataset_parseEnumValueRequest_ZeroQueryID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"queryId":0}`)
	req, ok := parseEnumValueRequest(c)
	assert.False(t, ok)
	assert.Nil(t, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound9C1_Dataset_parseEnumValueRequest_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodPost, "/", `{"queryId":1,"filter":[{"fieldId":"f1","operator":"eq","value":["v1"]}]}`)
	req, ok := parseEnumValueRequest(c)
	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Len(t, req.Filter, 1)
}

func TestRound9C1_Dataset_dedupeDatasetIDs(t *testing.T) {
	result := dedupeDatasetIDs([]int64{3, 1, 2, 3, 1, 0, -1})
	assert.Equal(t, []int64{3, 1, 2}, result)
}

func TestRound9C1_Dataset_dedupeDatasetIDs_Empty(t *testing.T) {
	result := dedupeDatasetIDs([]int64{})
	assert.Equal(t, []int64{}, result)
}
