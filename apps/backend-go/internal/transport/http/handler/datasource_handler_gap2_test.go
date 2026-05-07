package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestDatasourceHandler_ValidateByID_HidePw_And_GetSimpleDs(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	config := `{}`
	status := datasource.StatusSuccess
	createBy := "tester"
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 111, PID: int64PtrForDatasourceHandler(0), Name: "folder-ds", Type: datasource.TypeFolder, Configuration: &config, Status: &status, CreateBy: &createBy})

	validateW := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/validate/111", nil)
	validateResp := decodeDatasourceResp(t, validateW.Body.Bytes())
	assert.Equal(t, "000000", validateResp.Code)
	assert.Contains(t, string(validateResp.Data), `"Success"`)

	hideW := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/hidePw/111", nil)
	hideResp := decodeDatasourceResp(t, hideW.Body.Bytes())
	assert.Equal(t, "000000", hideResp.Code)
	assert.Contains(t, string(hideResp.Data), `"folder-ds"`)

	simpleW := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/simple/111", nil)
	simpleResp := decodeDatasourceResp(t, simpleW.Body.Bytes())
	assert.Equal(t, "000000", simpleResp.Code)
	assert.JSONEq(t, `{"id":"111","name":"folder-ds","type":"folder"}`, string(simpleResp.Data))
}

func TestDatasourceHandler_CheckRepeat_Rename_And_CreateFolder(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	config := `{"host":"127.0.0.1","port":3306,"dataBase":"demo","username":"root"}`
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 121, PID: int64PtrForDatasourceHandler(0), Name: "existing", Type: "MySQL", Configuration: &config})

	checkW := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/checkRepeat", map[string]any{
		"type":          "MySQL",
		"configuration": map[string]any{"host": "127.0.0.1", "port": 3306, "dataBase": "demo", "username": "root"},
	})
	checkResp := decodeDatasourceResp(t, checkW.Body.Bytes())
	assert.Equal(t, "000000", checkResp.Code)
	assert.Equal(t, "true", string(checkResp.Data))

	renameW := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/reName", map[string]any{"id": 121, "name": "renamed-ds"})
	renameResp := decodeDatasourceResp(t, renameW.Body.Bytes())
	assert.Equal(t, "000000", renameResp.Code)
	stored, err := repository.NewDatasourceRepository(env.db).GetByID(121)
	require.NoError(t, err)
	assert.Equal(t, "renamed-ds", stored.Name)

	createFolderW := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/createFolder", map[string]any{"name": "new-folder", "pid": 0})
	createFolderResp := decodeDatasourceResp(t, createFolderW.Body.Bytes())
	assert.Equal(t, "000000", createFolderResp.Code)
	var created datasource.CoreDatasource
	require.NoError(t, json.Unmarshal(createFolderResp.Data, &created))
	assert.Equal(t, datasource.TypeFolder, created.Type)
	assert.Equal(t, "new-folder", created.Name)
}

func TestDatasourceHandler_ShowFinishPage_SetShowFinishPageError_Types_And_ListSyncRecord(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 131, PID: int64PtrForDatasourceHandler(0), Name: "sync-ds", Type: "MySQL"})
	seedDatasourceTaskLogRecord(t, env.db, &auto.CoreDatasourceTaskLog{ID: 301, DsID: 131, TaskID: 9001, TaskStatus: "completed", PhysicalTableName: "orders", CreateTime: 100, EndTime: 200})

	showFinishW := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/showFinishPage", nil)
	showFinishResp := decodeDatasourceResp(t, showFinishW.Body.Bytes())
	assert.Equal(t, "000000", showFinishResp.Code)
	assert.Equal(t, "true", string(showFinishResp.Data))

	setFinishW := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/showFinishPage", nil)
	setFinishResp := decodeDatasourceResp(t, setFinishW.Body.Bytes())
	assert.Equal(t, "500000", setFinishResp.Code)

	typesW := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/types", nil)
	typesResp := decodeDatasourceResp(t, typesW.Body.Bytes())
	assert.Equal(t, "000000", typesResp.Code)
	assert.Contains(t, string(typesResp.Data), `"MySQL"`)
	assert.Contains(t, string(typesResp.Data), `"Excel"`)

	syncRecordW := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/syncRecord/131/0/0", nil)
	syncRecordResp := decodeDatasourceResp(t, syncRecordW.Body.Bytes())
	assert.Equal(t, "000000", syncRecordResp.Code)
	assert.Contains(t, string(syncRecordResp.Data), `"datasourceId":131`)
	assert.Contains(t, string(syncRecordResp.Data), `"current":1`)
	assert.Contains(t, string(syncRecordResp.Data), `"size":10`)
	assert.Contains(t, string(syncRecordResp.Data), `"orders"`)
}
