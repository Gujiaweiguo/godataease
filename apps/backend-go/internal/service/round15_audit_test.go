package service

import (
	"encoding/json"
	"testing"

	"dataease/backend/internal/domain/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// RecordPermissionMutationAudit tests (perm_audit_helper.go)
// ---------------------------------------------------------------------------

func TestRound15_Audit_PermAudit_NilReceiver(t *testing.T) {
	var h *permAuditHelper
	err := h.RecordPermissionMutationAudit("grant", PermissionMutationScope{ActorID: 1}, "dataset", 10, 20, 30, nil)
	assert.NoError(t, err, "nil receiver should return nil")
}

func TestRound15_Audit_PermAudit_NilAuditSvc(t *testing.T) {
	h := &permAuditHelper{auditSvc: nil}
	err := h.RecordPermissionMutationAudit("grant", PermissionMutationScope{ActorID: 1}, "dataset", 10, 20, 30, nil)
	assert.NoError(t, err, "nil auditSvc should return nil")
}

func TestRound15_Audit_PermAudit_HappyPath(t *testing.T) {
	auditSvc, db := setupAuditServiceRepoTest(t)

	h := &permAuditHelper{auditSvc: auditSvc}
	scope := PermissionMutationScope{ActorID: 42, OrgID: 7, Username: "alice"}
	err := h.RecordPermissionMutationAudit("grant", scope, "dataset", 100, 200, 300, map[string]interface{}{
		"extraKey": "extraVal",
	})
	require.NoError(t, err)

	var logs []audit.AuditLog
	require.NoError(t, db.Order("id ASC").Find(&logs).Error)
	require.Len(t, logs, 1)

	lo := logs[0]
	assert.Equal(t, "alice", *lo.Username)
	assert.Equal(t, int64(42), *lo.UserID)
	assert.Equal(t, string(audit.ResourceTypePermission), *lo.ResourceType)
	assert.Equal(t, audit.ActionTypeSystemConfig, lo.ActionType)
	assert.Equal(t, "grant", lo.ActionName)
	assert.Equal(t, audit.OperationUpdate, lo.Operation)
	assert.Equal(t, audit.StatusSuccess, lo.Status)
	assert.Equal(t, int64(7), *lo.OrganizationID)

	var before map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(*lo.BeforeValue), &before))
	assert.Equal(t, "dataset", before["targetType"])
	assert.Equal(t, float64(100), before["targetId"])
	assert.Equal(t, float64(200), before["permId"])

	var after map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(*lo.AfterValue), &after))
	assert.Equal(t, "grant", after["operation"])
	assert.Equal(t, "dataset", after["targetType"])
	assert.Equal(t, float64(100), after["targetId"])
	assert.Equal(t, float64(200), after["permId"])
	assert.Equal(t, float64(300), after["resourceId"])
	assert.Equal(t, "extraVal", after["extraKey"])
}

func TestRound15_Audit_PermAudit_SystemActorFallback(t *testing.T) {
	auditSvc, db := setupAuditServiceRepoTest(t)

	h := &permAuditHelper{auditSvc: auditSvc}
	scope := PermissionMutationScope{ActorID: 0, OrgID: 0, Username: ""}
	err := h.RecordPermissionMutationAudit("revoke", scope, "panel", 1, 2, 3, nil)
	require.NoError(t, err)

	var logs []audit.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1)

	assert.Equal(t, systemActor, *logs[0].Username)
	assert.Nil(t, logs[0].UserID)
	assert.Nil(t, logs[0].OrganizationID)
}

func TestRound15_Audit_PermAudit_ActorIDAsUsername(t *testing.T) {
	auditSvc, db := setupAuditServiceRepoTest(t)

	h := &permAuditHelper{auditSvc: auditSvc}
	scope := PermissionMutationScope{ActorID: 99, Username: ""}
	err := h.RecordPermissionMutationAudit("modify", scope, "dataset", 1, 2, 3, nil)
	require.NoError(t, err)

	var logs []audit.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1)

	assert.Equal(t, "99", *logs[0].Username)
	assert.Equal(t, int64(99), *logs[0].UserID)
}

// ---------------------------------------------------------------------------
// recordRoleAssignmentAudit tests (role_service.go)
// ---------------------------------------------------------------------------

func TestRound15_Audit_RoleAssignment_NilGovernanceSvc(t *testing.T) {
	s := &RoleService{governancePolicySvc: nil}
	s.recordRoleAssignmentAudit(1, 2, 3, audit.StatusSuccess, "")
}

func TestRound15_Audit_RoleAssignment_NilAuditSvc(t *testing.T) {
	s := &RoleService{
		governancePolicySvc: &GovernancePolicyService{auditSvc: nil},
	}
	s.recordRoleAssignmentAudit(1, 2, 3, audit.StatusSuccess, "")
}

func TestRound15_Audit_RoleAssignment_HappyPath(t *testing.T) {
	auditSvc, db := setupAuditServiceRepoTest(t)

	s := &RoleService{
		governancePolicySvc: &GovernancePolicyService{auditSvc: auditSvc},
	}
	s.recordRoleAssignmentAudit(10, 20, 30, audit.StatusSuccess, "assigned admin role")

	var logs []audit.AuditLog
	require.NoError(t, db.Order("id ASC").Find(&logs).Error)
	require.Len(t, logs, 1)

	lo := logs[0]
	assert.Equal(t, int64(20), *lo.UserID)
	assert.Equal(t, systemActor, *lo.Username)
	assert.Equal(t, string(audit.ResourceTypeUser), *lo.ResourceType)
	assert.Equal(t, audit.ActionTypeUserAction, lo.ActionType)
	assert.Equal(t, "分配用户角色", lo.ActionName)
	assert.Equal(t, audit.OperationCreate, lo.Operation)
	assert.Equal(t, audit.StatusSuccess, lo.Status)
	assert.Equal(t, int64(10), *lo.OrganizationID)

	var before map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(*lo.BeforeValue), &before))
	assert.Equal(t, float64(20), before["uid"])
	assert.Equal(t, float64(30), before["rid"])
	assert.Equal(t, float64(10), before["orgId"])

	var after map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(*lo.AfterValue), &after))
	assert.Equal(t, "assigned admin role", after["detail"])
}

func TestRound15_Audit_RoleAssignment_FailureStatus(t *testing.T) {
	auditSvc, db := setupAuditServiceRepoTest(t)

	s := &RoleService{
		governancePolicySvc: &GovernancePolicyService{auditSvc: auditSvc},
	}
	s.recordRoleAssignmentAudit(5, 6, 7, audit.StatusFailed, "role not found")

	var logs []audit.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1)

	assert.Equal(t, audit.StatusFailed, logs[0].Status)
	assert.Equal(t, "role not found", *logs[0].FailureReason)
}
