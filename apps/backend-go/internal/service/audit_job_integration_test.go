//go:build integration

package service

import (
	"context"
	"testing"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCleanupJob replicates the Run logic from jobs.NewAuditCleanupDefinition
// to avoid an import cycle (internal/job/jobs → internal/service).
func runCleanupJob(ctx context.Context, settingsSvc *SystemParamService, auditSvc *AuditService) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	settings, err := settingsSvc.QueryAuditAlertSettings()
	if err != nil {
		return err
	}
	_, err = auditSvc.DeleteAuditLogsBeforeDate(settings.RetentionDays)
	return err
}

func TestAuditCleanupJobIntegration_DeletesOnlyOldLogs(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	auditLogRepo := repository.NewAuditLogRepository(testDB)
	loginFailureRepo := repository.NewLoginFailureRepository(testDB)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(testDB)
	auditSvc := NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)

	settingsRepo := repository.NewSystemParamRepository(testDB)
	settingsSvc := NewSystemParamService(settingsRepo, auditSvc)

	settings := audit.DefaultAuditAlertSettings()
	settings.RetentionDays = 30
	require.NoError(t, settingsSvc.SaveAuditAlertSettings(settings))

	oldIDs := make([]int64, 3)
	for i := range oldIDs {
		log := &audit.AuditLog{
			ActionType: audit.ActionTypeUserAction,
			ActionName: "old-action",
			Operation:  audit.OperationCreate,
			Status:     audit.StatusSuccess,
		}
		require.NoError(t, testDB.Create(log).Error)
		oldIDs[i] = log.ID
	}
	oldTime := time.Now().AddDate(0, 0, -31).Format("2006-01-02 15:04:05")
	require.NoError(t, testDB.Exec(
		"UPDATE de_audit_log SET create_time = ? WHERE id IN (?, ?, ?)",
		oldTime, oldIDs[0], oldIDs[1], oldIDs[2],
	).Error)

	recentIDs := make([]int64, 2)
	for i := range recentIDs {
		log := &audit.AuditLog{
			ActionType: audit.ActionTypeUserAction,
			ActionName: "recent-action",
			Operation:  audit.OperationCreate,
			Status:     audit.StatusSuccess,
		}
		require.NoError(t, testDB.Create(log).Error)
		recentIDs[i] = log.ID
	}

	err := runCleanupJob(context.Background(), settingsSvc, auditSvc)
	require.NoError(t, err)

	for _, id := range oldIDs {
		var count int64
		testDB.Model(&audit.AuditLog{}).Where("id = ?", id).Count(&count)
		assert.Equal(t, int64(0), count, "old log id=%d should be deleted", id)
	}

	for _, id := range recentIDs {
		var found audit.AuditLog
		require.NoError(t, testDB.First(&found, id).Error)
		assert.Equal(t, "recent-action", found.ActionName)
	}
}

func TestAuditCleanupJobIntegration_CanceledContext(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	auditSvc := NewAuditService(
		repository.NewAuditLogRepository(testDB),
		repository.NewLoginFailureRepository(testDB),
		repository.NewAuditLogDetailRepository(testDB),
	)
	settingsSvc := NewSystemParamService(
		repository.NewSystemParamRepository(testDB), auditSvc,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runCleanupJob(ctx, settingsSvc, auditSvc)
	assert.Error(t, err)
}

func TestAuditCleanupJobIntegration_DefaultRetentionWhenNoSettingsSaved(t *testing.T) {
	cleanupTables(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}, "core_sys_setting")

	auditSvc := NewAuditService(
		repository.NewAuditLogRepository(testDB),
		repository.NewLoginFailureRepository(testDB),
		repository.NewAuditLogDetailRepository(testDB),
	)
	settingsSvc := NewSystemParamService(
		repository.NewSystemParamRepository(testDB), auditSvc,
	)

	log := &audit.AuditLog{
		ActionType: audit.ActionTypeUserAction,
		ActionName: "keep-me",
		Operation:  audit.OperationCreate,
		Status:     audit.StatusSuccess,
	}
	require.NoError(t, testDB.Create(log).Error)

	err := runCleanupJob(context.Background(), settingsSvc, auditSvc)
	require.NoError(t, err)

	var found audit.AuditLog
	require.NoError(t, testDB.First(&found, log.ID).Error)
	assert.Equal(t, "keep-me", found.ActionName)
}
