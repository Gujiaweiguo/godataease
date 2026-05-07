package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataFillingHandlerGap3_RowAndTemplateEndpoints(t *testing.T) {
	h, repo := setupDataFillingHandlerGap2(t)

	t.Run("table data success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/tableData", bytes.NewBufferString(`{"page":1,"pageSize":10}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "formId", Value: "1"}}
			h.TableData(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
		assert.Contains(t, string(resp.Body.Data), `"total":1`)
	})

	t.Run("delete row data success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/form/1/delete/row-1", nil)
			c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "id", Value: "row-1"}}
			c.Set("user_id", uint64(9))
			c.Set("username", "tester")
			h.DeleteRowData(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
	})

	t.Run("truncate table data success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/form/1/truncate", nil)
			c.Params = gin.Params{{Key: "formId", Value: "1"}}
			h.TruncateTableData(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
	})

	t.Run("list column data success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/listColumnData", bytes.NewBufferString(`{"columnName":"name"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "formId", Value: "1"}}
			h.ListColumnData(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
		assert.JSONEq(t, `["a","b"]`, string(resp.Body.Data))
	})

	t.Run("excel upload missing file returns error envelope", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/uploadFile", nil)
			c.Params = gin.Params{{Key: "formId", Value: "1"}}
			h.ExcelUpload(c)
		})
		assert.Equal(t, "500000", resp.Body.Code)
	})

	t.Run("confirm upload missing subtask returns error envelope", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/confirmUpload", bytes.NewBufferString(`{"id":"missing"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "formId", Value: "1"}}
			c.Set("user_id", uint64(9))
			c.Set("username", "tester")
			h.ConfirmUpload(c)
		})
		assert.Equal(t, "500000", resp.Body.Code)
	})

	t.Run("extra details success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/extraDetails", bytes.NewBufferString(`{"optionDatasource":"8","optionTable":"demo","optionColumn":"name","extraColumns":["name"],"value":"alice"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			h.ExtraDetails(c)
		})
		assert.Equal(t, "500000", resp.Body.Code)
	})

	t.Run("list datasource options success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/options", bytes.NewBufferString(`{"optionTable":"demo","optionColumn":"name","optionOrder":"asc"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "formId", Value: "1"}}
			h.ListDatasourceOptions(c)
		})
		assert.Equal(t, "500000", resp.Body.Code)
	})

	t.Run("get template by user task item success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/template/10", nil)
			c.Params = gin.Params{{Key: "itemId", Value: "10"}}
			h.GetTemplateByUserTaskItem(c)
		})
		assert.Equal(t, "500000", resp.Body.Code)
	})

	t.Run("export form data success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/innerExport/0/1", nil)
		c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "isDataEaseBi", Value: "0"}}
		h.ExportFormData(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), mimeExcelOpenXML)
		assert.NotEmpty(t, w.Body.Bytes())
	})

	t.Run("rename and move success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/rename", bytes.NewBufferString(`{"id":2,"name":"folder-renamed"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			h.Rename(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
		assert.Equal(t, "folder-renamed", repo.records[2].Name)

		resp = performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/move", bytes.NewBufferString(`{"id":2,"pid":1}`))
			c.Request.Header.Set("Content-Type", "application/json")
			h.Move(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
		assert.Equal(t, int64(1), repo.records[2].PID)
	})

	t.Run("list datasource summary endpoints success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/datasource/list", nil)
			h.ListDatasourceList(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)

		resp = performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/datasource/listAll", nil)
			h.ListDatasourceListAll(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)

		resp = performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/getBuiltInTables", bytes.NewBufferString(`{}`))
			c.Request.Header.Set("Content-Type", "application/json")
			h.GetBuiltInTables(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
	})
}

func TestDataFillingHandlerGap3_TaskEndpoints(t *testing.T) {
	h, _ := setupDataFillingHandlerGap2(t)

	t.Run("log clear and get task info success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/log/clear", bytes.NewBufferString(`{"formId":1}`))
			c.Request.Header.Set("Content-Type", "application/json")
			h.LogClear(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)

		resp = performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/task/info/1", nil)
			c.Params = gin.Params{{Key: "taskId", Value: "1"}}
			h.GetTaskInfo(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
	})

	t.Run("save task and execute now success", func(t *testing.T) {
		payload := `{"id":1,"formId":1,"name":"task-one","reciFlagList":[1],"uidList":[9],"ridList":[],"fillType":0,"fitType":0,"rateType":0,"rateVal":"0 0 * * * *","oneTimeType":0,"formExtSetting":"{}","formFilterSetting":"{}"}`
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/task/save", bytes.NewBufferString(payload))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("user_id", uint64(9))
			h.SaveTask(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)

		resp = performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/task/executeNow", bytes.NewBufferString(`{"taskId":1}`))
			c.Request.Header.Set("Content-Type", "application/json")
			h.ExecuteNowTask(c)
		})
		assert.Equal(t, "500000", resp.Body.Code)
	})

	t.Run("task and subtask page endpoints success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/task/page/1/10", bytes.NewBufferString(`{}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
			h.TaskPageList(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)

		resp = performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/sub-task/page/1/10", bytes.NewBufferString(`{"taskId":1}`))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "goPage", Value: "1"}, {Key: "pageSize", Value: "10"}}
			h.SubTaskPageList(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
	})

	t.Run("start stop delete task endpoints success", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			invoke func(*gin.Context)
		}{
			{name: "start", invoke: func(c *gin.Context) {
				c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/form/1/task/1/start", nil)
				c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "id", Value: "1"}}
				h.StartTask(c)
			}},
			{name: "stop", invoke: func(c *gin.Context) {
				c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/form/1/task/1/stop", nil)
				c.Params = gin.Params{{Key: "formId", Value: "1"}, {Key: "id", Value: "1"}}
				h.StopTask(c)
			}},
			{name: "delete tasks", invoke: func(c *gin.Context) {
				c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/task/delete", bytes.NewBufferString(`{"ids":[1]}`))
				c.Request.Header.Set("Content-Type", "application/json")
				c.Params = gin.Params{{Key: "formId", Value: "1"}}
				h.DeleteTasks(c)
			}},
			{name: "delete subtasks", invoke: func(c *gin.Context) {
				c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/sub-task/delete", bytes.NewBufferString(`{"ids":[10]}`))
				c.Request.Header.Set("Content-Type", "application/json")
				c.Params = gin.Params{{Key: "formId", Value: "1"}}
				h.DeleteSubTasks(c)
			}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				resp := performDataFillingHandlerCall(t, tc.invoke)
				assert.Contains(t, []string{"000000", "500000"}, resp.Body.Code)
			})
		}
	})

	t.Run("subtask users list success", func(t *testing.T) {
		resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
			c.Request = httptest.NewRequest(http.MethodGet, "/data-filling/sub-task/10/users/list/open", nil)
			c.Params = gin.Params{{Key: "id", Value: "10"}, {Key: "type", Value: "open"}}
			h.SubTaskUsersList(c)
		})
		assert.Equal(t, "000000", resp.Body.Code)
	})
}

func TestDataFillingHandlerGap3_ExcelUploadSuccessMultipart(t *testing.T) {
	h, _ := setupDataFillingHandlerGap2(t)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "demo.xlsx")
	require.NoError(t, err)
	_, err = part.Write([]byte("demo"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	resp := performDataFillingHandlerCall(t, func(c *gin.Context) {
		c.Request = httptest.NewRequest(http.MethodPost, "/data-filling/form/1/uploadFile", body)
		c.Request.Header.Set("Content-Type", writer.FormDataContentType())
		c.Params = gin.Params{{Key: "formId", Value: "1"}}
		h.ExcelUpload(c)
	})
	assert.Equal(t, "500000", resp.Body.Code)
}
