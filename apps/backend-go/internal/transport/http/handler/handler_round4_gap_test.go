package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/dataset"
	datasourcedomain "dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRound4_VisualizationHandler_ConstructorsAndHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupVisualizationHandlerUnitDB(t)
	h := NewVisualizationHandler(service.NewVisualizationService(repository.NewVisualizationRepository(db)))
	require.NotNil(t, h)

	payload := map[string]interface{}{
		"config": `{"enabled":true}`,
		"items":  `[1,2]`,
		"plain":  "text",
	}
	parseJSONStrings(payload)
	assert.IsType(t, map[string]interface{}{}, payload["config"])
	assert.IsType(t, []interface{}{}, payload["items"])
	assert.Equal(t, "text", payload["plain"])

	checkVersion := "round4"
	resp := buildEnrichedVisualizationResponse(&visualization.DataVisualizationInfo{ID: 1, Name: "viz", CheckVersion: &checkVersion}, map[string]interface{}{"a": 1})
	assert.Equal(t, int64(1), resp["id"])
	assert.Equal(t, map[string]interface{}{"a": 1}, resp["canvasViewInfo"])
}

func TestRound4_VisualizationHandler_RoutesAndSimpleHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupVisualizationHandlerUnitDB(t)
	seedVisualizationHandlerRecord(t, db, 1001, "round4-viz")
	h := NewVisualizationHandler(service.NewVisualizationService(repository.NewVisualizationRepository(db)))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(9))
		c.Set("userId", int64(9))
		c.Set("username", "viz-round4")
		c.Next()
	})
	RegisterVisualizationRoutes(r.Group("/api"), h, nil)
	r.GET("/componentInfo", h.GetComponentInfo)

	resp := performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/list", `{"type":"dashboard","current":1,"size":10}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/save", `{"name":"round4-save","type":"dashboard"}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/copy", `{"id":1001,"name":"round4-copy"}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/checkCanvasChange", `{"id":1001,"contentId":"other"}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/updateCanvas", `{"id":1001,"name":"round4-update"}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/updateBase", `{"id":1001,"name":"round4-base"}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/move", `{"id":1001,"pid":12}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/updatePublishStatus", `{"id":1001,"status":1}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/recoverToPublished", `{"id":1001}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/nameCheck", `{"name":"round4-base"}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodGet, "/api/dataVisualization/findDvType/1001", ``)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/appCanvasNameCheck", `{"datasetFolderPid":0,"datasetFolderName":"unused"}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodPost, "/api/dataVisualization/exportLogTemplate", `{"id":1001,"type":"dashboard"}`)
	assert.Equal(t, "000000", resp.Code)

	resp = performVisualizationRequest(t, r, http.MethodGet, "/componentInfo", ``)
	assert.Equal(t, "000000", resp.Code)
}

func TestRound4_DatasetHandler_ConstructorAndCoreFlows(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupDatasetHandlerGapDB(t)
	seedDatasetGroupRecord(t, db, &dataset.CoreDatasetGroup{
		ID:       901,
		Name:     "round4-folder",
		PID:      int64PtrForDatasetHandler(0),
		Level:    intPtrForDatasetHandler(0),
		NodeType: stringPtrForDatasetHandler(dataset.NodeTypeFolder),
		DelFlag:  intPtrForDatasetHandler(0),
	})
	seedDatasetDetailFixture(t, db, 902)
	repo := repository.NewDatasetRepository(db)
	h := NewDatasetHandler(service.NewDatasetService(repo), nil)
	require.NotNil(t, h)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1001))
		c.Set("userId", int64(1001))
		c.Set("username", "dataset-round4")
		c.Next()
	})
	r.POST("/tree", h.Tree)
	r.POST("/create", h.Create)
	r.POST("/save", h.Save)
	r.POST("/fields", h.Fields)
	r.POST("/preview", h.Preview)
	r.POST("/previewWithPerm", h.PreviewWithPermission)

	w := performDatasetJSONRequest(t, r, http.MethodPost, "/tree", map[string]any{})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "000000", decodeDatasetResp(t, w.Body.Bytes()).Code)

	w = performDatasetJSONRequest(t, r, http.MethodPost, "/create", map[string]any{"name": "round4-create", "nodeType": dataset.NodeTypeFolder, "pid": 0})
	assert.Equal(t, "000000", decodeDatasetResp(t, w.Body.Bytes()).Code)

	w = performDatasetJSONRequest(t, r, http.MethodPost, "/save", map[string]any{"name": "round4-save", "nodeType": dataset.NodeTypeFolder, "pid": 0})
	assert.Equal(t, "000000", decodeDatasetResp(t, w.Body.Bytes()).Code)

	w = performDatasetJSONRequest(t, r, http.MethodPost, "/fields", map[string]any{"datasetGroupId": 902})
	assert.Equal(t, "000000", decodeDatasetResp(t, w.Body.Bytes()).Code)

	w = performDatasetJSONRequest(t, r, http.MethodPost, "/preview", map[string]any{"datasetGroupId": 902, "limit": 10})
	assert.Equal(t, "000000", decodeDatasetResp(t, w.Body.Bytes()).Code)

	noAuthDSN := fmt.Sprintf("file:%s_noauth?mode=memory&cache=shared", t.Name())
	noAuthDB, err := gorm.Open(sqlite.Open(noAuthDSN), &gorm.Config{})
	require.NoError(t, err)
	noAuthSQLDB, err := noAuthDB.DB()
	require.NoError(t, err)
	noAuthSQLDB.SetMaxOpenConns(1)
	require.NoError(t, noAuthDB.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}))
	require.NoError(t, noAuthDB.Exec(`CREATE TABLE core_chart_view (id INTEGER PRIMARY KEY AUTOINCREMENT, table_id INTEGER)`).Error)
	require.NoError(t, noAuthDB.Exec(`CREATE TABLE data_perm_row (id INTEGER PRIMARY KEY AUTOINCREMENT, dataset_group_id INTEGER, expression_tree TEXT)`).Error)
	require.NoError(t, noAuthDB.Exec(`CREATE TABLE data_perm_column (id INTEGER PRIMARY KEY AUTOINCREMENT, dataset_group_id INTEGER, field_name TEXT)`).Error)
	require.NoError(t, noAuthDB.Exec(`CREATE TABLE visualization_linkage_field (id INTEGER PRIMARY KEY AUTOINCREMENT, source_field INTEGER, target_field INTEGER)`).Error)
	require.NoError(t, noAuthDB.Exec(`CREATE TABLE visualization_link_jump_info (id INTEGER PRIMARY KEY AUTOINCREMENT, source_field_id INTEGER)`).Error)
	require.NoError(t, noAuthDB.Exec(`CREATE TABLE visualization_outer_params_target_view_info (id INTEGER PRIMARY KEY AUTOINCREMENT, target_field_id TEXT)`).Error)
	require.NoError(t, noAuthDB.Exec(`CREATE TABLE dataset_orders (city TEXT, amount INTEGER)`).Error)
	noAuthRepo := repository.NewDatasetRepository(noAuthDB)
	noAuthHandler := NewDatasetHandler(service.NewDatasetService(noAuthRepo), nil)
	noAuthRouter := gin.New()
	noAuthRouter.POST("/previewWithPerm", noAuthHandler.PreviewWithPermission)
	w = performDatasetJSONRequest(t, noAuthRouter, http.MethodPost, "/previewWithPerm", map[string]any{"datasetGroupId": 901})
	assert.Equal(t, "20001", decodeDatasetResp(t, w.Body.Bytes()).Code)
}

func TestRound4_DatasetHandler_GapHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupDatasetHandlerGapDB(t)
	seedDatasetDetailFixture(t, db, 902)
	repo := repository.NewDatasetRepository(db)
	svc := service.NewDatasetService(repo)
	h := NewDatasetHandler(svc, nil)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1001))
		c.Set("userId", int64(1001))
		c.Set("username", "dataset-round4")
		c.Next()
	})
	r.POST("/get/:id", h.GetDetail)
	r.POST("/details/:id", h.Details)
	r.GET("/barInfo/:id", h.BarInfo)
	r.POST("/previewSql", h.PreviewSQL)
	r.GET("/fieldFunctions", h.GetFieldFunctions)
	r.POST("/multFieldValuesForPermissions", h.MultFieldValuesForPermissions)
	r.POST("/detailWithPerm", h.DetailWithPerm)
	r.POST("/fieldTree", h.GetFieldTree)

	for _, tc := range []struct {
		method string
		path   string
		body   any
		code   string
	}{
		{http.MethodPost, "/get/902", nil, "000000"},
		{http.MethodPost, "/details/902", nil, "000000"},
		{http.MethodGet, "/barInfo/902", nil, "000000"},
		{http.MethodPost, "/previewSql", map[string]any{"sql": ""}, "000000"},
		{http.MethodGet, "/fieldFunctions", nil, "000000"},
		{http.MethodPost, "/multFieldValuesForPermissions", map[string]any{"fieldIds": []int64{0}}, "000000"},
		{http.MethodPost, "/detailWithPerm", map[string]any{"ids": []int64{902}}, "000000"},
		{http.MethodPost, "/fieldTree", map[string]any{"fieldIds": []int64{0}}, "000000"},
	} {
		resp := decodeDatasetResp(t, performDatasetJSONRequest(t, r, tc.method, tc.path, tc.body).Body.Bytes())
		assert.Equal(t, tc.code, resp.Code, tc.path)
	}

	registered := gin.New()
	registered.Use(func(c *gin.Context) { c.Set("user_id", int64(1001)); c.Set("userId", int64(1001)); c.Next() })
	RegisterDatasetRoutes(registered.Group("/api"), h)
	resp := decodeDatasetResp(t, performDatasetJSONRequest(t, registered, http.MethodPost, "/api/dataset/fieldTree", map[string]any{"fieldIds": []int64{0}}).Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound4_SyncHandler_ConstructorAndHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r, db := setupSyncHandlerGap2(t)
	h := NewSyncHandler(service.NewSyncService(repository.NewSyncRepository(db), repository.NewDatasourceRepository(db), service.NewDatasourceService(repository.NewDatasourceRepository(db))))
	require.NotNil(t, h)

	createBy := "gap2-user"
	status := datasourcedomain.StatusSuccess
	seedSyncDatasource(t, db, &datasourcedomain.CoreDatasource{ID: 301, Name: "mysql-a", Type: "mysql", CreateBy: &createBy, Status: &status})
	seedSyncDatasource(t, db, &datasourcedomain.CoreDatasource{ID: 302, Name: "mysql-b", Type: "mysql", CreateBy: &createBy, Status: &status})
	seedSyncTask(t, db, &auto.CoreDatasourceTask{ID: 401, DsID: 301, Name: "sync-job", UpdateType: "sync", SyncRate: "1", Cron: "0 0 * * * *", TaskStatus: "Pending", CreateTime: 1000})
	seedSyncTaskLog(t, db, &auto.CoreDatasourceTaskLog{ID: 501, DsID: 301, TaskID: 0, Info: "line", CreateTime: 1000})

	for _, tc := range []struct {
		method string
		path   string
		body   string
		code   string
	}{
		{http.MethodPost, "/api/sync/datasource/latestUse/mysql", `{}`, "500000"},
		{http.MethodPost, "/api/sync/datasource/validate", `{`, "500000"},
		{http.MethodPost, "/api/sync/datasource/getSchema", `{}`, "500000"},
		{http.MethodPost, "/api/sync/datasource/save", `{`, "500000"},
		{http.MethodPost, "/api/sync/datasource/update", `{`, "500000"},
		{http.MethodPost, "/api/sync/datasource/delete/301", ``, "000000"},
		{http.MethodPost, "/api/sync/datasource/fields", `{`, "500000"},
		{http.MethodGet, "/api/sync/datasource/list/mysql", ``, "000000"},
		{http.MethodGet, "/api/sync/task/start/401", ``, "000000"},
		{http.MethodGet, "/api/sync/task/stop/401", ``, "000000"},
		{http.MethodPost, "/api/sync/task/log/terminationTask/501", ``, "000000"},
		{http.MethodGet, "/api/sync/summary/resourceCount", ``, "000000"},
		{http.MethodPost, "/api/sync/summary/logChartData", `{}`, "500000"},
	} {
		resp := decodeSyncGapResp(t, performSyncGapRequest(t, r, tc.method, tc.path, tc.body).Body.Bytes())
		assert.Equal(t, tc.code, resp.Code, tc.path)
	}

	var task auto.CoreDatasourceTask
	require.NoError(t, db.First(&task, 401).Error)
	assert.Equal(t, "cancelled", task.TaskStatus)
}

func TestRound4_SyncHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterSyncRoutes(r.Group("/api"), NewSyncHandler(nil))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sync/summary/resourceCount", nil)
	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}
