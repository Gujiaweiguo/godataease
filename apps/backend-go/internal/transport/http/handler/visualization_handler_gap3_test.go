package handler

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/visualization"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisualizationHandlerGap3_HelperFunctions(t *testing.T) {
	t.Run("build enriched response and flags", func(t *testing.T) {
		mobile := true
		status := 2
		nodeType := "leaf"
		typ := "dashboard"
		createTime := int64(11)
		updateTime := int64(22)
		contentID := "content-1"
		version := 3
		resp := buildEnrichedVisualizationResponse(&visualization.DataVisualizationInfo{
			ID:           8,
			Name:         "viz",
			NodeType:     &nodeType,
			Type:         &typ,
			Status:       &status,
			MobileLayout: &mobile,
			CreateTime:   &createTime,
			UpdateTime:   &updateTime,
			ContentID:    &contentID,
			Version:      &version,
		}, map[string]any{"k": "v"})
		assert.Equal(t, int64(8), resp["id"])
		assert.Equal(t, true, resp["mobileLayout"])
		assert.Equal(t, map[string]any{"k": "v"}, resp["canvasViewInfo"])
		assert.Equal(t, 1, visualizationExtraFlag(&mobile))
		assert.Equal(t, 1, visualizationPublishFlag(&status))
		assert.Equal(t, 0, visualizationExtraFlag(nil))
		assert.Equal(t, 0, visualizationPublishFlag(nil))
	})

	t.Run("build canvas info parses nested json strings", func(t *testing.T) {
		title := "Chart"
		xAxis := `{"name":"x"}`
		customAttr := `[1,2]`
		views := []chart.CoreChartView{{ID: 9, Title: &title, XAxis: &xAxis, CustomAttr: &customAttr}}
		info := buildCanvasViewInfo(views)
		entry, ok := info["9"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Chart", entry["title"])
		assert.IsType(t, map[string]any{}, entry["xAxis"])
		assert.IsType(t, []any{}, entry["customAttr"])
	})

	t.Run("resolve busi types and tree validators", func(t *testing.T) {
		types, err := resolveBusiTypes("dashboard-dataV")
		require.NoError(t, err)
		assert.Equal(t, []string{"dashboard", "dataV"}, types)
		_, err = resolveBusiTypes("bad")
		require.Error(t, err)

		leafOnly := true
		nodeTypeFolder := "folder"
		nodeTypeLeaf := "leaf"
		pid := int64(1)
		tree, err := buildVisualizationTree([]*visualization.DataVisualizationInfo{{ID: 1, Name: "root", NodeType: &nodeTypeFolder}, {ID: 2, Name: "child", PID: &pid, NodeType: &nodeTypeLeaf}}, nil)
		require.NoError(t, err)
		require.Len(t, tree, 1)
		assert.Len(t, tree[0].Children, 1)

		leafTree, err := buildVisualizationTree([]*visualization.DataVisualizationInfo{{ID: 2, Name: "child", PID: &pid, NodeType: &nodeTypeLeaf}}, &leafOnly)
		require.NoError(t, err)
		assert.Len(t, leafTree, 0)

		_, err = buildVisualizationTree([]*visualization.DataVisualizationInfo{{ID: 0, Name: "bad", NodeType: &nodeTypeLeaf}}, nil)
		require.Error(t, err)
		_, err = buildVisualizationTree([]*visualization.DataVisualizationInfo{{ID: 3, Name: "", NodeType: &nodeTypeLeaf}}, nil)
		require.Error(t, err)
		badType := "other"
		_, err = buildVisualizationTree([]*visualization.DataVisualizationInfo{{ID: 4, Name: "bad", NodeType: &badType}}, nil)
		require.Error(t, err)
		require.NoError(t, validateTreeNodes([]treeNode{{ID: "1", Name: "ok"}}))
		require.Error(t, validateTreeNodes([]treeNode{{ID: "", Name: "bad"}}))
		require.Error(t, validateTreeNodes([]treeNode{{ID: "1", Name: "leaf", Leaf: true, Children: []treeNode{{ID: "2", Name: "child"}}}}))
	})
}

func TestVisualizationHandlerGap3_getUpdateBy_And_RecordExportLogBadRequest(t *testing.T) {
	h := &VisualizationHandler{}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, "system", h.getUpdateBy(c))
	c.Set("userId", int64(99))
	assert.Equal(t, "99", h.getUpdateBy(c))

	c.Request = httptest.NewRequest(http.MethodPost, "/export", strings.NewReader(`{`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.recordExportLog(c, "app")
	var resp visualizationHandlerResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
}

func TestVisualizationHandlerGap3_HandlerErrorBranches(t *testing.T) {
	db := setupVisualizationHandlerUnitDB(t)
	router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) {
		r.POST("/findById", h.FindByID)
		r.POST("/tree", h.Tree)
		r.POST("/decompression", h.Decompression)
		r.POST("/appCanvasNameCheck", h.AppCanvasNameCheck)
		r.POST("/nameCheck", h.NameCheck)
	})

	resp := performVisualizationRequest(t, router, http.MethodPost, "/findById", `{`)
	assert.Equal(t, "500000", resp.Code)
	resp = performVisualizationRequest(t, router, http.MethodPost, "/tree", `{"busiFlag":"bad"}`)
	assert.Equal(t, "500000", resp.Code)
	resp = performVisualizationRequest(t, router, http.MethodPost, "/decompression", `{`)
	assert.Equal(t, "500000", resp.Code)
	resp = performVisualizationRequest(t, router, http.MethodPost, "/appCanvasNameCheck", `{`)
	assert.Equal(t, "500000", resp.Code)
	resp = performVisualizationRequest(t, router, http.MethodPost, "/nameCheck", `{`)
	assert.Equal(t, "500000", resp.Code)
}

func TestVisualizationHandlerGap3_DecompressionLocalFileErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupVisualizationHandlerUnitDB(t)
	router := setupVisualizationHandlerRouter(db, func(r *gin.Engine, h *VisualizationHandler) { r.POST("/local", h.DecompressionLocalFile) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/local", nil)
	router.ServeHTTP(w, req)
	var resp visualizationHandlerResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)

	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "bad.DET2")
	require.NoError(t, err)
	_, err = part.Write([]byte(`not-json`))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/local", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp.Code)
}
