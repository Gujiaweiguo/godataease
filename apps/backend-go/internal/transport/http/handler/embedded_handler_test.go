package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/embedded"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type embeddedHandlerTestEnv struct {
	r  *gin.Engine
	db *gorm.DB
}

type embeddedBridgeResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func setupEmbeddedHandlerTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	return setupEmbeddedHandlerTestEnv(t).r
}

func setupEmbeddedHandlerTestEnv(t *testing.T) *embeddedHandlerTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&embedded.CoreEmbedded{}))

	repo := repository.NewEmbeddedRepository(db)
	svc := service.NewEmbeddedService(repo)
	h := NewEmbeddedHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userId", int64(99))
		c.Set("orgId", int64(7))
		c.Set("user_id", int64(99))
		c.Next()
	})
	api := r.Group("/api")
	RegisterEmbeddedRoutes(api, h)

	return &embeddedHandlerTestEnv{r: r, db: db}
}

func performEmbeddedJSONRequest(t *testing.T, r *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
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

func decodeEmbeddedResp(t *testing.T, body []byte) embeddedBridgeResp {
	t.Helper()
	var resp embeddedBridgeResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func seedEmbedded(t *testing.T, db *gorm.DB, e *embedded.CoreEmbedded) {
	t.Helper()
	require.NoError(t, db.Create(e).Error)
}

func TestEmbeddedHandler_QueryGrid_Success(t *testing.T) {
	env := setupEmbeddedHandlerTestEnv(t)
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 1, Name: "match-app", AppId: "app_query", AppSecret: "abcdefghijklmnop", Domain: "https://foo.example.com", SecretLength: 16, CreateTime: 200})
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 2, Name: "other-app", AppId: "app_other", AppSecret: "qrstuvwxyzabcdef", Domain: "https://bar.example.com", SecretLength: 16, CreateTime: 100})

	w := performEmbeddedJSONRequest(t, env.r, http.MethodPost, "/api/embedded/pager/1/10", map[string]any{"keyword": "match"})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data embedded.EmbeddedPagerResponse
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.Len(t, data.List, 1)
	assert.Equal(t, int64(1), data.Total)
	assert.Equal(t, 1, data.Current)
	assert.Equal(t, 10, data.Size)
	assert.Equal(t, int64(1), data.List[0].ID)
	assert.Equal(t, "match-app", data.List[0].Name)
	assert.Equal(t, "abcd****mnop", data.List[0].AppSecret)
}

func TestEmbeddedHandler_QueryGrid_Empty(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodPost, "/api/embedded/pager/1/10", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data embedded.EmbeddedPagerResponse
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Empty(t, data.List)
	assert.Equal(t, int64(0), data.Total)
	assert.Equal(t, 1, data.Current)
	assert.Equal(t, 10, data.Size)
}

func TestEmbeddedHandler_QueryGrid_InvalidJSON(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodPost, "/api/embedded/pager/1/10", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestEmbeddedHandler_Create_Success(t *testing.T) {
	env := setupEmbeddedHandlerTestEnv(t)
	body := map[string]any{
		"name":         "create-app",
		"domain":       "https://create.example.com",
		"secretLength": 20,
	}

	w := performEmbeddedJSONRequest(t, env.r, http.MethodPost, "/api/embedded/create", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var createdID int64
	require.NoError(t, json.Unmarshal(resp.Data, &createdID))
	assert.NotZero(t, createdID)

	created, err := repository.NewEmbeddedRepository(env.db).GetByID(createdID)
	require.NoError(t, err)
	assert.Equal(t, "create-app", created.Name)
	assert.Equal(t, "https://create.example.com", created.Domain)
	assert.Equal(t, 20, created.SecretLength)
	assert.Equal(t, "99", created.UpdateBy)
	assert.Len(t, created.AppSecret, 20)
	assert.True(t, strings.HasPrefix(created.AppId, "app_"))
}

func TestEmbeddedHandler_Create_InvalidJSON(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodPost, "/api/embedded/create", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestEmbeddedHandler_Edit_Success(t *testing.T) {
	env := setupEmbeddedHandlerTestEnv(t)
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 11, Name: "before-edit", AppId: "app_edit", AppSecret: "beforesecret1234", Domain: "https://before.example.com", SecretLength: 16, UpdateBy: "old"})
	body := map[string]any{
		"id":           11,
		"name":         "after-edit",
		"domain":       "https://after.example.com",
		"secretLength": 24,
	}

	w := performEmbeddedJSONRequest(t, env.r, http.MethodPost, "/api/embedded/edit", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	updated, err := repository.NewEmbeddedRepository(env.db).GetByID(11)
	require.NoError(t, err)
	assert.Equal(t, "after-edit", updated.Name)
	assert.Equal(t, "https://after.example.com", updated.Domain)
	assert.Equal(t, 24, updated.SecretLength)
	assert.Equal(t, "99", updated.UpdateBy)
}

func TestEmbeddedHandler_Edit_InvalidJSON(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodPost, "/api/embedded/edit", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestEmbeddedHandler_Delete_Success(t *testing.T) {
	env := setupEmbeddedHandlerTestEnv(t)
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 21, Name: "delete-app", AppId: "app_delete", AppSecret: "delete-secret-123", Domain: "https://delete.example.com", SecretLength: 16})

	w := performEmbeddedJSONRequest(t, env.r, http.MethodPost, "/api/embedded/delete/21", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	_, err := repository.NewEmbeddedRepository(env.db).GetByID(21)
	assert.Error(t, err)
}

func TestEmbeddedHandler_Delete_InvalidID(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodPost, "/api/embedded/delete/not-a-number", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Equal(t, "Invalid ID", resp.Msg)
}

func TestEmbeddedHandler_BatchDelete_Success(t *testing.T) {
	env := setupEmbeddedHandlerTestEnv(t)
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 31, Name: "batch-a", AppId: "app_batch_a", AppSecret: "batch-secret-aaaa", Domain: "a.com", SecretLength: 16})
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 32, Name: "batch-b", AppId: "app_batch_b", AppSecret: "batch-secret-bbbb", Domain: "b.com", SecretLength: 16})

	w := performEmbeddedJSONRequest(t, env.r, http.MethodPost, "/api/embedded/batchDelete", map[string]any{"ids": []int64{31, 32}})

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var count int64
	require.NoError(t, env.db.Model(&embedded.CoreEmbedded{}).Where("id IN ?", []int64{31, 32}).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestEmbeddedHandler_BatchDelete_InvalidJSON(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodPost, "/api/embedded/batchDelete", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestEmbeddedHandler_Reset_Success(t *testing.T) {
	env := setupEmbeddedHandlerTestEnv(t)
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 41, Name: "reset-app", AppId: "app_reset", AppSecret: "old-secret-123456", Domain: "https://reset.example.com", SecretLength: 16, UpdateBy: "old"})
	body := map[string]any{
		"id":        41,
		"appSecret": "manual-reset-secret",
	}

	w := performEmbeddedJSONRequest(t, env.r, http.MethodPost, "/api/embedded/reset", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	updated, err := repository.NewEmbeddedRepository(env.db).GetByID(41)
	require.NoError(t, err)
	assert.Equal(t, "manual-reset-secret", updated.AppSecret)
	assert.Equal(t, "99", updated.UpdateBy)
}

func TestEmbeddedHandler_Reset_InvalidJSON(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodPost, "/api/embedded/reset", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestEmbeddedHandler_DomainList_Success(t *testing.T) {
	env := setupEmbeddedHandlerTestEnv(t)
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 51, Name: "domain-a", AppId: "app_domain_a", AppSecret: "domain-secret-aaaa", Domain: "https://foo.example.com, bar.example.com", SecretLength: 16})
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 52, Name: "domain-b", AppId: "app_domain_b", AppSecret: "domain-secret-bbbb", Domain: "bar.example.com;https://baz.example.com/", SecretLength: 16})

	w := performEmbeddedJSONRequest(t, env.r, http.MethodGet, "/api/embedded/domainList", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var domains []string
	require.NoError(t, json.Unmarshal(resp.Data, &domains))
	assert.ElementsMatch(t, []string{"https://foo.example.com", "bar.example.com", "https://baz.example.com"}, domains)
}

func TestEmbeddedHandler_InitIframe_Success(t *testing.T) {
	env := setupEmbeddedHandlerTestEnv(t)
	seedEmbedded(t, env.db, &embedded.CoreEmbedded{ID: 61, Name: "iframe-app", AppId: "appiframe", AppSecret: "iframe-secret-1234", Domain: "https://iframe.example.com,allowed.example.com", SecretLength: 16})
	body := map[string]any{
		"token":  "header.appId:appiframe.signature",
		"origin": "https://iframe.example.com",
	}

	w := performEmbeddedJSONRequest(t, env.r, http.MethodPost, "/api/embedded/initIframe", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var domains []string
	require.NoError(t, json.Unmarshal(resp.Data, &domains))
	assert.ElementsMatch(t, []string{"https://iframe.example.com", "allowed.example.com"}, domains)
}

func TestEmbeddedHandler_InitIframe_InvalidJSON(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodPost, "/api/embedded/initIframe", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestEmbeddedHandler_GetTokenArgs_Success(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodGet, "/api/embedded/getTokenArgs", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data embedded.TokenArgsResponse
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, int64(99), data.UserId)
	assert.Equal(t, int64(7), data.OrgId)
}

func TestEmbeddedHandler_GetLimitCount_Success(t *testing.T) {
	r := setupEmbeddedHandlerTestRouter(t)

	w := performEmbeddedJSONRequest(t, r, http.MethodGet, "/api/embedded/limitCount", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeEmbeddedResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var count int
	require.NoError(t, json.Unmarshal(resp.Data, &count))
	assert.Equal(t, 5, count)
}
