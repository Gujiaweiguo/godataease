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
