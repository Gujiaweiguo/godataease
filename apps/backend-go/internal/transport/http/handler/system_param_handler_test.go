package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type systemParamHandlerTestEnv struct {
	r  *gin.Engine
	db *gorm.DB
}

type systemParamBridgeResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type systemParamCoreSysSettingMirror struct {
	ID   int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Pkey string `gorm:"column:pkey"`
	Pval string `gorm:"column:pval"`
	Type string `gorm:"column:type"`
	Sort int    `gorm:"column:sort"`
}

func (systemParamCoreSysSettingMirror) TableName() string {
	return "core_sys_setting"
}

func setupSystemParamHandlerTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	return setupSystemParamHandlerTestEnv(t).r
}

func setupSystemParamHandlerTestEnv(t *testing.T) *systemParamHandlerTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&systemParamCoreSysSettingMirror{}))

	repo := repository.NewSystemParamRepository(db)
	svc := service.NewSystemParamService(repo, nil)
	h := NewSystemParamHandler(svc)

	r := gin.New()
	api := r.Group("/api")
	RegisterSystemParamRoutes(api, h)

	return &systemParamHandlerTestEnv{r: r, db: db}
}

func performSystemParamJSONRequest(t *testing.T, r *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
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

func decodeSystemParamResp(t *testing.T, body []byte) systemParamBridgeResp {
	t.Helper()
	var resp systemParamBridgeResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func seedSystemParamRows(t *testing.T, db *gorm.DB, rows ...systemParamCoreSysSettingMirror) {
	t.Helper()
	for _, row := range rows {
		require.NoError(t, db.Create(&row).Error)
	}
}

func TestSystemParamHandler_QueryBasic_Empty(t *testing.T) {
	r := setupSystemParamHandlerTestRouter(t)

	w := performSystemParamJSONRequest(t, r, http.MethodGet, "/api/sysParameter/basic/query", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeSystemParamResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.JSONEq(t, "[]", string(resp.Data))
}

func TestSystemParamHandler_SaveBasic_SuccessAndRoundTrip(t *testing.T) {
	env := setupSystemParamHandlerTestEnv(t)
	body := []map[string]any{
		{"pkey": "shareDisable", "pval": "true", "type": "bool", "sort": 2},
		{"pkey": "basic.defaultSort", "pval": "9", "type": "text", "sort": 1},
	}

	w := performSystemParamJSONRequest(t, env.r, http.MethodPost, "/api/sysParameter/basic/save", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeSystemParamResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var rows []systemParamCoreSysSettingMirror
	require.NoError(t, env.db.Order("sort ASC").Find(&rows).Error)
	require.Len(t, rows, 2)
	assert.Equal(t, "basic.defaultSort", rows[0].Pkey)
	assert.Equal(t, "basic.shareDisable", rows[1].Pkey)

	queryW := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/basic/query", nil)
	assert.Equal(t, http.StatusOK, queryW.Code)
	queryResp := decodeSystemParamResp(t, queryW.Body.Bytes())
	assert.Equal(t, "000000", queryResp.Code)

	var data []system.SettingItem
	require.NoError(t, json.Unmarshal(queryResp.Data, &data))
	require.Len(t, data, 2)
	assert.Equal(t, "basic.defaultSort", data[0].Pkey)
	assert.Equal(t, "9", data[0].Pval)
	assert.Equal(t, "basic.shareDisable", data[1].Pkey)
	assert.Equal(t, "true", data[1].Pval)
}

func TestSystemParamHandler_SaveBasic_InvalidJSON(t *testing.T) {
	r := setupSystemParamHandlerTestRouter(t)

	w := performSystemParamJSONRequest(t, r, http.MethodPost, "/api/sysParameter/basic/save", []byte("{"))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeSystemParamResp(t, w.Body.Bytes())
	assert.Equal(t, "500000", resp.Code)
	assert.Contains(t, resp.Msg, "Invalid request")
}

func TestSystemParamHandler_SaveOnlineMap_SuccessAndRoundTrip(t *testing.T) {
	env := setupSystemParamHandlerTestEnv(t)
	body := map[string]any{
		"mapType":      "gaode",
		"key":          "gaode-test-key",
		"securityCode": "gaode-sec",
	}

	w := performSystemParamJSONRequest(t, env.r, http.MethodPost, "/api/sysParameter/saveOnlineMap", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeSystemParamResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	queryW := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/queryOnlineMap", nil)
	assert.Equal(t, http.StatusOK, queryW.Code)
	queryResp := decodeSystemParamResp(t, queryW.Body.Bytes())
	assert.Equal(t, "000000", queryResp.Code)

	var data system.OnlineMapEditor
	require.NoError(t, json.Unmarshal(queryResp.Data, &data))
	assert.Equal(t, "gaode", data.MapType)
	assert.Equal(t, "gaode-test-key", data.Key)
	assert.Equal(t, "gaode-sec", data.SecurityCode)
}

func TestSystemParamHandler_QueryOnlineMapByType(t *testing.T) {
	env := setupSystemParamHandlerTestEnv(t)
	seedSystemParamRows(t, env.db,
		systemParamCoreSysSettingMirror{Pkey: "tencent.map.key", Pval: "tencent-key", Type: "text", Sort: 1},
		systemParamCoreSysSettingMirror{Pkey: "tencent.map.securityCode", Pval: "tencent-sec", Type: "text", Sort: 2},
	)

	w := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/queryOnlineMap/tencent", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeSystemParamResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var data system.OnlineMapEditor
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, "tencent", data.MapType)
	assert.Equal(t, "tencent-key", data.Key)
	assert.Equal(t, "tencent-sec", data.SecurityCode)
}

func TestSystemParamHandler_QuerySQLBot_EmptyReturnsNil(t *testing.T) {
	r := setupSystemParamHandlerTestRouter(t)

	w := performSystemParamJSONRequest(t, r, http.MethodGet, "/api/sysParameter/sqlbot", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeSystemParamResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, "null", string(resp.Data))
}

func TestSystemParamHandler_SaveSQLBot_SuccessAndRoundTrip(t *testing.T) {
	env := setupSystemParamHandlerTestEnv(t)
	body := map[string]any{
		"domain":  "https://sqlbot.example.com",
		"id":      "sqlbot-id",
		"enabled": true,
		"valid":   true,
	}

	w := performSystemParamJSONRequest(t, env.r, http.MethodPost, "/api/sysParameter/sqlbot", body)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeSystemParamResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	queryW := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/sqlbot", nil)
	assert.Equal(t, http.StatusOK, queryW.Code)
	queryResp := decodeSystemParamResp(t, queryW.Body.Bytes())
	assert.Equal(t, "000000", queryResp.Code)

	var data system.SQLBotConfig
	require.NoError(t, json.Unmarshal(queryResp.Data, &data))
	assert.Equal(t, "https://sqlbot.example.com", data.Domain)
	assert.Equal(t, "sqlbot-id", data.ID)
	assert.True(t, data.Enabled)
	assert.True(t, data.Valid)
}

func TestSystemParamHandler_ReadonlyEndpointsBundle(t *testing.T) {
	env := setupSystemParamHandlerTestEnv(t)
	seedSystemParamRows(t, env.db,
		systemParamCoreSysSettingMirror{Pkey: "basic.shareDisable", Pval: "true", Type: "bool", Sort: 1},
		systemParamCoreSysSettingMirror{Pkey: "basic.sharePeRequire", Pval: "false", Type: "bool", Sort: 2},
		systemParamCoreSysSettingMirror{Pkey: "basic.frontTimeOut", Pval: "invalid", Type: "text", Sort: 3},
		systemParamCoreSysSettingMirror{Pkey: "basic.defaultSort", Pval: "2", Type: "text", Sort: 4},
		systemParamCoreSysSettingMirror{Pkey: "basic.defaultOpen", Pval: "expand", Type: "text", Sort: 5},
		systemParamCoreSysSettingMirror{Pkey: "basic.defaultLogin", Pval: "3", Type: "text", Sort: 6},
	)

	t.Run("share base", func(t *testing.T) {
		w := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/shareBase", nil)
		resp := decodeSystemParamResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)

		var data system.ShareBase
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		assert.True(t, data.Disable)
		assert.False(t, data.PERequire)
	})

	t.Run("request timeout", func(t *testing.T) {
		w := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/requestTimeOut", nil)
		resp := decodeSystemParamResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)

		var timeout int
		require.NoError(t, json.Unmarshal(resp.Data, &timeout))
		assert.Equal(t, 60, timeout)
	})

	t.Run("default settings", func(t *testing.T) {
		w := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/defaultSettings", nil)
		resp := decodeSystemParamResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)

		var data map[string]any
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		assert.Equal(t, "2", data["defaultSort"])
		assert.Equal(t, "expand", data["defaultOpen"])
	})

	t.Run("ui", func(t *testing.T) {
		w := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/ui", nil)
		resp := decodeSystemParamResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)

		var data []map[string]any
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		require.Len(t, data, 3)
		assert.Equal(t, "community", data[0]["pkey"])
		assert.Equal(t, true, data[0]["pval"])
	})

	t.Run("default login", func(t *testing.T) {
		w := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/defaultLogin", nil)
		resp := decodeSystemParamResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)

		var data int
		require.NoError(t, json.Unmarshal(resp.Data, &data))
		assert.Equal(t, 3, data)
	})

	t.Run("i18n options", func(t *testing.T) {
		w := performSystemParamJSONRequest(t, env.r, http.MethodGet, "/api/sysParameter/i18nOptions", nil)
		resp := decodeSystemParamResp(t, w.Body.Bytes())
		assert.Equal(t, "000000", resp.Code)
		assert.JSONEq(t, "{}", string(resp.Data))
	})
}
