//go:build integration

package service

import (
	"context"
	"testing"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper: build real services wired to testDB
func newTestAuditAlertServices(t *testing.T) (*AuditAlertService, *AuditService, *SystemParamService) {
	t.Helper()

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	auditSvc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	settingsRepo := repository.NewSystemParamRepository(testDB)
	settingsSvc := NewSystemParamService(settingsRepo, auditSvc)

	alertSvc := NewAuditAlertService(settingsSvc, auditSvc)
	return alertSvc, auditSvc, settingsSvc
}

// ---------- Failed login threshold ----------

func TestAuditAlertServiceIntegration_FailedLoginThreshold(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	alertSvc, auditSvc, settingsSvc := newTestAuditAlertServices(t)

	// Persist alert settings: low threshold of 3
	settings := audit.DefaultAuditAlertSettings()
	settings.EnableAlerts = true
	settings.FailedLoginThreshold = 3
	settings.AlertOnPermissionChange = false
	settings.BatchOperationThreshold = 1000 // high so it doesn't fire
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	// Insert 3 login failures for "alertuser"
	for i := 0; i < 3; i++ {
		_, err := auditSvc.RecordLoginFailure(&audit.LoginFailureRequest{
			Username:      "alertuser",
			IPAddress:     strPtr("10.0.0.1"),
			FailureReason: strPtr("bad password"),
		})
		require.NoError(t, err)
	}

	// Run detection
	err := alertSvc.DetectAndAlert(context.Background())
	require.NoError(t, err)

	// Verify alert audit log created
	var alertLogs []audit.AuditLog
	require.NoError(t, testDB.Where("action_name LIKE ?", "安全告警: failed_login%").Find(&alertLogs).Error)
	require.Len(t, alertLogs, 1)
	assert.Equal(t, audit.ActionTypeSystemConfig, alertLogs[0].ActionType)
	require.NotNil(t, alertLogs[0].AfterValue)
	assert.Contains(t, *alertLogs[0].AfterValue, "alertuser")
	assert.Contains(t, *alertLogs[0].AfterValue, "3")
}

func TestAuditAlertServiceIntegration_FailedLoginBelowThreshold_NoAlert(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	alertSvc, auditSvc, settingsSvc := newTestAuditAlertServices(t)

	settings := audit.DefaultAuditAlertSettings()
	settings.EnableAlerts = true
	settings.FailedLoginThreshold = 5
	settings.AlertOnPermissionChange = false
	settings.BatchOperationThreshold = 1000
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	// Only 2 failures (below threshold)
	for i := 0; i < 2; i++ {
		_, err := auditSvc.RecordLoginFailure(&audit.LoginFailureRequest{
			Username:      "quietuser",
			FailureReason: strPtr("bad password"),
		})
		require.NoError(t, err)
	}

	err := alertSvc.DetectAndAlert(context.Background())
	require.NoError(t, err)

	var count int64
	testDB.Model(&audit.AuditLog{}).Where("action_name LIKE ?", "安全告警:%").Count(&count)
	assert.Equal(t, int64(0), count)
}

// ---------- Permission change alert ----------

func TestAuditAlertServiceIntegration_PermissionChangeAlert(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	alertSvc, auditSvc, settingsSvc := newTestAuditAlertServices(t)

	settings := audit.DefaultAuditAlertSettings()
	settings.EnableAlerts = true
	settings.AlertOnPermissionChange = true
	settings.FailedLoginThreshold = 100
	settings.BatchOperationThreshold = 9999
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	_, err := auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
		Username:     strPtr("admin"),
		ActionType:   audit.ActionTypePermissionChange,
		ActionName:   "修改角色权限",
		ResourceType: strPtr("ROLE"),
		Operation:    audit.OperationUpdate,
	})
	require.NoError(t, err)

	err = alertSvc.DetectAndAlert(context.Background())
	require.NoError(t, err)

	var alertLogs []audit.AuditLog
	require.NoError(t, testDB.Where("action_name LIKE ?", "安全告警: permission_change%").Find(&alertLogs).Error)
	require.Len(t, alertLogs, 1)
	require.NotNil(t, alertLogs[0].AfterValue)
	assert.Contains(t, *alertLogs[0].AfterValue, "修改角色权限")
}

func TestAuditAlertServiceIntegration_PermissionChange_DisabledFlag_NoAlert(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	alertSvc, auditSvc, settingsSvc := newTestAuditAlertServices(t)

	settings := audit.DefaultAuditAlertSettings()
	settings.EnableAlerts = true
	settings.AlertOnPermissionChange = false
	settings.FailedLoginThreshold = 100
	settings.BatchOperationThreshold = 9999
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	_, err := auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
		Username:   strPtr("admin"),
		ActionType: audit.ActionTypePermissionChange,
		ActionName: "修改角色权限",
		Operation:  audit.OperationUpdate,
	})
	require.NoError(t, err)

	err = alertSvc.DetectAndAlert(context.Background())
	require.NoError(t, err)

	var count int64
	testDB.Model(&audit.AuditLog{}).Where("action_name LIKE ?", "安全告警: permission_change%").Count(&count)
	assert.Equal(t, int64(0), count)
}

// ---------- Batch operation threshold ----------

func TestAuditAlertServiceIntegration_BatchOperationThreshold(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	alertSvc, auditSvc, settingsSvc := newTestAuditAlertServices(t)

	settings := audit.DefaultAuditAlertSettings()
	settings.EnableAlerts = true
	settings.AlertOnPermissionChange = false
	settings.FailedLoginThreshold = 100
	settings.BatchOperationThreshold = 5
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	// Create 5 user-action audit logs for "batchuser"
	userID := int64(42)
	for i := 0; i < 5; i++ {
		_, err := auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:       &userID,
			Username:     strPtr("batchuser"),
			ActionType:   audit.ActionTypeUserAction,
			ActionName:   "批量删除仪表板",
			ResourceType: strPtr("DASHBOARD"),
			Operation:    audit.OperationDelete,
		})
		require.NoError(t, err)
	}

	err := alertSvc.DetectAndAlert(context.Background())
	require.NoError(t, err)

	var alertLogs []audit.AuditLog
	require.NoError(t, testDB.Where("action_name LIKE ?", "安全告警: batch_operation%").Find(&alertLogs).Error)
	require.Len(t, alertLogs, 1)
	require.NotNil(t, alertLogs[0].AfterValue)
	assert.Contains(t, *alertLogs[0].AfterValue, "batchuser")
	assert.Contains(t, *alertLogs[0].AfterValue, "5")
}

func TestAuditAlertServiceIntegration_BatchOperationBelowThreshold_NoAlert(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	alertSvc, auditSvc, settingsSvc := newTestAuditAlertServices(t)

	settings := audit.DefaultAuditAlertSettings()
	settings.EnableAlerts = true
	settings.AlertOnPermissionChange = false
	settings.FailedLoginThreshold = 100
	settings.BatchOperationThreshold = 50
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	userID := int64(43)
	for i := 0; i < 3; i++ {
		_, err := auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:     &userID,
			Username:   strPtr("lightuser"),
			ActionType: audit.ActionTypeUserAction,
			ActionName: "普通操作",
			Operation:  audit.OperationCreate,
		})
		require.NoError(t, err)
	}

	err := alertSvc.DetectAndAlert(context.Background())
	require.NoError(t, err)

	var count int64
	testDB.Model(&audit.AuditLog{}).Where("action_name LIKE ?", "安全告警: batch_operation%").Count(&count)
	assert.Equal(t, int64(0), count)
}

// ---------- Alerts disabled short-circuit ----------

func TestAuditAlertServiceIntegration_AlertsDisabled_NoAlerts(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	alertSvc, auditSvc, settingsSvc := newTestAuditAlertServices(t)

	settings := audit.DefaultAuditAlertSettings()
	settings.EnableAlerts = false // disabled
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	// Insert data that would normally trigger all alert types
	for i := 0; i < 10; i++ {
		_, _ = auditSvc.RecordLoginFailure(&audit.LoginFailureRequest{
			Username:      "disableduser",
			FailureReason: strPtr("bad"),
		})
	}
	_, _ = auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
		Username:   strPtr("disableduser"),
		ActionType: audit.ActionTypePermissionChange,
		ActionName: "改权限",
		Operation:  audit.OperationUpdate,
	})

	err := alertSvc.DetectAndAlert(context.Background())
	require.NoError(t, err)

	var count int64
	testDB.Model(&audit.AuditLog{}).Where("action_name LIKE ?", "安全告警:%").Count(&count)
	assert.Equal(t, int64(0), count)
}

// ---------- Duplicate alert on repeated DetectAndAlert calls ----------

func TestAuditAlertServiceIntegration_DuplicateAlertOnRepeatedCalls(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	alertSvc, auditSvc, settingsSvc := newTestAuditAlertServices(t)

	settings := audit.DefaultAuditAlertSettings()
	settings.EnableAlerts = true
	settings.AlertOnPermissionChange = true
	settings.FailedLoginThreshold = 100
	settings.BatchOperationThreshold = 9999
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	_, err := auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
		Username:   strPtr("dupeuser"),
		ActionType: audit.ActionTypePermissionChange,
		ActionName: "修改角色权限",
		Operation:  audit.OperationUpdate,
	})
	require.NoError(t, err)

	err = alertSvc.DetectAndAlert(context.Background())
	require.NoError(t, err)

	err = alertSvc.DetectAndAlert(context.Background())
	require.NoError(t, err)

	var alertLogs []audit.AuditLog
	require.NoError(t, testDB.Where("action_name LIKE ?", "安全告警: permission_change%").Find(&alertLogs).Error)
	assert.Len(t, alertLogs, 2, "calling DetectAndAlert twice within the lookback window produces two alerts")
}

// ---------- Canceled context ----------

func TestAuditAlertServiceIntegration_CanceledContext(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	alertSvc, _, settingsSvc := newTestAuditAlertServices(t)

	settings := audit.DefaultAuditAlertSettings()
	settings.EnableAlerts = true
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := alertSvc.DetectAndAlert(ctx)
	assert.Error(t, err)
}
