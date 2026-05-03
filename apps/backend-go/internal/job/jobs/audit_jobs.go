package jobs

import (
	"context"

	"dataease/backend/internal/domain/audit"
	scheduler "dataease/backend/internal/job"
	"dataease/backend/internal/service"
)

const (
	AuditCleanupJobName    = "audit-cleanup"
	AuditCleanupJobSpec    = "0 0 2 * * *"
	AuditAlertCheckJobName = "audit-alert-check"
	AuditAlertCheckJobSpec = "0 */5 * * * *"
)

const (
	auditCleanupDescription    = "cleanup audit logs older than retention period"
	auditAlertCheckDescription = "check suspicious activity and dispatch alerts"
)

type auditSettingsReader interface {
	QueryAuditAlertSettings() (*audit.AuditAlertSettings, error)
}

type auditLogCleaner interface {
	DeleteAuditLogsBeforeDate(days int) (int64, error)
}

type auditAlertDetector interface {
	DetectAndAlert(ctx context.Context) error
}

func NewAuditCleanupDefinition(settingsSvc auditSettingsReader, auditSvc auditLogCleaner) scheduler.Definition {
	return scheduler.Definition{
		Metadata: scheduler.Metadata{
			Key:         AuditCleanupJobName,
			Spec:        AuditCleanupJobSpec,
			Description: auditCleanupDescription,
			Enabled:     true,
			Distributed: true,
		},
		Run: func(ctx context.Context) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			settings, err := settingsSvc.QueryAuditAlertSettings()
			if err != nil {
				return err
			}
			_, err = auditSvc.DeleteAuditLogsBeforeDate(settings.RetentionDays)
			return err
		},
	}
}

func NewAuditAlertCheckDefinition(alertSvc auditAlertDetector) scheduler.Definition {
	return scheduler.Definition{
		Metadata: scheduler.Metadata{
			Key:         AuditAlertCheckJobName,
			Spec:        AuditAlertCheckJobSpec,
			Description: auditAlertCheckDescription,
			Enabled:     true,
			Distributed: true,
		},
		Run: func(ctx context.Context) error {
			return alertSvc.DetectAndAlert(ctx)
		},
	}
}

func NewAuditDefinitions(settingsSvc *service.SystemParamService, auditSvc *service.AuditService, alertSvc *service.AuditAlertService) []scheduler.Definition {
	return []scheduler.Definition{
		NewAuditCleanupDefinition(settingsSvc, auditSvc),
		NewAuditAlertCheckDefinition(alertSvc),
	}
}
