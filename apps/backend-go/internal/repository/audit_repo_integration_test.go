//go:build integration
// +build integration

package repository

import (
	"fmt"
	"testing"
	"time"

	"dataease/backend/internal/domain/audit"
)

func int64PtrAudit(v int64) *int64 { return &v }
func strPtrAudit(v string) *string { return &v }

func TestAuditLogRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewAuditLogRepository(testDB)
	cleanupTables("de_audit_log")

	logEntry := &audit.AuditLog{
		UserID:       int64PtrAudit(1),
		Username:     strPtrAudit("testuser"),
		ActionType:   audit.ActionTypeUserAction,
		ResourceType: strPtrAudit("user"),
		ResourceID:   int64PtrAudit(1),
		Status:       audit.StatusSuccess,
	}

	err := repo.Create(logEntry)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if logEntry.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	found, err := repo.GetByID(logEntry.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Username == nil || *found.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%v'", found.Username)
	}
}

func TestAuditLogRepository_CreateBatch(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewAuditLogRepository(testDB)
	cleanupTables("de_audit_log")

	logs := make([]*audit.AuditLog, 5)
	for i := 0; i < 5; i++ {
		logs[i] = &audit.AuditLog{
			UserID:       int64PtrAudit(1),
			Username:     strPtrAudit(fmt.Sprintf("user%d", i)),
			ActionType:   audit.ActionTypeDataAccess,
			ResourceType: strPtrAudit("dashboard"),
		}
	}

	err := repo.CreateBatch(logs)
	if err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}
}

func TestAuditLogRepository_GetByUserID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewAuditLogRepository(testDB)
	cleanupTables("de_audit_log")

	for i := 0; i < 3; i++ {
		logEntry := &audit.AuditLog{
			UserID:     int64PtrAudit(100),
			Username:   strPtrAudit("testuser"),
			ActionType: audit.ActionTypeUserAction,
		}
		_ = repo.Create(logEntry)
	}

	logs, total, err := repo.GetByUserID(100, 1, 10)
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}
}

func TestAuditLogRepository_Query(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewAuditLogRepository(testDB)
	cleanupTables("de_audit_log")

	for i := 0; i < 3; i++ {
		logEntry := &audit.AuditLog{
			UserID:       int64PtrAudit(1),
			Username:     strPtrAudit("queryuser"),
			ActionType:   audit.ActionTypeDataAccess,
			ResourceType: strPtrAudit("dataset"),
		}
		_ = repo.Create(logEntry)
	}

	actionType := audit.ActionTypeDataAccess
	query := &audit.AuditLogQuery{
		ActionType: &actionType,
		Page:       1,
		PageSize:   10,
	}

	_, total, err := repo.Query(query)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
}

func TestAuditLogRepository_GetByIDs(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewAuditLogRepository(testDB)
	cleanupTables("de_audit_log")

	var ids []int64
	for i := 0; i < 3; i++ {
		logEntry := &audit.AuditLog{
			UserID:   int64PtrAudit(1),
			Username: strPtrAudit("test"),
		}
		_ = repo.Create(logEntry)
		ids = append(ids, logEntry.ID)
	}

	logs, err := repo.GetByIDs(ids)
	if err != nil {
		t.Fatalf("GetByIDs failed: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}
}

func TestLoginFailureRepository_Create(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewLoginFailureRepository(testDB)
	cleanupTables("de_login_failure")

	failure := &audit.LoginFailure{
		Username:  "faileduser",
		IPAddress: strPtrAudit("192.168.1.1"),
	}

	err := repo.Create(failure)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if failure.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}
}

func TestLoginFailureRepository_GetByUsername(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewLoginFailureRepository(testDB)
	cleanupTables("de_login_failure")

	for i := 0; i < 3; i++ {
		failure := &audit.LoginFailure{
			Username:  "getuser",
			IPAddress: strPtrAudit("192.168.1.1"),
		}
		_ = repo.Create(failure)
	}

	failures, err := repo.GetByUsername("getuser", 10)
	if err != nil {
		t.Fatalf("GetByUsername failed: %v", err)
	}

	if len(failures) != 3 {
		t.Errorf("Expected 3 failures, got %d", len(failures))
	}
}

func TestLoginFailureRepository_CountSinceTime(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewLoginFailureRepository(testDB)
	cleanupTables("de_login_failure")

	for i := 0; i < 3; i++ {
		failure := &audit.LoginFailure{
			Username:  "countuser",
			IPAddress: strPtrAudit("192.168.1.1"),
		}
		_ = repo.Create(failure)
	}

	count, err := repo.CountSinceTime("countuser", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("CountSinceTime failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestAuditLogDetailRepository_CreateAndGetByAuditLogID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewAuditLogDetailRepository(testDB)
	cleanupTables("de_audit_log_detail")

	detail := &audit.AuditLogDetail{
		AuditLogID:  1,
		DetailType:  strPtrAudit("field"),
		DetailKey:   strPtrAudit("status"),
		DetailValue: strPtrAudit("changed from active to inactive"),
	}

	err := repo.Create(detail)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	details, err := repo.GetByAuditLogID(1)
	if err != nil {
		t.Fatalf("GetByAuditLogID failed: %v", err)
	}

	if len(details) != 1 {
		t.Errorf("Expected 1 detail, got %d", len(details))
	}
}

func TestAuditLogDetailRepository_DeleteByAuditLogID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewAuditLogDetailRepository(testDB)
	cleanupTables("de_audit_log_detail")

	detail := &audit.AuditLogDetail{
		AuditLogID: 2,
		DetailKey:  strPtrAudit("test"),
	}
	_ = repo.Create(detail)

	err := repo.DeleteByAuditLogID(2)
	if err != nil {
		t.Fatalf("DeleteByAuditLogID failed: %v", err)
	}

	details, _ := repo.GetByAuditLogID(2)
	if len(details) != 0 {
		t.Error("Expected 0 details after delete")
	}
}
