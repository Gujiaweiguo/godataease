package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type datasetHandlerTestEnv struct {
	r  *gin.Engine
	db *gorm.DB
}

type datasetHandlerResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func setupDatasetHandlerTestEnv(t *testing.T) *datasetHandlerTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}))

	repo := repository.NewDatasetRepository(db)
	svc := service.NewDatasetService(repo)
	h := NewDatasetHandler(svc, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1001))
		c.Set("userId", int64(1001))
		c.Set("username", "dataset-tester")
		c.Next()
	})
	RegisterDatasetRoutes(r.Group("/api"), h)

	return &datasetHandlerTestEnv{r: r, db: db}
}

func performDatasetJSONRequest(t *testing.T, r *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
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

func decodeDatasetResp(t *testing.T, body []byte) datasetHandlerResp {
	t.Helper()

	var resp datasetHandlerResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func seedDatasetGroupRecord(t *testing.T, db *gorm.DB, group *dataset.CoreDatasetGroup) {
	t.Helper()
	require.NoError(t, db.Create(group).Error)
}

func int64PtrForDatasetHandler(v int64) *int64 {
	return &v
}

func intPtrForDatasetHandler(v int) *int {
	return &v
}

func stringPtrForDatasetHandler(v string) *string {
	return &v
}

func TestDatasetHandler_Tree_ReturnsEmptyArrayForEmptyDB(t *testing.T) {
	env := setupDatasetHandlerTestEnv(t)

	w := performDatasetJSONRequest(t, env.r, http.MethodPost, "/api/dataset/tree", map[string]any{})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data []dataset.TreeNode
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Empty(t, data)
}

func TestDatasetHandler_BarInfo_ReturnsSeededDatasetAuditFields(t *testing.T) {
	env := setupDatasetHandlerTestEnv(t)
	seedDatasetGroupRecord(t, env.db, &dataset.CoreDatasetGroup{
		ID:             101,
		Name:           "sales-folder",
		PID:            int64PtrForDatasetHandler(0),
		Level:          intPtrForDatasetHandler(0),
		NodeType:       stringPtrForDatasetHandler(dataset.NodeTypeFolder),
		CreateBy:       "2001",
		CreateTime:     1710000001,
		UpdateBy:       "2002",
		LastUpdateTime: 1710009999,
		DelFlag:        intPtrForDatasetHandler(0),
	})

	w := performDatasetJSONRequest(t, env.r, http.MethodGet, "/api/dataset/barInfo/101", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data dataset.BarInfo
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, int64(101), data.ID)
	assert.Equal(t, "sales-folder", data.Name)
	assert.Equal(t, dataset.NodeTypeFolder, data.NodeType)
	assert.Equal(t, "2001", data.CreateBy)
	assert.Equal(t, "2001", data.Creator)
	assert.Equal(t, int64(1710000001), data.CreateTime)
	assert.Equal(t, "2002", data.UpdateBy)
	assert.Equal(t, "2002", data.Updater)
	assert.Equal(t, int64(1710009999), data.LastUpdateTime)
	assert.False(t, data.IsCross)
	assert.Empty(t, data.DatasourceDTOList)
}

func TestDatasetHandler_Create_CreatesDatasetGroup(t *testing.T) {
	env := setupDatasetHandlerTestEnv(t)

	w := performDatasetJSONRequest(t, env.r, http.MethodPost, "/api/dataset/create", map[string]any{
		"name":     "new-group",
		"nodeType": dataset.NodeTypeFolder,
		"pid":      0,
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var created dataset.CoreDatasetGroup
	require.NoError(t, json.Unmarshal(resp.Data, &created))
	assert.NotZero(t, created.ID)
	assert.Equal(t, "new-group", created.Name)
	require.NotNil(t, created.PID)
	assert.Equal(t, int64(0), *created.PID)
	require.NotNil(t, created.Level)
	assert.Equal(t, 0, *created.Level)
	require.NotNil(t, created.NodeType)
	assert.Equal(t, dataset.NodeTypeFolder, *created.NodeType)
	require.NotNil(t, created.DelFlag)
	assert.Equal(t, 0, *created.DelFlag)

	stored, err := repository.NewDatasetRepository(env.db).GetGroupByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "new-group", stored.Name)
	require.NotNil(t, stored.NodeType)
	assert.Equal(t, dataset.NodeTypeFolder, *stored.NodeType)
}

func TestDatasetHandler_Rename_RenamesDatasetGroup(t *testing.T) {
	env := setupDatasetHandlerTestEnv(t)
	seedDatasetGroupRecord(t, env.db, &dataset.CoreDatasetGroup{
		ID:       201,
		Name:     "before-rename",
		PID:      int64PtrForDatasetHandler(0),
		Level:    intPtrForDatasetHandler(0),
		NodeType: stringPtrForDatasetHandler(dataset.NodeTypeFolder),
		DelFlag:  intPtrForDatasetHandler(0),
	})

	w := performDatasetJSONRequest(t, env.r, http.MethodPost, "/api/dataset/rename", map[string]any{
		"id":   201,
		"name": "after-rename",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var renamed dataset.CoreDatasetGroup
	require.NoError(t, json.Unmarshal(resp.Data, &renamed))
	assert.Equal(t, int64(201), renamed.ID)
	assert.Equal(t, "after-rename", renamed.Name)

	stored, err := repository.NewDatasetRepository(env.db).GetGroupByID(201)
	require.NoError(t, err)
	assert.Equal(t, "after-rename", stored.Name)
}

func TestDatasetHandler_Move_MovesDatasetGroupToNewParent(t *testing.T) {
	env := setupDatasetHandlerTestEnv(t)
	seedDatasetGroupRecord(t, env.db, &dataset.CoreDatasetGroup{
		ID:       301,
		Name:     "source-group",
		PID:      int64PtrForDatasetHandler(0),
		Level:    intPtrForDatasetHandler(0),
		NodeType: stringPtrForDatasetHandler(dataset.NodeTypeFolder),
		DelFlag:  intPtrForDatasetHandler(0),
	})
	seedDatasetGroupRecord(t, env.db, &dataset.CoreDatasetGroup{
		ID:       302,
		Name:     "target-parent",
		PID:      int64PtrForDatasetHandler(0),
		Level:    intPtrForDatasetHandler(0),
		NodeType: stringPtrForDatasetHandler(dataset.NodeTypeFolder),
		DelFlag:  intPtrForDatasetHandler(0),
	})

	w := performDatasetJSONRequest(t, env.r, http.MethodPost, "/api/dataset/move", map[string]any{
		"id":  301,
		"pid": 302,
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var moved dataset.CoreDatasetGroup
	require.NoError(t, json.Unmarshal(resp.Data, &moved))
	assert.Equal(t, int64(301), moved.ID)
	require.NotNil(t, moved.PID)
	assert.Equal(t, int64(302), *moved.PID)

	stored, err := repository.NewDatasetRepository(env.db).GetGroupByID(301)
	require.NoError(t, err)
	require.NotNil(t, stored.PID)
	assert.Equal(t, int64(302), *stored.PID)
}

func TestDatasetHandler_Delete_SoftDeletesDatasetGroup(t *testing.T) {
	env := setupDatasetHandlerTestEnv(t)
	seedDatasetGroupRecord(t, env.db, &dataset.CoreDatasetGroup{
		ID:       401,
		Name:     "delete-me",
		PID:      int64PtrForDatasetHandler(0),
		Level:    intPtrForDatasetHandler(0),
		NodeType: stringPtrForDatasetHandler(dataset.NodeTypeFolder),
		DelFlag:  intPtrForDatasetHandler(0),
	})

	w := performDatasetJSONRequest(t, env.r, http.MethodPost, "/api/dataset/delete/401", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasetResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Empty(t, resp.Data)

	var stored dataset.CoreDatasetGroup
	require.NoError(t, env.db.Unscoped().First(&stored, 401).Error)
	require.NotNil(t, stored.DelFlag)
	assert.Equal(t, 1, *stored.DelFlag)
}
