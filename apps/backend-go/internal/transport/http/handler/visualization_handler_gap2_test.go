package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"dataease/backend/internal/domain/visualization"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisualizationHandler_FindByID_EnrichesCanvasViewInfo(t *testing.T) {
	t.Parallel()

	db := setupVisualizationHandlerUnitDB(t)
	canvasStyle := `{"bg":"white"}`
	componentData := `[{"id":"101"}]`
	typ := "dashboard"
	nodeType := "leaf"
	status := 1
	createBy := "tester"
	contentID := "content-1"
	checkVersion := "v1"
	createTime := int64(111)
	updateTime := int64(222)
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{
		ID:              101,
		Name:            "Canvas A",
		Type:            &typ,
		NodeType:        &nodeType,
		Status:          &status,
		CreateBy:        &createBy,
		UpdateBy:        &createBy,
		CanvasStyleData: &canvasStyle,
		ComponentData:   &componentData,
		ContentID:       &contentID,
		CheckVersion:    &checkVersion,
		CreateTime:      &createTime,
		UpdateTime:      &updateTime,
	}).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_chart_view (id, title, scene_id, table_id, type, x_axis) VALUES (101, 'Chart A', 101, 9, 'bar', '{"foo":"bar"}')`).Error)

	router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/findById", h.FindByID) })
	resp := performVisualizationRequest(t, router, http.MethodPost, "/findById", `{"id":101}`)
	require.Equal(t, "000000", resp.Code)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, float64(101), data["id"])
	assert.Equal(t, "Canvas A", data["name"])
	assert.Equal(t, true, data["selfWatermarkStatus"])
	canvasViewInfo, ok := data["canvasViewInfo"].(map[string]interface{})
	require.True(t, ok)
	chartView, ok := canvasViewInfo["101"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Chart A", chartView["title"])
	assert.IsType(t, map[string]interface{}{}, chartView["xAxis"])
}

func TestVisualizationHandler_FindDvType_And_UpdateCheckVersion(t *testing.T) {
	t.Parallel()

	db := setupVisualizationHandlerUnitDB(t)
	typ := "dataV"
	nodeType := "leaf"
	status := 1
	createBy := "tester"
	checkVersion := "check-v2"
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{ID: 201, Name: "Screen", Type: &typ, NodeType: &nodeType, Status: &status, CreateBy: &createBy, UpdateBy: &createBy, CheckVersion: &checkVersion}).Error)
	router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) {
		r.GET("/findDvType/:id", h.FindDvType)
		r.GET("/updateCheckVersion/:id", h.UpdateCheckVersion)
	})

	findTypeResp := performVisualizationRequest(t, router, http.MethodGet, "/findDvType/201", ``)
	require.Equal(t, "000000", findTypeResp.Code)
	assert.Equal(t, `"dataV"`, string(findTypeResp.Data))

	checkVersionResp := performVisualizationRequest(t, router, http.MethodGet, "/updateCheckVersion/201", ``)
	require.Equal(t, "000000", checkVersionResp.Code)
	assert.Equal(t, `"check-v2"`, string(checkVersionResp.Data))
}

func TestVisualizationHandler_CheckCanvasChange_UpdateBase_And_RecoverToPublished(t *testing.T) {
	t.Parallel()

	db := setupVisualizationHandlerUnitDB(t)
	contentID := "content-a"
	checkVersion := "v1"
	nodeType := "leaf"
	typ := "dashboard"
	status := 0
	createBy := "tester"
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{ID: 301, Name: "Draft", Type: &typ, NodeType: &nodeType, Status: &status, CreateBy: &createBy, UpdateBy: &createBy, ContentID: &contentID, CheckVersion: &checkVersion}).Error)
	router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) {
		r.POST("/checkCanvasChange", h.CheckCanvasChange)
		r.POST("/updateBase", h.UpdateBase)
		r.POST("/recoverToPublished", h.RecoverToPublished)
	})

	changeResp := performVisualizationRequest(t, router, http.MethodPost, "/checkCanvasChange", `{"id":301,"contentId":"other"}`)
	require.Equal(t, "000000", changeResp.Code)
	assert.Equal(t, `"Repeat"`, string(changeResp.Data))

	updateResp := performVisualizationRequest(t, router, http.MethodPost, "/updateBase", `{"id":301,"name":"Renamed Draft"}`)
	require.Equal(t, "000000", updateResp.Code)
	assert.Empty(t, string(updateResp.Data))

	recoverResp := performVisualizationRequest(t, router, http.MethodPost, "/recoverToPublished", `{"id":301}`)
	require.Equal(t, "000000", recoverResp.Code)
	var item visualization.DataVisualizationInfo
	require.NoError(t, db.First(&item, 301).Error)
	assert.Equal(t, "Renamed Draft", item.Name)
	require.NotNil(t, item.Status)
	assert.Equal(t, 1, *item.Status)
}

func TestVisualizationHandler_FindCopyResource_ViewDetailList_ComponentInfo_AndExportLogs(t *testing.T) {
	t.Parallel()

	db := setupVisualizationHandlerUnitDB(t)
	pid := int64(-1)
	nodeType := "leaf"
	typ := "dashboard"
	status := 1
	createBy := "tester"
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{ID: 401, Name: "Template Canvas", PID: &pid, Type: &typ, NodeType: &nodeType, Status: &status, CreateBy: &createBy, UpdateBy: &createBy}).Error)
	require.NoError(t, db.Exec(`INSERT INTO snapshot_core_chart_view (id, title, scene_id, table_id, type) VALUES (501, 'Snapshot Chart', 401, 3, 'line')`).Error)

	router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) {
		r.GET("/findCopyResource/:dvId/:busiFlag", h.FindCopyResource)
		r.GET("/viewDetailList/:dvId", h.ViewDetailList)
		r.GET("/componentInfo", h.GetComponentInfo)
		r.POST("/exportLogApp", h.ExportLogApp)
		r.POST("/exportLogPDF", h.ExportLogPDF)
		r.POST("/exportLogImg", h.ExportLogImg)
	})

	copyResp := performVisualizationRequest(t, router, http.MethodGet, "/findCopyResource/401/dashboard", ``)
	require.Equal(t, "000000", copyResp.Code)
	assert.Contains(t, string(copyResp.Data), `"Template Canvas"`)

	viewResp := performVisualizationRequest(t, router, http.MethodGet, "/viewDetailList/401", ``)
	require.Equal(t, "000000", viewResp.Code)
	assert.Contains(t, string(viewResp.Data), `"Snapshot Chart"`)

	componentResp := performVisualizationRequest(t, router, http.MethodGet, "/componentInfo", ``)
	require.Equal(t, "000000", componentResp.Code)
	assert.Empty(t, string(componentResp.Data))

	for _, path := range []string{"/exportLogApp", "/exportLogPDF", "/exportLogImg"} {
		resp := performVisualizationRequest(t, router, http.MethodPost, path, `{"id":401,"type":"dashboard"}`)
		require.Equal(t, "000000", resp.Code)
		assert.Empty(t, string(resp.Data))
	}
}
