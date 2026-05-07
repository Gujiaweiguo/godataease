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
	datasourcedomain "dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type syncGapResp struct {
	Code string          `json:"code"`
	Data json.RawMessage `json:"data"`
}

func setupSyncHandlerGap2(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&auto.CoreDatasourceTask{}, &auto.CoreDatasourceTaskLog{}, &datasourcedomain.CoreDatasource{}))
	syncRepo := repository.NewSyncRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	dsSvc := service.NewDatasourceService(dsRepo)
	h := NewSyncHandler(service.NewSyncService(syncRepo, dsRepo, dsSvc))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("username", "gap2-user")
		c.Next()
	})
	RegisterSyncRoutes(r.Group("/api"), h)
	return r, db
}

func decodeSyncGapResp(t *testing.T, body []byte) syncGapResp {
	t.Helper()
	var resp syncGapResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func performSyncGapRequest(t *testing.T, r *gin.Engine, method, url, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, url, bytes.NewBufferString(body))
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func seedSyncTask(t *testing.T, db *gorm.DB, task *auto.CoreDatasourceTask) {
	t.Helper()
	require.NoError(t, db.Create(task).Error)
}

func seedSyncTaskLog(t *testing.T, db *gorm.DB, log *auto.CoreDatasourceTaskLog) {
	t.Helper()
	require.NoError(t, db.Create(log).Error)
}

func seedSyncDatasource(t *testing.T, db *gorm.DB, ds *datasourcedomain.CoreDatasource) {
	t.Helper()
	require.NoError(t, db.Create(ds).Error)
}

func TestSyncHandlerGap2_TaskCRUDAndPager(t *testing.T) {
	r, db := setupSyncHandlerGap2(t)
	seedSyncDatasource(t, db, &datasourcedomain.CoreDatasource{ID: 1, Name: "source-ds", Type: "mysql"})

	w := performSyncGapRequest(t, r, http.MethodPost, "/api/sync/task/add", `{"name":"task-a","schedulerType":"CRON","schedulerConf":"0 0 * * * *","source":{"datasourceId":"1"},"target":{"datasourceId":"1","tableName":"demo"}}`)
	resp := decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performSyncGapRequest(t, r, http.MethodPost, "/api/sync/task/pager/1/10", `{}`)
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var page struct {
		Records []map[string]any `json:"records"`
		Total   int64            `json:"total"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &page))
	require.Len(t, page.Records, 1)
	assert.Equal(t, int64(1), page.Total)
	taskID := page.Records[0]["id"].(string)

	w = performSyncGapRequest(t, r, http.MethodGet, "/api/sync/task/get/"+taskID, "")
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var task map[string]any
	require.NoError(t, json.Unmarshal(resp.Data, &task))
	assert.Equal(t, "task-a", task["name"])

	w = performSyncGapRequest(t, r, http.MethodPost, "/api/sync/task/update", `{"id":"`+taskID+`","name":"task-b","schedulerType":"CRON","schedulerConf":"0 5 * * * *","source":{"datasourceId":"1"},"target":{"datasourceId":"1","tableName":"demo2"}}`)
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performSyncGapRequest(t, r, http.MethodGet, "/api/sync/task/get/"+taskID, "")
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	require.NoError(t, json.Unmarshal(resp.Data, &task))
	assert.Equal(t, "task-b", task["name"])

	w = performSyncGapRequest(t, r, http.MethodPost, "/api/sync/task/remove/"+taskID, "")
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestSyncHandlerGap2_TaskLogsAndResourceCount(t *testing.T) {
	r, db := setupSyncHandlerGap2(t)
	seedSyncDatasource(t, db, &datasourcedomain.CoreDatasource{ID: 11, Name: "mysql-a", Type: "mysql"})
	seedSyncDatasource(t, db, &datasourcedomain.CoreDatasource{ID: 12, Name: "folder-a", Type: "folder"})
	seedSyncTask(t, db, &auto.CoreDatasourceTask{ID: 21, DsID: 11, Name: "sync-job", UpdateType: "sync", SyncRate: "1", Cron: "0 0 * * * *", TaskStatus: "Pending", CreateTime: 1000})
	seedSyncTask(t, db, &auto.CoreDatasourceTask{ID: 22, DsID: 11, Name: "sync-job-2", UpdateType: "sync", SyncRate: "1", Cron: "0 1 * * * *", TaskStatus: "Pending", CreateTime: 2000})
	seedSyncTaskLog(t, db, &auto.CoreDatasourceTaskLog{ID: 31, DsID: 11, TaskID: 21, StartTime: 100, EndTime: 120, TaskStatus: "Success", PhysicalTableName: "job-a", Info: "line1\nline2", CreateTime: 1000})
	seedSyncTaskLog(t, db, &auto.CoreDatasourceTaskLog{ID: 32, DsID: 11, TaskID: 21, StartTime: 200, EndTime: 220, TaskStatus: "Failed", PhysicalTableName: "job-a", Info: "boom", CreateTime: 2000})

	w := performSyncGapRequest(t, r, http.MethodPost, "/api/sync/task/log/pager/1/10", `{"jobId":"21"}`)
	resp := decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var page struct {
		Records []map[string]any `json:"records"`
		Total   int64            `json:"total"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &page))
	require.Len(t, page.Records, 2)
	assert.Equal(t, int64(2), page.Total)

	w = performSyncGapRequest(t, r, http.MethodGet, "/api/sync/task/log/detail/31/1", "")
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var detail map[string]any
	require.NoError(t, json.Unmarshal(resp.Data, &detail))
	assert.Equal(t, float64(1), detail["fromLineNum"])
	assert.Equal(t, float64(2), detail["toLineNum"])

	w = performSyncGapRequest(t, r, http.MethodGet, "/api/sync/summary/resourceCount", "")
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var count map[string]any
	require.NoError(t, json.Unmarshal(resp.Data, &count))
	assert.Equal(t, float64(2), count["jobCount"])
	assert.Equal(t, float64(1), count["datasourceCount"])
	assert.Equal(t, float64(2), count["jobLogCount"])

	w = performSyncGapRequest(t, r, http.MethodPost, "/api/sync/task/log/delete/31", "")
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performSyncGapRequest(t, r, http.MethodPost, "/api/sync/task/log/clear", `{"jobId":"21"}`)
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestSyncHandlerGap2_DatasourceFlows(t *testing.T) {
	r, db := setupSyncHandlerGap2(t)
	createBy := "gap2-user"
	status := datasourcedomain.StatusSuccess
	seedSyncDatasource(t, db, &datasourcedomain.CoreDatasource{ID: 41, Name: "mysql-one", Type: "mysql", CreateBy: &createBy, Status: &status})
	seedSyncDatasource(t, db, &datasourcedomain.CoreDatasource{ID: 42, Name: "mysql-two", Type: "mysql", CreateBy: &createBy, Status: &status})

	w := performSyncGapRequest(t, r, http.MethodPost, "/api/sync/datasource/source/pager/1/10", `{}`)
	resp := decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var page struct {
		Records []map[string]any `json:"records"`
		Total   int64            `json:"total"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &page))
	assert.Equal(t, int64(2), page.Total)

	w = performSyncGapRequest(t, r, http.MethodGet, "/api/sync/datasource/get/41", "")
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var ds map[string]any
	require.NoError(t, json.Unmarshal(resp.Data, &ds))
	assert.Equal(t, "41", ds["id"])

	w = performSyncGapRequest(t, r, http.MethodPost, "/api/sync/datasource/batchDel", `["41","42"]`)
	resp = decodeSyncGapResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var count int64
	require.NoError(t, db.Model(&datasourcedomain.CoreDatasource{}).Where("COALESCE(del_flag, 0) = 0").Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
