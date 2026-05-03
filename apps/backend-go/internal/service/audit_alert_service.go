package service

import (
	"context"
	"fmt"
	"time"

	"dataease/backend/internal/domain/audit"
)

type AlertType string

const (
	AlertTypeFailedLogin      AlertType = "failed_login"
	AlertTypePermissionChange AlertType = "permission_change"
	AlertTypeBatchOperation   AlertType = "batch_operation"
)

const auditAlertLookbackWindow = time.Hour

type AlertEvent struct {
	Type       AlertType `json:"type"`
	Username   string    `json:"username"`
	Details    string    `json:"details"`
	DetectedAt time.Time `json:"detectedAt"`
}

type AlertNotifier interface {
	Send(event AlertEvent) error
}

type LogNotifier struct {
	auditService *AuditService
}

func NewLogNotifier(auditService *AuditService) *LogNotifier {
	return &LogNotifier{auditService: auditService}
}

func (n *LogNotifier) Send(event AlertEvent) error {
	if n == nil || n.auditService == nil {
		return fmt.Errorf("audit service is nil")
	}
	actionName := fmt.Sprintf("安全告警: %s", event.Type)
	detail := event.Details
	_, err := n.auditService.CreateAuditLog(&audit.AuditLogCreateRequest{
		Username:   stringPtrOrNil(event.Username),
		ActionType: audit.ActionTypeSystemConfig,
		ActionName: actionName,
		Operation:  audit.OperationCreate,
		AfterValue: &detail,
	})
	return err
}

type AuditAlertService struct {
	settingsSvc *SystemParamService
	auditSvc    *AuditService
	notifiers   []AlertNotifier
}

func NewAuditAlertService(settingsSvc *SystemParamService, auditSvc *AuditService) *AuditAlertService {
	return &AuditAlertService{
		settingsSvc: settingsSvc,
		auditSvc:    auditSvc,
		notifiers:   []AlertNotifier{NewLogNotifier(auditSvc)},
	}
}

func (s *AuditAlertService) Notify(event AlertEvent) error {
	for _, notifier := range s.notifiers {
		if notifier == nil {
			continue
		}
		if err := notifier.Send(event); err != nil {
			return err
		}
	}
	return nil
}

func (s *AuditAlertService) DetectAndAlert(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	settings, err := s.settingsSvc.QueryAuditAlertSettings()
	if err != nil {
		return err
	}
	if !settings.EnableAlerts {
		return nil
	}

	now := time.Now()
	if err := s.checkFailedLogins(ctx, settings, now); err != nil {
		return err
	}
	if settings.AlertOnPermissionChange {
		if err := s.checkPermissionChanges(ctx, now); err != nil {
			return err
		}
	}
	if err := s.checkBatchOperations(ctx, settings, now); err != nil {
		return err
	}

	return nil
}

func (s *AuditAlertService) checkFailedLogins(ctx context.Context, settings *audit.AuditAlertSettings, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.auditSvc == nil || s.auditSvc.loginFailureRepo == nil {
		return nil
	}
	failures, err := s.auditSvc.loginFailureRepo.ListSinceTime(now.Add(-auditAlertLookbackWindow))
	if err != nil {
		return err
	}
	counts := make(map[string]int)
	for _, failure := range failures {
		counts[failure.Username]++
	}
	for username, count := range counts {
		if count < settings.FailedLoginThreshold {
			continue
		}
		event := AlertEvent{
			Type:       AlertTypeFailedLogin,
			Username:   username,
			Details:    fmt.Sprintf("用户 %s 在最近1小时内登录失败 %d 次，超过阈值 %d", username, count, settings.FailedLoginThreshold),
			DetectedAt: now,
		}
		if err := s.Notify(event); err != nil {
			return err
		}
	}
	return nil
}

func (s *AuditAlertService) checkPermissionChanges(ctx context.Context, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	logs, err := s.queryRecentAuditLogs(&audit.AuditLogQuery{
		ActionType: actionTypePtr(audit.ActionTypePermissionChange),
		StartTime:  timePtr(now.Add(-auditAlertLookbackWindow)),
		Page:       1,
		PageSize:   100,
	})
	if err != nil {
		return err
	}
	for _, log := range logs {
		event := AlertEvent{
			Type:       AlertTypePermissionChange,
			Username:   derefString(log.Username),
			Details:    fmt.Sprintf("检测到权限变更操作：%s", log.ActionName),
			DetectedAt: now,
		}
		if err := s.Notify(event); err != nil {
			return err
		}
	}
	return nil
}

func (s *AuditAlertService) checkBatchOperations(ctx context.Context, settings *audit.AuditAlertSettings, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if settings.BatchOperationThreshold <= 0 {
		return nil
	}
	counts := make(map[string]int)
	page := 1
	for {
		logs, total, err := s.queryRecentAuditLogsPage(&audit.AuditLogQuery{
			StartTime: timePtr(now.Add(-auditAlertLookbackWindow)),
			Page:      page,
			PageSize:  100,
		})
		if err != nil {
			return err
		}
		for _, log := range logs {
			username := derefString(log.Username)
			if username == "" {
				continue
			}
			counts[username]++
		}
		if len(logs) == 0 || int64(page*100) >= total {
			break
		}
		page++
	}
	for username, count := range counts {
		if count < settings.BatchOperationThreshold {
			continue
		}
		event := AlertEvent{
			Type:       AlertTypeBatchOperation,
			Username:   username,
			Details:    fmt.Sprintf("用户 %s 在最近1小时内执行了 %d 次操作，超过阈值 %d", username, count, settings.BatchOperationThreshold),
			DetectedAt: now,
		}
		if err := s.Notify(event); err != nil {
			return err
		}
	}
	return nil
}

func (s *AuditAlertService) queryRecentAuditLogs(query *audit.AuditLogQuery) ([]*audit.AuditLog, error) {
	logs, _, err := s.queryRecentAuditLogsPage(query)
	return logs, err
}

func (s *AuditAlertService) queryRecentAuditLogsPage(query *audit.AuditLogQuery) ([]*audit.AuditLog, int64, error) {
	if s == nil || s.auditSvc == nil {
		return nil, 0, nil
	}
	result, err := s.auditSvc.QueryAuditLogs(query)
	if err != nil {
		return nil, 0, err
	}
	logs, _ := result.List.([]*audit.AuditLog)
	return logs, result.Total, nil
}

func actionTypePtr(v audit.ActionType) *audit.ActionType {
	return &v
}

func timePtr(v time.Time) *time.Time {
	return &v
}

func stringPtrOrNil(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
