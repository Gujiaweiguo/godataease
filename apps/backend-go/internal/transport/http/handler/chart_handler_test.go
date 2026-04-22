package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChartHandler_Query_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/query", h.Query)

	req := httptest.NewRequest("POST", "/chart/query", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp bridgeCodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
}

func TestChartHandler_Data_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/data", h.Data)

	req := httptest.NewRequest("POST", "/chart/data", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp bridgeCodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
}

func TestChartHandler_Data_SuccessReturnsTopLevelConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	render := "antv"
	chartType := "bar"
	xAxis := `[{"id":"2","dataeaseName":"category","originName":"category","name":"分类"}]`
	yAxis := `[{"id":"5","dataeaseName":"sales_amount","originName":"sales_amount","name":"销售额","summary":"sum"}]`
	repo := &fakeBridgeChartRepo{
		charts: map[int64]*chart.CoreChartView{5: {ID: 5, Render: &render, Type: &chartType, XAxis: &xAxis, YAxis: &yAxis}},
	}
	h := NewChartHandler(service.NewChartService(repo), nil)
	r := gin.New()
	r.POST("/chart/data", h.Data)

	req := httptest.NewRequest("POST", "/chart/data", strings.NewReader(`{"id":5}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["code"])
	viewMap, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "antv", viewMap["render"])
	assert.Equal(t, "bar", viewMap["type"])
	_, hasXAxis := viewMap["xAxis"]
	assert.True(t, hasXAxis)
	chartData, hasData := viewMap["data"]
	if hasData && chartData != nil {
		dataMap, ok := chartData.(map[string]interface{})
		require.True(t, ok)
		points, exists := dataMap["data"]
		if exists {
			pointList, ok := points.([]interface{})
			require.True(t, ok)
			assert.Len(t, pointList, 0)
		}
	}
}

func TestChartHandler_GetChart_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.GET("/chart/:id", h.GetChart)

	req := httptest.NewRequest("GET", "/chart/not-a-number", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp bridgeCodeResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
}

func TestChartHandler_GetChart_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	title := "sales-overview"
	repo := &fakeBridgeChartRepo{charts: map[int64]*chart.CoreChartView{101: {ID: 101, Title: &title}}}
	h := NewChartHandler(service.NewChartService(repo), nil)
	r := gin.New()
	r.GET("/chart/:id", h.GetChart)

	req := httptest.NewRequest("GET", "/chart/101", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp bridgeAnyResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, float64(101), resp.Data["id"])
	assert.Equal(t, title, resp.Data["title"])
}

func TestChartHandler_CheckSameDataSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("same table id", func(t *testing.T) {
		sharedTableID := int64(99)
		repo := &fakeBridgeChartRepo{charts: map[int64]*chart.CoreChartView{
			1: {ID: 1, TableID: &sharedTableID},
			2: {ID: 2, TableID: &sharedTableID},
		}}
		h := NewChartHandler(service.NewChartService(repo), nil)
		r := gin.New()
		r.GET("/chart/checkSameDataSet/:viewIdSource/:viewIdTarget", h.CheckSameDataSet)

		req := httptest.NewRequest("GET", "/chart/checkSameDataSet/1/2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp struct {
			Code string `json:"code"`
			Data bool   `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		assert.True(t, resp.Data)
	})

	t.Run("different table id", func(t *testing.T) {
		tableA := int64(99)
		tableB := int64(100)
		repo := &fakeBridgeChartRepo{charts: map[int64]*chart.CoreChartView{
			1: {ID: 1, TableID: &tableA},
			2: {ID: 2, TableID: &tableB},
		}}
		h := NewChartHandler(service.NewChartService(repo), nil)
		r := gin.New()
		r.GET("/chart/checkSameDataSet/:viewIdSource/:viewIdTarget", h.CheckSameDataSet)

		req := httptest.NewRequest("GET", "/chart/checkSameDataSet/1/2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp struct {
			Code string `json:"code"`
			Data bool   `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "000000", resp.Code)
		assert.False(t, resp.Data)
	})
}
