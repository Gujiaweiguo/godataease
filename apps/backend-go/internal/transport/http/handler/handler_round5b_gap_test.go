package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/share"
	"dataease/backend/internal/domain/static"
	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type round5bCoreShare struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Creator       int64      `gorm:"column:creator;index"`
	ResourceID    int64      `gorm:"column:resource_id;index"`
	ResourceType  string     `gorm:"column:resource_type;size:50"`
	Time          *time.Time `gorm:"column:time"`
	Exp           int64      `gorm:"column:exp"`
	UUID          string     `gorm:"column:uuid;size:64;uniqueIndex"`
	Pwd           string     `gorm:"column:pwd;size:255"`
	AutoPwd       bool       `gorm:"column:auto_pwd;default:true"`
	TicketRequire bool       `gorm:"column:ticket_require;default:false"`
}

func (round5bCoreShare) TableName() string { return "core_share" }

type round5bCoreShareTicket struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UUID       string     `gorm:"column:uuid;size:64;index"`
	Ticket     string     `gorm:"column:ticket;size:64;uniqueIndex"`
	Exp        int64      `gorm:"column:exp"`
	Args       string     `gorm:"column:args;type:text"`
	AccessTime *time.Time `gorm:"column:access_time"`
}

func (round5bCoreShareTicket) TableName() string { return "core_share_ticket" }

// ---------------------------------------------------------------------------
// Shared helpers for round5b tests
// ---------------------------------------------------------------------------

type round5bBridgeResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func round5bDecode(t *testing.T, body []byte) round5bBridgeResp {
	t.Helper()
	var resp round5bBridgeResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func round5bOpenDB(t *testing.T, name string, models ...interface{}) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	if len(models) > 0 {
		require.NoError(t, db.AutoMigrate(models...))
	}
	return db
}

func writeTestFile(dir, name string, content []byte) error {
	return os.WriteFile(filepath.Join(dir, name), content, 0o644)
}

func round5bReq(t *testing.T, r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
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

// ===================================================================
// SHARE HANDLER TESTS
// ===================================================================

func setupRound5bShareEnv(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db := round5bOpenDB(t, dbName, &round5bCoreShare{}, &round5bCoreShareTicket{})
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	h := NewShareHandler(svc)
	r := gin.New()
	RegisterShareRoutes(r.Group("/api"), h)
	return r, db
}

func TestRound5B_Share_NewShareHandler_NilService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewShareHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

func TestRound5B_Share_Create_Success(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	body := map[string]interface{}{
		"resourceId":   1,
		"resourceType": "panel",
		"exp":          0,
		"autoPwd":      true,
	}
	w := round5bReq(t, r, http.MethodPost, "/api/share/create", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Share_Create_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/share/create", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound5B_Share_Create_MissingRequired(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/share/create", map[string]interface{}{})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound5B_Share_Validate_Success(t *testing.T) {
	r, db := setupRound5bShareEnv(t)
	// Seed a share record via repo
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 10, ResourceType: "panel"}, 1)
	require.NoError(t, err)

	w := round5bReq(t, r, http.MethodPost, "/api/share/validate", map[string]interface{}{
		"uuid": created.UUID,
	})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Share_Validate_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/share/validate", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound5B_Share_Revoke_InvalidID(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodDelete, "/api/share/revoke/notanumber", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound5B_Share_Revoke_NotFound(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodDelete, "/api/share/revoke/99999", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, false, data["success"])
}

func TestRound5B_Share_Revoke_Success(t *testing.T) {
	r, db := setupRound5bShareEnv(t)
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 20, ResourceType: "panel"}, 1)
	require.NoError(t, err)

	w := round5bReq(t, r, http.MethodDelete, fmt.Sprintf("/api/share/revoke/%d", created.ID), nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Share_Status_InvalidResourceID(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/share/status/notanumber", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound5B_Share_Status_NotShared(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/share/status/99999", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	// data should be false (not shared)
	assert.Equal(t, "false", strings.TrimSpace(string(resp.Data)))
}

func TestRound5B_Share_Status_Shared(t *testing.T) {
	r, db := setupRound5bShareEnv(t)
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 30, ResourceType: "panel"}, 1)
	require.NoError(t, err)

	w := round5bReq(t, r, http.MethodGet, "/api/share/status/30", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, "true", strings.TrimSpace(string(resp.Data)))
}

func TestRound5B_Share_Detail_InvalidResourceID(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/share/detail/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound5B_Share_Detail_NotFound(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/share/detail/99999", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, "null", strings.TrimSpace(string(resp.Data)))
}

func TestRound5B_Share_Detail_Success(t *testing.T) {
	r, db := setupRound5bShareEnv(t)
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 40, ResourceType: "screen"}, 1)
	require.NoError(t, err)

	w := round5bReq(t, r, http.MethodGet, "/api/share/detail/40", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var detail share.ShareDetailResponse
	require.NoError(t, json.Unmarshal(resp.Data, &detail))
	assert.NotEmpty(t, detail.UUID)
}

func TestRound5B_Share_Switcher_InvalidResourceID(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/share/switcher/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound5B_Share_Switcher_Success(t *testing.T) {
	r, db := setupRound5bShareEnv(t)
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 50, ResourceType: "panel"}, 1)
	require.NoError(t, err)

	// Switch toggles share status
	w := round5bReq(t, r, http.MethodPost, "/api/share/switcher/50", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func setupRound5bShareEnvWithUser(t *testing.T, userID int64) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db := round5bOpenDB(t, dbName, &round5bCoreShare{}, &round5bCoreShareTicket{})
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	h := NewShareHandler(svc)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserID, userID)
		c.Next()
	})
	RegisterShareRoutes(r.Group("/api"), h)
	return r, db
}

func TestRound5B_Share_EditUUID_Success(t *testing.T) {
	r, db := setupRound5bShareEnvWithUser(t, 1)
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 60, ResourceType: "panel"}, 1)
	require.NoError(t, err)

	body := map[string]interface{}{
		"resourceId": 60,
		"uuid":       "newuuid123456",
	}
	w := round5bReq(t, r, http.MethodPost, "/api/share/editUuid", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Share_EditExp_Success(t *testing.T) {
	r, db := setupRound5bShareEnvWithUser(t, 1)
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 70, ResourceType: "panel"}, 1)
	require.NoError(t, err)

	body := map[string]interface{}{
		"resourceId": 70,
		"exp":        0,
	}
	w := round5bReq(t, r, http.MethodPost, "/api/share/editExp", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Share_EditPwd_Success(t *testing.T) {
	r, db := setupRound5bShareEnvWithUser(t, 1)
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 80, ResourceType: "panel"}, 1)
	require.NoError(t, err)

	body := map[string]interface{}{
		"resourceId": 80,
		"pwd":        "",
		"autoPwd":    true,
	}
	w := round5bReq(t, r, http.MethodPost, "/api/share/editPwd", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Share_CreateTicket_Success(t *testing.T) {
	r, db := setupRound5bShareEnv(t)
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 90, ResourceType: "panel"}, 1)
	require.NoError(t, err)

	body := map[string]interface{}{
		"ticket": "ticket001",
		"uuid":   created.UUID,
	}
	w := round5bReq(t, r, http.MethodPost, "/api/share/ticket/create", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Share_CreateTicket_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/share/ticket/create", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound5B_Share_ValidateTicket_Success(t *testing.T) {
	r, db := setupRound5bShareEnv(t)
	repo := repository.NewShareRepository(db)
	svc := service.NewShareService(repo)
	created, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 100, ResourceType: "panel"}, 1)
	require.NoError(t, err)

	ticket, err := svc.CreateTicket(&share.TicketCreateRequest{Ticket: "ticket002", UUID: created.UUID})
	require.NoError(t, err)

	body := map[string]interface{}{
		"ticket": ticket.Ticket,
		"uuid":   created.UUID,
	}
	w := round5bReq(t, r, http.MethodPost, "/api/share/ticket/validate", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Share_ValidateTicket_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bShareEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/share/ticket/validate", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "10001", resp.Code)
}

func TestRound5B_Share_RegisterShareRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterShareRoutes(r.Group("/api"), NewShareHandler(nil))
	routes := r.Routes()
	paths := make(map[string]bool)
	for _, rt := range routes {
		paths[rt.Path] = true
	}
	assert.True(t, paths["/api/share/create"])
	assert.True(t, paths["/api/share/validate"])
	assert.True(t, paths["/api/share/revoke/:id"])
	assert.True(t, paths["/api/share/status/:resourceId"])
	assert.True(t, paths["/api/share/detail/:resourceId"])
	assert.True(t, paths["/api/share/switcher/:resourceId"])
	assert.True(t, paths["/api/share/editUuid"])
	assert.True(t, paths["/api/share/editExp"])
	assert.True(t, paths["/api/share/editPwd"])
	assert.True(t, paths["/api/share/ticket/create"])
	assert.True(t, paths["/api/share/ticket/validate"])
}

func TestRound5B_Share_ShareUserID_NoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	uid := shareUserID(c)
	assert.Equal(t, int64(0), uid)
}

func TestRound5B_Share_ShareUserID_Int64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserID, int64(42))
	uid := shareUserID(c)
	assert.Equal(t, int64(42), uid)
}

func TestRound5B_Share_ShareUserID_Uint64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserID, uint64(99))
	uid := shareUserID(c)
	assert.Equal(t, int64(99), uid)
}

func TestRound5B_Share_ShareUserID_UserIDKey_Int64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", int64(77))
	uid := shareUserID(c)
	assert.Equal(t, int64(77), uid)
}

func TestRound5B_Share_ShareUserID_UserIDKey_Uint64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint64(88))
	uid := shareUserID(c)
	assert.Equal(t, int64(88), uid)
}

// ===================================================================
// MENU HANDLER TESTS
// ===================================================================

func setupRound5bMenuEnv(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db := round5bOpenDB(t, dbName, &menu.CoreMenu{})
	repo := repository.NewMenuRepository(db)
	svc := service.NewMenuService(repo)
	h := NewMenuHandler(svc)
	r := gin.New()
	api := r.Group("/api")
	RegisterMenuReadRoutes(api, h)
	RegisterMenuWriteRoutes(api, h)
	return r, db
}

func TestRound5B_Menu_NewMenuHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMenuHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

func TestRound5B_Menu_Query_Empty(t *testing.T) {
	r, _ := setupRound5bMenuEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/menu/query", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Menu_ApplyMenuTitles_NilMeta(t *testing.T) {
	menus := []*menu.MenuVO{
		{Name: "test", Meta: nil},
	}
	applyMenuTitles(menus, "zh-CN")
	assert.Nil(t, menus[0].Meta)
}

func TestRound5B_Menu_ApplyMenuTitles_WithTitle(t *testing.T) {
	menus := []*menu.MenuVO{
		{Name: "test-menu", Meta: &menu.MenuMeta{Title: "custom-title"}},
	}
	applyMenuTitles(menus, "zh-CN")
	assert.Equal(t, "custom-title", menus[0].Meta.Title)
}

func TestRound5B_Menu_ApplyMenuTitles_NilMenu(t *testing.T) {
	menus := []*menu.MenuVO{nil}
	applyMenuTitles(menus, "en")
	// no panic
}

func TestRound5B_Menu_ApplyMenuTitles_Children(t *testing.T) {
	menus := []*menu.MenuVO{
		{
			Name: "parent",
			Meta: &menu.MenuMeta{Title: "parent"},
			Children: []*menu.MenuVO{
				{Name: "child", Meta: &menu.MenuMeta{Title: "child-title"}},
			},
		},
	}
	applyMenuTitles(menus, "zh-CN")
	assert.Equal(t, "parent", menus[0].Meta.Title)
	assert.Equal(t, "child-title", menus[0].Children[0].Meta.Title)
}

func TestRound5B_Menu_Create_Success(t *testing.T) {
	r, db := setupRound5bMenuEnv(t)
	body := map[string]interface{}{
		"name": "round5b-menu",
		"path": "/round5b-menu",
		"type": 0,
	}
	w := round5bReq(t, r, http.MethodPost, "/api/menu/create", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var createdID int64
	require.NoError(t, json.Unmarshal(resp.Data, &createdID))
	assert.NotZero(t, createdID)

	found, err := repository.NewMenuRepository(db).GetByID(createdID)
	require.NoError(t, err)
	assert.Equal(t, "round5b-menu", found.Name)
}

func TestRound5B_Menu_Update_Success(t *testing.T) {
	r, db := setupRound5bMenuEnv(t)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 101, Pid: 0, Name: "old-name", Path: "/old", MenuSort: 1}).Error)

	body := map[string]interface{}{
		"id":   101,
		"name": "new-name",
		"path": "/new",
	}
	w := round5bReq(t, r, http.MethodPost, "/api/menu/update", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Menu_Delete_Success(t *testing.T) {
	r, db := setupRound5bMenuEnv(t)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 201, Pid: 0, Name: "del-me", Path: "/del", MenuSort: 1}).Error)

	w := round5bReq(t, r, http.MethodPost, "/api/menu/delete/201", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Menu_Delete_InvalidID(t *testing.T) {
	r, _ := setupRound5bMenuEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/menu/delete/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_Menu_UpdateSort_Success(t *testing.T) {
	r, db := setupRound5bMenuEnv(t)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 301, Pid: 0, Name: "sort-me", Path: "/sort", MenuSort: 1}).Error)

	w := round5bReq(t, r, http.MethodPost, "/api/menu/updateSort", map[string]interface{}{"id": 301, "sort": 42})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	found, err := repository.NewMenuRepository(db).GetByID(301)
	require.NoError(t, err)
	assert.Equal(t, 42, found.MenuSort)
}

func TestRound5B_Menu_UpdateHidden_Success(t *testing.T) {
	r, db := setupRound5bMenuEnv(t)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 401, Pid: 0, Name: "hide-me", Path: "/hide", MenuSort: 1, Hidden: false}).Error)

	w := round5bReq(t, r, http.MethodPost, "/api/menu/updateHidden", map[string]interface{}{"id": 401, "hidden": true})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	found, err := repository.NewMenuRepository(db).GetByID(401)
	require.NoError(t, err)
	assert.True(t, found.Hidden)
}

func TestRound5B_Menu_Detail_Success(t *testing.T) {
	r, db := setupRound5bMenuEnv(t)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 501, Pid: 0, Name: "detail-me", Path: "/detail", MenuSort: 7}).Error)

	w := round5bReq(t, r, http.MethodGet, "/api/menu/detail/501", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var m menu.CoreMenu
	require.NoError(t, json.Unmarshal(resp.Data, &m))
	assert.Equal(t, int64(501), m.ID)
	assert.Equal(t, "detail-me", m.Name)
}

func TestRound5B_Menu_Detail_InvalidID(t *testing.T) {
	r, _ := setupRound5bMenuEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/menu/detail/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_Menu_RegisterMenuRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMenuHandler(nil)
	api := r.Group("/api")
	RegisterMenuReadRoutes(api, h)
	RegisterMenuWriteRoutes(api, h)
	routes := r.Routes()
	paths := make(map[string]bool)
	for _, rt := range routes {
		paths[rt.Path] = true
	}
	assert.True(t, paths["/api/menu/query"])
	assert.True(t, paths["/api/menu/detail/:id"])
	assert.True(t, paths["/api/menu/create"])
	assert.True(t, paths["/api/menu/update"])
	assert.True(t, paths["/api/menu/delete/:id"])
	assert.True(t, paths["/api/menu/updateSort"])
	assert.True(t, paths["/api/menu/updateHidden"])
}

// ===================================================================
// STATIC HANDLER TESTS
// ===================================================================

func setupRound5bStaticEnv(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db := round5bOpenDB(t, dbName, &static.StaticResource{}, &static.Store{}, &static.Typeface{})
	staticRepo := repository.NewStaticRepository(db)
	storeRepo := repository.NewStoreRepository(db)
	typefaceRepo := repository.NewTypefaceRepository(db)
	svc := service.NewStaticService(staticRepo, storeRepo, typefaceRepo)
	h := NewStaticHandler(svc)
	r := gin.New()
	RegisterStaticRoutes(r.Group("/api"), h)
	return r, db
}

func TestRound5B_Static_NewStaticHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.service)
}

func TestRound5B_Static_ListResources_Success(t *testing.T) {
	r, _ := setupRound5bStaticEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/staticResource/list", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Static_GetResource_NotFound(t *testing.T) {
	r, _ := setupRound5bStaticEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/staticResource/nonexistent-id", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_Static_ListStores_Success(t *testing.T) {
	r, _ := setupRound5bStaticEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/store/list", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Static_ListTypefaces_Success(t *testing.T) {
	_, db := setupRound5bStaticEnv(t)
	repo := repository.NewTypefaceRepository(db)
	svc := service.NewStaticService(nil, nil, repo)
	h := NewStaticHandler(svc)
	r := gin.New()
	r.GET("/typeface/list", h.ListTypefaces)
	w := round5bReq(t, r, http.MethodGet, "/typeface/list", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Static_ListFont_Success(t *testing.T) {
	_, db := setupRound5bStaticEnv(t)
	repo := repository.NewTypefaceRepository(db)
	svc := service.NewStaticService(nil, nil, repo)
	h := NewStaticHandler(svc)
	r := gin.New()
	r.GET("/typeface/listFont", h.ListFont)
	w := round5bReq(t, r, http.MethodGet, "/typeface/listFont", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Static_DefaultFont_Empty(t *testing.T) {
	_, db := setupRound5bStaticEnv(t)
	repo := repository.NewTypefaceRepository(db)
	svc := service.NewStaticService(nil, nil, repo)
	h := NewStaticHandler(svc)
	r := gin.New()
	r.GET("/typeface/defaultFont", h.DefaultFont)
	w := round5bReq(t, r, http.MethodGet, "/typeface/defaultFont", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, "{}", strings.TrimSpace(string(resp.Data)))
}

func TestRound5B_Static_XpackModel_ReturnsFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	r := gin.New()
	r.GET("/xpackModel", h.XpackModel)
	w := round5bReq(t, r, http.MethodGet, "/xpackModel", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, "false", strings.TrimSpace(string(resp.Data)))
}

func TestRound5B_Static_Upload_MissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	r := gin.New()
	r.POST("/staticResource/upload/:fileId", h.Upload)

	req := httptest.NewRequest(http.MethodPost, "/staticResource/upload/1234567890", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_Static_Upload_InvalidExt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	r := gin.New()
	r.POST("/staticResource/upload/:fileId", h.Upload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "malicious.exe")
	require.NoError(t, err)
	_, err = part.Write([]byte("exe-data"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/staticResource/upload/1234567890", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "not allowed")
}

func TestRound5B_Static_Upload_EmptyFileID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	r := gin.New()
	r.POST("/staticResource/upload/:fileId", h.Upload)

	w := round5bReq(t, r, http.MethodPost, "/staticResource/upload/", nil)
	// With no fileId param, Gin won't match the route - 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRound5B_Static_FindResourceAsBase64_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	r := gin.New()
	r.POST("/findResourceAsBase64", h.FindResourceAsBase64)

	w := round5bReq(t, r, http.MethodPost, "/findResourceAsBase64", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_Static_FindResourceAsBase64_EmptyPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStaticHandler(nil)
	r := gin.New()
	r.POST("/findResourceAsBase64", h.FindResourceAsBase64)

	body := `{"resourcePathList":[]}`
	req := httptest.NewRequest(http.MethodPost, "/findResourceAsBase64", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_Static_FindResourceAsBase64_ExistingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	staticDir := t.TempDir()
	t.Setenv("STATIC_RESOURCE_DIR", staticDir)

	require.NoError(t, writeTestFile(staticDir, "test.png", []byte("png-data")))

	h := NewStaticHandler(nil)
	r := gin.New()
	r.POST("/findResourceAsBase64", h.FindResourceAsBase64)

	body := `{"resourcePathList":["test.png"]}`
	req := httptest.NewRequest(http.MethodPost, "/findResourceAsBase64", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	var data map[string]string
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.NotEmpty(t, data["test.png"])
}

func TestRound5B_Static_RegisterStaticRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterStaticRoutes(r.Group("/api"), NewStaticHandler(nil))
	routes := r.Routes()
	paths := make(map[string]bool)
	for _, rt := range routes {
		paths[rt.Path] = true
	}
	assert.True(t, paths["/api/staticResource/list"])
	assert.True(t, paths["/api/staticResource/:id"])
	assert.True(t, paths["/api/staticResource/upload/:fileId"])
	assert.True(t, paths["/api/staticResource/findResourceAsBase64"])
	assert.True(t, paths["/api/store/list"])
	assert.True(t, paths["/api/xpackModel"])
}

// ===================================================================
// SYSTEM VARIABLE HANDLER TESTS
// ===================================================================

func setupRound5bSysVarEnv(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db := round5bOpenDB(t, dbName, &system.SysVariable{}, &system.SysVariableValue{})
	repo := repository.NewSystemVariableRepository(db)
	svc := service.NewSystemVariableService(repo)
	h := NewSystemVariableHandler(svc)
	r := gin.New()
	RegisterSystemVariableRoutes(r.Group("/api"), h)
	return r, db
}

func TestRound5B_SystemVar_NewHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSystemVariableHandler(nil)
	assert.NotNil(t, h)
}

func TestRound5B_SystemVar_Create_Success(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	body := map[string]interface{}{
		"type": "text",
		"name": "test-var",
	}
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/create", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_Create_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/create", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_Edit_Success(t *testing.T) {
	r, db := setupRound5bSysVarEnv(t)
	repo := repository.NewSystemVariableRepository(db)
	svc := service.NewSystemVariableService(repo)
	created, err := svc.Create(&system.SysVariable{Type: "text", Name: "edit-me"})
	require.NoError(t, err)

	body := map[string]interface{}{
		"id":   created.ID,
		"type": "text",
		"name": "edited-var",
	}
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/edit", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_Edit_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/edit", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_Detail_Success(t *testing.T) {
	r, db := setupRound5bSysVarEnv(t)
	repo := repository.NewSystemVariableRepository(db)
	svc := service.NewSystemVariableService(repo)
	created, err := svc.Create(&system.SysVariable{Type: "text", Name: "detail-me"})
	require.NoError(t, err)

	w := round5bReq(t, r, http.MethodGet, fmt.Sprintf("/api/sysVariable/detail/%d", created.ID), nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_Detail_InvalidID(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/sysVariable/detail/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_Delete_Success(t *testing.T) {
	r, db := setupRound5bSysVarEnv(t)
	repo := repository.NewSystemVariableRepository(db)
	svc := service.NewSystemVariableService(repo)
	created, err := svc.Create(&system.SysVariable{Type: "text", Name: "delete-me"})
	require.NoError(t, err)

	w := round5bReq(t, r, http.MethodGet, fmt.Sprintf("/api/sysVariable/delete/%d", created.ID), nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_Delete_InvalidID(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/sysVariable/delete/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_Query_Success(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/query", map[string]interface{}{})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_Query_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/query", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_CreateValue_Success(t *testing.T) {
	r, db := setupRound5bSysVarEnv(t)
	repo := repository.NewSystemVariableRepository(db)
	svc := service.NewSystemVariableService(repo)
	created, err := svc.Create(&system.SysVariable{Type: "text", Name: "val-parent"})
	require.NoError(t, err)

	body := map[string]interface{}{
		"sysVariableId": created.ID,
		"value":         "val1",
		"valueDesc":     "Value 1",
	}
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/create", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_CreateValue_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/create", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_EditValue_Success(t *testing.T) {
	r, db := setupRound5bSysVarEnv(t)
	repo := repository.NewSystemVariableRepository(db)
	svc := service.NewSystemVariableService(repo)
	created, err := svc.Create(&system.SysVariable{Type: "text", Name: "edit-val-parent"})
	require.NoError(t, err)
	val, err := svc.CreateValue(&system.SysVariableValue{SysVariableID: created.ID, Value: "v1"})
	require.NoError(t, err)

	body := map[string]interface{}{
		"id":            val.ID,
		"sysVariableId": created.ID,
		"value":         "v2",
		"valueDesc":     "Updated",
	}
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/edit", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_EditValue_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/edit", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_DeleteValue_InvalidID(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/sysVariable/value/delete/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_DeleteValue_Success(t *testing.T) {
	r, db := setupRound5bSysVarEnv(t)
	repo := repository.NewSystemVariableRepository(db)
	svc := service.NewSystemVariableService(repo)
	created, err := svc.Create(&system.SysVariable{Type: "text", Name: "del-val-parent"})
	require.NoError(t, err)
	val, err := svc.CreateValue(&system.SysVariableValue{SysVariableID: created.ID, Value: "to-delete"})
	require.NoError(t, err)

	w := round5bReq(t, r, http.MethodGet, fmt.Sprintf("/api/sysVariable/value/delete/%d", val.ID), nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_SelectedValues_InvalidID(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodGet, "/api/sysVariable/value/selected/bad", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_SelectedValues_Success(t *testing.T) {
	r, db := setupRound5bSysVarEnv(t)
	repo := repository.NewSystemVariableRepository(db)
	svc := service.NewSystemVariableService(repo)
	created, err := svc.Create(&system.SysVariable{Type: "text", Name: "sel-val-parent"})
	require.NoError(t, err)

	w := round5bReq(t, r, http.MethodGet, fmt.Sprintf("/api/sysVariable/value/selected/%d", created.ID), nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_SelectedValuePage_InvalidPage(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/selected/0/10", map[string]interface{}{})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_SelectedValuePage_InvalidSize(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/selected/1/0", map[string]interface{}{})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_SelectedValuePage_Success(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/selected/1/10", map[string]interface{}{})
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_SelectedValuePage_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/selected/1/10", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_BatchDeleteValues_Success(t *testing.T) {
	r, db := setupRound5bSysVarEnv(t)
	repo := repository.NewSystemVariableRepository(db)
	svc := service.NewSystemVariableService(repo)
	created, err := svc.Create(&system.SysVariable{Type: "text", Name: "batch-del-parent"})
	require.NoError(t, err)
	val1, err := svc.CreateValue(&system.SysVariableValue{SysVariableID: created.ID, Value: "bv1"})
	require.NoError(t, err)
	val2, err := svc.CreateValue(&system.SysVariableValue{SysVariableID: created.ID, Value: "bv2"})
	require.NoError(t, err)

	body := []int64{val1.ID, val2.ID}
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/batchDel", body)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
}

func TestRound5B_SystemVar_BatchDeleteValues_InvalidJSON(t *testing.T) {
	r, _ := setupRound5bSysVarEnv(t)
	w := round5bReq(t, r, http.MethodPost, "/api/sysVariable/value/batchDel", []byte("{bad"))
	assert.Equal(t, http.StatusOK, w.Code)
	resp := round5bDecode(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
}

func TestRound5B_SystemVar_RegisterSystemVariableRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterSystemVariableRoutes(r.Group("/api"), NewSystemVariableHandler(nil))
	routes := r.Routes()
	paths := make(map[string]bool)
	for _, rt := range routes {
		paths[rt.Path] = true
	}
	assert.True(t, paths["/api/sysVariable/create"])
	assert.True(t, paths["/api/sysVariable/edit"])
	assert.True(t, paths["/api/sysVariable/detail/:id"])
	assert.True(t, paths["/api/sysVariable/delete/:id"])
	assert.True(t, paths["/api/sysVariable/query"])
	assert.True(t, paths["/api/sysVariable/value/create"])
	assert.True(t, paths["/api/sysVariable/value/edit"])
	assert.True(t, paths["/api/sysVariable/value/delete/:id"])
	assert.True(t, paths["/api/sysVariable/value/selected/:id"])
	assert.True(t, paths["/api/sysVariable/value/selected/:page/:limit"])
	assert.True(t, paths["/api/sysVariable/value/batchDel"])
}
