package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncHandler_InvalidInputSmoke(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSyncRoutes(r.Group("/api"), NewSyncHandler(nil))

	t.Run("task_get_invalid_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sync/task/get/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})

	t.Run("task_log_invalid_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sync/task/log/detail/abc/0", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})

	t.Run("task_pager_invalid_page", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/sync/task/pager/0/10", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})

	t.Run("datasource_batch_del_invalid_id", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/sync/datasource/batchDel", strings.NewReader("[\"bad\"]"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "500000", resp["code"])
	})
}

func TestSyncHandler_InvalidPageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSyncRoutes(r.Group("/api"), NewSyncHandler(nil))

	tests := []struct {
		name   string
		method string
		url    string
		body   string
	}{
		{name: "source_pager_zero_page", method: "POST", url: "/api/sync/datasource/source/pager/0/10", body: "{}"},
		{name: "source_pager_negative_page", method: "POST", url: "/api/sync/datasource/source/pager/-1/10", body: "{}"},
		{name: "source_pager_non_numeric_page", method: "POST", url: "/api/sync/datasource/source/pager/abc/10", body: "{}"},
		{name: "source_pager_zero_size", method: "POST", url: "/api/sync/datasource/source/pager/1/0", body: "{}"},
		{name: "source_pager_negative_size", method: "POST", url: "/api/sync/datasource/source/pager/1/-1", body: "{}"},
		{name: "source_pager_non_numeric_size", method: "POST", url: "/api/sync/datasource/source/pager/1/abc", body: "{}"},
		{name: "target_pager_zero_page", method: "POST", url: "/api/sync/datasource/target/pager/0/10", body: "{}"},
		{name: "target_pager_negative_page", method: "POST", url: "/api/sync/datasource/target/pager/-1/10", body: "{}"},
		{name: "target_pager_non_numeric_page", method: "POST", url: "/api/sync/datasource/target/pager/abc/10", body: "{}"},
		{name: "target_pager_zero_size", method: "POST", url: "/api/sync/datasource/target/pager/1/0", body: "{}"},
		{name: "target_pager_negative_size", method: "POST", url: "/api/sync/datasource/target/pager/1/-1", body: "{}"},
		{name: "target_pager_non_numeric_size", method: "POST", url: "/api/sync/datasource/target/pager/1/abc", body: "{}"},
		{name: "task_pager_zero_page", method: "POST", url: "/api/sync/task/pager/0/10", body: "{}"},
		{name: "task_pager_negative_page", method: "POST", url: "/api/sync/task/pager/-1/10", body: "{}"},
		{name: "task_pager_non_numeric_page", method: "POST", url: "/api/sync/task/pager/abc/10", body: "{}"},
		{name: "task_pager_zero_size", method: "POST", url: "/api/sync/task/pager/1/0", body: "{}"},
		{name: "task_pager_negative_size", method: "POST", url: "/api/sync/task/pager/1/-1", body: "{}"},
		{name: "task_pager_non_numeric_size", method: "POST", url: "/api/sync/task/pager/1/abc", body: "{}"},
		{name: "task_log_pager_zero_page", method: "POST", url: "/api/sync/task/log/pager/0/10", body: "{}"},
		{name: "task_log_pager_negative_page", method: "POST", url: "/api/sync/task/log/pager/-1/10", body: "{}"},
		{name: "task_log_pager_non_numeric_page", method: "POST", url: "/api/sync/task/log/pager/abc/10", body: "{}"},
		{name: "task_log_pager_zero_size", method: "POST", url: "/api/sync/task/log/pager/1/0", body: "{}"},
		{name: "task_log_pager_negative_size", method: "POST", url: "/api/sync/task/log/pager/1/-1", body: "{}"},
		{name: "task_log_pager_non_numeric_size", method: "POST", url: "/api/sync/task/log/pager/1/abc", body: "{}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSyncHandlerErrorResponse(t, r, tt.method, tt.url, tt.body)
		})
	}
}

func TestSyncHandler_InvalidIDParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSyncRoutes(r.Group("/api"), NewSyncHandler(nil))

	tests := []struct {
		name   string
		method string
		url    string
	}{
		{name: "validate_datasource_by_id_non_numeric", method: "GET", url: "/api/sync/datasource/validate/abc"},
		{name: "validate_datasource_by_id_zero", method: "GET", url: "/api/sync/datasource/validate/0"},
		{name: "validate_datasource_by_id_negative", method: "GET", url: "/api/sync/datasource/validate/-1"},
		{name: "get_datasource_non_numeric", method: "GET", url: "/api/sync/datasource/get/abc"},
		{name: "get_datasource_zero", method: "GET", url: "/api/sync/datasource/get/0"},
		{name: "get_datasource_negative", method: "GET", url: "/api/sync/datasource/get/-1"},
		{name: "delete_datasource_non_numeric", method: "POST", url: "/api/sync/datasource/delete/abc"},
		{name: "delete_datasource_zero", method: "POST", url: "/api/sync/datasource/delete/0"},
		{name: "delete_datasource_negative", method: "POST", url: "/api/sync/datasource/delete/-1"},
		{name: "list_datasource_tables_non_numeric", method: "GET", url: "/api/sync/datasource/table/list/abc"},
		{name: "list_datasource_tables_zero", method: "GET", url: "/api/sync/datasource/table/list/0"},
		{name: "list_datasource_tables_negative", method: "GET", url: "/api/sync/datasource/table/list/-1"},
		{name: "get_task_non_numeric", method: "GET", url: "/api/sync/task/get/abc"},
		{name: "get_task_zero", method: "GET", url: "/api/sync/task/get/0"},
		{name: "get_task_negative", method: "GET", url: "/api/sync/task/get/-1"},
		{name: "remove_task_non_numeric", method: "POST", url: "/api/sync/task/remove/abc"},
		{name: "remove_task_zero", method: "POST", url: "/api/sync/task/remove/0"},
		{name: "remove_task_negative", method: "POST", url: "/api/sync/task/remove/-1"},
		{name: "execute_task_non_numeric", method: "GET", url: "/api/sync/task/execute/abc"},
		{name: "execute_task_zero", method: "GET", url: "/api/sync/task/execute/0"},
		{name: "execute_task_negative", method: "GET", url: "/api/sync/task/execute/-1"},
		{name: "start_task_non_numeric", method: "GET", url: "/api/sync/task/start/abc"},
		{name: "start_task_zero", method: "GET", url: "/api/sync/task/start/0"},
		{name: "start_task_negative", method: "GET", url: "/api/sync/task/start/-1"},
		{name: "stop_task_non_numeric", method: "GET", url: "/api/sync/task/stop/abc"},
		{name: "stop_task_zero", method: "GET", url: "/api/sync/task/stop/0"},
		{name: "stop_task_negative", method: "GET", url: "/api/sync/task/stop/-1"},
		{name: "delete_task_log_non_numeric", method: "POST", url: "/api/sync/task/log/delete/abc"},
		{name: "delete_task_log_zero", method: "POST", url: "/api/sync/task/log/delete/0"},
		{name: "delete_task_log_negative", method: "POST", url: "/api/sync/task/log/delete/-1"},
		{name: "terminate_task_non_numeric", method: "POST", url: "/api/sync/task/log/terminationTask/abc"},
		{name: "terminate_task_zero", method: "POST", url: "/api/sync/task/log/terminationTask/0"},
		{name: "terminate_task_negative", method: "POST", url: "/api/sync/task/log/terminationTask/-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSyncHandlerErrorResponse(t, r, tt.method, tt.url, "")
		})
	}
}

func TestSyncHandler_TaskLogDetailInvalidFromLineNum(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSyncRoutes(r.Group("/api"), NewSyncHandler(nil))

	tests := []struct {
		name string
		url  string
	}{
		{name: "task_log_detail_non_numeric_from_line_num", url: "/api/sync/task/log/detail/1/abc"},
		{name: "task_log_detail_blank_like_from_line_num", url: "/api/sync/task/log/detail/1/%20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSyncHandlerErrorResponse(t, r, "GET", tt.url, "")
		})
	}
}

func TestSyncHandler_InvalidJSONBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSyncRoutes(r.Group("/api"), NewSyncHandler(nil))

	tests := []struct {
		name string
		url  string
		body string
	}{
		{name: "validate_datasource_empty_body", url: "/api/sync/datasource/validate", body: ""},
		{name: "validate_datasource_malformed_json", url: "/api/sync/datasource/validate", body: "{"},
		{name: "save_datasource_empty_body", url: "/api/sync/datasource/save", body: ""},
		{name: "save_datasource_malformed_json", url: "/api/sync/datasource/save", body: "{"},
		{name: "update_datasource_empty_body", url: "/api/sync/datasource/update", body: ""},
		{name: "update_datasource_malformed_json", url: "/api/sync/datasource/update", body: "{"},
		{name: "get_datasource_fields_empty_body", url: "/api/sync/datasource/fields", body: ""},
		{name: "get_datasource_fields_malformed_json", url: "/api/sync/datasource/fields", body: "{"},
		{name: "add_task_empty_body", url: "/api/sync/task/add", body: ""},
		{name: "add_task_malformed_json", url: "/api/sync/task/add", body: "{"},
		{name: "update_task_empty_body", url: "/api/sync/task/update", body: ""},
		{name: "update_task_malformed_json", url: "/api/sync/task/update", body: "{"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSyncHandlerErrorResponse(t, r, "POST", tt.url, tt.body)
		})
	}
}

func TestSyncHandler_InvalidBatchDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSyncRoutes(r.Group("/api"), NewSyncHandler(nil))

	t.Run("datasource_batch_delete_non_numeric_ids", func(t *testing.T) {
		assertSyncHandlerErrorResponse(t, r, "POST", "/api/sync/datasource/batchDel", "[\"bad\"]")
	})

	t.Run("task_batch_delete_non_numeric_ids", func(t *testing.T) {
		assertSyncHandlerErrorResponse(t, r, "POST", "/api/sync/task/batch/del", "[\"bad\"]")
	})

	t.Run("datasource_batch_delete_empty_array_reaches_route", func(t *testing.T) {
		assertSyncHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/sync/datasource/batchDel", "[]")
	})

	t.Run("task_batch_delete_empty_array_reaches_route", func(t *testing.T) {
		assertSyncHandlerRouteReachableWithRecoveredPanic(t, r, "POST", "/api/sync/task/batch/del", "[]")
	})
}

func TestSyncHandler_NilServiceRoutesReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSyncRoutes(r.Group("/api"), NewSyncHandler(nil))

	tests := []struct {
		name   string
		method string
		url    string
		body   string
	}{
		{name: "latest_use_route_registered", method: "POST", url: "/api/sync/datasource/latestUse/mysql", body: "{}"},
		{name: "get_schemas_route_registered", method: "POST", url: "/api/sync/datasource/getSchema", body: "{}"},
		{name: "list_datasource_by_type_route_registered", method: "GET", url: "/api/sync/datasource/list/mysql"},
		{name: "clear_task_log_route_registered", method: "POST", url: "/api/sync/task/log/clear", body: "{}"},
		{name: "resource_count_route_registered", method: "GET", url: "/api/sync/summary/resourceCount"},
		{name: "log_chart_data_route_registered", method: "POST", url: "/api/sync/summary/logChartData", body: "{}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSyncHandlerRouteReachableWithRecoveredPanic(t, r, tt.method, tt.url, tt.body)
		})
	}
}

func assertSyncHandlerErrorResponse(t *testing.T, r *gin.Engine, method, url, body string) {
	t.Helper()

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, url, reader)
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "500000", resp["code"])
}

func assertSyncHandlerRouteReachableWithRecoveredPanic(t *testing.T, r *gin.Engine, method, url, body string) {
	t.Helper()

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, url, reader)
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()

	didPanic := false
	func() {
		defer func() {
			if recover() != nil {
				didPanic = true
			}
		}()
		r.ServeHTTP(w, req)
	}()

	assert.NotEqual(t, 404, w.Code)
	assert.True(t, didPanic || w.Code == 200)
}
