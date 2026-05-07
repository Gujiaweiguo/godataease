package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDataFillingHandlerGap2(t *testing.T) (*DataFillingHandler, *serviceTestDataFillingRepoBridge) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "root-form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: `[{"type":"input","settings":{"name":"Name","mapping":{"columnName":"name","type":"nvarchar"}}}]`}
	repo.records[2] = &datafillingdomain.DataFillingForm{ID: 2, Name: "folder", NodeType: datafillingdomain.NodeTypeFolder, PID: 0}
	taskRepo := &userTaskHandlerFakeTaskRepo{records: map[int64]*datafillingdomain.DataFillingTask{1: {ID: 1, FormID: 1, Name: "task-one", ReciFlagList: "[1]", UIDList: "[9]", RIDList: "[]", RateVal: "0 0 * * * *"}}}
	subTaskRepo := &userTaskHandlerFakeSubTaskRepo{records: map[int64]*datafillingdomain.DataFillingSubTask{10: {ID: 10, TaskID: 1, Status: datafillingdomain.SubTaskStatusActive}}}
	subInstanceRepo := &userTaskHandlerFakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, PID: 10, TaskID: 1, UID: 9, FormID: 1, Status: datafillingdomain.SubInstanceStatusOpen}}}
	svc := service.NewDataFillingService(repo, &serviceTestDataFillingDatasourceServiceBridge{}, &serviceTestDataFillingDDLBridge{}, &serviceTestCommitLogRepoBridge{}, taskRepo, subTaskRepo, subInstanceRepo, nil)
	svc.SetDatasourceConnectionProvider(&serviceTestDatasourceConnProviderBridge{})
	return NewDataFillingHandler(svc), repo
}

func TestDataFillingHandlerGap2_GetDeleteUpdate(t *testing.T) {
	h, repo := setupDataFillingHandlerGap2(t)

	t.Run("get success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/get/1", nil)
			c.Params = gin.Params{{Key: "id", Value: "1"}}
			h.Get(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
		var got datafillingdomain.DataFillingForm
		require.NoError(t, json.Unmarshal(resp.Body.Data, &got))
		assert.Equal(t, int64(1), got.ID)
	})

	t.Run("delete success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/delete/2", nil)
			c.Params = gin.Params{{Key: "id", Value: "2"}}
			h.Delete(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
		_, exists := repo.records[2]
		assert.False(t, exists)
	})

	t.Run("update bad json", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/update", bytes.NewBufferString("{"))
			c.Request.Header.Set("Content-Type", "application/json")
			h.Update(c)
		})
		assert.Equal(t, "10001", resp.Body.Code)
	})

	t.Run("update success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/update", bytes.NewBufferString(`{"id":1,"name":"updated-form","nodeType":"form","tableName":"df_demo","datasourceId":8,"forms":"[]"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("user_id", uint64(9))
			h.Update(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
		assert.Equal(t, "updated-form", repo.records[1].Name)
	})
}

func TestDataFillingHandlerGap2_TableTreeAndLogs(t *testing.T) {
	h, repo := setupDataFillingHandlerGap2(t)
	repo.records[3] = &datafillingdomain.DataFillingForm{ID: 3, Name: "child-form", PID: 2, NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_child", DatasourceID: 8, Forms: "[]"}

	t.Run("table data bad form id", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/0/tableData", bytes.NewBufferString(`{}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "formId", Value: "0"}}
			h.TableData(c)
		})
		assert.Equal(t, "10001", resp.Body.Code)
	})

	t.Run("tree success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/tree", bytes.NewBufferString(`{}`))
			c.Request.Header.Set("Content-Type", "application/json")
			h.Tree(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
		var tree []datafillingdomain.TreeNode
		require.NoError(t, json.Unmarshal(resp.Body.Data, &tree))
		assert.NotEmpty(t, tree)
	})

	t.Run("save row data success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/rowData/save", bytes.NewBufferString(`{"name":"alice"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "formId", Value: "1"}}
			c.Set("user_id", uint64(9))
			c.Set("username", "tester")
			h.SaveRowData(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
	})

	t.Run("batch delete row data success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/batch-delete", bytes.NewBufferString(`{"ids":["r1","r2"]}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "formId", Value: "1"}}
			c.Set("user_id", uint64(9))
			c.Set("username", "tester")
			h.BatchDeleteRowData(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
	})

	t.Run("log page invalid pagination", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/log/page/0/10", bytes.NewBufferString(`{"formId":1}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "goPage", Value: "0"}, {Key: "pageSize", Value: "10"}}
			h.LogPage(c)
		})
		assert.Equal(t, "10001", resp.Body.Code)
	})

	t.Run("log page success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/log/page/1/10", bytes.NewBufferString(`{"formId":1}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
			h.LogPage(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
		var page map[string]any
		require.NoError(t, json.Unmarshal(resp.Body.Data, &page))
		assert.Equal(t, float64(1), page["total"])
	})

	t.Run("excel template success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/form/1/excelTemplate", nil)
		c.Params = gin.Params{{Key: "formId", Value: "1"}}
		h.ExcelTemplate(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), mimeExcelOpenXML)
		assert.NotEmpty(t, w.Body.Bytes())
	})
}
