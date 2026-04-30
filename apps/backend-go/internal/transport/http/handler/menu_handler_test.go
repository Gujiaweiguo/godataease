package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type menuHandlerTestEnv struct {
	r  *gin.Engine
	db *gorm.DB
}

type menuBridgeResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func setupMenuHandlerTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	return setupMenuHandlerTestEnv(t).r
}

func setupMenuHandlerTestEnv(t *testing.T) *menuHandlerTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&menu.CoreMenu{}))

	repo := repository.NewMenuRepository(db)
	svc := service.NewMenuService(repo)
	h := NewMenuHandler(svc)

	r := gin.New()
	api := r.Group("/api")
	RegisterMenuReadRoutes(api, h)
	RegisterMenuWriteRoutes(api, h)

	return &menuHandlerTestEnv{r: r, db: db}
}

func (e *menuHandlerTestEnv) closeDB(t *testing.T) {
	t.Helper()
	sqlDB, err := e.db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func performJSONRequest(t *testing.T, r *gin.Engine, method string, path string, body interface{}) *httptest.ResponseRecorder {
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

func decodeBridgeResp(t *testing.T, body []byte) menuBridgeResp {
	t.Helper()
	var resp menuBridgeResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func seedMenu(t *testing.T, db *gorm.DB, m *menu.CoreMenu) {
	t.Helper()
	require.NoError(t, db.Create(m).Error)
}

func TestMenuHandler_Query_Success(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	seedMenu(t, env.db, &menu.CoreMenu{ID: 1, Pid: 0, Type: 0, Name: "custom-root", Path: "/root", MenuSort: 1, Icon: "home"})
	seedMenu(t, env.db, &menu.CoreMenu{ID: 2, Pid: 1, Type: 0, Name: "custom-child", Path: "/child", MenuSort: 2})

	w := performJSONRequest(t, env.r, http.MethodGet, "/api/menu/query", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data []menu.MenuVO
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.Len(t, data, 1)
	assert.Equal(t, int64(1), data[0].ID)
	assert.Equal(t, "/root", data[0].Path)
	require.NotNil(t, data[0].Meta)
	assert.Equal(t, "custom-root", data[0].Meta.Title)
	require.Len(t, data[0].Children, 1)
	assert.Equal(t, "child", data[0].Children[0].Path)
}

func TestMenuHandler_Query_Empty(t *testing.T) {
	r := setupMenuHandlerTestRouter(t)

	w := performJSONRequest(t, r, http.MethodGet, "/api/menu/query", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.True(t, len(resp.Data) == 0 || string(resp.Data) == "null")
}

func TestMenuHandler_Query_Error(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	env.closeDB(t)

	w := performJSONRequest(t, env.r, http.MethodGet, "/api/menu/query", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "closed")
}

func TestMenuHandler_Create_Success(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	body := map[string]interface{}{
		"pid":          0,
		"type":         0,
		"name":         "menu-create",
		"component":    "views/menu/index",
		"menuSort":     3,
		"icon":         "menu",
		"path":         "/menu-create",
		"hidden":       true,
		"inLayout":     true,
		"auth":         true,
		"menuLocation": "side",
		"menuType":     "menu",
		"actionConfig": map[string]interface{}{"method": "POST"},
	}

	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/create", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var createdID int64
	require.NoError(t, json.Unmarshal(resp.Data, &createdID))
	assert.NotZero(t, createdID)

	created, err := repository.NewMenuRepository(env.db).GetByID(createdID)
	require.NoError(t, err)
	assert.Equal(t, "menu-create", created.Name)
	assert.Equal(t, "/menu-create", created.Path)
	assert.Equal(t, 3, created.MenuSort)
	assert.True(t, created.Hidden)
	assert.Equal(t, "POST", created.ActionConfig["method"])
}

func TestMenuHandler_Create_InvalidJSON(t *testing.T) {
	r := setupMenuHandlerTestRouter(t)

	w := performJSONRequest(t, r, http.MethodPost, "/api/menu/create", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestMenuHandler_Create_MissingRequiredFields(t *testing.T) {
	r := setupMenuHandlerTestRouter(t)
	body := map[string]interface{}{
		"path": "/missing-name",
	}

	w := performJSONRequest(t, r, http.MethodPost, "/api/menu/create", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestMenuHandler_Create_Error(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	env.closeDB(t)

	body := map[string]interface{}{
		"name": "menu-create-error",
		"path": "/menu-create-error",
	}
	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/create", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "closed")
}

func TestMenuHandler_Update_Success(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	seedMenu(t, env.db, &menu.CoreMenu{ID: 11, Pid: 0, Type: 0, Name: "before-update", Path: "/before-update", MenuSort: 1})

	body := map[string]interface{}{
		"id":           11,
		"pid":          0,
		"type":         0,
		"name":         "after-update",
		"component":    "views/after/index",
		"menuSort":     9,
		"icon":         "edit",
		"path":         "/after-update",
		"hidden":       true,
		"inLayout":     true,
		"auth":         true,
		"menuLocation": "header",
		"menuType":     "menu",
		"actionConfig": map[string]interface{}{"confirm": true},
	}

	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/update", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	updated, err := repository.NewMenuRepository(env.db).GetByID(11)
	require.NoError(t, err)
	assert.Equal(t, "after-update", updated.Name)
	assert.Equal(t, "/after-update", updated.Path)
	assert.Equal(t, 9, updated.MenuSort)
	assert.True(t, updated.Hidden)
	assert.Equal(t, true, updated.ActionConfig["confirm"])
}

func TestMenuHandler_Update_InvalidID(t *testing.T) {
	r := setupMenuHandlerTestRouter(t)
	body := map[string]interface{}{
		"name": "missing-id",
		"path": "/missing-id",
	}

	w := performJSONRequest(t, r, http.MethodPost, "/api/menu/update", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestMenuHandler_Update_Error(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	env.closeDB(t)

	body := map[string]interface{}{
		"id":   101,
		"name": "update-error",
		"path": "/update-error",
	}
	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/update", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "closed")
}

func TestMenuHandler_Delete_Success(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	seedMenu(t, env.db, &menu.CoreMenu{ID: 21, Pid: 0, Type: 0, Name: "delete-leaf", Path: "/delete-leaf", MenuSort: 1})

	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/delete/21", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	_, err := repository.NewMenuRepository(env.db).GetByID(21)
	assert.Error(t, err)
}

func TestMenuHandler_Delete_InvalidID(t *testing.T) {
	r := setupMenuHandlerTestRouter(t)

	w := performJSONRequest(t, r, http.MethodPost, "/api/menu/delete/not-a-number", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Equal(t, "Invalid menu ID", resp.Msg)
}

func TestMenuHandler_Delete_Error(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	seedMenu(t, env.db, &menu.CoreMenu{ID: 30, Pid: 0, Type: 0, Name: "parent", Path: "/parent", MenuSort: 1})
	seedMenu(t, env.db, &menu.CoreMenu{ID: 31, Pid: 30, Type: 0, Name: "child", Path: "/parent/child", MenuSort: 2})

	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/delete/30", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Equal(t, service.ErrMenuHasChildren.Error(), resp.Msg)
}

func TestMenuHandler_UpdateSort_Success(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	seedMenu(t, env.db, &menu.CoreMenu{ID: 41, Pid: 0, Type: 0, Name: "sort-target", Path: "/sort-target", MenuSort: 1})

	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/updateSort", map[string]interface{}{"id": 41, "sort": 88})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	updated, err := repository.NewMenuRepository(env.db).GetByID(41)
	require.NoError(t, err)
	assert.Equal(t, 88, updated.MenuSort)
}

func TestMenuHandler_UpdateSort_InvalidJSON(t *testing.T) {
	r := setupMenuHandlerTestRouter(t)

	w := performJSONRequest(t, r, http.MethodPost, "/api/menu/updateSort", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestMenuHandler_UpdateSort_Error(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	env.closeDB(t)

	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/updateSort", map[string]interface{}{"id": 41, "sort": 88})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "closed")
}

func TestMenuHandler_UpdateHidden_Success(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	seedMenu(t, env.db, &menu.CoreMenu{ID: 51, Pid: 0, Type: 0, Name: "hidden-target", Path: "/hidden-target", MenuSort: 1, Hidden: false})

	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/updateHidden", map[string]interface{}{"id": 51, "hidden": true})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	updated, err := repository.NewMenuRepository(env.db).GetByID(51)
	require.NoError(t, err)
	assert.True(t, updated.Hidden)
}

func TestMenuHandler_UpdateHidden_InvalidJSON(t *testing.T) {
	r := setupMenuHandlerTestRouter(t)

	w := performJSONRequest(t, r, http.MethodPost, "/api/menu/updateHidden", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestMenuHandler_UpdateHidden_Error(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	env.closeDB(t)

	w := performJSONRequest(t, env.r, http.MethodPost, "/api/menu/updateHidden", map[string]interface{}{"id": 51, "hidden": true})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "closed")
}

func TestMenuHandler_Detail_Success(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	seedMenu(t, env.db, &menu.CoreMenu{ID: 61, Pid: 0, Type: 0, Name: "detail-menu", Path: "/detail-menu", MenuSort: 7, Hidden: true, ActionConfig: menu.JSON{"method": "GET"}})

	w := performJSONRequest(t, env.r, http.MethodGet, "/api/menu/detail/61", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data menu.CoreMenu
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, int64(61), data.ID)
	assert.Equal(t, "detail-menu", data.Name)
	assert.True(t, data.Hidden)
	assert.Equal(t, "GET", data.ActionConfig["method"])
}

func TestMenuHandler_Detail_InvalidID(t *testing.T) {
	r := setupMenuHandlerTestRouter(t)

	w := performJSONRequest(t, r, http.MethodGet, "/api/menu/detail/not-a-number", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Equal(t, "Invalid menu ID", resp.Msg)
}

func TestMenuHandler_Detail_NotFound(t *testing.T) {
	r := setupMenuHandlerTestRouter(t)

	w := performJSONRequest(t, r, http.MethodGet, "/api/menu/detail/999", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "record not found")
}

func TestMenuHandler_Detail_Error(t *testing.T) {
	env := setupMenuHandlerTestEnv(t)
	env.closeDB(t)

	w := performJSONRequest(t, env.r, http.MethodGet, "/api/menu/detail/61", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBridgeResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "closed")
}
