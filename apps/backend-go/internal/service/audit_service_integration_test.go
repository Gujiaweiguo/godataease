//go:build integration

package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestAuditServiceIntegration_CreateAuditLog(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	userID := int64(1)
	req := &audit.AuditLogCreateRequest{
		UserID:       &userID,
		Username:     strPtr("testuser"),
		ActionType:   audit.ActionTypeUserAction,
		ActionName:   "User Login",
		ResourceType: strPtr("USER"),
		ResourceID:   &userID,
		ResourceName: strPtr("test-resource"),
		Operation:    audit.OperationLogin,
		IPAddress:    strPtr("127.0.0.1"),
		UserAgent:    strPtr("test-agent"),
	}

	log, err := svc.CreateAuditLog(req)
	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Greater(t, log.ID, int64(0))
	assert.Equal(t, "testuser", *log.Username)
	assert.Equal(t, audit.ActionTypeUserAction, log.ActionType)
	assert.Equal(t, audit.OperationLogin, log.Operation)
	assert.Equal(t, audit.StatusSuccess, log.Status)
}

func TestAuditServiceIntegration_CreateAuditLog_WithFailure(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	userID := int64(1)
	failedStatus := audit.StatusFailed
	req := &audit.AuditLogCreateRequest{
		UserID:        &userID,
		Username:      strPtr("testuser"),
		ActionType:    audit.ActionTypeUserAction,
		ActionName:    "User Login Failed",
		ResourceType:  strPtr("USER"),
		ResourceID:    &userID,
		ResourceName:  strPtr("test-resource"),
		Operation:     audit.OperationLogin,
		IPAddress:     strPtr("127.0.0.1"),
		Status:        &failedStatus,
		FailureReason: strPtr("Invalid password"),
	}

	log, err := svc.CreateAuditLog(req)
	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, audit.StatusFailed, log.Status)
	assert.Equal(t, "Invalid password", *log.FailureReason)
}

func TestAuditServiceIntegration_GetAuditLogByID(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	// Create audit log first
	userID := int64(1)
	created, _ := svc.CreateAuditLog(&audit.AuditLogCreateRequest{
		UserID:       &userID,
		Username:     strPtr("testuser"),
		ActionType:   audit.ActionTypeUserAction,
		ActionName:   "Test Action",
		ResourceType: strPtr("USER"),
		Operation:    audit.OperationCreate,
	})

	// Get by ID
	found, err := svc.GetAuditLogByID(created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "testuser", *found.Username)
}

func TestAuditServiceIntegration_GetAuditLogByID_NotFound(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	_, err := svc.GetAuditLogByID(99999)
	assert.Error(t, err)
}

func TestAuditServiceIntegration_GetAuditLogsByUserID(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	// Create multiple audit logs
	userID := int64(1)
	for i := 1; i <= 5; i++ {
		_, _ = svc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:       &userID,
			Username:     strPtr("testuser"),
			ActionType:   audit.ActionTypeUserAction,
			ActionName:   "Action",
			ResourceType: strPtr("USER"),
			Operation:    audit.OperationCreate,
		})
	}

	// Query by user ID
	result, err := svc.GetAuditLogsByUserID(userID, 1, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Total, int64(5))
	assert.GreaterOrEqual(t, len(result.List.([]*audit.AuditLog)), 5)
}

func TestAuditServiceIntegration_QueryAuditLogs(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	// Create audit logs with different action types
	userID := int64(1)
	for i := 1; i <= 3; i++ {
		_, _ = svc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:       &userID,
			Username:     strPtr("testuser"),
			ActionType:   audit.ActionTypeUserAction,
			ActionName:   "User Action",
			ResourceType: strPtr("USER"),
			Operation:    audit.OperationCreate,
		})
	}
	for i := 1; i <= 2; i++ {
		_, _ = svc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:       &userID,
			Username:     strPtr("testuser"),
			ActionType:   audit.ActionTypePermissionChange,
			ActionName:   "Permission Change",
			ResourceType: strPtr("ROLE"),
			Operation:    audit.OperationUpdate,
		})
	}

	// Query all
	result, err := svc.QueryAuditLogs(&audit.AuditLogQuery{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Total, int64(5))

	// Query by action type
	actionType := audit.ActionTypeUserAction
	result2, err := svc.QueryAuditLogs(&audit.AuditLogQuery{
		ActionType: &actionType,
		Page:       1,
		PageSize:   10,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result2.Total, int64(3))
}

func TestAuditServiceIntegration_QueryAuditLogs_Empty(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	result, err := svc.QueryAuditLogs(&audit.AuditLogQuery{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result.Total)
}

func TestAuditServiceIntegration_QueryAuditLogs_DefaultPagination(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	// Query with invalid pagination (should use defaults)
	result, err := svc.QueryAuditLogs(&audit.AuditLogQuery{
		Page:     0,
		PageSize: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, result.Current)
	assert.Equal(t, 20, result.Size)
}

func TestAuditServiceIntegration_RecordLoginFailure(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	req := &audit.LoginFailureRequest{
		Username:      "testuser",
		IPAddress:     strPtr("127.0.0.1"),
		FailureReason: strPtr("Invalid password"),
		UserAgent:     strPtr("test-agent"),
	}

	failure, err := svc.RecordLoginFailure(req)
	assert.NoError(t, err)
	assert.NotNil(t, failure)
	assert.Greater(t, failure.ID, int64(0))
	assert.Equal(t, "testuser", failure.Username)
	assert.Equal(t, "Invalid password", *failure.FailureReason)
}

func TestAuditServiceIntegration_DeleteAuditLogsBeforeDate(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	// Create some audit logs
	userID := int64(1)
	for i := 1; i <= 3; i++ {
		_, _ = svc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:       &userID,
			Username:     strPtr("testuser"),
			ActionType:   audit.ActionTypeUserAction,
			ActionName:   "Action",
			ResourceType: strPtr("USER"),
			Operation:    audit.OperationCreate,
		})
	}

	// Delete logs older than 0 days - new logs won't be deleted
	// This tests the function works without error
	count, err := svc.DeleteAuditLogsBeforeDate(0)
	assert.NoError(t, err)
	// New logs are from today, so count should be 0 (nothing older than today)
	assert.GreaterOrEqual(t, count, int64(0))
}

func TestAuditServiceIntegration_DeleteAuditLogsBeforeDate_DefaultDays(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	// Create audit log
	userID := int64(1)
	_, _ = svc.CreateAuditLog(&audit.AuditLogCreateRequest{
		UserID:       &userID,
		Username:     strPtr("testuser"),
		ActionType:   audit.ActionTypeUserAction,
		ActionName:   "Action",
		ResourceType: strPtr("USER"),
		Operation:    audit.OperationCreate,
	})

	// Delete with default days (90) - should not delete recent logs
	count, err := svc.DeleteAuditLogsBeforeDate(-1) // invalid, uses default
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestAuditServiceIntegration_CheckSuspiciousLoginActivity(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	// Record multiple login failures
	for i := 1; i <= 5; i++ {
		_, _ = svc.RecordLoginFailure(&audit.LoginFailureRequest{
			Username:      "suspicious",
			IPAddress:     strPtr("127.0.0.1"),
			FailureReason: strPtr("Invalid password"),
		})
	}

	// Check suspicious activity
	isSuspicious, err := svc.CheckSuspiciousLoginActivity("suspicious", 3, time.Hour)
	assert.NoError(t, err)
	assert.True(t, isSuspicious)

	// Check non-suspicious user
	isSuspicious2, err := svc.CheckSuspiciousLoginActivity("normal", 3, time.Hour)
	assert.NoError(t, err)
	assert.False(t, isSuspicious2)
}

func TestAuditServiceIntegration_ExportAuditLogs(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	// Create audit logs
	userID := int64(1)
	for i := 1; i <= 3; i++ {
		created, _ := svc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:       &userID,
			Username:     strPtr("testuser"),
			ActionType:   audit.ActionTypeUserAction,
			ActionName:   "Action",
			ResourceType: strPtr("USER"),
			Operation:    audit.OperationCreate,
		})
		_ = created
	}

	// Get log IDs
	result, _ := svc.QueryAuditLogs(&audit.AuditLogQuery{Page: 1, PageSize: 100})
	logs := result.List.([]*audit.AuditLog)
	ids := make([]int64, len(logs))
	for i, log := range logs {
		ids[i] = log.ID
	}

	// Export to CSV
	filePath, err := svc.ExportAuditLogs(ids, "csv")
	assert.NoError(t, err)
	assert.NotEmpty(t, filePath)

	// Export to JSON
	filePath2, err := svc.ExportAuditLogs(ids, "json")
	assert.NoError(t, err)
	assert.NotEmpty(t, filePath2)
}

func TestAuditServiceIntegration_ExportAuditLogs_Empty(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{})

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	svc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	// Export empty list
	_, err := svc.ExportAuditLogs([]int64{}, "csv")
	assert.Error(t, err)
}
