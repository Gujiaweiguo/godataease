package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type templateHandlerTestEnv struct {
	r  *gin.Engine
	db *gorm.DB
}

type templateBridgeResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type templateCoreVisualizationTemplateMirror struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name          string     `gorm:"column:name;size:255" json:"name"`
	Pid           int64      `gorm:"column:pid;index" json:"pid"`
	Level         int        `gorm:"column:level" json:"level"`
	DvType        string     `gorm:"column:dv_type;size:50" json:"dvType"`
	NodeType      string     `gorm:"column:node_type;size:50" json:"nodeType"`
	CreateBy      string     `gorm:"column:create_by;size:255" json:"createBy"`
	CreateTime    *time.Time `gorm:"column:create_time" json:"createTime"`
	Snapshot      string     `gorm:"column:snapshot;type:longtext" json:"snapshot"`
	TemplateType  string     `gorm:"column:template_type;size:50" json:"templateType"`
	TemplateStyle string     `gorm:"column:template_style;type:longtext" json:"templateStyle"`
	TemplateData  string     `gorm:"column:template_data;type:longtext" json:"templateData"`
	DynamicData   string     `gorm:"column:dynamic_data;type:longtext" json:"dynamicData"`
	AppData       string     `gorm:"column:app_data;type:longtext" json:"appData"`
	UseCount      int        `gorm:"column:use_count;default:0" json:"useCount"`
	Version       int        `gorm:"column:version;default:3" json:"version"`
}

func (templateCoreVisualizationTemplateMirror) TableName() string {
	return "core_visualization_template"
}

func setupTemplateHandlerTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	return setupTemplateHandlerTestEnv(t).r
}

func setupTemplateHandlerTestEnv(t *testing.T) *templateHandlerTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))

	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userId", int64(42))
		c.Set("userName", "creator-user")
		c.Next()
	})
	api := r.Group("/api")
	RegisterTemplateRoutes(api, h)

	return &templateHandlerTestEnv{r: r, db: db}
}

func performTemplateJSONRequest(t *testing.T, r *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody []byte
	switch v := body.(type) {
	case nil:
		reqBody = nil
	case []byte:
		reqBody = v
	default:
		var err error
		reqBody, err = json.Marshal(v)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func decodeTemplateResp(t *testing.T, body []byte) templateBridgeResp {
	t.Helper()
	var resp templateBridgeResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func seedTemplateRow(t *testing.T, db *gorm.DB, row templateCoreVisualizationTemplateMirror) {
	t.Helper()
	require.NoError(t, db.Create(&row).Error)
}

func seedTemplateCategoryMapRow(t *testing.T, db *gorm.DB, row auto.VisualizationTemplateCategoryMap) {
	t.Helper()
	require.NoError(t, db.Create(&row).Error)
}

func TestTemplateHandler_Create_SuccessUsesUserName(t *testing.T) {
	env := setupTemplateHandlerTestEnv(t)
	body := map[string]any{
		"name":          "template-create",
		"pid":           101,
		"level":         2,
		"dvType":        "dashboard",
		"nodeType":      "template",
		"snapshot":      "snap-1",
		"templateType":  "user",
		"templateStyle": "style-1",
		"templateData":  "{\"v\":1}",
		"dynamicData":   "dynamic-1",
		"appData":       "app-1",
	}

	w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/template/create", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeTemplateResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var created templateCoreVisualizationTemplateMirror
	require.NoError(t, env.db.First(&created).Error)
	assert.Equal(t, "template-create", created.Name)
	assert.Equal(t, int64(101), created.Pid)
	assert.Equal(t, "creator-user", created.CreateBy)
	assert.Equal(t, 3, created.Version)
	assert.Equal(t, 0, created.UseCount)
	assert.NotNil(t, created.CreateTime)
	var data map[string]any
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "template-create", data["name"])
	assert.Equal(t, "creator-user", data["createBy"])
	assert.Equal(t, "snap-1", data["snapshot"])
	assert.Equal(t, "style-1", data["templateStyle"])
	assert.Equal(t, "{\"v\":1}", data["templateData"])
	assert.NotEmpty(t, data["createTime"])
}

func TestTemplateHandler_Save_CreateBranch_PersistsCategoryMappings(t *testing.T) {
	env := setupTemplateHandlerTestEnv(t)
	body := map[string]any{
		"name":         "save-create",
		"pid":          11,
		"level":        2,
		"dvType":       "dashboard",
		"nodeType":     "template",
		"templateType": "user",
		"templateData": "create-data",
		"categories":   []string{"201", "202", "201"},
	}

	w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/save", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeTemplateResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var created templateCoreVisualizationTemplateMirror
	require.NoError(t, env.db.First(&created).Error)
	assert.Equal(t, "creator-user", created.CreateBy)
	assert.Equal(t, int64(11), created.Pid)

	var mappings []auto.VisualizationTemplateCategoryMap
	require.NoError(t, env.db.Order("category_id asc").Find(&mappings).Error)
	require.Len(t, mappings, 2)
	assert.Equal(t, "201", mappings[0].CategoryID)
	assert.Equal(t, fmt.Sprintf("%d", created.ID), mappings[0].TemplateID)
	assert.Equal(t, "202", mappings[1].CategoryID)
	assert.Equal(t, fmt.Sprintf("%d", created.ID), mappings[1].TemplateID)
	assert.NotEmpty(t, mappings[0].ID)
	assert.NotEmpty(t, mappings[1].ID)
	var data map[string]any
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, float64(created.ID), data["id"])
	assert.Equal(t, "save-create", data["name"])
	assert.Equal(t, "creator-user", data["createBy"])
	assert.Equal(t, "create-data", data["templateData"])
	assert.NotEmpty(t, data["createTime"])
}

func TestTemplateHandler_Save_UpdateBranch_UpdatesTemplateAndCategories(t *testing.T) {
	env := setupTemplateHandlerTestEnv(t)
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 301, Name: "before-save-update", Pid: 21, Level: 2, DvType: "dashboard", NodeType: "template", CreateBy: "old-user", TemplateType: "user", TemplateData: "old-data", TemplateStyle: "old-style", DynamicData: "old-dynamic", AppData: "old-app", Version: 3})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-1", CategoryID: "21", TemplateID: "301"})
	body := map[string]any{
		"id":            301,
		"name":          "after-save-update",
		"nodeType":      "template",
		"snapshot":      "new-snap",
		"templateStyle": "new-style",
		"templateData":  "new-data",
		"dynamicData":   "new-dynamic",
		"appData":       "new-app",
		"categories":    []string{"31", "32", "31"},
	}

	w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/save", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeTemplateResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var updated templateCoreVisualizationTemplateMirror
	require.NoError(t, env.db.First(&updated, 301).Error)
	assert.Equal(t, "after-save-update", updated.Name)
	assert.Equal(t, int64(31), updated.Pid)
	assert.Equal(t, "new-snap", updated.Snapshot)
	assert.Equal(t, "new-style", updated.TemplateStyle)
	assert.Equal(t, "new-data", updated.TemplateData)
	assert.Equal(t, "new-dynamic", updated.DynamicData)
	assert.Equal(t, "new-app", updated.AppData)
	assert.Equal(t, "old-user", updated.CreateBy)

	var mappings []auto.VisualizationTemplateCategoryMap
	require.NoError(t, env.db.Where("template_id = ?", "301").Order("category_id asc").Find(&mappings).Error)
	require.Len(t, mappings, 2)
	assert.Equal(t, []string{"31", "32"}, []string{mappings[0].CategoryID, mappings[1].CategoryID})
	assert.JSONEq(t, `{"id":301,"name":"after-save-update","pid":31,"level":2,"dvType":"dashboard","nodeType":"template","createBy":"old-user","snapshot":"new-snap","templateType":"user","templateStyle":"new-style","templateData":"new-data","dynamicData":"new-dynamic","appData":"new-app","useCount":0,"version":3}`, string(resp.Data))
}

func TestTemplateHandler_Get_InvalidID(t *testing.T) {
	r := setupTemplateHandlerTestRouter(t)

	w := performTemplateJSONRequest(t, r, http.MethodGet, "/api/template/get/not-a-number", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeTemplateResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Equal(t, "Invalid template ID", resp.Msg)
}

func TestTemplateHandler_Get_NotFound(t *testing.T) {
	r := setupTemplateHandlerTestRouter(t)

	w := performTemplateJSONRequest(t, r, http.MethodGet, "/api/template/get/9999", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeTemplateResp(t, w.Body.Bytes())
	assert.Equal(t, "40001", resp.Code)
	assert.Contains(t, resp.Msg, "Failed to get template")
	assert.Contains(t, strings.ToLower(resp.Msg), "record not found")
}

func TestTemplateHandler_List_CategoryScopedListing(t *testing.T) {
	env := setupTemplateHandlerTestEnv(t)
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 401, Name: "category-folder", Pid: 0, Level: 1, DvType: "dashboard", NodeType: "folder", TemplateType: "user", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 402, Name: "direct-child", Pid: 401, Level: 2, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 403, Name: "mapped-template", Pid: 0, Level: 2, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 404, Name: "other-dvtype", Pid: 401, Level: 2, DvType: "screen", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-401", CategoryID: "401", TemplateID: "403"})

	w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/template/list", map[string]any{
		"categoryId": "401",
		"dvType":     "dashboard",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeTemplateResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data struct {
		List  []templateCoreVisualizationTemplateMirror `json:"list"`
		Total int64                                     `json:"total"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.Len(t, data.List, 2)
	assert.Equal(t, int64(2), data.Total)
	assert.ElementsMatch(t, []int64{402, 403}, []int64{data.List[0].ID, data.List[1].ID})
	for _, item := range data.List {
		assert.NotEqual(t, "folder", item.NodeType)
		assert.Equal(t, "dashboard", item.DvType)
	}
}

func TestTemplateHandler_Delete_SuccessRemovesTemplateAndCategoryMappings(t *testing.T) {
	env := setupTemplateHandlerTestEnv(t)
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 501, Name: "delete-me", Pid: 51, Level: 2, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-501-a", CategoryID: "51", TemplateID: "501"})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-501-b", CategoryID: "52", TemplateID: "501"})

	w := performTemplateJSONRequest(t, env.r, http.MethodDelete, "/api/template/delete/501", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeTemplateResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Empty(t, resp.Data)

	var templateCount int64
	require.NoError(t, env.db.Model(&templateCoreVisualizationTemplateMirror{}).Where("id = ?", 501).Count(&templateCount).Error)
	assert.Equal(t, int64(0), templateCount)

	var mappingCount int64
	require.NoError(t, env.db.Model(&auto.VisualizationTemplateCategoryMap{}).Where("template_id = ?", "501").Count(&mappingCount).Error)
	assert.Equal(t, int64(0), mappingCount)
}

func TestTemplateHandler_DeleteWithCategory_UnlinksOnlyWhenOtherCategoriesRemain(t *testing.T) {
	env := setupTemplateHandlerTestEnv(t)
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 601, Name: "unlink-only", Pid: 61, Level: 2, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-601-a", CategoryID: "61", TemplateID: "601"})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-601-b", CategoryID: "62", TemplateID: "601"})

	w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/delete/601/61", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeTemplateResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var templateCount int64
	require.NoError(t, env.db.Model(&templateCoreVisualizationTemplateMirror{}).Where("id = ?", 601).Count(&templateCount).Error)
	assert.Equal(t, int64(1), templateCount)

	var mappings []auto.VisualizationTemplateCategoryMap
	require.NoError(t, env.db.Where("template_id = ?", "601").Order("category_id asc").Find(&mappings).Error)
	require.Len(t, mappings, 1)
	assert.Equal(t, "62", mappings[0].CategoryID)
}

func TestTemplateHandler_NameCheck_Semantics(t *testing.T) {
	env := setupTemplateHandlerTestEnv(t)
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 701, Name: "duplicate-name", Pid: 0, Level: 1, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})

	t.Run("create duplicate returns existAll", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/nameCheck", map[string]any{
			"name":    "duplicate-name",
			"optType": "insert",
		})
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		assert.JSONEq(t, `"existAll"`, string(resp.Data))
	})

	t.Run("update same id returns none", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/nameCheck", map[string]any{
			"name":    "duplicate-name",
			"id":      "701",
			"optType": "update",
		})
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		assert.JSONEq(t, `"none"`, string(resp.Data))
	})

	t.Run("blank name returns none", func(t *testing.T) {
		w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/nameCheck", map[string]any{
			"name":    "   ",
			"optType": "insert",
		})
		resp := decodeTemplateResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		assert.JSONEq(t, `"none"`, string(resp.Data))
	})
}

func TestTemplateHandler_BatchDelete_MergesIDsAndIgnoresBadStringIDs(t *testing.T) {
	env := setupTemplateHandlerTestEnv(t)
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 801, Name: "batch-a", Pid: 0, Level: 1, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 802, Name: "batch-b", Pid: 0, Level: 1, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateRow(t, env.db, templateCoreVisualizationTemplateMirror{ID: 803, Name: "batch-c", Pid: 0, Level: 1, DvType: "dashboard", NodeType: "template", TemplateType: "user", Version: 3})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-801", CategoryID: "81", TemplateID: "801"})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-802", CategoryID: "82", TemplateID: "802"})
	seedTemplateCategoryMapRow(t, env.db, auto.VisualizationTemplateCategoryMap{ID: "map-803", CategoryID: "83", TemplateID: "803"})

	w := performTemplateJSONRequest(t, env.r, http.MethodPost, "/api/templateManage/batchDelete", map[string]any{
		"ids":         []int64{801},
		"templateIds": []string{"802", "bad", "0", "-4"},
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeTemplateResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var remainingIDs []int64
	require.NoError(t, env.db.Model(&templateCoreVisualizationTemplateMirror{}).Order("id asc").Pluck("id", &remainingIDs).Error)
	assert.Equal(t, []int64{803}, remainingIDs)

	var mappingTemplateIDs []string
	require.NoError(t, env.db.Model(&auto.VisualizationTemplateCategoryMap{}).Order("template_id asc").Pluck("template_id", &mappingTemplateIDs).Error)
	assert.Equal(t, []string{"803"}, mappingTemplateIDs)
}
