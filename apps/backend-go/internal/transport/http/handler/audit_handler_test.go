//go:build integration

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openAuditHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := envOrDefault("TEST_DB_HOST", "localhost")
	port := envOrDefault("TEST_DB_PORT", "3306")
	user := envOrDefault("TEST_DB_USER", "root")
	password := envOrDefault("TEST_DB_PASSWORD", "Admin168")
	name := envOrDefault("TEST_DB_NAME", "dataease_test")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&audit.AuditLog{}, &audit.AuditLogDetail{}, &audit.LoginFailure{}))

	require.NoError(t, db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error)
	require.NoError(t, db.Exec("DELETE FROM de_audit_log_detail").Error)
	require.NoError(t, db.Exec("DELETE FROM de_audit_log").Error)
	require.NoError(t, db.Exec("DELETE FROM de_login_failure").Error)
	require.NoError(t, db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error)

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func newAuditHandlerForTest(t *testing.T) *AuditHandler {
	t.Helper()
	db := openAuditHandlerTestDB(t)
	auditSvc := service.NewAuditService(
		repository.NewAuditLogRepository(db),
		repository.NewLoginFailureRepository(db),
		repository.NewAuditLogDetailRepository(db),
	)
	return NewAuditHandler(auditSvc)
}

func TestAuditHandler_CreateAuditLog_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newAuditHandlerForTest(t)
	r := gin.New()
	r.POST("/audit/log", h.CreateAuditLog)

	body := `{"actionType":"USER_ACTION","actionName":"Create User","operation":"CREATE","username":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/audit/log", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code string         `json:"code"`
		Data audit.AuditLog `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, audit.ActionTypeUserAction, resp.Data.ActionType)
	assert.Equal(t, "Create User", resp.Data.ActionName)
	if assert.NotNil(t, resp.Data.Username) {
		assert.Equal(t, "alice", *resp.Data.Username)
	}
	assert.Equal(t, audit.StatusSuccess, resp.Data.Status)
	assert.NotZero(t, resp.Data.ID)
	assert.WithinDuration(t, time.Now(), resp.Data.CreateTime, 5*time.Second)
}

func TestAuditHandler_QueryAuditLogs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newAuditHandlerForTest(t)
	db := openAuditHandlerTestDB(t)
	repo := repository.NewAuditLogRepository(db)
	username := "query-user"
	resourceType := string(audit.ResourceTypeDataset)
	for i := range 2 {
		err := repo.Create(&audit.AuditLog{
			Username:     &username,
			ActionType:   audit.ActionTypeUserAction,
			ActionName:   fmt.Sprintf("Action %d", i+1),
			ResourceType: &resourceType,
			Operation:    audit.OperationExport,
			Status:       audit.StatusSuccess,
			CreateTime:   time.Now().Add(time.Duration(i) * time.Second),
		})
		require.NoError(t, err)
	}

	r := gin.New()
	r.GET("/audit/list", h.QueryAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/audit/list?username=query-user&page=1&pageSize=1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code string `json:"code"`
		Data struct {
			List    []audit.AuditLog `json:"list"`
			Total   int64            `json:"total"`
			Current int              `json:"current"`
			Size    int              `json:"size"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Len(t, resp.Data.List, 1)
	assert.Equal(t, int64(2), resp.Data.Total)
	assert.Equal(t, 1, resp.Data.Current)
	assert.Equal(t, 1, resp.Data.Size)
	if assert.NotNil(t, resp.Data.List[0].Username) {
		assert.Equal(t, username, *resp.Data.List[0].Username)
	}
}

func TestAuditHandler_GetAuditLogByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newAuditHandlerForTest(t)
	db := openAuditHandlerTestDB(t)
	repo := repository.NewAuditLogRepository(db)
	username := "detail-user"
	resourceType := string(audit.ResourceTypeDataset)
	log := &audit.AuditLog{
		Username:     &username,
		ActionType:   audit.ActionTypeUserAction,
		ActionName:   "Read Audit",
		ResourceType: &resourceType,
		Operation:    audit.OperationExport,
		Status:       audit.StatusSuccess,
		CreateTime:   time.Now(),
	}
	require.NoError(t, repo.Create(log))

	r := gin.New()
	r.GET("/audit/:id", h.GetAuditLogByID)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/audit/%d", log.ID), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code string         `json:"code"`
		Data audit.AuditLog `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "000000", resp.Code)
	assert.Equal(t, log.ID, resp.Data.ID)
	assert.Equal(t, "Read Audit", resp.Data.ActionName)
	if assert.NotNil(t, resp.Data.Username) {
		assert.Equal(t, username, *resp.Data.Username)
	}
}

func TestAuditHandler_GetAuditLogByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newAuditHandlerForTest(t)
	r := gin.New()
	r.GET("/audit/:id", h.GetAuditLogByID)

	req := httptest.NewRequest(http.MethodGet, "/audit/999999999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "50001", resp.Code)
	assert.Equal(t, "Audit log not found", resp.Msg)
}
