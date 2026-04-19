package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupExport2AppCheckHandlerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE core_chart_view (
		id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, scene_id INTEGER, table_id INTEGER,
		type TEXT
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE core_dataset_group (
		id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, pid INTEGER, level INTEGER,
		node_type TEXT, type TEXT, del_flag INTEGER, create_by TEXT, create_time INTEGER,
		update_by TEXT, last_update_time INTEGER
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE core_dataset_table (
		id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, datasource_id INTEGER,
		dataset_group_id INTEGER, table_name TEXT, type TEXT
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE core_dataset_table_field (
		id INTEGER PRIMARY KEY AUTOINCREMENT, datasource_id INTEGER,
		dataset_table_id INTEGER, dataset_group_id INTEGER, chart_id INTEGER,
		origin_name TEXT, name TEXT, dataease_name TEXT, field_short_name TEXT,
		group_type TEXT, type TEXT, de_type INTEGER, de_extract_type INTEGER,
		ext_field INTEGER, checked INTEGER, params TEXT
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE core_datasource (
		id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, type TEXT, pid INTEGER,
		description TEXT, configuration TEXT, create_time INTEGER, update_time INTEGER,
		status TEXT, del_flag INTEGER
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE core_datasource_task (
		id INTEGER PRIMARY KEY AUTOINCREMENT, ds_id INTEGER, name TEXT,
		update_type TEXT, start_time INTEGER, sync_rate TEXT, cron TEXT,
		simple_cron_value INTEGER, simple_cron_type TEXT, end_limit TEXT,
		end_time INTEGER, create_time INTEGER, last_exec_time INTEGER,
		last_exec_status TEXT, extra_data TEXT, task_status TEXT
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE visualization_linkage (
		id INTEGER PRIMARY KEY AUTOINCREMENT, dv_id INTEGER, source_view_id INTEGER,
		target_view_id INTEGER, update_time INTEGER, update_people TEXT,
		linkage_active INTEGER, ext1 TEXT, ext2 TEXT, copy_from INTEGER, copy_id INTEGER
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE visualization_linkage_field (
		id INTEGER PRIMARY KEY AUTOINCREMENT, linkage_id INTEGER, source_field INTEGER,
		target_field INTEGER, update_time INTEGER, copy_from INTEGER, copy_id INTEGER
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE visualization_link_jump (
		id INTEGER PRIMARY KEY AUTOINCREMENT, source_dv_id INTEGER, source_view_id INTEGER,
		link_jump_info TEXT, checked INTEGER, copy_from INTEGER, copy_id INTEGER
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE visualization_link_jump_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT, link_jump_id INTEGER, link_type TEXT,
		jump_type TEXT, target_dv_id INTEGER, source_field_id INTEGER,
		content TEXT, checked INTEGER, attach_params INTEGER, copy_from INTEGER,
		copy_id INTEGER, window_size TEXT
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE visualization_link_jump_target_view_info (
		target_id INTEGER PRIMARY KEY AUTOINCREMENT, link_jump_info_id INTEGER,
		source_field_active_id INTEGER, target_view_id TEXT, target_field_id TEXT,
		copy_from INTEGER, copy_id INTEGER, target_type TEXT
	)`).Error)

	require.NoError(t, db.Exec(`INSERT INTO core_datasource (id, name, type, configuration, create_time, update_time) VALUES (1, 'test_ds', 'MySQL', '{}', 1, 1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_group (id, name, node_type, create_by, create_time, update_by, last_update_time) VALUES (100, 'grp1', 'dataset', 'admin', 1, 'admin', 1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_table (id, name, datasource_id, dataset_group_id, type) VALUES (200, 't1', 1, 100, 'db')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, group_type, type, de_type) VALUES (300, 100, 'f1', 'field1', 'd', 'VARCHAR', 0)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_chart_view (id, title, scene_id, table_id, type) VALUES (400, 'chart1', 500, 100, 'bar')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_datasource_task (id, ds_id, name, update_type, sync_rate, task_status) VALUES (600, 1, 'task1', 'all', '0', 'Success')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO visualization_linkage (id, dv_id, source_view_id, target_view_id, linkage_active) VALUES (700, 500, 400, 401, 1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO visualization_linkage_field (id, linkage_id, source_field, target_field) VALUES (800, 700, 300, 301)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO visualization_link_jump (id, source_dv_id, source_view_id, checked) VALUES (900, 500, 400, 1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO visualization_link_jump_info (id, link_jump_id, link_type, jump_type, source_field_id, checked) VALUES (1000, 900, 'inner', '_blank', 300, 1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO visualization_link_jump_target_view_info (target_id, link_jump_info_id, source_field_active_id, target_view_id, target_field_id) VALUES (1100, 1000, 300, '999', '888')`).Error)

	return db
}

func TestVisualizationHandler_Decompression_OuterTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	repo := repository.NewVisualizationRepository(db)
	h := NewVisualizationHandler(service.NewVisualizationService(repo))

	r := gin.New()
	r.POST("/dataVisualization/decompression", h.Decompression)

	body := `{
		"newFrom":"new_outer_template",
		"name":"Imported Panel",
		"type":"dashboard",
		"canvasStyleData":"{\"scale\":100}",
		"componentData":"[{\"id\":\"view_100\",\"component\":\"VChart\"}]",
		"dynamicData":"{\"view_100\":{\"title\":\"Sales\",\"type\":\"bar\",\"tableId\":5}}"
	}`
	req := httptest.NewRequest(http.MethodPost, "/dataVisualization/decompression", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code string                              `json:"code"`
		Data visualization.DecompressionResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, "Imported Panel", resp.Data.Name)
	assert.Equal(t, "dashboard", resp.Data.Type)
	assert.NotEmpty(t, resp.Data.ID)
	require.Len(t, resp.Data.CanvasViewInfo, 1)

	for viewID, view := range resp.Data.CanvasViewInfo {
		assert.Contains(t, resp.Data.ComponentData, viewID)
		assert.NotContains(t, resp.Data.ComponentData, "view_100")
		assert.Equal(t, "Sales", view["title"])
		assert.Equal(t, "template", view["dataFrom"])
		assert.Equal(t, float64(5), view["sourceTableId"])
		assert.Nil(t, view["tableId"])
	}
}

func TestVisualizationHandler_Decompression_OuterTemplate_APIAlias(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	repo := repository.NewVisualizationRepository(db)
	h := NewVisualizationHandler(service.NewVisualizationService(repo))

	r := gin.New()
	RegisterVisualizationRoutes(r.Group("/api"), h, nil)

	body := `{
		"newFrom":"new_outer_template",
		"name":"Imported Panel",
		"type":"dashboard",
		"canvasStyleData":"{\"scale\":100}",
		"componentData":"[{\"id\":\"view_100\",\"component\":\"VChart\"}]",
		"dynamicData":"{\"view_100\":{\"title\":\"Sales\",\"type\":\"bar\",\"tableId\":5}}"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/dataVisualization/decompression", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
}

func TestVisualizationHandler_Export2AppCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupExport2AppCheckHandlerDB(t)
	repo := repository.NewVisualizationRepository(db)
	h := NewVisualizationHandler(service.NewVisualizationService(repo))

	r := gin.New()
	r.POST("/dataVisualization/export2AppCheck", h.Export2AppCheck)

	body := `{"dvId":500,"viewIds":[400],"dsIds":[100]}`
	req := httptest.NewRequest(http.MethodPost, "/dataVisualization/export2AppCheck", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code string                                `json:"code"`
		Data visualization.Export2AppCheckResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.True(t, resp.Data.CheckStatus)
	assert.Equal(t, "success", resp.Data.CheckMes)
	assert.Len(t, resp.Data.ChartViewsInfo, 1)
	assert.Len(t, resp.Data.DatasetGroupsInfo, 1)
	assert.Len(t, resp.Data.DatasourceInfo, 1)
	assert.Len(t, resp.Data.LinkJumps, 1)
}

func TestVisualizationHandler_AppCanvasNameCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupExport2AppCheckHandlerDB(t)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_group (id, name, pid, node_type, create_by, create_time, update_by, last_update_time) VALUES (2000, 'Folder A', 10, 'folder', 'admin', 1, 'admin', 1)`).Error)

	repo := repository.NewVisualizationRepository(db)
	svc := service.NewVisualizationService(repo)
	svc.SetDatasetRepository(repository.NewDatasetRepository(db))
	h := NewVisualizationHandler(svc)

	r := gin.New()
	r.POST("/dataVisualization/appCanvasNameCheck", h.AppCanvasNameCheck)

	body := `{"datasetFolderPid":10,"datasetFolderName":"Folder A"}`
	req := httptest.NewRequest(http.MethodPost, "/dataVisualization/appCanvasNameCheck", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Code string `json:"code"`
		Data string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, "repeat", resp.Data)
}

func TestVisualizationHandler_ExportLogTemplate_NoAuditService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	repo := repository.NewVisualizationRepository(db)
	h := NewVisualizationHandler(service.NewVisualizationService(repo))

	r := gin.New()
	r.POST("/dataVisualization/exportLogTemplate", h.ExportLogTemplate)

	body := `{"id":123,"type":"dashboard"}`
	req := httptest.NewRequest(http.MethodPost, "/dataVisualization/exportLogTemplate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
