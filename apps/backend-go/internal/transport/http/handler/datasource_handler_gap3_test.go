package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/domain/datasource"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatasourceHandlerGap3_ValidateAndDeleteVariants(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	config := `{"host":"127.0.0.1","port":3306,"dataBase":"demo","username":"root"}`
	status := datasource.StatusSuccess
	createBy := "tester"
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 141, PID: int64PtrForDatasourceHandler(0), Name: "mysql-ds", Type: "MySQL", Configuration: &config, Status: &status, CreateBy: &createBy})

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/validate", map[string]any{"datasourceId": 141})
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/validate", []byte(`{`))
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)

	w = performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/perDelete/141", nil)
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/delete/bad", nil)
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestDatasourceHandlerGap3_TableAndReadOnlyEndpoints(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 151, PID: int64PtrForDatasourceHandler(0), Name: "excel-ds", Type: datasource.TypeExcel})

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/tree", nil)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/tables", map[string]any{"datasourceId": 151})
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/tableField", map[string]any{"datasourceId": 151, "tableName": "Sheet1"})
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)

	w = performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/previewData", map[string]any{"datasourceId": 151, "tableName": "Sheet1", "limit": 10})
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)

	w = performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/schema", nil)
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)

	w = performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/latestUse", nil)
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	w = performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/types", nil)
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Contains(t, string(resp.Data), `"PostgreSQL"`)

	w = performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/syncRecord/bad/0/0", nil)
	resp = decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestDatasourceHandlerGap3_JSONAndUploadErrorEndpoints(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	for _, path := range []string{"/api/ds/syncApiTable", "/api/ds/syncApiDs", "/api/ds/loadRemoteFile", "/api/ds/checkApiDatasource"} {
		w := performDatasourceJSONRequest(t, env.r, http.MethodPost, path, []byte(`{`))
		resp := decodeDatasourceResp(t, w.Body.Bytes())
		assert.Equal(t, "500000", resp.Code, path)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/ds/uploadFile", bytes.NewReader(nil))
	env.r.ServeHTTP(w, req)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestDatasourceHandlerGap3_HelperFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, int64(0), getCurrentUserID(c))
	assert.Equal(t, "", getCurrentUsername(c))
	c.Set("userId", int64(77))
	c.Set("username", "tester")
	assert.Equal(t, int64(77), getCurrentUserID(c))
	assert.Equal(t, "tester", getCurrentUsername(c))
}

func TestDatasourceHandlerGap3_TypesPayloadShape(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	w := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/types", nil)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	require.Equal(t, "000000", resp.Code)
	var items []map[string]string
	require.NoError(t, json.Unmarshal(resp.Data, &items))
	require.NotEmpty(t, items)
	assert.Equal(t, "MySQL", items[0]["type"])
}
