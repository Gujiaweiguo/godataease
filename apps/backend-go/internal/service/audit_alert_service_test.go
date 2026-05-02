package service

import (
	"context"
	"testing"
	"time"

	"dataease/backend/internal/domain/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureNotifier struct {
	events []AlertEvent
	err    error
}

func (n *captureNotifier) Send(event AlertEvent) error {
	if n.err != nil {
		return n.err
	}
	n.events = append(n.events, event)
	return nil
}

func TestLogNotifier_Send(t *testing.T) {
	auditSvc, _ := setupAuditServiceRepoTest(t)
	notifier := NewLogNotifier(auditSvc)

	err := notifier.Send(AlertEvent{
		Type:       AlertTypeFailedLogin,
		Username:   "tester",
		Details:    "登录失败次数过多",
		DetectedAt: time.Now(),
	})
	require.NoError(t, err)

	result, err := auditSvc.QueryAuditLogs(&audit.AuditLogQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	logs := result.List.([]*audit.AuditLog)
	require.Len(t, logs, 1)
	assert.Equal(t, audit.ActionTypeSystemConfig, logs[0].ActionType)
	assert.Equal(t, "安全告警: failed_login", logs[0].ActionName)
	require.NotNil(t, logs[0].Username)
	assert.Equal(t, "tester", *logs[0].Username)
	require.NotNil(t, logs[0].AfterValue)
	assert.Equal(t, "登录失败次数过多", *logs[0].AfterValue)
}

func TestAuditAlertService_DetectAndAlert(t *testing.T) {
	t.Run("skips when alerts disabled", func(t *testing.T) {
		auditSvc, _ := setupAuditServiceRepoTest(t)
		settingsRepo := NewMockSystemParamRepository()
		settings := audit.DefaultAuditAlertSettings()
		settings.EnableAlerts = false
		data, err := settings.ToJSON()
		require.NoError(t, err)
		settingsRepo.settingsByKey[auditAlertSettingsKey] = string(data)

		svc := NewAuditAlertService(NewSystemParamService(settingsRepo, nil), auditSvc)
		capture := &captureNotifier{}
		svc.notifiers = []AlertNotifier{capture}

		err = svc.DetectAndAlert(context.Background())
		require.NoError(t, err)
		assert.Empty(t, capture.events)
	})

	t.Run("detects failed logins permission changes and batch operations", func(t *testing.T) {
		auditSvc, db := setupAuditServiceRepoTest(t)
		now := time.Now()
		for i := 0; i < 3; i++ {
			require.NoError(t, db.Create(&audit.LoginFailure{Username: "alice", CreateTime: now.Add(-10 * time.Minute)}).Error)
		}
		for i := 0; i < 2; i++ {
			require.NoError(t, db.Create(&audit.AuditLog{
				Username:   strPtr("alice"),
				ActionType: audit.ActionTypePermissionChange,
				ActionName: "更新权限",
				Operation:  audit.OperationUpdate,
				Status:     audit.StatusSuccess,
				CreateTime: now.Add(-5 * time.Minute),
			}).Error)
		}
		for i := 0; i < 4; i++ {
			require.NoError(t, db.Create(&audit.AuditLog{
				Username:   strPtr("alice"),
				ActionType: audit.ActionTypeUserAction,
				ActionName: "批量操作",
				Operation:  audit.OperationUpdate,
				Status:     audit.StatusSuccess,
				CreateTime: now.Add(-3 * time.Minute),
			}).Error)
		}

		settingsRepo := NewMockSystemParamRepository()
		settings := audit.DefaultAuditAlertSettings()
		settings.FailedLoginThreshold = 3
		settings.BatchOperationThreshold = 4
		data, err := settings.ToJSON()
		require.NoError(t, err)
		settingsRepo.settingsByKey[auditAlertSettingsKey] = string(data)

		svc := NewAuditAlertService(NewSystemParamService(settingsRepo, nil), auditSvc)
		capture := &captureNotifier{}
		svc.notifiers = []AlertNotifier{capture}

		err = svc.DetectAndAlert(context.Background())
		require.NoError(t, err)
		require.Len(t, capture.events, 4)
		assert.Equal(t, AlertTypeFailedLogin, capture.events[0].Type)
		assert.Equal(t, AlertTypePermissionChange, capture.events[1].Type)
		assert.Equal(t, AlertTypePermissionChange, capture.events[2].Type)
		assert.Equal(t, AlertTypeBatchOperation, capture.events[3].Type)
	})
}
