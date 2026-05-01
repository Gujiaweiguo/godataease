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

type datasourceHandlerTestEnv struct {
	r  *gin.Engine
	db *gorm.DB
}

type datasourceHandlerBridgeResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type datasourceHandlerTreeNode struct {
	ID       string                      `json:"id"`
	Name     string                      `json:"name"`
	PID      string                      `json:"pid"`
	Type     string                      `json:"type"`
	Leaf     bool                        `json:"leaf"`
	Children []datasourceHandlerTreeNode `json:"children"`
}

func setupDatasourceHandlerTestEnv(t *testing.T) *datasourceHandlerTestEnv {
	return setupDatasourceHandlerTestEnvWithUser(t, 1001, "datasource-tester")
}

func setupDatasourceHandlerTestEnvWithUser(t *testing.T, userID int64, username string) *datasourceHandlerTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
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

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("userId", userID)
		c.Set("username", username)
		c.Next()
	})
	RegisterDatasourceRoutes(r.Group("/api"), h, nil)

	return &datasourceHandlerTestEnv{r: r, db: db}
}

func performDatasourceJSONRequest(t *testing.T, r *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
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

func decodeDatasourceResp(t *testing.T, body []byte) datasourceHandlerBridgeResp {
	t.Helper()
	var resp datasourceHandlerBridgeResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func seedDatasourceRecord(t *testing.T, db *gorm.DB, ds *datasource.CoreDatasource) {
	t.Helper()
	require.NoError(t, db.Create(ds).Error)
}

func seedDatasourceTableRecord(t *testing.T, db *gorm.DB, table *auto.CoreDatasetTable) {
	t.Helper()
	require.NoError(t, db.Create(table).Error)
}

func seedDatasourceTaskLogRecord(t *testing.T, db *gorm.DB, log *auto.CoreDatasourceTaskLog) {
	t.Helper()
	require.NoError(t, db.Create(log).Error)
}

func datasourceHandlerFindTreeNode(nodes []datasourceHandlerTreeNode, id string) *datasourceHandlerTreeNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
		if found := datasourceHandlerFindTreeNode(nodes[i].Children, id); found != nil {
			return found
		}
	}
	return nil
}

func datasourceHandlerFindTableInfo(tables []datasource.TableInfo, tableName string) *datasource.TableInfo {
	for i := range tables {
		if tables[i].TableName == tableName {
			return &tables[i]
		}
	}
	return nil
}

func int64PtrForDatasourceHandler(v int64) *int64 {
	return &v
}

func TestDatasourceHandler_List_SuccessPagination(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 1, PID: int64PtrForDatasourceHandler(0), Name: "ds-1", Type: "MySQL"})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 2, PID: int64PtrForDatasourceHandler(0), Name: "ds-2", Type: "MySQL"})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 3, PID: int64PtrForDatasourceHandler(0), Name: "ds-3", Type: "MySQL"})

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/list", map[string]any{"current": 2, "size": 1})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data datasource.ListResponse
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.Len(t, data.List, 1)
	assert.Equal(t, int64(3), data.Total)
	assert.Equal(t, 2, data.Current)
	assert.Equal(t, 1, data.Size)
	assert.Equal(t, int64(2), data.List[0].ID)
	assert.Equal(t, "ds-2", data.List[0].Name)
}

func TestDatasourceHandler_Tree_ReturnsRootChildrenStructure(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 10, PID: int64PtrForDatasourceHandler(0), Name: "folder-root", Type: datasource.TypeFolder})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 20, PID: int64PtrForDatasourceHandler(10), Name: "folder-child", Type: datasource.TypeFolder})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 30, PID: int64PtrForDatasourceHandler(20), Name: "leaf-ds", Type: "MySQL"})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 40, PID: int64PtrForDatasourceHandler(0), Name: "top-leaf", Type: "PostgreSQL"})

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/tree", map[string]any{})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data []datasourceHandlerTreeNode
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.Len(t, data, 1)
	assert.Equal(t, "0", data[0].ID)
	assert.Equal(t, "root", data[0].Name)

	folderNode := datasourceHandlerFindTreeNode(data, "10")
	require.NotNil(t, folderNode)
	assert.Equal(t, "folder-root", folderNode.Name)
	assert.False(t, folderNode.Leaf)

	childNode := datasourceHandlerFindTreeNode(data, "20")
	require.NotNil(t, childNode)
	assert.Equal(t, "folder-child", childNode.Name)
	assert.False(t, childNode.Leaf)

	leafNode := datasourceHandlerFindTreeNode(data, "30")
	require.NotNil(t, leafNode)
	assert.Equal(t, "leaf-ds", leafNode.Name)
	assert.True(t, leafNode.Leaf)

	topLeafNode := datasourceHandlerFindTreeNode(data, "40")
	require.NotNil(t, topLeafNode)
	assert.Equal(t, "top-leaf", topLeafNode.Name)
	assert.True(t, topLeafNode.Leaf)
}

func TestDatasourceHandler_Get_ReturnsSanitizedDatasourcePayload(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	config := `{"password":"secret","host":"127.0.0.1"}`
	status := datasource.StatusSuccess
	createBy := "7"
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{
		ID:            31,
		PID:           int64PtrForDatasourceHandler(0),
		Name:          "get-ds",
		Type:          "MySQL",
		Configuration: &config,
		Status:        &status,
		CreateBy:      &createBy,
	})

	w := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/31", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data map[string]any
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, "31", data["id"])
	assert.Equal(t, "0", data["pid"])
	assert.Equal(t, "get-ds", data["name"])
	assert.Equal(t, "MySQL", data["type"])
	assert.Equal(t, config, data["configuration"])
	assert.Equal(t, createBy, data["creator"])
	assert.Equal(t, status, data["status"])
}

func TestDatasourceHandler_Save_CreatingFolderNormalizesPIDAndConfiguration(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/save", map[string]any{
		"name":     "new-folder",
		"nodeType": datasource.TypeFolder,
		"pid":      -99,
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var created datasource.CoreDatasource
	require.NoError(t, json.Unmarshal(resp.Data, &created))
	assert.NotZero(t, created.ID)
	require.NotNil(t, created.PID)
	assert.Equal(t, int64(0), *created.PID)
	require.NotNil(t, created.Configuration)
	assert.Equal(t, "{}", *created.Configuration)
	assert.Equal(t, datasource.TypeFolder, created.Type)

	stored, err := repository.NewDatasourceRepository(env.db).GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, stored.PID)
	assert.Equal(t, int64(0), *stored.PID)
	require.NotNil(t, stored.Configuration)
	assert.Equal(t, "{}", *stored.Configuration)
}

func TestDatasourceHandler_Update_DuplicateNameErrorEnvelope(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 51, PID: int64PtrForDatasourceHandler(0), Name: "existing-name", Type: datasource.TypeFolder})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 52, PID: int64PtrForDatasourceHandler(0), Name: "target-name", Type: datasource.TypeFolder})

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/update", map[string]any{
		"id":   52,
		"pid":  0,
		"name": "existing-name",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Equal(t, "Failed: datasource name already exists", resp.Msg)

	stored, err := repository.NewDatasourceRepository(env.db).GetByID(52)
	require.NoError(t, err)
	assert.Equal(t, "target-name", stored.Name)
}

func TestDatasourceHandler_Move_RejectsMovingIntoDescendantFolder(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 61, PID: int64PtrForDatasourceHandler(0), Name: "root-folder", Type: datasource.TypeFolder})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 62, PID: int64PtrForDatasourceHandler(61), Name: "child-folder", Type: datasource.TypeFolder})

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/move", map[string]any{
		"id":  61,
		"pid": 62,
	})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Equal(t, "Failed: destination folder cannot be child of current datasource", resp.Msg)

	stored, err := repository.NewDatasourceRepository(env.db).GetByID(61)
	require.NoError(t, err)
	require.NotNil(t, stored.PID)
	assert.Equal(t, int64(0), *stored.PID)
}

func TestDatasourceHandler_Delete_RecursivelySoftDeletesDescendants(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 71, PID: int64PtrForDatasourceHandler(0), Name: "delete-root", Type: datasource.TypeFolder})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 72, PID: int64PtrForDatasourceHandler(71), Name: "delete-child-folder", Type: datasource.TypeFolder})
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 73, PID: int64PtrForDatasourceHandler(72), Name: "delete-leaf", Type: "MySQL"})

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/delete/71", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	type datasourceHandlerDeletionState struct {
		ID      int64 `json:"id"`
		DelFlag int   `json:"del_flag"`
	}
	var states []datasourceHandlerDeletionState
	require.NoError(t, env.db.Model(&datasource.CoreDatasource{}).
		Select("id, COALESCE(del_flag, 0) AS del_flag").
		Where("id IN ?", []int64{71, 72, 73}).
		Order("id ASC").
		Find(&states).Error)
	require.Len(t, states, 3)
	assert.Equal(t, 1, states[0].DelFlag)
	assert.Equal(t, 1, states[1].DelFlag)
	assert.Equal(t, 1, states[2].DelFlag)
}

func TestDatasourceHandler_TableStatus_MapsTaskLogStates(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	seedDatasourceRecord(t, env.db, &datasource.CoreDatasource{ID: 81, PID: int64PtrForDatasourceHandler(0), Name: "status-ds", Type: "MySQL"})

	seedDatasourceTableRecord(t, env.db, &auto.CoreDatasetTable{ID: 101, Name: "Pending Table", PhysicalTableName: "pending_table", DatasourceID: 81, DatasetGroupID: 1, Type: "db"})
	seedDatasourceTableRecord(t, env.db, &auto.CoreDatasetTable{ID: 102, Name: "Running Table", PhysicalTableName: "running_table", DatasourceID: 81, DatasetGroupID: 1, Type: "db"})
	seedDatasourceTableRecord(t, env.db, &auto.CoreDatasetTable{ID: 103, Name: "Success Table", PhysicalTableName: "success_table", DatasourceID: 81, DatasetGroupID: 1, Type: "db"})
	seedDatasourceTableRecord(t, env.db, &auto.CoreDatasetTable{ID: 104, Name: "Error Table", PhysicalTableName: "error_table", DatasourceID: 81, DatasetGroupID: 1, Type: "db"})
	seedDatasourceTableRecord(t, env.db, &auto.CoreDatasetTable{ID: 105, Name: "Cancelled Table", PhysicalTableName: "cancelled_table", DatasourceID: 81, DatasetGroupID: 1, Type: "db"})
	seedDatasourceTableRecord(t, env.db, &auto.CoreDatasetTable{ID: 106, Name: "Warning Table", PhysicalTableName: "warning_table", DatasourceID: 81, DatasetGroupID: 1, Type: "db"})

	seedDatasourceTaskLogRecord(t, env.db, &auto.CoreDatasourceTaskLog{ID: 201, DsID: 81, TaskID: 1, TaskStatus: "queued", PhysicalTableName: "pending_table", CreateTime: 110})
	seedDatasourceTaskLogRecord(t, env.db, &auto.CoreDatasourceTaskLog{ID: 202, DsID: 81, TaskID: 2, TaskStatus: "running", PhysicalTableName: "running_table", StartTime: 220, CreateTime: 210})
	seedDatasourceTaskLogRecord(t, env.db, &auto.CoreDatasourceTaskLog{ID: 203, DsID: 81, TaskID: 3, TaskStatus: "completed", PhysicalTableName: "success_table", StartTime: 300, EndTime: 330, CreateTime: 290})
	seedDatasourceTaskLogRecord(t, env.db, &auto.CoreDatasourceTaskLog{ID: 204, DsID: 81, TaskID: 4, TaskStatus: "error", PhysicalTableName: "error_table", StartTime: 400, EndTime: 440, CreateTime: 390})
	seedDatasourceTaskLogRecord(t, env.db, &auto.CoreDatasourceTaskLog{ID: 205, DsID: 81, TaskID: 5, TaskStatus: "cancelled", PhysicalTableName: "cancelled_table", StartTime: 500, EndTime: 550, CreateTime: 490})

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/tableStatus", map[string]any{"datasourceId": 81})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var tables []datasource.TableInfo
	require.NoError(t, json.Unmarshal(resp.Data, &tables))
	require.Len(t, tables, 6)

	pending := datasourceHandlerFindTableInfo(tables, "pending_table")
	require.NotNil(t, pending)
	assert.Equal(t, datasource.TableStatusPending, pending.Status)
	assert.Equal(t, int64(110), pending.LastUpdate)

	running := datasourceHandlerFindTableInfo(tables, "running_table")
	require.NotNil(t, running)
	assert.Equal(t, datasource.TableStatusUnderExecution, running.Status)
	assert.Equal(t, int64(220), running.LastUpdate)

	completed := datasourceHandlerFindTableInfo(tables, "success_table")
	require.NotNil(t, completed)
	assert.Equal(t, datasource.TableStatusCompleted, completed.Status)
	assert.Equal(t, int64(330), completed.LastUpdate)

	failed := datasourceHandlerFindTableInfo(tables, "error_table")
	require.NotNil(t, failed)
	assert.Equal(t, datasource.TableStatusError, failed.Status)
	assert.Equal(t, int64(440), failed.LastUpdate)

	cancelled := datasourceHandlerFindTableInfo(tables, "cancelled_table")
	require.NotNil(t, cancelled)
	assert.Equal(t, datasource.TableStatusCancelled, cancelled.Status)
	assert.Equal(t, int64(550), cancelled.LastUpdate)

	warning := datasourceHandlerFindTableInfo(tables, "warning_table")
	require.NotNil(t, warning)
	assert.Equal(t, datasource.TableStatusWarning, warning.Status)
	assert.Equal(t, int64(0), warning.LastUpdate)
}

func TestDatasourceHandler_ShowFinishPage_ReturnsTrueForNewUser(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)

	w := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/showFinishPage", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var show bool
	require.NoError(t, json.Unmarshal(resp.Data, &show))
	assert.True(t, show)
}

func TestDatasourceHandler_ShowFinishPage_ReturnsFalseWhenFinishPageRecordExists(t *testing.T) {
	env := setupDatasourceHandlerTestEnv(t)
	require.NoError(t, env.db.Create(&auto.CoreDsFinishPage{ID: 1001}).Error)

	w := performDatasourceJSONRequest(t, env.r, http.MethodGet, "/api/ds/showFinishPage", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	showBody := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", showBody.Code)

	var show bool
	require.NoError(t, json.Unmarshal(showBody.Data, &show))
	assert.False(t, show)
}


func TestDatasourceHandler_LatestUse_ReturnsEmptyWhenCurrentUsernameHasNoDatasources(t *testing.T) {
	env := setupDatasourceHandlerTestEnvWithUser(t, 1001, "datasource-without-records")

	w := performDatasourceJSONRequest(t, env.r, http.MethodPost, "/api/ds/latestUse", map[string]any{})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeDatasourceResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var types []string
	require.NoError(t, json.Unmarshal(resp.Data, &types))
	assert.Empty(t, types)
}
