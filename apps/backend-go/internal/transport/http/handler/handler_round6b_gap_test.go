package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Shared round6b helpers
// ---------------------------------------------------------------------------

type round6bBridgeResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func round6bSetupDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:round6b_%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	return db
}

func round6bDecodeResp(t *testing.T, body []byte) round6bBridgeResp {
	t.Helper()
	var resp round6bBridgeResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func round6bJSONRequest(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
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

// ---------------------------------------------------------------------------
// Datasource handler tests
// ---------------------------------------------------------------------------

func TestRound6B_Ds_NewDatasourceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	assert.NotNil(t, h)
	assert.Equal(t, svc, h.service)
}

func TestRound6B_Ds_List_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.POST("/ds/list", h.List)
	w := round6bJSONRequest(t, r, http.MethodPost, "/ds/list", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_Validate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.POST("/ds/validate", h.Validate)
	w := round6bJSONRequest(t, r, http.MethodPost, "/ds/validate", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_ValidateByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.GET("/ds/validate/:id", h.ValidateByID)
	w := round6bJSONRequest(t, r, http.MethodGet, "/ds/validate/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_Tree_BadPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.POST("/ds/tree", h.Tree)
	w := round6bJSONRequest(t, r, http.MethodPost, "/ds/tree", []byte(`{"current":"abc"}`))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Contains(t, []string{"000000", "500000"}, resp.Code)
}

func TestRound6B_Ds_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.GET("/ds/:id", h.Get)
	w := round6bJSONRequest(t, r, http.MethodGet, "/ds/abc", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_HidePw_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.GET("/ds/hidePw/:id", h.HidePw)
	w := round6bJSONRequest(t, r, http.MethodGet, "/ds/hidePw/notid", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_GetSimpleDs_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.GET("/ds/simple/:id", h.GetSimpleDs)
	w := round6bJSONRequest(t, r, http.MethodGet, "/ds/simple/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_PerDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.POST("/ds/perDelete/:id", h.PerDelete)
	w := round6bJSONRequest(t, r, http.MethodPost, "/ds/perDelete/abc", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_SyncApiTable_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.POST("/ds/syncApiTable", h.SyncApiTable)
	w := round6bJSONRequest(t, r, http.MethodPost, "/ds/syncApiTable", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_SyncApiDs_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.POST("/ds/syncApiDs", h.SyncApiDs)
	w := round6bJSONRequest(t, r, http.MethodPost, "/ds/syncApiDs", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_LoadRemoteFile_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.POST("/ds/loadRemoteFile", h.LoadRemoteFile)
	w := round6bJSONRequest(t, r, http.MethodPost, "/ds/loadRemoteFile", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_CheckAPIDatasource_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.POST("/ds/checkApiDatasource", h.CheckAPIDatasource)
	w := round6bJSONRequest(t, r, http.MethodPost, "/ds/checkApiDatasource", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_Types_ResponseShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.GET("/ds/types", h.Types)
	w := round6bJSONRequest(t, r, http.MethodGet, "/ds/types", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var items []map[string]string
	require.NoError(t, json.Unmarshal(resp.Data, &items))
	assert.Len(t, items, 5)
	typeSet := make(map[string]bool)
	for _, item := range items {
		typeSet[item["type"]] = true
	}
	assert.True(t, typeSet["MySQL"])
	assert.True(t, typeSet["PostgreSQL"])
	assert.True(t, typeSet["Excel"])
}

func TestRound6B_Ds_ListSyncRecord_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasetTable{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDsFinishPage{}))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	r := gin.New()
	r.POST("/ds/syncRecord/:dsId/:page/:limit", h.ListSyncRecord)
	w := round6bJSONRequest(t, r, http.MethodPost, "/ds/syncRecord/abc/0/0", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Ds_GetCurrentUserID_And_GetCurrentUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, int64(0), getCurrentUserID(c))
	assert.Equal(t, "", getCurrentUsername(c))
	c.Set("userId", int64(42))
	c.Set("username", "round6b-user")
	assert.Equal(t, int64(42), getCurrentUserID(c))
	assert.Equal(t, "round6b-user", getCurrentUsername(c))
}

// ---------------------------------------------------------------------------
// Template handler tests
// ---------------------------------------------------------------------------

func TestRound6B_Tpl_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/create", h.Create)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/create", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_Save_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/save", h.Save)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/save", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.GET("/tpl/get/:id", h.Get)
	w := round6bJSONRequest(t, r, http.MethodGet, "/tpl/get/notid", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_List_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/list", h.List)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/list", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/update", h.Update)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/update", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.DELETE("/tpl/delete/:id", h.Delete)
	w := round6bJSONRequest(t, r, http.MethodDelete, "/tpl/delete/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_ListCategories_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.GET("/tpl/categories", h.ListCategories)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tpl/categories", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound6B_Tpl_DeleteWithCategory_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/delete/:id/:categoryId", h.DeleteWithCategory)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/delete/abc/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_SearchTemplates_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.GET("/tpl/search", h.SearchTemplates)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tpl/search", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound6B_Tpl_NameCheck_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/nameCheck", h.NameCheck)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/nameCheck", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_CategoryTemplateNameCheck_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/categoryTemplateNameCheck", h.CategoryTemplateNameCheck)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/categoryTemplateNameCheck", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_BatchDelete_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/batchDelete", h.BatchDelete)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/batchDelete", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_BatchUpdate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/batchUpdate", h.BatchUpdate)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/batchUpdate", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_FindCategoriesByTemplateIds_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := round6bSetupDB(t)
	require.NoError(t, db.AutoMigrate(&templateCoreVisualizationTemplateMirror{}, &auto.VisualizationTemplateCategoryMap{}))
	repo := repository.NewTemplateRepository(db)
	svc := service.NewTemplateService(repo)
	h := NewTemplateHandler(svc)
	r := gin.New()
	r.POST("/tpl/findCategoriesByTemplateIds", h.FindCategoriesByTemplateIds)
	w := round6bJSONRequest(t, r, http.MethodPost, "/tpl/findCategoriesByTemplateIds", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Tpl_TemplateCreateBy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, "0", templateCreateBy(c))
	c.Set("userName", "alice")
	assert.Equal(t, "alice", templateCreateBy(c))
}

// ---------------------------------------------------------------------------
// Chart handler tests
// ---------------------------------------------------------------------------

func TestRound6B_Chart_Query_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/query", h.Query)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/query", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_ViewOption_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.GET("/chart/viewOption/:resourceId", h.ViewOption)
	w := round6bJSONRequest(t, r, http.MethodGet, "/chart/viewOption/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_ChartBaseInfo_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.GET("/chart/baseInfo/:id/:resourceTable", h.ChartBaseInfo)
	w := round6bJSONRequest(t, r, http.MethodGet, "/chart/baseInfo/bad/core", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_Data_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/data", h.Data)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/data", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_Data_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/data", h.Data)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/data", map[string]any{"id": "bad"})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_SaveFromMap_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/save", h.SaveFromMap)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/save", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_ListByDQ_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/listByDQ/:id/:chartId", h.ListByDQ)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/listByDQ/bad/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_CopyField_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/copyField/:id/:chartId", h.CopyField)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/copyField/bad/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_DeleteField_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/deleteField/:id", h.DeleteField)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/deleteField/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_DeleteFieldByChart_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/deleteFieldByChart/:chartId", h.DeleteFieldByChart)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/deleteFieldByChart/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_GetChart_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.GET("/chart/:id", h.GetChart)
	w := round6bJSONRequest(t, r, http.MethodGet, "/chart/notid", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_GetDetail_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.GET("/chart/detail/:id", h.GetDetail)
	w := round6bJSONRequest(t, r, http.MethodGet, "/chart/detail/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_GetFieldData_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/getFieldData/:fieldId/:fieldType", h.GetFieldData)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/getFieldData/bad/text", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_GetDrillFieldData_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/getDrillFieldData/:fieldId", h.GetDrillFieldData)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/getDrillFieldData/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_InnerExportDetails_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.POST("/chart/innerExportDetails", h.InnerExportDetails)
	w := round6bJSONRequest(t, r, http.MethodPost, "/chart/innerExportDetails", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_CheckSameDataSet_InvalidSourceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.GET("/chart/checkSameDataSet/:viewIdSource/:viewIdTarget", h.CheckSameDataSet)
	w := round6bJSONRequest(t, r, http.MethodGet, "/chart/checkSameDataSet/bad/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_CheckSameDataSet_InvalidTargetID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ChartHandler{}
	r := gin.New()
	r.GET("/chart/checkSameDataSet/:viewIdSource/:viewIdTarget", h.CheckSameDataSet)
	w := round6bJSONRequest(t, r, http.MethodGet, "/chart/checkSameDataSet/1/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Chart_chartDataIDFromMap(t *testing.T) {
	id, ok := chartDataIDFromMap(nil)
	assert.False(t, ok)
	assert.Equal(t, int64(0), id)

	id, ok = chartDataIDFromMap(map[string]interface{}{})
	assert.False(t, ok)
	assert.Equal(t, int64(0), id)

	id, ok = chartDataIDFromMap(map[string]interface{}{"id": "123"})
	assert.True(t, ok)
	assert.Equal(t, int64(123), id)

	id, ok = chartDataIDFromMap(map[string]interface{}{"id": "abc"})
	assert.False(t, ok)
	assert.Equal(t, int64(0), id)

	id, ok = chartDataIDFromMap(map[string]interface{}{"id": float64(456)})
	assert.True(t, ok)
	assert.Equal(t, int64(456), id)

	id, ok = chartDataIDFromMap(map[string]interface{}{"id": 789})
	assert.True(t, ok)
	assert.Equal(t, int64(789), id)

	id, ok = chartDataIDFromMap(map[string]interface{}{"id": int64(999)})
	assert.True(t, ok)
	assert.Equal(t, int64(999), id)

	id, ok = chartDataIDFromMap(map[string]interface{}{"id": true})
	assert.False(t, ok)
	assert.Equal(t, int64(0), id)
}

func TestRound6B_Chart_chartDataResultCountFromMap(t *testing.T) {
	rc, ok := chartDataResultCountFromMap(nil)
	assert.False(t, ok)
	assert.Equal(t, 0, rc)

	rc, ok = chartDataResultCountFromMap(map[string]interface{}{})
	assert.False(t, ok)
	assert.Equal(t, 0, rc)

	rc, ok = chartDataResultCountFromMap(map[string]interface{}{"resultCount": float64(5)})
	assert.True(t, ok)
	assert.Equal(t, 5, rc)

	rc, ok = chartDataResultCountFromMap(map[string]interface{}{"resultCount": 10})
	assert.True(t, ok)
	assert.Equal(t, 10, rc)

	rc, ok = chartDataResultCountFromMap(map[string]interface{}{"resultCount": int64(15)})
	assert.True(t, ok)
	assert.Equal(t, 15, rc)

	rc, ok = chartDataResultCountFromMap(map[string]interface{}{"resultCount": "20"})
	assert.True(t, ok)
	assert.Equal(t, 20, rc)

	rc, ok = chartDataResultCountFromMap(map[string]interface{}{"resultCount": "abc"})
	assert.False(t, ok)
	assert.Equal(t, 0, rc)

	rc, ok = chartDataResultCountFromMap(map[string]interface{}{"resultCount": true})
	assert.False(t, ok)
	assert.Equal(t, 0, rc)
}

func TestRound6B_Chart_mergeChartViewIntoMap(t *testing.T) {
	mergeChartViewIntoMap(nil, nil)
	mergeChartViewIntoMap(make(map[string]interface{}), nil)
}

func TestRound6B_Chart_decodeChartViewJSONField(t *testing.T) {
	assert.Equal(t, 42, decodeChartViewJSONField("xAxis", 42))
	assert.Equal(t, "", decodeChartViewJSONField("xAxis", ""))
	assert.Equal(t, "   ", decodeChartViewJSONField("xAxis", "   "))
	assert.Equal(t, "value", decodeChartViewJSONField("unknown", "value"))
	result := decodeChartViewJSONField("xAxis", `[{"id":"1"}]`)
	arr, ok := result.([]interface{})
	require.True(t, ok)
	require.Len(t, arr, 1)
	assert.Equal(t, "{bad", decodeChartViewJSONField("xAxis", "{bad"))
}

func TestRound6B_Chart_isChartViewJSONField(t *testing.T) {
	for _, k := range []string{"xAxis", "yAxis", "customAttr", "customStyle", "customFilter", "senior", "drillFields", "snapshot"} {
		assert.True(t, isChartViewJSONField(k), k)
	}
	for _, k := range []string{"id", "title", "unknown", "name", "type"} {
		assert.False(t, isChartViewJSONField(k), k)
	}
}

// ---------------------------------------------------------------------------
// Threshold handler tests
// ---------------------------------------------------------------------------

func TestRound6B_Thresh_Save_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/save", []byte("{"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_Edit_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/edit", []byte("{"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_Pager_InvalidPageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/pager/0/10", []byte("{}"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
	resp = performThresholdRequest(t, engine, http.MethodPost, "/threshold/pager/1/0", []byte("{}"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_Pager_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/pager/1/10", []byte("{"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_FormInfo_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodGet, "/threshold/formInfo/bad/core", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_SwitchEnable_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/switch", []byte("{"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_Delete_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/delete/core", []byte("{"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_BatchReci_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/batchReci", []byte("{"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_InstancePager_InvalidPageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/instancePager/0/10", []byte("{}"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_InstancePager_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/instancePager/1/10", []byte("{"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_Preview_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodPost, "/threshold/preview", []byte("{"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_AnyThreshold_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodGet, "/threshold/anyThreshold/bad/core", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

func TestRound6B_Thresh_DeleteWithChart_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _, _ := newThresholdHandlerTestEnv(t)
	resp := performThresholdRequest(t, engine, http.MethodGet, "/threshold/deleteWithChart/bad/core", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10001", resp.Body.Code)
}

// ---------------------------------------------------------------------------
// Embedded handler tests
// ---------------------------------------------------------------------------

func TestRound6B_Embedded_QueryGrid_DefaultPageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := setupEmbeddedHandlerTestEnv(t)
	w := round6bJSONRequest(t, env.r, http.MethodPost, "/api/embedded/pager/0/0", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound6B_Embedded_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := setupEmbeddedHandlerTestEnv(t)
	w := round6bJSONRequest(t, env.r, http.MethodPost, "/api/embedded/create", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Embedded_Edit_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := setupEmbeddedHandlerTestEnv(t)
	w := round6bJSONRequest(t, env.r, http.MethodPost, "/api/embedded/edit", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Embedded_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := setupEmbeddedHandlerTestEnv(t)
	w := round6bJSONRequest(t, env.r, http.MethodPost, "/api/embedded/delete/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Embedded_BatchDelete_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := setupEmbeddedHandlerTestEnv(t)
	w := round6bJSONRequest(t, env.r, http.MethodPost, "/api/embedded/batchDelete", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Embedded_Reset_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := setupEmbeddedHandlerTestEnv(t)
	w := round6bJSONRequest(t, env.r, http.MethodPost, "/api/embedded/reset", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Embedded_InitIframe_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := setupEmbeddedHandlerTestEnv(t)
	w := round6bJSONRequest(t, env.r, http.MethodPost, "/api/embedded/initIframe", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound6B_Embedded_toString(t *testing.T) {
	assert.Equal(t, "hello", toString("hello"))
	assert.Equal(t, "42", toString(int64(42)))
	assert.Equal(t, "7", toString(7))
	assert.Equal(t, "", toString(3.14))
	assert.Equal(t, "", toString(true))
}

func TestRound6B_Embedded_toInt64(t *testing.T) {
	assert.Equal(t, int64(42), toInt64(int64(42)))
	assert.Equal(t, int64(7), toInt64(7))
	assert.Equal(t, int64(99), toInt64(float64(99)))
	assert.Equal(t, int64(55), toInt64("55"))
	assert.Equal(t, int64(0), toInt64("abc"))
	assert.Equal(t, int64(0), toInt64(true))
}

func TestRound6B_Embedded_getUpdateBy_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h := &EmbeddedHandler{}
	assert.Equal(t, "system", h.getUpdateBy(c))
}

func TestRound6B_Embedded_getCurrentUser_NoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h := &EmbeddedHandler{}
	uid, oid := h.getCurrentUser(c)
	assert.Equal(t, int64(1), uid)
	assert.Equal(t, int64(1), oid)
}

// ---------------------------------------------------------------------------
// Audit handler tests (non-integration, handler-level only)
// ---------------------------------------------------------------------------

func TestRound6B_Audit_GetAuditAlertSettings_NilSvc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuditHandler{}
	r := gin.New()
	r.GET("/audit/settings", h.GetAuditAlertSettings)
	w := round6bJSONRequest(t, r, http.MethodGet, "/audit/settings", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound6B_Audit_SaveAuditAlertSettings_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuditHandler{}
	r := gin.New()
	r.PUT("/audit/settings", h.SaveAuditAlertSettings)
	w := round6bJSONRequest(t, r, http.MethodPut, "/audit/settings", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Audit_CleanupNow_NilSvc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuditHandler{}
	r := gin.New()
	r.POST("/audit/cleanup", h.CleanupNow)
	w := round6bJSONRequest(t, r, http.MethodPost, "/audit/cleanup", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound6B_Audit_TestNotification_NilSvc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuditHandler{}
	r := gin.New()
	r.POST("/audit/test-notification", h.TestNotification)
	w := round6bJSONRequest(t, r, http.MethodPost, "/audit/test-notification", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound6B_Audit_CreateAuditLog_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuditHandler{}
	r := gin.New()
	r.POST("/audit/log", h.CreateAuditLog)
	w := round6bJSONRequest(t, r, http.MethodPost, "/audit/log", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Audit_GetAuditLogByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuditHandler{}
	r := gin.New()
	r.GET("/audit/:id", h.GetAuditLogByID)
	w := round6bJSONRequest(t, r, http.MethodGet, "/audit/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Audit_GetAuditLogsByUserID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuditHandler{}
	r := gin.New()
	r.GET("/audit/user/:userId", h.GetAuditLogsByUserID)
	w := round6bJSONRequest(t, r, http.MethodGet, "/audit/user/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Audit_ExportAuditLogs_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuditHandler{}
	r := gin.New()
	r.POST("/audit/export", h.ExportAuditLogs)
	w := round6bJSONRequest(t, r, http.MethodPost, "/audit/export", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound6B_Audit_RecordLoginFailure_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuditHandler{}
	r := gin.New()
	r.POST("/audit/login-failure", h.RecordLoginFailure)
	w := round6bJSONRequest(t, r, http.MethodPost, "/audit/login-failure", []byte("{"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round6bDecodeResp(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}
