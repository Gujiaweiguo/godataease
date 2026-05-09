package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ---------- Shared helpers ----------

func newRound8Ctx(t *testing.T, method, path, body string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return w, c
}

func parseRound8Resp(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func openRound8DB(t *testing.T, name string, models ...interface{}) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:round8_%s_%s?mode=memory&cache=shared", name, strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	for _, m := range models {
		require.NoError(t, db.AutoMigrate(m))
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	return db
}

// ---------- RelationHandler tests ----------

func TestRound8_RelationHandler_GetDatasourceRelationship_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.GetDatasourceRelationship(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RelationHandler_GetDatasourceRelationship_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetDatasourceRelationship(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
	assert.Contains(t, resp["msg"], "Failed")
}

func TestRound8_RelationHandler_GetDatasetRelationship_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	h.GetDatasetRelationship(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RelationHandler_GetPanelRelationship_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "-1"}}
	h.GetPanelRelationship(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RelationHandler_CheckPermission_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "0"}}
	h.CheckPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RelationHandler_CheckPermission_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	h.CheckPermission(c)
	resp := parseRound8Resp(t, w)
	// nil service → explicit nil guard → error response
	assert.Equal(t, "500000", resp["code"])
}

// ---------- GeoHandler tests ----------

func TestRound8_GeoHandler_ListAreas_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewGeoHandler(nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.ListAreas(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_GeoHandler_GetArea_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewGeoHandler(nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "123"}}
	h.GetArea(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_GeoHandler_Save_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewGeoHandler(nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Request.Header.Del("Content-Type")
	h.Save(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
	assert.Contains(t, resp["msg"], "file is required")
}

func TestRound8_GeoHandler_Delete_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewGeoHandler(nil)
	w, c := newRound8Ctx(t, http.MethodDelete, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "area1"}}
	h.Delete(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

// ---------- MapHandler tests ----------

func TestRound8_MapHandler_GetWorldTree_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMapHandler(nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.GetWorldTree(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------- DataPermissionHandler tests ----------

func TestRound8_DataPermission_RowPermissionPager_InvalidDatasetID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{
		{Key: "datasetId", Value: "bad"},
		{Key: "page", Value: "1"},
		{Key: "limit", Value: "10"},
	}
	h.RowPermissionPager(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_DataPermission_RowPermissionPager_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{
		{Key: "datasetId", Value: "1"},
		{Key: "page", Value: "abc"},
		{Key: "limit", Value: "10"},
	}
	h.RowPermissionPager(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_DataPermission_RowPermissionPager_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{
		{Key: "datasetId", Value: "1"},
		{Key: "page", Value: "1"},
		{Key: "limit", Value: "10"},
	}
	h.RowPermissionPager(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
}

func TestRound8_DataPermission_RowPermissionPagerByTarget_InvalidTargetID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{
		{Key: "datasetId", Value: "1"},
		{Key: "targetType", Value: "role"},
		{Key: "targetId", Value: "bad"},
		{Key: "page", Value: "1"},
		{Key: "limit", Value: "10"},
	}
	h.RowPermissionPagerByTarget(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_DataPermission_SaveRowPermission_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveRowPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
	assert.Contains(t, resp["msg"], "Invalid request")
}

func TestRound8_DataPermission_DeleteRowPermission_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.DeleteRowPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_DataPermission_ColumnPermissionPager_InvalidDatasetID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{
		{Key: "datasetId", Value: "bad"},
		{Key: "page", Value: "1"},
		{Key: "limit", Value: "10"},
	}
	h.ColumnPermissionPager(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_DataPermission_ColumnPermissionPager_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{
		{Key: "datasetId", Value: "1"},
		{Key: "page", Value: "1"},
		{Key: "limit", Value: "10"},
	}
	h.ColumnPermissionPager(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
}

func TestRound8_DataPermission_SaveColumnPermission_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveColumnPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_DataPermission_DeleteColumnPermission_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.DeleteColumnPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_DataPermission_SaveRowPermission_UnsupportedFilterType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"id":0,"datasetId":1,"filterType":"logic","targetId":1}`)
	h.SaveRowPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_DataPermission_DeleteRowPermission_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"id":1}`)
	h.DeleteRowPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
}

func TestRound8_DataPermission_SaveColumnPermission_UnsupportedRuleType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	// ruleType "show" is not a supported permission type (only "disable"/"mask" are), so the service returns error
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"id":0,"datasetId":1,"fieldName":"col1","ruleType":"show"}`)
	h.SaveColumnPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_DataPermission_DeleteColumnPermission_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(service.NewDataPermissionAdminService(&fakeRowPermissionHandlerStore{}, &fakeColumnPermissionHandlerStore{}, &fakeDataPermissionFieldProvider{}, nil))
	// fake store GetByID returns nil, so service reports "not found"
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"id":1}`)
	h.DeleteColumnPermission(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

// ---------- VisualizationBackgroundHandler tests ----------

func TestRound8_VisBackgroundHandler_FindAll_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "visbg", &auto.VisualizationBackground{})
	repo := repository.NewVisualizationBackgroundRepository(db)
	h := NewVisualizationBackgroundHandler(repo)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.FindAll(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
	assert.Equal(t, map[string]interface{}{}, resp["data"])
}

func TestRound8_VisBackgroundHandler_FindAll_WithData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "visbgdata", &auto.VisualizationBackground{})
	require.NoError(t, db.Create(&auto.VisualizationBackground{
		ID: "bg1", Name: "Gradient Blue", Classification: "gradient", Sort: 1,
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationBackground{
		ID: "bg2", Name: "Solid Red", Classification: "solid", Sort: 2,
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationBackground{
		ID: "bg3", Name: "Gradient Green", Classification: "gradient", Sort: 3,
	}).Error)

	repo := repository.NewVisualizationBackgroundRepository(db)
	h := NewVisualizationBackgroundHandler(repo)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.FindAll(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
	assert.Contains(t, data, "gradient")
	assert.Contains(t, data, "solid")
}

// ---------- PdfTemplateHandler tests ----------

func TestRound8_PdfTemplateHandler_QueryAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPdfTemplateHandler()
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.QueryAll(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
	assert.Equal(t, []interface{}{}, resp["data"])
}

// ---------- RoleHandler remaining tests ----------

func TestRound8_RoleHandler_Query_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/query", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RoleHandler_Query_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/query", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
}

func TestRound8_RoleHandler_Detail_DirectInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/delete/bad", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RoleHandler_Edit_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/edit", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RoleHandler_MountExternalUser_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/mountExternalUser", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RoleHandler_BeforeUnmountInfo_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/beforeUnmountInfo", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RoleHandler_SearchExternalUser_MissingOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouterWithOrg(t, 0)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/role/searchExternalUser/testuser", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RoleHandler_OptionForUser_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/user/option", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RoleHandler_SelectedForUser_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/role/user/selected", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RoleHandler_UpdateLastRolePolicy_InvalidPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouter(t)
	body := []byte(`{"orgId":1,"policy":"INVALID_POLICY"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/governance/last-role-policy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_RoleHandler_GetLastRolePolicy_MissingOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRoleHandlerTestRouterWithOrg(t, 0)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/governance/last-role-policy", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

// ---------- OrgHandler remaining tests ----------

func TestRound8_OrgHandler_UpdateOrg_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupOrgHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/organization/update", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_OrgHandler_UpdateOrgStatus_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupOrgHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/organization/updateStatus", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_OrgHandler_GetOrgByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupOrgHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/organization/info/bad", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_OrgHandler_CheckOrgName_MissingParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupOrgHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/organization/checkName", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
	assert.Contains(t, resp["msg"], "orgName is required")
}

func TestRound8_OrgHandler_GetChildOrgs_InvalidParentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupOrgHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/organization/children/bad", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_OrgHandler_TransferUser_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupOrgHandlerTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/organization/transfer-user", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

// ---------- ThresholdHandler remaining tests ----------

func TestRound8_ThresholdHandler_Edit_BadRequest(t *testing.T) {
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/edit", []byte{})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound8_ThresholdHandler_SwitchEnable_BadRequest(t *testing.T) {
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/switch", []byte{})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound8_ThresholdHandler_BatchReci_BadRequest(t *testing.T) {
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/batchReci", []byte{})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound8_ThresholdHandler_InstancePager_InvalidPage(t *testing.T) {
	engine, _, _ := newThresholdHandlerTestEnv(t)
	body, _ := json.Marshal(map[string]any{"keyword": ""})
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/instancePager/0/10", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound8_ThresholdHandler_Preview_BadRequest(t *testing.T) {
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/preview", []byte{})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound8_ThresholdHandler_DeleteWithChart_InvalidChartID(t *testing.T) {
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodGet, "/threshold/deleteWithChart/bad/core", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

// ---------- StoreHandler remaining tests ----------

func TestRound8_StoreHandler_ResourceTypeFromString(t *testing.T) {
	assert.Equal(t, int32(1), resourceTypeFromString("PANEL"))
	assert.Equal(t, int32(1), resourceTypeFromString("DATA_VIZ"))
	assert.Equal(t, int32(2), resourceTypeFromString("SCREEN"))
	assert.Equal(t, int32(2), resourceTypeFromString("DATA_VIZ_SCREEN"))
	assert.Equal(t, int32(3), resourceTypeFromString("REPORT"))
	assert.Equal(t, int32(1), resourceTypeFromString("UNKNOWN"))
}

func TestRound8_StoreHandler_GetUserID_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, int64(0), getUserID(c))
}

func TestRound8_StoreHandler_GetUserID_Int64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(middleware.ContextUserID, int64(42))
	assert.Equal(t, int64(42), getUserID(c))
}

// ---------- FontHandler remaining tests ----------

func TestRound8_FontHandler_List_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fontlist", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.List(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
	assert.Equal(t, []interface{}{}, resp["data"])
}

func TestRound8_FontHandler_Create_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fontcreate", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.Create(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_FontHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fontcreateok", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"name":"TestFont","fileName":"test.ttf","fileTransName":"uuid.ttf","isDefault":false,"size":12.5,"sizeType":"KB"}`)
	h.Create(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "TestFont", data["name"])
}

func TestRound8_FontHandler_Create_DuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fontdup", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	require.NoError(t, repo.CreateFont(&auto.CoreFont{Name: "DupFont", UpdateTime: 1}))
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"name":"DupFont","fileName":"test.ttf"}`)
	h.Create(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
	assert.Contains(t, resp["msg"], "重名")
}

func TestRound8_FontHandler_Edit_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fontedit", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.Edit(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_FontHandler_Edit_DelegatesToCreateWhenNoID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fonteditcreate", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"id":0,"name":"NewFont","fileName":"new.ttf","fileTransName":"uuid.ttf","isDefault":false,"size":10.0,"sizeType":"KB"}`)
	h.Edit(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
}

func TestRound8_FontHandler_Edit_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fonteditnf", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"id":999,"name":"NoFont"}`)
	h.Edit(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_FontHandler_DefaultFont_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fontdefault", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.DefaultFont(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "000000", resp["code"])
	assert.Equal(t, []interface{}{}, resp["data"])
}

func TestRound8_FontHandler_UploadFile_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fontupload", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Request.Header.Del("Content-Type")
	h.UploadFile(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_FontHandler_IsAllowedFontDownloadExtension(t *testing.T) {
	assert.True(t, isAllowedFontDownloadExtension("font.ttf"))
	assert.True(t, isAllowedFontDownloadExtension("font.TTF"))
	assert.True(t, isAllowedFontDownloadExtension("font.otf"))
	assert.True(t, isAllowedFontDownloadExtension("font.woff"))
	assert.True(t, isAllowedFontDownloadExtension("font.woff2"))
	assert.False(t, isAllowedFontDownloadExtension("font.exe"))
	assert.False(t, isAllowedFontDownloadExtension("font.txt"))
	assert.False(t, isAllowedFontDownloadExtension("font"))
}

func TestRound8_FontHandler_ResolveSafeFontFilePath(t *testing.T) {
	safe, ok := resolveSafeFontFilePath("/fonts", "demo.ttf")
	assert.True(t, ok)
	assert.Equal(t, "/fonts/demo.ttf", safe)

	_, ok = resolveSafeFontFilePath("/fonts", "")
	assert.False(t, ok)

	_, ok = resolveSafeFontFilePath("/fonts", "../escape.ttf")
	assert.False(t, ok)

	_, ok = resolveSafeFontFilePath("/fonts", "/absolute/path.ttf")
	assert.False(t, ok)

	_, ok = resolveSafeFontFilePath("/fonts", "file.exe")
	assert.False(t, ok)
}

func TestRound8_FontHandler_FontToDTO(t *testing.T) {
	f := &auto.CoreFont{
		ID: 1, Name: "Test", FileName: "test.ttf", FileTransName: "uuid.ttf",
		IsDefault: true, IsBuiltIn: false, Size: 12.5, SizeType: "KB",
	}
	dto := fontToDTO(f)
	assert.Equal(t, int64(1), dto.ID)
	assert.Equal(t, "Test", dto.Name)
	assert.Equal(t, "test.ttf", dto.FileName)
	assert.True(t, dto.IsDefault)
	assert.False(t, dto.IsBuiltIn)
	assert.Equal(t, 12.5, dto.Size)
}

func TestRound8_FontHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fontdelinv", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Delete(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_FontHandler_Delete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openRound8DB(t, "fontdelnf", &auto.CoreFont{})
	repo := repository.NewTypefaceRepository(db)
	h := NewFontHandler(repo)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "99999"}}
	h.Delete(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_FontHandler_Download_EmptyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFontHandler(nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "file", Value: ""}}
	h.Download(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

// ---------- AuthHandler remaining tests ----------

func TestRound8_AuthHandler_Logout_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.Logout(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_AuthHandler_Refresh_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Request.Header.Set("Authorization", "Bearer some-token")
	h.Refresh(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound8_AuthHandler_LocalLogin_EncryptedPassthrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(service.NewAuthService(nil, nil, nil, nil))
	r := gin.New()
	r.POST("/login/localLogin", h.LocalLogin)

	body := `{"name":"admin","pwd":"admin123"}`
	req := httptest.NewRequest(http.MethodPost, "/login/localLogin", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

// ---------- HandlerSharedHelpers tests ----------

func TestRound8_ParseInt64Value(t *testing.T) {
	val, ok := parseInt64Value(json.Number("42"))
	assert.True(t, ok)
	assert.Equal(t, int64(42), val)

	val, ok = parseInt64Value(float64(99))
	assert.True(t, ok)
	assert.Equal(t, int64(99), val)

	val, ok = parseInt64Value(int64(7))
	assert.True(t, ok)
	assert.Equal(t, int64(7), val)

	val, ok = parseInt64Value(int(3))
	assert.True(t, ok)
	assert.Equal(t, int64(3), val)

	val, ok = parseInt64Value("123")
	assert.True(t, ok)
	assert.Equal(t, int64(123), val)

	_, ok = parseInt64Value("not-a-number")
	assert.False(t, ok)

	_, ok = parseInt64Value(nil)
	assert.False(t, ok)

	_, ok = parseInt64Value(true)
	assert.False(t, ok)
}

func TestRound8_ParsePageParams_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "page", Value: "bad"}, {Key: "limit", Value: "10"}}

	page, size, ok := parsePageParams(c)
	assert.False(t, ok)
	assert.Equal(t, 0, page)
	assert.Equal(t, 0, size)
}

func TestRound8_ParseIDParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	id, ok := parseIDParam(c, "id")
	assert.True(t, ok)
	assert.Equal(t, int64(42), id)
}

func TestRound8_ParseIDParamBadRequest_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	id, ok := parseIDParamBadRequest(c, "id")
	assert.False(t, ok)
	assert.Equal(t, int64(0), id)
}

func TestRound8_ParseIDList(t *testing.T) {
	ids, err := parseIDList([]string{"1", "2", "3"})
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, ids)

	_, err = parseIDList([]string{"1", "bad", "3"})
	assert.Error(t, err)

	_, err = parseIDList([]string{"0"})
	assert.Error(t, err)
}

func TestRound8_IsEOFBindError(t *testing.T) {
	assert.True(t, isEOFBindError(fmt.Errorf("unexpected EOF")))
	assert.True(t, isEOFBindError(fmt.Errorf("something EOF something")))
	assert.False(t, isEOFBindError(fmt.Errorf("other error")))
	assert.False(t, isEOFBindError(nil))
}

func TestRound8_FirstNonEmptyParam(t *testing.T) {
	assert.Equal(t, "a", firstNonEmptyParam("a", "b"))
	assert.Equal(t, "b", firstNonEmptyParam("", "b"))
	assert.Equal(t, "", firstNonEmptyParam("", ""))
	assert.Equal(t, "x", firstNonEmptyParam("", " ", "x"))
}

// ---------- AuditHandler non-integration tests ----------

func TestRound8_AuditHandler_GetAuditAlertSettings_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.GetAuditAlertSettings(c)
	resp := parseRound8Resp(t, w)
	// nil systemParamService panics, recovered by recoverServicePanic → CodeInternalError
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_AuditHandler_SaveAuditAlertSettings_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveAuditAlertSettings(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "10001", resp["code"])
}

func TestRound8_AuditHandler_CleanupNow_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	h.CleanupNow(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_AuditHandler_TestNotification_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "")
	h.TestNotification(c)
	resp := parseRound8Resp(t, w)
	// auditAlertService.Notify returns error (no panic), handler uses InternalError → CodeServerErr "40001"
	assert.Equal(t, "40001", resp["code"])
}

func TestRound8_AuditHandler_CreateAuditLog_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.CreateAuditLog(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "10001", resp["code"])
}

func TestRound8_AuditHandler_GetAuditLogByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.GetAuditLogByID(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "10001", resp["code"])
}

func TestRound8_AuditHandler_GetAuditLogsByUserID_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "userId", Value: "bad"}}
	h.GetAuditLogsByUserID(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "10001", resp["code"])
}

func TestRound8_AuditHandler_ExportAuditLogs_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.ExportAuditLogs(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "10001", resp["code"])
}

func TestRound8_AuditHandler_ExportAuditLogs_EmptyIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", `{"ids":[]}`)
	h.ExportAuditLogs(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "10001", resp["code"])
	assert.Contains(t, resp["msg"], "No audit log IDs")
}

func TestRound8_AuditHandler_DeleteAuditLogsRetention_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodDelete, "/", "")
	h.DeleteAuditLogsRetention(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "500000", resp["code"])
}

func TestRound8_AuditHandler_RecordLoginFailure_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodPost, "/", "not-json")
	h.RecordLoginFailure(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "10001", resp["code"])
}

func TestRound8_AuditHandler_DownloadExportFile_EmptyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/", "")
	h.DownloadExportFile(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "10001", resp["code"])
}

func TestRound8_AuditHandler_DownloadExportFile_PathTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuditHandler(nil, nil)
	w, c := newRound8Ctx(t, http.MethodGet, "/?path=/etc/passwd", "")
	h.DownloadExportFile(c)
	resp := parseRound8Resp(t, w)
	assert.Equal(t, "10001", resp["code"])
}

// ---------- ParseDatasetPagerParams direct test ----------

func TestRound8_ParseDatasetPagerParams_InvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{
		{Key: "datasetId", Value: "1"},
		{Key: "page", Value: "1"},
		{Key: "limit", Value: "bad"},
	}
	datasetID, page, size, ok := parseDatasetPagerParams(c)
	assert.False(t, ok)
	assert.Equal(t, int64(0), datasetID)
	assert.Equal(t, 0, page)
	assert.Equal(t, 0, size)
}

// ---------- PermissionScopeHelper direct test ----------

func TestRound8_BuildPermissionScope_Admin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("role", "admin")
	c.Set("user_id", uint64(1))

	scope, err := buildPermissionScope(c)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), scope.OrgID)
	assert.Equal(t, int64(1), scope.ActorID)
}

func TestRound8_BuildPermissionScope_NonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("role", "viewer")
	c.Set("user_id", uint64(5))
	c.Set("org_id", uint64(10))

	scope, err := buildPermissionScope(c)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), scope.OrgID)
	assert.Equal(t, int64(5), scope.ActorID)
}

func TestRound8_BuildPermissionScope_MissingOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("role", "viewer")
	c.Set("user_id", uint64(5))
	// no org_id set

	_, err := buildPermissionScope(c)
	assert.ErrorIs(t, err, service.ErrInvalidOrgContext)
}

// ---------- parseThresholdPageParams direct test ----------

func TestRound8_ParseThresholdPageParams_InvalidGoPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "goPage", Value: "0"}, {Key: "pageSize", Value: "10"}}
	goPage, pageSize, ok := parseThresholdPageParams(c)
	assert.False(t, ok)
	assert.Equal(t, 0, goPage)
	assert.Equal(t, 0, pageSize)
}

func TestRound8_ParseThresholdPageParams_InvalidPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "-5"}}
	goPage, pageSize, ok := parseThresholdPageParams(c)
	assert.False(t, ok)
	assert.Equal(t, 0, goPage)
	assert.Equal(t, 0, pageSize)
}

// ---------- Nil-handler constructor tests ----------

func TestRound8_NewRelationHandler_Nil(t *testing.T) {
	h := NewRelationHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.relationService)
}

func TestRound8_NewGeoHandler_Nil(t *testing.T) {
	h := NewGeoHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

func TestRound8_NewMapHandler_Nil(t *testing.T) {
	h := NewMapHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

func TestRound8_NewPdfTemplateHandler(t *testing.T) {
	h := NewPdfTemplateHandler()
	assert.NotNil(t, h)
}

func TestRound8_NewVisualizationBackgroundHandler(t *testing.T) {
	h := NewVisualizationBackgroundHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.repo)
}

func TestRound8_NewDataPermissionHandler(t *testing.T) {
	h := NewDataPermissionHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

func TestRound8_RegisterDataPermissionRoutes_NilHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterDataPermissionRoutes(r.Group("/api"), nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/dataset/rowPermissions/pager/1/1/10", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
