package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =====================================================================
// ThresholdHandler remaining branches
// =====================================================================

func TestRound9C2_Threshold_SwitchEnable_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"enable":true}`)
	h.SwitchEnable(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Threshold_BatchReci_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"ids":[1]}`)
	h.BatchReci(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Threshold_DeleteWithChart_InvalidResourceTable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewThresholdHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "chartId", Value: "1"}, {Key: "resourceTable", Value: ""}}
	h.DeleteWithChart(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// =====================================================================
// StoreHandler remaining branches
// =====================================================================

func TestRound9C2_Store_Execute_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStoreHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Execute(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9C2_Store_Execute_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStoreHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"type":"PANEL"}`)
	h.Execute(c)
	assertCode(t, w, response.CodeUnauthorized)
}

func TestRound9C2_Store_Execute_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStoreHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"type":"PANEL"}`)
	c.Set("userId", int64(1))
	h.Execute(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Store_Favorited_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStoreHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Favorited(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_Store_Favorited_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStoreHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("userId", int64(1))
	h.Favorited(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Store_Query_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStoreHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Query(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9C2_Store_Query_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStoreHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"type":"PANEL"}`)
	h.Query(c)
	assertCode(t, w, response.CodeSuccess)
}

func TestRound9C2_Store_Query_NilRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStoreHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"type":"PANEL"}`)
	c.Set("userId", int64(1))
	h.Query(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Store_resourceTypeFromString(t *testing.T) {
	assert.Equal(t, int32(1), resourceTypeFromString("PANEL"))
	assert.Equal(t, int32(1), resourceTypeFromString("DATA_VIZ"))
	assert.Equal(t, int32(2), resourceTypeFromString("SCREEN"))
	assert.Equal(t, int32(2), resourceTypeFromString("DATA_VIZ_SCREEN"))
	assert.Equal(t, int32(3), resourceTypeFromString("REPORT"))
	assert.Equal(t, int32(1), resourceTypeFromString("unknown"))
}

// =====================================================================
// OuterParamsHandler remaining branches
// =====================================================================

func TestRound9C2_OuterParams_QueryWithVisualizationId_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOuterParamsHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "1"}}
	h.QueryWithVisualizationId(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_OuterParams_GetOuterParamsInfo_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOuterParamsHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "1"}}
	h.GetOuterParamsInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_OuterParams_QueryDsWithVisualizationId_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOuterParamsHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dvId", Value: "1"}}
	h.QueryDsWithVisualizationId(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// GeoHandler remaining branches
// =====================================================================

func TestRound9C2_Geo_ListAreas_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListAreas(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Geo_GetArea_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetArea(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Geo_Save_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewGeoHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Request.Header.Del("Content-Type")
	h.Save(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// ChartHandler remaining branches
// =====================================================================

func TestRound9C2_Chart_GetFieldData_InvalidFieldID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewChartHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "fieldId", Value: "bad"}, {Key: "fieldType", Value: "text"}}
	h.GetFieldData(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Chart_GetFieldData_NilDatasetSvc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewChartHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "fieldId", Value: "1"}, {Key: "fieldType", Value: "text"}}
	h.GetFieldData(c)
	assertCode(t, w, response.CodeSuccess)
}

func TestRound9C2_Chart_GetDrillFieldData_InvalidFieldID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewChartHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "fieldId", Value: "bad"}}
	h.GetDrillFieldData(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Chart_GetDrillFieldData_NilDatasetSvc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewChartHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "fieldId", Value: "1"}}
	h.GetDrillFieldData(c)
	assertCode(t, w, response.CodeSuccess)
}

func TestRound9C2_Chart_InnerExportDetails_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewChartHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.InnerExportDetails(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Chart_InnerExportDetails_NilExportSvc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewChartHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"viewName":"test"}`)
	h.InnerExportDetails(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// TicketHandler remaining branches
// =====================================================================

func TestRound9C2_Ticket_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTicketHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Create(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Ticket_Create_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTicketHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"uuid":"test","ticket":"t1"}`)
	h.Create(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Ticket_Delete_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTicketHandler(nil)
	w, c := newRound9Ctx(t, http.MethodDelete, "/", "")
	c.Params = gin.Params{{Key: "id", Value: ""}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Ticket_Delete_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTicketHandler(nil)
	w, c := newRound9Ctx(t, http.MethodDelete, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "t1"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// ShareHandler remaining branches
// =====================================================================

func TestRound9C2_Share_Switcher_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewShareHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Switcher(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_Share_Switcher_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewShareHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"autoCreate":true}`)
	h.Switcher(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_Share_Revoke_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewShareHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.Revoke(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// RelationHandler remaining branches
// =====================================================================

func TestRound9C2_Relation_GetDatasetRelationship_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1}`)
	h.GetDatasetRelationship(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Relation_GetDatasetRelationship_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.GetDatasetRelationship(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_Relation_GetPanelRelationship_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1}`)
	h.GetPanelRelationship(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Relation_GetPanelRelationship_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRelationHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.GetPanelRelationship(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// =====================================================================
// MsgCenterHandler remaining branches
// =====================================================================

func TestRound9C2_MsgCenter_Read_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMsgCenterHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "msgId", Value: "1"}}
	h.Read(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_MsgCenter_ReadBatch_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMsgCenterHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.ReadBatch(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// EngineHandler remaining branches
// =====================================================================

func TestRound9C2_Engine_GetEngine_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewEngineHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.GetEngine(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Engine_SupportSetKey_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewEngineHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.SupportSetKey(c)
	resp := parseRound9Resp(t, w)
	// SupportSetKey() never dereferences the receiver, so nil service returns success
	assert.Equal(t, response.CodeSuccess, resp["code"])
}

// =====================================================================
// EmbeddedHandler remaining branches
// =====================================================================

func TestRound9C2_Embedded_BatchDelete_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewEmbeddedHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.BatchDelete(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Embedded_BatchDelete_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewEmbeddedHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `["app1","app2"]`)
	h.BatchDelete(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Embedded_DomainList_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewEmbeddedHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.DomainList(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// DataPermissionHandler remaining branches
// =====================================================================

func TestRound9C2_DataPermission_RowPermissionPager_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.RowPermissionPager(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_DataPermission_RowPermissionPager_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.RowPermissionPager(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_DataPermission_ColumnPermissionPager_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.ColumnPermissionPager(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_DataPermission_ColumnPermissionPager_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataPermissionHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.ColumnPermissionPager(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// WatermarkHandler remaining branches
// =====================================================================

func TestRound9C2_Watermark_getAuthenticatedUserID_NoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	result := getAuthenticatedUserID(c)
	assert.Equal(t, uint64(0), result)
}

func TestRound9C2_Watermark_getAuthenticatedUserID_Uint64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", uint64(42))
	result := getAuthenticatedUserID(c)
	assert.Equal(t, uint64(42), result)
}

func TestRound9C2_Watermark_getAuthenticatedUserID_Int64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", int64(42))
	result := getAuthenticatedUserID(c)
	assert.Equal(t, uint64(42), result)
}

func TestRound9C2_Watermark_getAuthenticatedUserID_Int64Negative(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", int64(-1))
	result := getAuthenticatedUserID(c)
	assert.Equal(t, uint64(0), result)
}

func TestRound9C2_Watermark_getAuthenticatedUserID_Uint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", uint(42))
	result := getAuthenticatedUserID(c)
	assert.Equal(t, uint64(42), result)
}

func TestRound9C2_Watermark_getAuthenticatedUserID_Int(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", 42)
	result := getAuthenticatedUserID(c)
	assert.Equal(t, uint64(42), result)
}

func TestRound9C2_Watermark_getAuthenticatedUserID_IntNegative(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", -1)
	result := getAuthenticatedUserID(c)
	assert.Equal(t, uint64(0), result)
}

func TestRound9C2_Watermark_getAuthenticatedUserID_String(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", "42")
	result := getAuthenticatedUserID(c)
	assert.Equal(t, uint64(42), result)
}

func TestRound9C2_Watermark_getAuthenticatedUserID_InvalidString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Set("userId", "bad")
	result := getAuthenticatedUserID(c)
	assert.Equal(t, uint64(0), result)
}

// =====================================================================
// RoleMenuHandler remaining branches
// =====================================================================

func TestRound9C2_RoleMenu_SaveRoleMenuAuth_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleMenuHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveRoleMenuAuth(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_RoleMenu_SaveRoleMenuAuth_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRoleMenuHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"roleId":1,"menuIds":[1,2]}`)
	h.SaveRoleMenuAuth(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// ResourceGovernanceHandler remaining branches
// =====================================================================

func TestRound9C2_ResourceGovernance_BackfillResources_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewResourceGovernanceHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.BackfillResources(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_ResourceGovernance_BackfillResources_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewResourceGovernanceHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"resourceType":"chart"}`)
	h.BackfillResources(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// MapHandler remaining branches
// =====================================================================

func TestRound9C2_Map_GetWorldTree_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMapHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.GetWorldTree(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// LinkageHandler remaining branches
// =====================================================================

func TestRound9C2_Linkage_RemoveLinkage_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkageHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.RemoveLinkage(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Linkage_RemoveLinkage_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLinkageHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"sourceDvId":"1","sourceViewId":"1"}`)
	h.RemoveLinkage(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// LicenseHandler remaining branches
// =====================================================================

func TestRound9C2_License_Revert_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLicenseHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.Revert(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// FrontendCompatHandler remaining branches
// =====================================================================

func TestRound9C2_FrontendCompat_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewFrontendCompatHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	assert.NotNil(t, h)
}

// =====================================================================
// StaticHandler remaining branches
// =====================================================================

func TestRound9C2_Static_DefaultFont_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.DefaultFont(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Static_ListResources_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListResources(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Static_ListStores_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListStores(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Static_ListTypefaces_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListTypefaces(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Static_ListFont_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListFont(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Static_Upload_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "fileId", Value: "test"}}
	c.Request.Header.Del("Content-Type")
	h.Upload(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// DataFillingUserTaskHandler remaining branches
// =====================================================================

func TestRound9C2_UserTask_UserTaskList_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UserTaskList(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_UserTask_UserTaskList_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	h.UserTaskList(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_UserTask_UserTaskTodoCount_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.UserTaskTodoCount(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_UserTask_UserTaskSave_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UserTaskSave(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_UserTask_UserTaskSave_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"formId":1,"taskId":1}`)
	h.UserTaskSave(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_UserTask_UserTaskAppend_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UserTaskAppend(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_UserTask_UserTaskAppend_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"formId":1,"taskId":1}`)
	h.UserTaskAppend(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_UserTask_UserTaskDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.UserTaskDelete(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_UserTask_UserTaskDelete_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.UserTaskDelete(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_UserTask_UserTaskConfirmUpload_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.UserTaskConfirmUpload(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_UserTask_UserTaskConfirmUpload_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserTaskHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.UserTaskConfirmUpload(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// =====================================================================
// DataFillingHandler remaining branches
// =====================================================================

func TestRound9C2_DataFilling_ListDatasourceList_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListDatasourceList(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_DataFilling_ListDatasourceListAll_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListDatasourceListAll(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_DataFilling_GetBuiltInTables_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.GetBuiltInTables(c)
	assertCode(t, w, response.CodeSuccess)
}

// =====================================================================
// DatasetHandler remaining branches
// =====================================================================

func TestRound9C2_Dataset_Rename_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Rename(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9C2_Dataset_PreviewWithPermission_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"datasetGroupId":1}`)
	round9SetUser(c, 1)
	h.PreviewWithPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9C2_Dataset_EnumValueObj_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"queryId":1}`)
	h.EnumValueObj(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// =====================================================================
// Route registration tests for remaining handlers
// =====================================================================

func TestRound9C2_Map_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api")
	h := NewMapHandler(nil)
	RegisterMapRoutes(group, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/map/worldTree", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound9C2_Engine_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api")
	h := NewEngineHandler(nil)
	RegisterEngineRoutes(group, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/engine/getEngine", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound9C2_Ticket_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTicketHandler(nil)
	RegisterTicketRoutes(r, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ticket/validate/test-ticket", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound9C2_Geo_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api")
	h := NewGeoHandler(nil)
	RegisterGeoRoutes(group, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/geometry/areaList", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound9C2_Store_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewStoreHandler(nil)
	RegisterStoreRoutes(r, h, false)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/store/query", bytes.NewBufferString(`{"type":"PANEL"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound9C2_Watermark_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewWatermarkHandler(nil)
	RegisterWatermarkRoutes(r, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/watermark/find", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =====================================================================
// FrontendCompat menu helpers
// =====================================================================

func TestRound9C2_FrontendCompat_safeName(t *testing.T) {
	assert.Equal(t, "test", safeName("test", "/path"))
	assert.Equal(t, "", safeName("", ""))
}

// =====================================================================
// CompatibilityBridge datasource routes
// =====================================================================

func TestRound9C2_CompatBridge_createFolderCompat_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasourceHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.createFolderCompat(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}
