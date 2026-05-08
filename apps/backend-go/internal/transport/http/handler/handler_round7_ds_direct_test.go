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

func round7Setup(t *testing.T) (*DatasourceHandler, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:round7ds_%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(
		&datasource.CoreDatasource{},
		&auto.CoreDatasetTable{},
		&auto.CoreDatasourceTaskLog{},
		&auto.CoreDsFinishPage{},
	))
	repo := repository.NewDatasourceRepository(db)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	return h, db
}

func round7Ctx(t *testing.T, method string, body string, userID int64, username string, params gin.Params) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, "/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, "/", nil)
	}
	c.Request = req
	c.Set("userId", userID)
	c.Set("user_id", userID)
	c.Set("username", username)
	c.Params = params
	return w, c
}

func round7Decode(t *testing.T, body []byte) datasourceHandlerBridgeResp {
	t.Helper()
	var resp datasourceHandlerBridgeResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func TestRound7DsDirect_NewDatasourceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := repository.NewDatasourceRepository(nil)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)
	assert.NotNil(t, h)
	assert.Equal(t, svc, h.service)
}

func TestRound7DsDirect_List_EmptyName(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, `{"name":""}`, 1, "tester", nil)
	h.List(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7DsDirect_Validate_MissingBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, ``, 1, "tester", nil)
	h.Validate(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_ValidateByID_Success(t *testing.T) {
	h, db := round7Setup(t)
	status := datasource.StatusSuccess
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 501, PID: int64PtrForDatasourceHandler(0), Name: "validate-ds", Type: "MySQL", Status: &status})

	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "501"}})
	h.ValidateByID(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7DsDirect_ValidateByID_BadID(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "abc"}})
	h.ValidateByID(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_ValidateByID_NotFound(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "99999"}})
	h.ValidateByID(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7DsDirect_Tree_EmptyBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, `{}`, 1, "tester", nil)
	h.Tree(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7DsDirect_Get_Success(t *testing.T) {
	h, db := round7Setup(t)
	config := `{"host":"127.0.0.1"}`
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 601, PID: int64PtrForDatasourceHandler(0), Name: "get-ds", Type: "MySQL", Configuration: &config})

	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "601"}})
	h.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7DsDirect_Get_BadID(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "notanumber"}})
	h.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_Get_NotFound(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "88888"}})
	h.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_HidePw_Success(t *testing.T) {
	h, db := round7Setup(t)
	config := `{"password":"secret123","host":"10.0.0.1"}`
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 701, PID: int64PtrForDatasourceHandler(0), Name: "hidepw-ds", Type: "MySQL", Configuration: &config})

	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "701"}})
	h.HidePw(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7DsDirect_HidePw_BadID(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "xyz"}})
	h.HidePw(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_GetSimpleDs_Success(t *testing.T) {
	h, db := round7Setup(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 801, PID: int64PtrForDatasourceHandler(0), Name: "simple-ds", Type: "PostgreSQL"})

	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "801"}})
	h.GetSimpleDs(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Contains(t, string(resp.Data), `"simple-ds"`)
	assert.Contains(t, string(resp.Data), `"PostgreSQL"`)
}

func TestRound7DsDirect_GetSimpleDs_BadID(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{{Key: "id", Value: "nope"}})
	h.GetSimpleDs(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_PerDelete_Success(t *testing.T) {
	h, db := round7Setup(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 901, PID: int64PtrForDatasourceHandler(0), Name: "perdel-ds", Type: "MySQL"})

	w, c := round7Ctx(t, http.MethodPost, "", 1, "tester", gin.Params{{Key: "id", Value: "901"}})
	h.PerDelete(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7DsDirect_PerDelete_BadID(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, "", 1, "tester", gin.Params{{Key: "id", Value: "bad"}})
	h.PerDelete(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_SyncApiTable_MissingBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, ``, 1, "tester", nil)
	h.SyncApiTable(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_SyncApiTable_WithBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, `{"tableName":"test_table"}`, 1, "tester", nil)
	h.SyncApiTable(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.NotEmpty(t, resp.Code)
}

func TestRound7DsDirect_SyncApiDs_MissingBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, ``, 1, "tester", nil)
	h.SyncApiDs(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_SyncApiDs_WithBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, `{"datasourceId":"1"}`, 1, "tester", nil)
	h.SyncApiDs(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.NotEmpty(t, resp.Code)
}

func TestRound7DsDirect_LoadRemoteFile_MissingBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, ``, 1, "tester", nil)
	h.LoadRemoteFile(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_LoadRemoteFile_WithBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, `{"url":"https://example.com/file.xlsx","userName":"","passwd":"","datasourceId":0}`, 1, "tester", nil)
	h.LoadRemoteFile(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.NotEmpty(t, resp.Code)
}

func TestRound7DsDirect_CheckAPIDatasource_MissingBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, ``, 1, "tester", nil)
	h.CheckAPIDatasource(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_CheckAPIDatasource_WithBody(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodPost, `{"url":"https://api.example.com/data"}`, 1, "tester", nil)
	h.CheckAPIDatasource(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.NotEmpty(t, resp.Code)
}

func TestRound7DsDirect_Types(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", nil)
	h.Types(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Contains(t, string(resp.Data), `"MySQL"`)
	assert.Contains(t, string(resp.Data), `"Excel"`)
}

func TestRound7DsDirect_ListSyncRecord_Success(t *testing.T) {
	h, db := round7Setup(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 1001, PID: int64PtrForDatasourceHandler(0), Name: "sync-ds", Type: "MySQL"})
	seedDatasourceTaskLogRecord(t, db, &auto.CoreDatasourceTaskLog{ID: 4001, DsID: 1001, TaskID: 1, TaskStatus: "completed", PhysicalTableName: "orders", CreateTime: 100, EndTime: 200})

	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{
		{Key: "dsId", Value: "1001"},
		{Key: "page", Value: "1"},
		{Key: "limit", Value: "10"},
	})
	h.ListSyncRecord(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Contains(t, string(resp.Data), `"datasourceId":1001`)
}

func TestRound7DsDirect_ListSyncRecord_BadDsID(t *testing.T) {
	h, _ := round7Setup(t)
	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{
		{Key: "dsId", Value: "bad"},
		{Key: "page", Value: "1"},
		{Key: "limit", Value: "10"},
	})
	h.ListSyncRecord(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound7DsDirect_ListSyncRecord_DefaultPagination(t *testing.T) {
	h, db := round7Setup(t)
	seedDatasourceRecord(t, db, &datasource.CoreDatasource{ID: 1002, PID: int64PtrForDatasourceHandler(0), Name: "sync-ds2", Type: "MySQL"})

	w, c := round7Ctx(t, http.MethodGet, "", 1, "tester", gin.Params{
		{Key: "dsId", Value: "1002"},
		{Key: "page", Value: "0"},
		{Key: "limit", Value: "0"},
	})
	h.ListSyncRecord(c)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round7Decode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound7DsDirect_getCurrentUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	assert.Equal(t, int64(0), getCurrentUserID(c))

	c.Set("userId", int64(42))
	assert.Equal(t, int64(42), getCurrentUserID(c))

	c.Set("userId", "not-an-int64")
	assert.Equal(t, int64(0), getCurrentUserID(c))
}

func TestRound7DsDirect_getCurrentUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	assert.Equal(t, "", getCurrentUsername(c))

	c.Set("username", "alice")
	assert.Equal(t, "alice", getCurrentUsername(c))

	c.Set("username", 12345)
	assert.Equal(t, "", getCurrentUsername(c))
}
