package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domainauth "dataease/backend/internal/domain/auth"
	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/domain/datasource"
	domainsync "dataease/backend/internal/domain/syncmodule"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Shared round9 helpers ----------

func newRound9Ctx(t *testing.T, method, path, body string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return w, c
}

func parseRound9Resp(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func assertCode(t *testing.T, w *httptest.ResponseRecorder, expected string) {
	t.Helper()
	resp := parseRound9Resp(t, w)
	assert.Equal(t, expected, resp["code"])
}

func round9SetUser(c *gin.Context, userID uint64) {
	c.Set("user_id", userID)
}

func round9SetOrg(c *gin.Context, orgID int64) {
	c.Set("org_id", orgID)
}

func round9SetUsername(c *gin.Context, name string) {
	c.Set("username", name)
}

// =====================================================================
// DataFillingHandler tests (28 functions < 70%)
// =====================================================================

// --- Save ---

func TestRound9A_DataFilling_Save_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	round9SetUser(c, 1)
	h.Save(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_Save_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"test"}`)
	round9SetUser(c, 1)
	h.Save(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Get ---

func TestRound9A_DataFilling_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Get(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_Get_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.Get(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Update ---

func TestRound9A_DataFilling_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Update(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_Update_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"test"}`)
	h.Update(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Delete ---

func TestRound9A_DataFilling_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Delete(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_Delete_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- TableData ---

func TestRound9A_DataFilling_TableData_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.TableData(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_TableData_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.TableData(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_TableData_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"page":1}`)
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.TableData(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveRowData ---

func TestRound9A_DataFilling_SaveRowData_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.SaveRowData(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_SaveRowData_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.SaveRowData(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- DeleteRowData ---

func TestRound9A_DataFilling_DeleteRowData_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.DeleteRowData(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_DeleteRowData_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "id", Value: "10"}}
	h.DeleteRowData(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- BatchDeleteRowData ---

func TestRound9A_DataFilling_BatchDeleteRowData_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.BatchDeleteRowData(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_BatchDeleteRowData_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.BatchDeleteRowData(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- TruncateTableData ---

func TestRound9A_DataFilling_TruncateTableData_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.TruncateTableData(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_TruncateTableData_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.TruncateTableData(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListColumnData ---

func TestRound9A_DataFilling_ListColumnData_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.ListColumnData(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_ListColumnData_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.ListColumnData(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- ExcelTemplate ---

func TestRound9A_DataFilling_ExcelTemplate_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.ExcelTemplate(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_ExcelTemplate_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.ExcelTemplate(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ExcelUpload ---

func TestRound9A_DataFilling_ExcelUpload_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.ExcelUpload(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_ExcelUpload_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Request.Header.Del("Content-Type")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.ExcelUpload(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ConfirmUpload ---

func TestRound9A_DataFilling_ConfirmUpload_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.ConfirmUpload(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_ConfirmUpload_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.ConfirmUpload(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- ExtraDetails ---

func TestRound9A_DataFilling_ExtraDetails_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.ExtraDetails(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_ExtraDetails_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"ids":[1,2]}`)
	h.ExtraDetails(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListDatasourceOptions ---

func TestRound9A_DataFilling_ListDatasourceOptions_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.ListDatasourceOptions(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_ListDatasourceOptions_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.ListDatasourceOptions(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- GetTemplateByUserTaskItem ---

func TestRound9A_DataFilling_GetTemplateByUserTaskItem_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "itemId", Value: "bad"}}
	h.GetTemplateByUserTaskItem(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_GetTemplateByUserTaskItem_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "itemId", Value: "1"}}
	h.GetTemplateByUserTaskItem(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ExportFormData ---

func TestRound9A_DataFilling_ExportFormData_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.ExportFormData(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_ExportFormData_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "isDataEaseBi", Value: "0"}, {Key: "formId", Value: "1"}}
	h.ExportFormData(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- LogPage ---

func TestRound9A_DataFilling_LogPage_InvalidPageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "goPage", Value: "0"}, {Key: "pageSize", Value: "10"}}
	h.LogPage(c)
	// parseThresholdPageParams rejects goPage=0
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_DataFilling_LogPage_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.LogPage(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_LogPage_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"formId":1}`)
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.LogPage(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- LogClear ---

func TestRound9A_DataFilling_LogClear_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.LogClear(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_LogClear_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"formId":1}`)
	h.LogClear(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetTaskInfo ---

func TestRound9A_DataFilling_GetTaskInfo_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "taskId", Value: "bad"}}
	h.GetTaskInfo(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_GetTaskInfo_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "taskId", Value: "1"}}
	h.GetTaskInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveTask ---

func TestRound9A_DataFilling_SaveTask_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveTask(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_SaveTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"task1"}`)
	h.SaveTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ExecuteNowTask ---

func TestRound9A_DataFilling_ExecuteNowTask_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.ExecuteNowTask(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_ExecuteNowTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"taskId":1}`)
	h.ExecuteNowTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- TaskPageList ---

func TestRound9A_DataFilling_TaskPageList_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}, {Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.TaskPageList(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_TaskPageList_InvalidPageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "goPage", Value: "0"}, {Key: "pageSize", Value: "10"}}
	h.TaskPageList(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_DataFilling_TaskPageList_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.TaskPageList(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- StartTask ---

func TestRound9A_DataFilling_StartTask_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}, {Key: "id", Value: "1"}}
	h.StartTask(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_StartTask_InvalidTaskID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "id", Value: "bad"}}
	h.StartTask(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- StopTask ---

func TestRound9A_DataFilling_StopTask_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}, {Key: "id", Value: "1"}}
	h.StopTask(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_StopTask_InvalidTaskID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "id", Value: "bad"}}
	h.StopTask(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- DeleteTasks ---

func TestRound9A_DataFilling_DeleteTasks_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.DeleteTasks(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_DeleteTasks_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.DeleteTasks(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- SubTaskPageList ---

func TestRound9A_DataFilling_SubTaskPageList_InvalidPageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "goPage", Value: "0"}, {Key: "pageSize", Value: "10"}}
	h.SubTaskPageList(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_DataFilling_SubTaskPageList_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
	h.SubTaskPageList(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- DeleteSubTasks ---

func TestRound9A_DataFilling_DeleteSubTasks_InvalidFormID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "formId", Value: "bad"}}
	h.DeleteSubTasks(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_DeleteSubTasks_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	c.Params = gin.Params{{Key: "formId", Value: "1"}}
	h.DeleteSubTasks(c)
	assertCode(t, w, response.CodeBadRequest)
}

// --- SubTaskUsersList ---

func TestRound9A_DataFilling_SubTaskUsersList_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}, {Key: "type", Value: "user"}}
	h.SubTaskUsersList(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_SubTaskUsersList_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "type", Value: "user"}}
	h.SubTaskUsersList(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Rename ---

func TestRound9A_DataFilling_Rename_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Rename(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_Rename_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"new-name"}`)
	h.Rename(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Move ---

func TestRound9A_DataFilling_Move_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Move(c)
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_Move_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"pid":0}`)
	h.Move(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Tree ---

func TestRound9A_DataFilling_Tree_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Tree(c)
	// shouldBindOptionalJSON returns error for bad json
	assertCode(t, w, response.CodeBadRequest)
}

func TestRound9A_DataFilling_Tree_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	// shouldBindOptionalJSON tolerates empty body for optional fields
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	h.Tree(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListDatasourceList ---

func TestRound9A_DataFilling_ListDatasourceList_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListDatasourceList(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListDatasourceListAll ---

func TestRound9A_DataFilling_ListDatasourceListAll_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ListDatasourceListAll(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetBuiltInTables ---

func TestRound9A_DataFilling_GetBuiltInTables_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.GetBuiltInTables(c)
	resp := parseRound9Resp(t, w)
	// nil service → panic → recovered → success with nil data
	assert.Equal(t, response.CodeSuccess, resp["code"])
}

// --- Constructor & route registration ---

func TestRound9A_DataFilling_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDataFillingHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

func TestRound9A_DataFilling_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDataFillingHandler(nil)
	RegisterDataFillingRoutes(r, h, nil, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/data-filling/tree", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// DatasetHandler tests (18 functions < 70%)
// =====================================================================

// --- Tree ---

func TestRound9A_Dataset_Tree_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Tree(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_Tree_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"nodeType":"folder"}`)
	h.Tree(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Fields ---

func TestRound9A_Dataset_Fields_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Fields(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_Fields_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"datasetId":1}`)
	h.Fields(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Preview ---

func TestRound9A_Dataset_Preview_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Preview(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_Preview_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"datasetGroupId":1}`)
	h.Preview(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- PreviewWithPermission ---

func TestRound9A_Dataset_PreviewWithPermission_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.PreviewWithPermission(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_PreviewWithPermission_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"datasetGroupId":1}`)
	h.PreviewWithPermission(c)
	resp := parseRound9Resp(t, w)
	assert.Equal(t, response.CodeUnauthorized, resp["code"])
}

// --- Save ---

func TestRound9A_Dataset_Save_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Save(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- Create ---

func TestRound9A_Dataset_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Create(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- Rename ---

func TestRound9A_Dataset_Rename_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Rename(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- Move ---

func TestRound9A_Dataset_Move_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.Move(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- Delete ---

func TestRound9A_Dataset_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Delete(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- PerDelete ---

func TestRound9A_Dataset_PerDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.PerDelete(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetDetail ---

func TestRound9A_Dataset_GetDetail_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.GetDetail(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Details ---

func TestRound9A_Dataset_Details_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.Details(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DsDetails ---

func TestRound9A_Dataset_DsDetails_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.DsDetails(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- GetSQLParams ---

func TestRound9A_Dataset_GetSQLParams_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.GetSQLParams(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- BarInfo ---

func TestRound9A_Dataset_BarInfo_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.BarInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_BarInfo_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.BarInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetDatasetTotal ---

func TestRound9A_Dataset_GetDatasetTotal_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.GetDatasetTotal(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_GetDatasetTotal_NoID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"other":1}`)
	h.GetDatasetTotal(c)
	assertCode(t, w, response.CodeSuccess)
	resp := parseRound9Resp(t, w)
	assert.Equal(t, float64(0), resp["data"])
}

// --- PreviewSQL ---

func TestRound9A_Dataset_PreviewSQL_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.PreviewSQL(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_PreviewSQL_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"sql":"SELECT 1"}`)
	h.PreviewSQL(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- EnumValueObj ---

func TestRound9A_Dataset_EnumValueObj_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.EnumValueObj(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- EnumValueDs ---

func TestRound9A_Dataset_EnumValueDs_MissingFieldID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.EnumValueDs(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- EnumValue ---

func TestRound9A_Dataset_EnumValue_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.EnumValue(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- ListByDatasetGroup ---

func TestRound9A_Dataset_ListByDatasetGroup_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "datasetId", Value: "bad"}}
	h.ListByDatasetGroup(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_ListByDatasetGroup_NilHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var h *DatasetHandler
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "datasetId", Value: "1"}}
	h.ListByDatasetGroup(c)
	assertCode(t, w, response.CodeSuccess)
}

// --- ListWithPermissions ---

func TestRound9A_Dataset_ListWithPermissions_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "datasetId", Value: "bad"}}
	h.ListWithPermissions(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveField ---

func TestRound9A_Dataset_SaveField_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveField(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_SaveField_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"field1"}`)
	h.SaveField(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetFieldFunctions ---

func TestRound9A_Dataset_GetFieldFunctions_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.GetFieldFunctions(c)
	assertCode(t, w, response.CodeSuccess)
	resp := parseRound9Resp(t, w)
	assert.Equal(t, []interface{}{}, resp["data"])
}

// --- MultFieldValuesForPermissions ---

func TestRound9A_Dataset_MultFieldValuesForPermissions_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.MultFieldValuesForPermissions(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- CopilotFields ---

func TestRound9A_Dataset_CopilotFields_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.CopilotFields(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_CopilotFields_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.CopilotFields(c)
	resp := parseRound9Resp(t, w)
	assert.Equal(t, response.CodeUnauthorized, resp["code"])
}

func TestRound9A_Dataset_CopilotFields_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	round9SetUser(c, 1)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.CopilotFields(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListFieldsByDsIds ---

func TestRound9A_Dataset_ListFieldsByDsIds_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.ListFieldsByDsIds(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_ListFieldsByDsIds_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"dsIds":[1,2]}`)
	h.ListFieldsByDsIds(c)
	assertCode(t, w, response.CodeSuccess)
}

// --- DetailWithPerm ---

func TestRound9A_Dataset_DetailWithPerm_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.DetailWithPerm(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- ExportDataset ---

func TestRound9A_Dataset_ExportDataset_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.ExportDataset(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_ExportDataset_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"dataEaseBi":false,"viewName":"test"}`)
	h.ExportDataset(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetFieldTree ---

func TestRound9A_Dataset_GetFieldTree_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.GetFieldTree(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

// --- DeleteDatasetField ---

func TestRound9A_Dataset_DeleteDatasetField_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.DeleteDatasetField(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_DeleteDatasetField_NilBoth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.DeleteDatasetField(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DeleteDatasetFieldByChart ---

func TestRound9A_Dataset_DeleteDatasetFieldByChart_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.DeleteDatasetFieldByChart(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Dataset_DeleteDatasetFieldByChart_NilBoth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDatasetHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.DeleteDatasetFieldByChart(c)
	assertCode(t, w, response.CodeInternalError)
}

// =====================================================================
// SyncHandler tests (16 functions < 70%)
// =====================================================================

// --- SourceDatasourcePager ---

func TestRound9A_Sync_SourceDatasourcePager_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "page", Value: "0"}, {Key: "limit", Value: "10"}}
	h.SourceDatasourcePager(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_SourceDatasourcePager_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "page", Value: "1"}, {Key: "limit", Value: "10"}}
	h.SourceDatasourcePager(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- TargetDatasourcePager ---

func TestRound9A_Sync_TargetDatasourcePager_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "page", Value: "0"}, {Key: "limit", Value: "10"}}
	h.TargetDatasourcePager(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_TargetDatasourcePager_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "page", Value: "1"}, {Key: "limit", Value: "10"}}
	h.TargetDatasourcePager(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- LatestUse ---

func TestRound9A_Sync_LatestUse_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "sourceType", Value: "mysql"}}
	h.LatestUse(c)
	assertCode(t, w, response.CodeSuccess)
}

func TestRound9A_Sync_LatestUse_WithUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	round9SetUsername(c, "admin")
	c.Params = gin.Params{{Key: "sourceType", Value: "mysql"}}
	h.LatestUse(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ValidateDatasource ---

func TestRound9A_Sync_ValidateDatasource_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.ValidateDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Sync_ValidateDatasource_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"type":"mysql","configuration":"{}"}`)
	h.ValidateDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ValidateDatasourceByID ---

func TestRound9A_Sync_ValidateDatasourceByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.ValidateDatasourceByID(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_ValidateDatasourceByID_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.ValidateDatasourceByID(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetSchemas ---

func TestRound9A_Sync_GetSchemas_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.GetSchemas(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SaveDatasource ---

func TestRound9A_Sync_SaveDatasource_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SaveDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Sync_SaveDatasource_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"test","type":"mysql"}`)
	h.SaveDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetDatasource ---

func TestRound9A_Sync_GetDatasource_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.GetDatasource(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_GetDatasource_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- UpdateDatasource ---

func TestRound9A_Sync_UpdateDatasource_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UpdateDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Sync_UpdateDatasource_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"test","type":"mysql"}`)
	h.UpdateDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DeleteDatasource ---

func TestRound9A_Sync_DeleteDatasource_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.DeleteDatasource(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_DeleteDatasource_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.DeleteDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- BatchDeleteDatasource ---

func TestRound9A_Sync_BatchDeleteDatasource_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.BatchDeleteDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Sync_BatchDeleteDatasource_InvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `["bad","ids"]`)
	h.BatchDeleteDatasource(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetDatasourceFields ---

func TestRound9A_Sync_GetDatasourceFields_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.GetDatasourceFields(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Sync_GetDatasourceFields_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"datasourceId":"1"}`)
	h.GetDatasourceFields(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListDatasourceByType ---

func TestRound9A_Sync_ListDatasourceByType_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "type", Value: "mysql"}}
	h.ListDatasourceByType(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ListDatasourceTables ---

func TestRound9A_Sync_ListDatasourceTables_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dsId", Value: "bad"}}
	h.ListDatasourceTables(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_ListDatasourceTables_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "dsId", Value: "1"}}
	h.ListDatasourceTables(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- TaskPager ---

func TestRound9A_Sync_TaskPager_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "current", Value: "0"}, {Key: "size", Value: "10"}}
	h.TaskPager(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_TaskPager_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "current", Value: "1"}, {Key: "size", Value: "10"}}
	h.TaskPager(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetTask ---

func TestRound9A_Sync_GetTask_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "taskId", Value: "bad"}}
	h.GetTask(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_GetTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "taskId", Value: "1"}}
	h.GetTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- AddTask ---

func TestRound9A_Sync_AddTask_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.AddTask(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Sync_AddTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"name":"task1"}`)
	h.AddTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- UpdateTask ---

func TestRound9A_Sync_UpdateTask_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UpdateTask(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Sync_UpdateTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"task1"}`)
	h.UpdateTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- RemoveTask ---

func TestRound9A_Sync_RemoveTask_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "taskId", Value: "bad"}}
	h.RemoveTask(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_RemoveTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "taskId", Value: "1"}}
	h.RemoveTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- BatchDeleteTasks ---

func TestRound9A_Sync_BatchDeleteTasks_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.BatchDeleteTasks(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ExecuteTask ---

func TestRound9A_Sync_ExecuteTask_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.ExecuteTask(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_ExecuteTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.ExecuteTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- StartTask ---

func TestRound9A_Sync_StartTask_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.StartTask(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_StartTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.StartTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- StopTask ---

func TestRound9A_Sync_StopTask_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.StopTask(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_StopTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.StopTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- TaskLogPager ---

func TestRound9A_Sync_TaskLogPager_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "current", Value: "0"}, {Key: "size", Value: "10"}}
	h.TaskLogPager(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_TaskLogPager_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "current", Value: "1"}, {Key: "size", Value: "10"}}
	h.TaskLogPager(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- TaskLogDetail ---

func TestRound9A_Sync_TaskLogDetail_InvalidLogID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "logId", Value: "bad"}, {Key: "fromLineNum", Value: "0"}}
	h.TaskLogDetail(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_TaskLogDetail_InvalidFromLineNum(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "logId", Value: "1"}, {Key: "fromLineNum", Value: "bad"}}
	h.TaskLogDetail(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_TaskLogDetail_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "logId", Value: "1"}, {Key: "fromLineNum", Value: "0"}}
	h.TaskLogDetail(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DeleteTaskLog ---

func TestRound9A_Sync_DeleteTaskLog_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "logId", Value: "bad"}}
	h.DeleteTaskLog(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_DeleteTaskLog_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "logId", Value: "1"}}
	h.DeleteTaskLog(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ClearTaskLog ---

func TestRound9A_Sync_ClearTaskLog_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.ClearTaskLog(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_Sync_ClearTaskLog_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	// ClearTaskLog uses isEOFBindError so empty body with EOF is tolerated
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{}`)
	h.ClearTaskLog(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- TerminateTask ---

func TestRound9A_Sync_TerminateTask_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "logId", Value: "bad"}}
	h.TerminateTask(c)
	resp := parseRound9Resp(t, w)
	assert.NotEqual(t, response.CodeSuccess, resp["code"])
}

func TestRound9A_Sync_TerminateTask_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "logId", Value: "1"}}
	h.TerminateTask(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ResourceCount ---

func TestRound9A_Sync_ResourceCount_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ResourceCount(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- LogChartData ---

func TestRound9A_Sync_LogChartData_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.LogChartData(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Constructor ---

func TestRound9A_Sync_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSyncHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

// =====================================================================
// UserHandler tests (15 functions < 70%)
// =====================================================================

// --- ListUsers ---

func TestRound9A_User_ListUsers_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.ListUsers(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_ListUsers_NoOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"current":1,"size":10}`)
	h.ListUsers(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_ListUsers_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"current":1,"size":10}`)
	round9SetOrg(c, 1)
	h.ListUsers(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- CreateUser ---

func TestRound9A_User_CreateUser_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.CreateUser(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_CreateUser_NoOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"account":"test","name":"test"}`)
	h.CreateUser(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_CreateUser_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"account":"test","name":"test"}`)
	round9SetOrg(c, 1)
	h.CreateUser(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- UpdateUser ---

func TestRound9A_User_UpdateUser_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.UpdateUser(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_UpdateUser_NoOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"test"}`)
	h.UpdateUser(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_UpdateUser_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"name":"test"}`)
	round9SetOrg(c, 1)
	h.UpdateUser(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DeleteUser ---

func TestRound9A_User_DeleteUser_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.DeleteUser(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_DeleteUser_NoOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.DeleteUser(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_DeleteUser_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	round9SetOrg(c, 1)
	h.DeleteUser(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetUserOptions ---

func TestRound9A_User_GetUserOptions_NoOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.GetUserOptions(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_GetUserOptions_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	round9SetOrg(c, 1)
	h.GetUserOptions(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetUserInfo ---

func TestRound9A_User_GetUserInfo_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.GetUserInfo(c)
	resp := parseRound9Resp(t, w)
	assert.Equal(t, response.CodeUnauthorized, resp["code"])
}

func TestRound9A_User_GetUserInfo_NoBootstrap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	round9SetUser(c, 1)
	h.GetUserInfo(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- PersonInfo ---

func TestRound9A_User_PersonInfo_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.PersonInfo(c)
	resp := parseRound9Resp(t, w)
	assert.Equal(t, response.CodeUnauthorized, resp["code"])
}

func TestRound9A_User_PersonInfo_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	round9SetUser(c, 1)
	h.PersonInfo(c)
	assertCode(t, w, response.CodeSuccess)
	resp := parseRound9Resp(t, w)
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "de", data["model"])
}

// --- IPInfo ---

func TestRound9A_User_IPInfo_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.IPInfo(c)
	resp := parseRound9Resp(t, w)
	assert.Equal(t, response.CodeUnauthorized, resp["code"])
}

func TestRound9A_User_IPInfo_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	round9SetUser(c, 1)
	h.IPInfo(c)
	assertCode(t, w, response.CodeSuccess)
	resp := parseRound9Resp(t, w)
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "ip")
}

// --- resolveWatermarkIdentity ---

func TestRound9A_User_ResolveWatermarkIdentity_NilLoader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	account, name := h.resolveWatermarkIdentity(0, "fallback")
	assert.Equal(t, "fallback", account)
	assert.Equal(t, "fallback", name)
}

func TestRound9A_User_ResolveWatermarkIdentity_EmptyFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	account, name := h.resolveWatermarkIdentity(0, "")
	assert.Equal(t, "admin", account)
	assert.Equal(t, "admin", name)
}

// --- SwitchOrg ---

func TestRound9A_User_SwitchOrg_NoSwitcher(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.SwitchOrg(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_SwitchOrg_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	h.SetAuthService(nil) // explicit nil
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.SwitchOrg(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_SwitchOrg_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	// Set a mock switchOrg function
	var captured bool
	h.switchOrg = func(userID int64, targetOrgID int64, requestLanguage string) (*domainauth.TokenVO, error) {
		captured = true
		return nil, fmt.Errorf("switch error")
	}
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	h.SwitchOrg(c)
	resp := parseRound9Resp(t, w)
	// userID = 0 → unauthorized
	assert.Equal(t, response.CodeUnauthorized, resp["code"])
	assert.False(t, captured)
}

// --- SwitchLanguage ---

func TestRound9A_User_SwitchLanguage_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SwitchLanguage(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_SwitchLanguage_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"lang":"zh-CN"}`)
	h.SwitchLanguage(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_SwitchLanguage_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"lang":"zh-CN"}`)
	round9SetUser(c, 1)
	h.SwitchLanguage(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- DownloadExcelTemplate ---

func TestRound9A_User_DownloadExcelTemplate_NilImportService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.DownloadExcelTemplate(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- BatchImportUsers ---

func TestRound9A_User_BatchImportUsers_NilImportService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	h.BatchImportUsers(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_BatchImportUsers_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Request.Header.Del("Content-Type")
	h.BatchImportUsers(c)
	// nil userImportService short-circuits
	assertCode(t, w, response.CodeInternalError)
}

// --- DownloadErrorRecord ---

func TestRound9A_User_DownloadErrorRecord_NilImportService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.DownloadErrorRecord(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_DownloadErrorRecord_EmptyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	c.Params = gin.Params{{Key: "key", Value: ""}}
	// nil userImportService short-circuits first
	h.DownloadErrorRecord(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ClearErrorRecord ---

func TestRound9A_User_ClearErrorRecord_NilImportService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.ClearErrorRecord(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- GetDefaultPassword ---

func TestRound9A_User_GetDefaultPassword_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodGet, "/", "")
	h.GetDefaultPassword(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- ResetPasswordCompat ---

func TestRound9A_User_ResetPasswordCompat_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "uid", Value: "1"}}
	h.ResetPasswordCompat(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_ResetPasswordCompat_InvalidUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "uid", Value: "bad"}}
	h.ResetPasswordCompat(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_ResetPasswordCompat_NoOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "")
	c.Params = gin.Params{{Key: "uid", Value: "1"}}
	// no org set → requireCurrentOrg fails
	h.ResetPasswordCompat(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- SwitchEnable ---

func TestRound9A_User_SwitchEnable_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"status":1}`)
	h.SwitchEnable(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_SwitchEnable_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", "not-json")
	h.SwitchEnable(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_SwitchEnable_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"status":1}`)
	h.SwitchEnable(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_SwitchEnable_MissingStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1}`)
	h.SwitchEnable(c)
	assertCode(t, w, response.CodeInternalError)
}

func TestRound9A_User_SwitchEnable_NoOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	w, c := newRound9Ctx(t, http.MethodPost, "/", `{"id":1,"status":1}`)
	// no org set → requireCurrentOrg fails
	h.SwitchEnable(c)
	assertCode(t, w, response.CodeInternalError)
}

// --- Constructor ---

func TestRound9A_User_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.userService)
	assert.Nil(t, h.userImportService)
}

func TestRound9A_User_SetAuthService_Nil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(nil, nil)
	h.SetAuthService(nil)
	assert.Nil(t, h.buildBootstrap)
	assert.Nil(t, h.switchOrg)
}

// --- requireCurrentOrg direct ---

func TestRound9A_User_RequireCurrentOrg_NoOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	orgID, ok := requireCurrentOrg(c)
	assert.False(t, ok)
	assert.Equal(t, int64(0), orgID)
}

func TestRound9A_User_RequireCurrentOrg_WithOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("org_id", int64(5))
	orgID, ok := requireCurrentOrg(c)
	assert.True(t, ok)
	assert.Equal(t, int64(5), orgID)
}

// --- int64Ptr direct ---

func TestRound9A_User_Int64Ptr(t *testing.T) {
	p := int64Ptr(42)
	assert.NotNil(t, p)
	assert.Equal(t, int64(42), *p)
}

// =====================================================================
// Route registration tests
// =====================================================================

func TestRound9A_Dataset_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDatasetHandler(nil, nil)
	RegisterDatasetRoutes(r.Group("/api"), h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/dataset/tree", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound9A_Sync_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSyncHandler(nil)
	RegisterSyncRoutes(r.Group("/api"), h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sync/summary/resourceCount", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRound9A_User_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewUserHandler(nil, nil)
	RegisterUserRoutes(r.Group("/api"), h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/system/user/defaultPwd", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// Silence unused imports (compile-time verification)
// =====================================================================

var _ = datafillingdomain.CreateFormRequest{}
var _ = datasource.ListRequest{}
var _ = domainsync.TaskInfo{}
var _ = middleware.ContextUserID
var _ = fmt.Sprintf
var _ = strings.ReplaceAll
