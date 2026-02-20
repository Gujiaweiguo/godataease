package audit

import (
	"testing"
)

func TestAuditConstants(t *testing.T) {
	if ActionTypeUserAction != ActionType("USER_ACTION") {
		t.Errorf("ActionTypeUserAction should be 'USER_ACTION'")
	}
	if ActionTypePermissionChange != ActionType("PERMISSION_CHANGE") {
		t.Errorf("ActionTypePermissionChange should be 'PERMISSION_CHANGE'")
	}
	if ActionTypeDataAccess != ActionType("DATA_ACCESS") {
		t.Errorf("ActionTypeDataAccess should be 'DATA_ACCESS'")
	}
	if ActionTypeSystemConfig != ActionType("SYSTEM_CONFIG") {
		t.Errorf("ActionTypeSystemConfig should be 'SYSTEM_CONFIG'")
	}

	if OperationCreate != Operation("CREATE") {
		t.Errorf("OperationCreate should be 'CREATE'")
	}
	if OperationUpdate != Operation("UPDATE") {
		t.Errorf("OperationUpdate should be 'UPDATE'")
	}
	if OperationDelete != Operation("DELETE") {
		t.Errorf("OperationDelete should be 'DELETE'")
	}

	if StatusSuccess != Status("SUCCESS") {
		t.Errorf("StatusSuccess should be 'SUCCESS'")
	}
	if StatusFailed != Status("FAILED") {
		t.Errorf("StatusFailed should be 'FAILED'")
	}
}

func TestAuditLogTableName(t *testing.T) {
	a := AuditLog{}
	if a.TableName() != "de_audit_log" {
		t.Errorf("TableName should be 'de_audit_log', got '%s'", a.TableName())
	}
}

func TestAuditLogDetailTableName(t *testing.T) {
	ad := AuditLogDetail{}
	if ad.TableName() != "de_audit_log_detail" {
		t.Errorf("TableName should be 'de_audit_log_detail', got '%s'", ad.TableName())
	}
}

func TestLoginFailureTableName(t *testing.T) {
	lf := LoginFailure{}
	if lf.TableName() != "de_login_failure" {
		t.Errorf("TableName should be 'de_login_failure', got '%s'", lf.TableName())
	}
}
