package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type auditHandlerTestEnv struct {
	r  *gin.Engine
	db *gorm.DB
}

type auditHandlerResp struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func auditStringPtr(v string) *string {
	return &v
}

func setupAuditHandlerTestEnv(t *testing.T) *auditHandlerTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&audit.AuditLog{}, &audit.AuditLogDetail{}, &audit.LoginFailure{}, &systemParamCoreSysSettingMirror{}))

	auditSvc := service.NewAuditService(
		repository.NewAuditLogRepository(db),
		repository.NewLoginFailureRepository(db),
		repository.NewAuditLogDetailRepository(db),
	)
	paramSvc := service.NewSystemParamService(repository.NewSystemParamRepository(db), auditSvc)
	h := NewAuditHandler(auditSvc, paramSvc)

	r := gin.New()
	RegisterAuditRoutes(r.Group(""), h)
	return &auditHandlerTestEnv{r: r, db: db}
}

func performAuditRequest(t *testing.T, r *gin.Engine, method string, path string, body any) *httptest.ResponseRecorder {
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

func decodeAuditHandlerResp(t *testing.T, body []byte) auditHandlerResp {
	t.Helper()
	var resp auditHandlerResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func TestAuditHandlerGetAuditAlertSettingsDefault(t *testing.T) {
	env := setupAuditHandlerTestEnv(t)
	w := performAuditRequest(t, env.r, http.MethodGet, "/audit/settings", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeAuditHandlerResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)

	var settings audit.AuditAlertSettings
	require.NoError(t, json.Unmarshal(resp.Data, &settings))
	assert.Equal(t, *audit.DefaultAuditAlertSettings(), settings)
}

func TestAuditHandlerSaveAuditAlertSettingsRoundTrip(t *testing.T) {
	env := setupAuditHandlerTestEnv(t)
	body := map[string]any{
		"retentionDays":            45,
		"cleanupFrequency":         "daily",
		"enableAlerts":             true,
		"failedLoginThreshold":     6,
		"alertOnPermissionChange":  true,
		"alertOnSensitiveAccess":   false,
		"batchOperationThreshold":  75,
		"enableEmailNotification":  true,
		"notificationEmail":        "audit@example.com",
		"enableSystemNotification": true,
		"defaultExportFormat":      "json",
		"exportLimit":              2048,
	}

	saveW := performAuditRequest(t, env.r, http.MethodPut, "/audit/settings", body)
	assert.Equal(t, http.StatusOK, saveW.Code)
	assert.Equal(t, "000000", decodeAuditHandlerResp(t, saveW.Body.Bytes()).Code)

	queryW := performAuditRequest(t, env.r, http.MethodGet, "/audit/settings", nil)
	queryResp := decodeAuditHandlerResp(t, queryW.Body.Bytes())
	assert.Equal(t, "000000", queryResp.Code)

	var settings audit.AuditAlertSettings
	require.NoError(t, json.Unmarshal(queryResp.Data, &settings))
	assert.Equal(t, 45, settings.RetentionDays)
	assert.Equal(t, "daily", settings.CleanupFrequency)
	assert.True(t, settings.EnableEmailNotification)
	assert.Equal(t, "audit@example.com", settings.NotificationEmail)
	assert.Equal(t, "json", settings.DefaultExportFormat)
	assert.Equal(t, 2048, settings.ExportLimit)
}

func TestAuditHandlerSaveAuditAlertSettingsInvalidPayload(t *testing.T) {
	env := setupAuditHandlerTestEnv(t)

	invalidJSON := performAuditRequest(t, env.r, http.MethodPut, "/audit/settings", []byte("{"))
	assert.Equal(t, "10001", decodeAuditHandlerResp(t, invalidJSON.Body.Bytes()).Code)

	semanticInvalid := performAuditRequest(t, env.r, http.MethodPut, "/audit/settings", map[string]any{
		"retentionDays":            1,
		"cleanupFrequency":         "daily",
		"enableAlerts":             true,
		"failedLoginThreshold":     5,
		"alertOnPermissionChange":  true,
		"alertOnSensitiveAccess":   false,
		"batchOperationThreshold":  10,
		"enableEmailNotification":  false,
		"notificationEmail":        "",
		"enableSystemNotification": true,
		"defaultExportFormat":      "csv",
		"exportLimit":              100,
	})
	assert.Equal(t, "40001", decodeAuditHandlerResp(t, semanticInvalid.Body.Bytes()).Code)
}

func TestAuditHandlerCleanupNow(t *testing.T) {
	env := setupAuditHandlerTestEnv(t)
	oldTime := time.Now().AddDate(0, 0, -10)
	recentTime := time.Now().AddDate(0, 0, -1)
	require.NoError(t, env.db.Create(&audit.AuditLog{Username: auditStringPtr("old"), ActionType: audit.ActionTypeUserAction, ActionName: "Old", Operation: audit.OperationCreate, Status: audit.StatusSuccess, CreateTime: oldTime}).Error)
	require.NoError(t, env.db.Create(&audit.AuditLog{Username: auditStringPtr("recent"), ActionType: audit.ActionTypeUserAction, ActionName: "Recent", Operation: audit.OperationCreate, Status: audit.StatusSuccess, CreateTime: recentTime}).Error)

	settings := audit.DefaultAuditAlertSettings()
	settings.RetentionDays = 7
	settings.CleanupFrequency = "daily"
	saveW := performAuditRequest(t, env.r, http.MethodPut, "/audit/settings", settings)
	assert.Equal(t, "000000", decodeAuditHandlerResp(t, saveW.Body.Bytes()).Code)

	w := performAuditRequest(t, env.r, http.MethodPost, "/audit/cleanup", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeAuditHandlerResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.JSONEq(t, `{"deleted":1,"retentionDays":7}`, string(resp.Data))

	var logs []audit.AuditLog
	require.NoError(t, env.db.Order("action_name ASC").Find(&logs).Error)
	require.Len(t, logs, 2)
	assert.NotEqual(t, "Old", logs[0].ActionName)
	assert.NotEqual(t, "Old", logs[1].ActionName)
}

func TestAuditHandlerTestNotification(t *testing.T) {
	env := setupAuditHandlerTestEnv(t)
	w := performAuditRequest(t, env.r, http.MethodPost, "/audit/test-notification", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeAuditHandlerResp(t, w.Body.Bytes())
	assert.Equal(t, "000000", resp.Code)
	assert.Contains(t, string(resp.Data), `"type":"batch_operation"`)
	assert.Contains(t, string(resp.Data), `"username":"system"`)

	var logs []audit.AuditLog
	require.NoError(t, env.db.Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.Equal(t, "安全告警: batch_operation", logs[0].ActionName)
	require.NotNil(t, logs[0].AfterValue)
	assert.Equal(t, "测试审计告警通知", *logs[0].AfterValue)
}
