package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInferOperation(t *testing.T) {
	tests := []struct {
		method string
		want   audit.Operation
	}{
		{"POST", audit.OperationCreate},
		{"PUT", audit.OperationUpdate},
		{"PATCH", audit.OperationUpdate},
		{"DELETE", audit.OperationDelete},
		{"GET", audit.OperationExport},
		{"UNKNOWN", audit.OperationCreate},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			assert.Equal(t, tt.want, inferOperation(tt.method))
		})
	}
}

func TestPtrString_Empty(t *testing.T) {
	assert.Nil(t, ptrString(""))
}

func TestPtrString_NonEmpty(t *testing.T) {
	assert.Equal(t, "hello", *ptrString("hello"))
}

func TestGetUsername_Set(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("username", "alice")
	assert.Equal(t, "alice", GetUsername(c))
}

func TestGetUsername_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Equal(t, "", GetUsername(c))
}

func TestGetRole_Set(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("role", "admin")
	assert.Equal(t, "admin", GetRole(c))
}

func TestGetRole_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Equal(t, "", GetRole(c))
}

func TestSetAuditService_Nil(t *testing.T) {
	orig := auditService
	t.Cleanup(func() { auditService = orig })

	assert.NotPanics(t, func() { SetAuditService(nil) })
	assert.Nil(t, auditService)
}

func TestSetAuditService_NonNil(t *testing.T) {
	orig := auditService
	t.Cleanup(func() { auditService = orig })

	svc := &service.AuditService{}
	assert.NotPanics(t, func() { SetAuditService(svc) })
	assert.Equal(t, svc, auditService)
}

func TestAuditLog_NilService_CallsNext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orig := auditService
	t.Cleanup(func() { auditService = orig })
	auditService = nil

	var nextCalled atomic.Bool
	r := gin.New()
	r.GET("/test", AuditLog(AuditConfig{
		ActionType:   audit.ActionTypeUserAction,
		ActionName:   "test-action",
		ResourceType: audit.ResourceTypeUser,
	}), func(c *gin.Context) {
		nextCalled.Store(true)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.True(t, nextCalled.Load(), "AuditLog should call c.Next()")
}

func TestAuditLog_WithErrors_SetsStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orig := auditService
	t.Cleanup(func() { auditService = orig })

	auditService = nil

	r := gin.New()
	r.GET("/test", AuditLog(AuditConfig{
		ActionType:   audit.ActionTypeUserAction,
		ActionName:   "test-action",
		ResourceType: audit.ResourceTypeUser,
	}), func(c *gin.Context) {
		_ = c.Error(assert.AnError)
		c.Status(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestGetUserID_Set(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", uint64(42))
	assert.Equal(t, uint64(42), GetUserID(c))
}

func TestGetUserID_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Equal(t, uint64(0), GetUserID(c))
}

func TestGetOrgID_Uint64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("org_id", uint64(7))
	assert.Equal(t, int64(7), GetOrgID(c))
}

func TestGetOrgID_Int64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("org_id", int64(99))
	assert.Equal(t, int64(99), GetOrgID(c))
}

func TestGetOrgID_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Equal(t, int64(0), GetOrgID(c))
}
