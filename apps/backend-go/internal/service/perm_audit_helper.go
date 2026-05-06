package service

import (
	"encoding/json"
	"strconv"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/repository"

	"gorm.io/gorm"
)

type permissionMutationAuditor interface {
	RecordPermissionMutationAudit(operation string, scope PermissionMutationScope, targetType string, targetID, permID, resourceID int64, details map[string]interface{}) error
}

type permAuditHelper struct {
	auditSvc *AuditService
}

func newPermAuditHelperFromDB(db *gorm.DB) *permAuditHelper {
	if db == nil {
		return nil
	}
	return &permAuditHelper{auditSvc: NewAuditService(
		repository.NewAuditLogRepository(db),
		repository.NewLoginFailureRepository(db),
		repository.NewAuditLogDetailRepository(db),
	)}
}

func (h *permAuditHelper) RecordPermissionMutationAudit(operation string, scope PermissionMutationScope, targetType string, targetID, permID, resourceID int64, details map[string]interface{}) error {
	if h == nil || h.auditSvc == nil {
		return nil
	}
	beforeValue, _ := json.Marshal(map[string]interface{}{
		"targetType": targetType,
		"targetId":   targetID,
		"permId":     permID,
	})
	afterPayload := map[string]interface{}{
		"operation":  operation,
		"targetType": targetType,
		"targetId":   targetID,
		"permId":     permID,
		"resourceId": resourceID,
	}
	for key, value := range details {
		afterPayload[key] = value
	}
	afterValue, _ := json.Marshal(afterPayload)
	resourceType := string(audit.ResourceTypePermission)
	username := scope.Username
	if username == "" {
		username = strconv.FormatInt(scope.ActorID, 10)
		if scope.ActorID <= 0 {
			username = systemActor
		}
	}
	_, err := h.auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
		UserID:         governanceInt64PtrIfPositive(scope.ActorID),
		Username:       governanceStringPtrIfNotEmpty(username),
		ActionType:     audit.ActionTypeSystemConfig,
		ActionName:     operation,
		ResourceType:   &resourceType,
		ResourceID:     governanceInt64PtrIfPositive(resourceID),
		OrganizationID: governanceInt64PtrIfPositive(scope.OrgID),
		Operation:      audit.OperationUpdate,
		Status:         ptrStatus(audit.StatusSuccess),
		BeforeValue:    governanceStringPtrIfNotEmpty(string(beforeValue)),
		AfterValue:     governanceStringPtrIfNotEmpty(string(afterValue)),
	})
	return err
}
