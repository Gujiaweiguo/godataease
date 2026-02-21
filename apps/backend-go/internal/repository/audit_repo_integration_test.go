//go:build integration
// +build integration

package repository

import (
	"fmt"
	"testing"
	"time"

	"dataease/backend/internal/domain/audit"
)

func TestAuditLogRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewAuditLogRepository(testDB)
	cleanupTables("sys_audit_log")

	log := &audit.AuditLog{
		UserID:       1,
		Username:     "testuser",
		ActionType:   "login",
		ResourceType: "user",
		ResourceID:   "1",
		Status:       "success",
		CreateTime:   time.Now().UnixMilli(),
	}

	err := repo.Create(log)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if log.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	found, err := repo.GetByID(log.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", found.Username)
	}
}

func TestAuditLogRepository_CreateBatch(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewAuditLogRepository(testDB)
	cleanupTables("sys_audit_log")

	logs := make([]*audit.AuditLog, 5)
	for i := 0; i < 5; i++ {
		logs[i] = &audit.AuditLog{
			UserID:       1,
			Username:     fmt.Sprintf("user%d", i),
			ActionType:   "create",
			ResourceType: "dashboard",
			CreateTime:   time.Now().UnixMilli(),
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
	cleanupTables("sys_audit_log")

	for i := 0; i < 3; i++ {
		log := &audit.AuditLog{
			UserID:     100,
			Username:   "testuser",
			ActionType: "login",
			CreateTime: time.Now().UnixMilli(),
		}
		_ = repo.Create(log)
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
	cleanupTables("sys_audit_log")

	for i := 0; i < 3; i++ {
		log := &audit.AuditLog{
			UserID:       1,
			Username:     "queryuser",
			ActionType:   "query",
			ResourceType: "dataset",
			CreateTime:   time.Now().UnixMilli(),
		}
		_ = repo.Create(log)
	}

	actionType := "query"
	query := &audit.AuditLogQuery{
		ActionType: &actionType,
		Page:       1,
		PageSize:   10,
	}

	logs, total, err := repo.Query(query)
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
	cleanupTables("sys_audit_log")

	var ids []int64
	for i := 0; i < 3; i++ {
		log := &audit.AuditLog{
			UserID:     1,
			Username:   "test",
			CreateTime: time.Now().UnixMilli(),
		}
		_ = repo.Create(log)
		ids = append(ids, log.ID)
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
	cleanupTables("sys_login_failure")

	failure := &audit.LoginFailure{
		Username:   "faileduser",
		IP:         "192.168.1.1",
		CreateTime: time.Now().UnixMilli(),
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
	cleanupTables("sys_login_failure")

	for i := 0; i < 3; i++ {
		failure := &audit.LoginFailure{
			Username:   "getuser",
			IP:         "192.168.1.1",
			CreateTime: time.Now().UnixMilli(),
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
	cleanupTables("sys_login_failure")

	for i := 0; i < 3; i++ {
		failure := &audit.LoginFailure{
			Username:   "countuser",
			IP:         "192.168.1.1",
			CreateTime: time.Now().UnixMilli(),
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
	cleanupTables("sys_audit_log_detail")

	detail := &audit.AuditLogDetail{
		AuditLogID:   1,
		FieldName:    "status",
		OldValue:     "active",
		NewValue:     "inactive",
		ChangeDetail: "User status changed",
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
	cleanupTables("sys_audit_log_detail")

	detail := &audit.AuditLogDetail{
		AuditLogID: 2,
		FieldName:  "test",
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
