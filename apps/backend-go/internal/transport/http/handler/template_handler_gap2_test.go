package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateHandlerGap2_UpdateSearchBatchAndCategories(t *testing.T) {
	env := setupTemplateHandlerTestEnv(t)
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 901, Name: "before-update", Pid: 91, Level: 2, DvType: "dashboard", NodeType: "template", TemplateType: "user", Snapshot: "old", TemplateStyle: "s1", TemplateData: "d1", DynamicData: "dy1", AppData: "a1", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 902, Name: "market-folder", Pid: 0, Level: 0, DvType: "dashboard", NodeType: "folder", TemplateType: "self", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 903, Name: "market-panel", Pid: 902, Level: 2, DvType: "dashboard", NodeType: "template", TemplateType: "self", Snapshot: "snap-panel", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 904, Name: "market-screen", Pid: 902, Level: 2, DvType: "dataV", NodeType: "template", TemplateType: "self", Snapshot: "snap-screen", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 905, Name: "batch-one", Pid: 91, Level: 2, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 906, Name: "batch-two", Pid: 91, Level: 2, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-903", CategoryID: "902", TemplateID: "903"})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-904", CategoryID: "902", TemplateID: "904"})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-905", CategoryID: "91", TemplateID: "905"})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-906", CategoryID: "91", TemplateID: "906"})

	t.Run("update", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/template/update", map[string]any{"id": 901, "name": "after-update", "snapshot": "new-snap", "templateStyle": "s2", "templateData": "d2", "dynamicData": "dy2", "appData": "a2"})
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		var row templateCoreVisualizationTemplateMirror
		require.NoError(t, env.db.First(&row, 901).Error)
		assert.Equal(t, "after-update", row.Name)
		assert.Equal(t, "new-snap", row.Snapshot)
	})

	t.Run("search templates", func(t *testing.T) {
		h := NewTemplateHandler(service.NewTemplateService(repository.NewTemplateRepository(env.db)))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/templateManage/search?pid=902&dvType=dashboard", nil)
		h.SearchTemplates(c)
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		var data struct {
			List []templateCoreVisualizationTemplateMirror `json:"list"`
		}
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		assert.NotEmpty(t, data.List)
		ids := make([]int64, 0, len(data.List))
		for _, item := range data.List {
			ids = append(ids, item.ID)
		}
		assert.Contains(t, ids, int64(903))
	})

	t.Run("delete category", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/deleteCategory/999", nil)
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		assert.JSONEq(t, `"success"`, string(resp.Data))
	})

	t.Run("batch update and find categories", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/batchUpdate", map[string]any{"templateIds": []string{"905", "906"}, "categories": []string{"902", "903"}})
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		w = performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/findCategoriesByTemplateIds", map[string]any{"templateArray": []string{"905", "906"}})
		resp = decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		var categories []string
		require.NoError(t, json.Unmarshal(resp.Data, &categories))
		assert.ElementsMatch(t, []string{"902", "903"}, categories)
	})

	t.Run("market endpoints", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodGet, "/api/templateMarket/searchRecommend", nil)
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		var market map[string]any
		require.NoError(t, json.Unmarshal(resp.Data, &market))
		contents := market["contents"].([]any)
		categories := market["categories"].([]any)
		assert.NotEmpty(t, contents)
		assert.NotEmpty(t, categories)
		first := contents[0].(map[string]any)
		assert.Contains(t, []string{"PANEL", "SCREEN"}, first["templateType"])

		w = performTemplateJSONRequest(t, env.r, http.MethodGet, "/api/templateMarket/searchPreview", nil)
		resp = decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		require.NoError(t, json.Unmarshal(resp.Data, &market))
		grouped := market["contents"].([]any)
		assert.NotEmpty(t, grouped)
		firstGroup := grouped[0].(map[string]any)
		assert.Contains(t, firstGroup, "category")
		assert.Contains(t, firstGroup, "contents")
	})
}
