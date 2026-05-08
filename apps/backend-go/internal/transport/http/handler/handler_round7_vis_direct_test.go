package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ---------- DB setup helpers ----------

func setupRound7VisDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:round7vis_" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(&visualization.DataVisualizationInfo{}))

	statements := []string{
		`CREATE TABLE IF NOT EXISTS core_opt_recent (uid INTEGER, resource_id INTEGER, time INTEGER)`,
		`CREATE TABLE IF NOT EXISTS core_store (uid INTEGER, resource_id INTEGER)`,
		`CREATE TABLE IF NOT EXISTS core_dataset_group (id INTEGER PRIMARY KEY, name TEXT, pid INTEGER, node_type TEXT, create_by TEXT, del_flag INTEGER)`,
		`CREATE TABLE IF NOT EXISTS core_datasource (id INTEGER PRIMARY KEY, name TEXT, type TEXT, create_by TEXT)`,
		`CREATE TABLE IF NOT EXISTS core_chart_view (id INTEGER PRIMARY KEY, title TEXT, scene_id INTEGER, table_id INTEGER, type TEXT, render TEXT, result_count INTEGER, result_mode TEXT, x_axis TEXT, x_axis_ext TEXT, y_axis TEXT, y_axis_ext TEXT, ext_stack TEXT, ext_bubble TEXT, ext_label TEXT, ext_tooltip TEXT, custom_attr TEXT, custom_attr_mobile TEXT, custom_style TEXT, custom_style_mobile TEXT, custom_filter TEXT, drill_fields TEXT, senior TEXT, create_by TEXT, create_time INTEGER, update_time INTEGER, snapshot TEXT, style_priority INTEGER, chart_type TEXT, is_plugin INTEGER, data_from TEXT, view_fields TEXT, refresh_view_enable INTEGER, refresh_unit TEXT, refresh_time INTEGER, linkage_active INTEGER, jump_active INTEGER, copy_from INTEGER, copy_id INTEGER, aggregate INTEGER, flow_map_start_name TEXT, flow_map_end_name TEXT, ext_color TEXT, sort_priority TEXT)`,
		`CREATE TABLE IF NOT EXISTS snapshot_core_chart_view (id INTEGER PRIMARY KEY, title TEXT, scene_id INTEGER, table_id INTEGER, type TEXT, render TEXT, result_count INTEGER, result_mode TEXT, x_axis TEXT, x_axis_ext TEXT, y_axis TEXT, y_axis_ext TEXT, ext_stack TEXT, ext_bubble TEXT, ext_label TEXT, ext_tooltip TEXT, custom_attr TEXT, custom_attr_mobile TEXT, custom_style TEXT, custom_style_mobile TEXT, custom_filter TEXT, drill_fields TEXT, senior TEXT, create_by TEXT, create_time INTEGER, update_time INTEGER, snapshot TEXT, style_priority INTEGER, chart_type TEXT, is_plugin INTEGER, data_from TEXT, view_fields TEXT, refresh_view_enable INTEGER, refresh_unit TEXT, refresh_time INTEGER, linkage_active INTEGER, jump_active INTEGER, copy_from INTEGER, copy_id INTEGER, aggregate INTEGER, flow_map_start_name TEXT, flow_map_end_name TEXT, ext_color TEXT, sort_priority TEXT)`,
		`CREATE TABLE IF NOT EXISTS visualization_subject (id INTEGER PRIMARY KEY, name TEXT, type TEXT, details TEXT, create_time INTEGER, update_time INTEGER, create_by TEXT, update_by TEXT)`,
		`CREATE TABLE IF NOT EXISTS visualization_background (id INTEGER PRIMARY KEY, name TEXT, classification TEXT, url TEXT, content TEXT)`,
	}
	for _, s := range statements {
		require.NoError(t, db.Exec(s).Error)
	}

	t.Cleanup(func() {
		sqlDB, dbErr := db.DB()
		require.NoError(t, dbErr)
		require.NoError(t, sqlDB.Close())
	})

	return db
}

func newRound7VisHandler(db *gorm.DB) *VisualizationHandler {
	repo := repository.NewVisualizationRepository(db)
	return NewVisualizationHandler(service.NewVisualizationService(repo))
}

func newRound7VisCtx(t *testing.T, method, path, body string) (*httptest.ResponseRecorder, *gin.Context) {
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

func seedRound7VisRecord(t *testing.T, db *gorm.DB, id int64, name string) {
	t.Helper()
	dashType := "dashboard"
	leafType := "leaf"
	status := 0
	createBy := "tester"
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{
		ID: id, Name: name, Type: &dashType, NodeType: &leafType, Status: &status, CreateBy: &createBy, UpdateBy: &createBy,
	}).Error)
}

// ---------- Tests ----------

func TestRound7VisDirect_NewVisualizationHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)
	assert.NotNil(t, h)
	assert.NotNil(t, h.service)
}

func TestRound7VisDirect_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	seedRound7VisRecord(t, db, 101, "ListPanel")
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"type":"dashboard","current":1,"size":10}`)
	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string          `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_SaveCanvas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"name":"New Canvas","type":"dashboard"}`)
	c.Set(middleware.ContextUserID, int64(1))
	h.SaveCanvas(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_FindDvType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	seedRound7VisRecord(t, db, 201, "DvTypePanel")
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "201"}}
	h.FindDvType(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_FindDvType_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.FindDvType(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7VisDirect_Copy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	seedRound7VisRecord(t, db, 301, "CopySource")
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":301,"name":"Copied"}`)
	c.Set(middleware.ContextUserID, int64(1))
	h.Copy(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_CheckCanvasChange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":1}`)
	h.CheckCanvasChange(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound7VisDirect_UpdateCanvas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	seedRound7VisRecord(t, db, 401, "BeforeCanvas")
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":401,"name":"AfterCanvas"}`)
	c.Set(middleware.ContextUserID, int64(1))
	h.UpdateCanvas(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_UpdateBase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	seedRound7VisRecord(t, db, 501, "BeforeBase")
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":501,"name":"AfterBase"}`)
	c.Set(middleware.ContextUserID, int64(1))
	h.UpdateBase(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_Move(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	seedRound7VisRecord(t, db, 601, "Movable")
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":601,"pid":99}`)
	c.Set(middleware.ContextUserID, int64(1))
	h.Move(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_UpdatePublishStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	seedRound7VisRecord(t, db, 701, "Publishable")
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":701,"status":1}`)
	c.Set(middleware.ContextUserID, int64(1))
	h.UpdatePublishStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string          `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_DeleteLogic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	seedRound7VisRecord(t, db, 801, "Deletable")
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "801"}}
	c.Set(middleware.ContextUserID, int64(1))
	h.DeleteLogic(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_DeleteLogic_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	h.DeleteLogic(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7VisDirect_NameCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	seedRound7VisRecord(t, db, 901, "TakenName")
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"name":"TakenName"}`)
	h.NameCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_AppCanvasNameCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_group (id, name, node_type, create_by) VALUES (2001, 'FolderA', 'folder', 'admin')`).Error)
	h := newRound7VisHandler(db)
	svc := service.NewVisualizationService(repository.NewVisualizationRepository(db))
	svc.SetDatasetRepository(repository.NewDatasetRepository(db))
	h = NewVisualizationHandler(svc)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"datasetFolderPid":10,"datasetFolderName":"FolderA"}`)
	h.AppCanvasNameCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_GetComponentInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodGet, "/", "")
	h.GetComponentInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7VisDirect_Export2AppCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"dvId":1}`)
	h.Export2AppCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound7VisDirect_ExportLogApp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":1,"type":"dashboard"}`)
	c.Set(middleware.ContextUserID, int64(1))
	c.Set(middleware.ContextUserName, "admin")
	h.ExportLogApp(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound7VisDirect_ExportLogTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":1,"type":"dashboard"}`)
	c.Set(middleware.ContextUserID, int64(1))
	h.ExportLogTemplate(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound7VisDirect_ExportLogPDF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":1,"type":"dashboard"}`)
	c.Set(middleware.ContextUserID, int64(1))
	h.ExportLogPDF(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound7VisDirect_ExportLogImg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	w, c := newRound7VisCtx(t, http.MethodPost, "/", `{"id":1,"type":"dashboard"}`)
	c.Set(middleware.ContextUserID, int64(1))
	h.ExportLogImg(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound7VisDirect_Decompression(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	body := `{"newFrom":"new_outer_template","name":"DecompPanel","type":"dashboard","canvasStyleData":"{\"scale\":100}","componentData":"[]","dynamicData":"{}"}`
	w, c := newRound7VisCtx(t, http.MethodPost, "/", body)
	h.Decompression(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string      `json:"code"`
		Data interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

// ---------- Pure helpers ----------

func TestRound7VisDirect_parseJSONStrings(t *testing.T) {
	m := map[string]interface{}{
		"plain":    "hello",
		"obj":      `{"key":"val"}`,
		"arr":      `[1,2,3]`,
		"notJson":  `{broken`,
		"emptyStr": "",
		"num":      float64(42),
	}
	parseJSONStrings(m)

	assert.Equal(t, "hello", m["plain"])
	assert.Equal(t, map[string]interface{}{"key": "val"}, m["obj"])
	assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, m["arr"])
	assert.Equal(t, `{broken`, m["notJson"]) // unchanged
	assert.Equal(t, "", m["emptyStr"])
	assert.Equal(t, float64(42), m["num"])
}

func TestRound7VisDirect_resolveBusiTypes(t *testing.T) {
	tests := []struct {
		input string
		want  []string
		err   bool
	}{
		{"dashboard", []string{"dashboard"}, false},
		{"dataV", []string{"dataV"}, false},
		{"panel", []string{"dashboard"}, false},
		{"screen", []string{"dataV"}, false},
		{"", []string{"dashboard", "dataV"}, false},
		{"dashboard-dataV", []string{"dashboard", "dataV"}, false},
		{"unknown", nil, true},
	}
	for _, tt := range tests {
		got, err := resolveBusiTypes(tt.input)
		if tt.err {
			assert.Error(t, err, "input=%q", tt.input)
		} else {
			assert.NoError(t, err, "input=%q", tt.input)
			assert.Equal(t, tt.want, got, "input=%q", tt.input)
		}
	}
}

func TestRound7VisDirect_visualizationExtraFlag(t *testing.T) {
	t1 := true
	assert.Equal(t, 1, visualizationExtraFlag(&t1))
	f1 := false
	assert.Equal(t, 0, visualizationExtraFlag(&f1))
	assert.Equal(t, 0, visualizationExtraFlag(nil))
}

func TestRound7VisDirect_visualizationPublishFlag(t *testing.T) {
	s1 := 1
	assert.Equal(t, 1, visualizationPublishFlag(&s1))
	s0 := 0
	assert.Equal(t, 0, visualizationPublishFlag(&s0))
	assert.Equal(t, 0, visualizationPublishFlag(nil))
}

func TestRound7VisDirect_buildEnrichedVisualizationResponse(t *testing.T) {
	name := "TestViz"
	pid := int64(5)
	status := 1
	vType := "dashboard"
	nodeType := "leaf"
	componentData := `[{"id":"c1"}]`
	canvasStyleData := `{"scale":100}`
	mobileLayout := true
	version := 3
	contentID := "cid123"
	createTime := int64(1000)
	updateTime := int64(2000)

	v := &visualization.DataVisualizationInfo{
		ID: 42, Name: name, PID: &pid, Status: &status, Type: &vType,
		NodeType: &nodeType, ComponentData: &componentData, CanvasStyleData: &canvasStyleData,
		MobileLayout: &mobileLayout, Version: &version, ContentID: &contentID,
		CreateTime: &createTime, UpdateTime: &updateTime,
	}

	canvasViewInfo := map[string]interface{}{"v1": map[string]interface{}{"title": "Chart1"}}
	result := buildEnrichedVisualizationResponse(v, canvasViewInfo)

	assert.Equal(t, int64(42), result["id"])
	assert.Equal(t, name, result["name"])
	assert.Equal(t, &pid, result["pid"])
	assert.Equal(t, &status, result["status"])
	assert.Equal(t, &vType, result["type"])
	assert.Equal(t, true, result["mobileLayout"])
	assert.Equal(t, &version, result["version"])
	assert.Equal(t, &contentID, result["contentId"])
	assert.Equal(t, true, result["selfWatermarkStatus"])
	assert.Equal(t, 9, result["weight"])
	assert.Equal(t, canvasViewInfo, result["canvasViewInfo"])
	assert.Equal(t, int64(1000), result["createTime"])
	assert.Equal(t, int64(2000), result["updateTime"])
}

func TestRound7VisDirect_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRound7VisDB(t)
	h := newRound7VisHandler(db)

	tests := []struct {
		name   string
		method func(c *gin.Context)
	}{
		{"List", h.List},
		{"SaveCanvas", h.SaveCanvas},
		{"Copy", h.Copy},
		{"CheckCanvasChange", h.CheckCanvasChange},
		{"UpdateCanvas", h.UpdateCanvas},
		{"UpdateBase", h.UpdateBase},
		{"Move", h.Move},
		{"UpdatePublishStatus", h.UpdatePublishStatus},
		{"NameCheck", h.NameCheck},
		{"Export2AppCheck", h.Export2AppCheck},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newRound7VisCtx(t, http.MethodPost, "/", `{invalid`)
			tt.method(c)
			assert.Equal(t, http.StatusOK, w.Code)
			var resp struct {
				Code string `json:"code"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, "500000", resp.Code)
		})
	}
}
