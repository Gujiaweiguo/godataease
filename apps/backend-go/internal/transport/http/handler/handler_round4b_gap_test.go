package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/template"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRound4B_DatasourceHandler_CoreAndHelperCoverage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := setupDatasourceHandlerTestEnv(t)
	h := NewDatasourceHandler(service.NewDatasourceService(repository.NewDatasourceRepository(env.db)))
	require.NotNil(t, h)

	config := `{}`
	status := datasource.StatusSuccess
	createBy := "tester"
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 211, PID: int64PtrForDatasourceHandler(0), Name: "folder-ds", Type: datasource.TypeFolder, Configuration: &config, Status: &status, CreateBy: &createBy})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 212, PID: int64PtrForDatasourceHandler(0), Name: "mysql-ds", Type: "MySQL", Configuration: &config, Status: &status, CreateBy: &createBy})

	t.Run("list invalid json", func(t *testing.T) {
		w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/list", []byte(`{`))
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("validate success", func(t *testing.T) {
		w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/validate", map[string]any{"datasourceId": 211})
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		assert.Contains(t, string(resp.Data), `"Success"`)
	})

	t.Run("validate by id invalid", func(t *testing.T) {
		w := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/validate/bad", nil)
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("tree invalid json", func(t *testing.T) {
		w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/tree", []byte(`{`))
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("get hidepw simple success", func(t *testing.T) {
		for _, path := range []string{"/api/ds/212", "/api/ds/hidePw/212", "/api/ds/simple/212"} {
			w := performDatasourceJSONRequest(t, env.r, http.MethodGet, path, nil)
			resp := decodeDatasourceResp(t, w.Body.Bytes())
			assert.Equal(t, "000000", resp.Code, path)
		}
	})

	t.Run("per delete success", func(t *testing.T) {
		w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/perDelete/212", nil)
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		assert.Equal(t, "false", string(resp.Data))
	})

	t.Run("sync api table invalid json", func(t *testing.T) {
		w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/syncApiTable", []byte(`{`))
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("sync api ds missing grpc config", func(t *testing.T) {
		w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/syncApiDs", map[string]any{"datasourceId": "212"})
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "500000", resp.Code)
		assert.Contains(t, resp.Msg, "seatunnel grpc address")
	})

	t.Run("load remote file invalid json", func(t *testing.T) {
		w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/loadRemoteFile", []byte(`{`))
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("check api datasource success", func(t *testing.T) {
		payload := base64.StdEncoding.EncodeToString([]byte(`{"name":"orders"}`))
		w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/checkApiDatasource", map[string]any{"data": payload, "type": "apiStructure"})
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		assert.Contains(t, string(resp.Data), `"showApiStructure":true`)
		assert.Contains(t, string(resp.Data), `"type":"table"`)
	})

	t.Run("types and sync record", func(t *testing.T) {
		seedDatasourceTaskLogRecord(t, env.db, &auto.CoreDatasourceTaskLog{ID: 401, DsID: 212, TaskID: 2, TaskStatus: "completed", PhysicalTableName: "orders", CreateTime: 1, EndTime: 2})

		typesW := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/types", nil)
		typesResp := decodeDatasourceResp(t, typesW.Body.Bytes())
		assert.Equal(t, "000000", typesResp.Code)
		assert.Contains(t, string(typesResp.Data), `"Excel"`)

		syncW := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/syncRecord/212/0/0", nil)
		syncResp := decodeDatasourceResp(t, syncW.Body.Bytes())
		assert.Equal(t, "000000", syncResp.Code)
		assert.Contains(t, string(syncResp.Data), `"datasourceId":212`)
		assert.Contains(t, string(syncResp.Data), `"current":1`)
		assert.Contains(t, string(syncResp.Data), `"size":10`)
	})

	t.Run("helper current user values", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		assert.Equal(t, int64(0), getCurrentUserID(c))
		assert.Equal(t, "", getCurrentUsername(c))
		c.Set(middleware.ContextUserID, int64(99))
		c.Set("username", "round4b")
		assert.Equal(t, int64(99), getCurrentUserID(c))
		assert.Equal(t, "round4b", getCurrentUsername(c))
	})
}

func TestRound4B_TemplateHandler_RoutesAndHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := setupTemplateHandlerTestEnv(t)
	h := NewTemplateHandler(service.NewTemplateService(repository.NewTemplateRepository(env.db)))
	require.NotNil(t, h)

	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 1001, Name: "cat-a", Pid: 0, Level: 0, DvType: "dashboard", NodeType: template.NodeTypeFolder, TemplateType: "self", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 1002, Name: "panel-a", Pid: 1001, Level: 2, DvType: "dashboard", NodeType: "template", TemplateType: "self", Snapshot: "snap-a", Version: 3})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-1002", CategoryID: "1001", TemplateID: "1002"})

	t.Run("templateCreateBy username and fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextUserID, int64(55))
		assert.Equal(t, "55", templateCreateBy(c))
		c.Set(middleware.ContextUserName, "admin")
		assert.Equal(t, "admin", templateCreateBy(c))
	})

	t.Run("list invalid json", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/template/list", []byte(`{`))
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "10001", resp.Code)
	})

	t.Run("update invalid json", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/template/update", []byte(`{`))
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "10001", resp.Code)
	})

	t.Run("search market family success", func(t *testing.T) {
		for _, path := range []string{"/api/templateMarket/search", "/api/templateMarket/searchRecommend", "/api/templateMarket/searchPreview"} {
			w := performTemplateJSONRequest(t, env.r, http.MethodGet, path, nil)
			resp := decodeTemplateResp(t, w.Body.Bytes())
			assert.Equal(t, "000000", resp.Code, path)
			assert.Contains(t, string(resp.Data), `"categories"`)
		}
	})

	t.Run("delete category false then success", func(t *testing.T) {
		busyW := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/deleteCategory/1001", nil)
		busyResp := decodeTemplateResp(t, busyW.Body.Bytes())
		assert.Equal(t, "000000", busyResp.Code)
		assert.JSONEq(t, `"failed"`, string(busyResp.Data))

		require.NoError(t, env.db.Delete(&auto.VisualizationTemplateCategoryMap{}, "template_id = ?", "1002").Error)
		require.NoError(t, env.db.Delete(&templateCoreVisualizationTemplateMirror{}, 1002).Error)
		okW := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/deleteCategory/1001", nil)
		okResp := decodeTemplateResp(t, okW.Body.Bytes())
		assert.Equal(t, "000000", okResp.Code)
		assert.JSONEq(t, `"success"`, string(okResp.Data))
	})

	t.Run("find categories invalid json", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/findCategoriesByTemplateIds", []byte(`{`))
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "10001", resp.Code)
	})

	t.Run("market template type helper", func(t *testing.T) {
		assert.Equal(t, "SCREEN", marketTemplateType("dataV"))
		assert.Equal(t, "PANEL", marketTemplateType("dashboard"))
	})
}

func TestRound4B_ChartHandler_GapCoverage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("constructor query and chart detail routes", func(t *testing.T) {
		title := "round4b-chart"
		render := "antv"
		repo := &fakeBridgeChartRepo{charts: map[int64]*chart.CoreChartView{501: {ID: 501, Title: &title, Render: &render}}}
		h := NewChartHandler(service.NewChartService(repo), nil)
		require.NotNil(t, h)

		r := gin.New()
		r.POST("/chart/query", h.Query)
		r.GET("/chart/viewOption/:resourceId", h.ViewOption)
		r.GET("/chart/chartBaseInfo/:id/:resourceTable", h.ChartBaseInfo)
		r.POST("/chart/save", h.SaveFromMap)
		r.GET("/chart/:id", h.GetChart)
		r.GET("/chart/detail/:id", h.GetDetail)

		repo.viewOptions = map[int64][]chart.ViewSelectorVO{10: {{ID: 1, Title: "keep"}, {ID: 2, Title: "drop"}}}
		repo.componentData = map[int64]string{10: `component-1`}
		repo.chartBaseInfo = map[string]*chart.ChartBaseVO{"panel:501": {ChartID: 501, ChartName: "base-info", ResourceType: "panel"}}

		queryW := httptest.NewRecorder()
		queryReq := httptest.NewRequest(http.MethodPost, "/chart/query", strings.NewReader(`{"id":501}`))
		queryReq.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(queryW, queryReq)
		queryResp := bridgeAnyResp{}
		require.NoError(t, json.Unmarshal(queryW.Body.Bytes(), &queryResp))
		assert.Equal(t, "000000", queryResp.Code)
		assert.Equal(t, title, queryResp.Data["title"])

		viewW := httptest.NewRecorder()
		r.ServeHTTP(viewW, httptest.NewRequest(http.MethodGet, "/chart/viewOption/10", nil))
		viewResp := struct {
			Code string                 `json:"code"`
			Data []chart.ViewSelectorVO `json:"data"`
		}{}
		require.NoError(t, json.Unmarshal(viewW.Body.Bytes(), &viewResp))
		assert.Equal(t, "000000", viewResp.Code)
		require.Len(t, viewResp.Data, 1)
		assert.Equal(t, int64(1), viewResp.Data[0].ID)

		baseW := httptest.NewRecorder()
		r.ServeHTTP(baseW, httptest.NewRequest(http.MethodGet, "/chart/chartBaseInfo/501/panel", nil))
		baseResp := bridgeAnyResp{}
		require.NoError(t, json.Unmarshal(baseW.Body.Bytes(), &baseResp))
		assert.Equal(t, "000000", baseResp.Code)
		assert.Equal(t, float64(501), baseResp.Data["chartId"])

		saveReqBody := `{"id":501,"title":"updated","type":"bar","xAxis":[{"id":"x1","name":"分类"}],"aggregate":true,"resultCount":20}`
		saveW := httptest.NewRecorder()
		saveReq := httptest.NewRequest(http.MethodPost, "/chart/save", strings.NewReader(saveReqBody))
		saveReq.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(saveW, saveReq)
		saveResp := bridgeAnyResp{}
		require.NoError(t, json.Unmarshal(saveW.Body.Bytes(), &saveResp))
		assert.Equal(t, "000000", saveResp.Code)
		assert.Equal(t, "updated", *repo.charts[501].Title)
		require.NotNil(t, repo.charts[501].XAxis)
		assert.Contains(t, *repo.charts[501].XAxis, `"x1"`)
		require.NotNil(t, repo.charts[501].Aggregate)
		assert.True(t, *repo.charts[501].Aggregate)

		for _, path := range []string{"/chart/501", "/chart/detail/501"} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			resp := bridgeAnyResp{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, "000000", resp.Code, path)
		}
	})

	t.Run("helper functions and export delegate", func(t *testing.T) {
		viewType := "bar"
		xAxis := `[{"id":"1","name":"分类"}]`
		customAttr := `{"legend":true}`
		view := &chart.CoreChartView{Type: &viewType, XAxis: &xAxis, CustomAttr: &customAttr}
		mapped := map[string]interface{}{"keep": "value"}
		mergeChartViewIntoMap(mapped, view)
		assert.Equal(t, "value", mapped["keep"])
		assert.IsType(t, []interface{}{}, mapped["xAxis"])
		assert.IsType(t, map[string]interface{}{}, mapped["customAttr"])
		assert.True(t, isChartViewJSONField("xAxis"))
		assert.False(t, isChartViewJSONField("title"))

		h := &ChartHandler{}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/chartData/innerExportDataSetDetails", strings.NewReader(`{`))
		c.Request.Header.Set("Content-Type", "application/json")
		h.InnerExportDataSetDetails(c)
		resp := bridgeCodeResp{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp.Code)
	})

	t.Run("chart data routes registration", func(t *testing.T) {
		fieldName := "city"
		originName := "city"
		groupType := "d"
		fieldType := "string"
		deType := 1
		checked := true
		chartID := int64(601)
		repo := &fakeBridgeChartRepo{
			fieldRegistry: map[int64]*dataset.CoreDatasetTableField{11: {ID: 11, DatasetGroupID: 71, ChartID: &chartID, Name: &fieldName, OriginName: &originName, GroupType: &groupType, Type: &fieldType, DeType: &deType, Checked: &checked}},
			chartFields:   map[int64][]*dataset.CoreDatasetTableField{601: {{ID: 11, DatasetGroupID: 71, ChartID: &chartID, Name: &fieldName, OriginName: &originName, GroupType: &groupType, Type: &fieldType, DeType: &deType, Checked: &checked}}},
		}
		h := NewChartHandler(service.NewChartService(repo), nil)
		r := gin.New()
		group := r.Group("/chartData")
		RegisterChartDataRoutes(group, h, nil)

		dataReq := httptest.NewRequest(http.MethodPost, "/chartData/getData", strings.NewReader(`{"id":0}`))
		dataReq.Header.Set("Content-Type", "application/json")
		dataW := httptest.NewRecorder()
		r.ServeHTTP(dataW, dataReq)
		dataResp := bridgeCodeResp{}
		require.NoError(t, json.Unmarshal(dataW.Body.Bytes(), &dataResp))
		assert.Equal(t, "500000", dataResp.Code)

		copyReq := httptest.NewRequest(http.MethodPost, "/chart/copyField/11/601", nil)
		copyW := httptest.NewRecorder()
		copyC, _ := gin.CreateTestContext(copyW)
		copyC.Params = gin.Params{{Key: "id", Value: "11"}, {Key: "chartId", Value: "601"}}
		copyC.Request = copyReq
		h.CopyField(copyC)
		copyResp := bridgeCodeResp{}
		require.NoError(t, json.Unmarshal(copyW.Body.Bytes(), &copyResp))
		assert.Equal(t, "000000", copyResp.Code)

		delReq := httptest.NewRequest(http.MethodPost, "/chart/deleteField/11", nil)
		delW := httptest.NewRecorder()
		delC, _ := gin.CreateTestContext(delW)
		delC.Params = gin.Params{{Key: "id", Value: "11"}}
		delC.Request = delReq
		h.DeleteField(delC)
		delResp := bridgeCodeResp{}
		require.NoError(t, json.Unmarshal(delW.Body.Bytes(), &delResp))
		assert.Equal(t, "000000", delResp.Code)

		delByChartReq := httptest.NewRequest(http.MethodPost, "/chart/deleteFieldByChart/601", nil)
		delByChartW := httptest.NewRecorder()
		delByChartC, _ := gin.CreateTestContext(delByChartW)
		delByChartC.Params = gin.Params{{Key: "chartId", Value: "601"}}
		delByChartC.Request = delByChartReq
		h.DeleteFieldByChart(delByChartC)
		delByChartResp := bridgeCodeResp{}
		require.NoError(t, json.Unmarshal(delByChartW.Body.Bytes(), &delByChartResp))
		assert.Equal(t, "000000", delByChartResp.Code)
	})
}
