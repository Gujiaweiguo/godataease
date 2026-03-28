package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuditServiceRepoTest(t *testing.T) (*AuditService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}))

	auditLogRepo := repository.NewAuditLogRepository(db)
	loginFailureRepo := repository.NewLoginFailureRepository(db)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(db)

	return NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo), db
}

func setupClosedAuditServiceRepoTest(t *testing.T) *AuditService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewAuditService(
		repository.NewAuditLogRepository(db),
		repository.NewLoginFailureRepository(db),
		repository.NewAuditLogDetailRepository(db),
	)
}

func TestAuditService_CreateAuditLog(t *testing.T) {
	t.Run("defaults status to success", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)
		userID := int64(7)

		log, err := svc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:     &userID,
			Username:   strPtr("audit-user"),
			ActionType: audit.ActionTypeUserAction,
			ActionName: "Create User",
			Operation:  audit.OperationCreate,
		})
		require.NoError(t, err)
		require.NotNil(t, log)
		assert.Equal(t, audit.StatusSuccess, log.Status)
		assert.Nil(t, log.FailureReason)
		assert.NotZero(t, log.ID)
	})

	t.Run("uses provided status and failure reason", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)
		failed := audit.StatusFailed

		log, err := svc.CreateAuditLog(&audit.AuditLogCreateRequest{
			Username:      strPtr("audit-user"),
			ActionType:    audit.ActionTypePermissionChange,
			ActionName:    "Update Role",
			Operation:     audit.OperationUpdate,
			Status:        &failed,
			FailureReason: strPtr("permission denied"),
		})
		require.NoError(t, err)
		require.NotNil(t, log)
		assert.Equal(t, audit.StatusFailed, log.Status)
		require.NotNil(t, log.FailureReason)
		assert.Equal(t, "permission denied", *log.FailureReason)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedAuditServiceRepoTest(t)

		log, err := svc.CreateAuditLog(&audit.AuditLogCreateRequest{
			ActionType: audit.ActionTypeUserAction,
			ActionName: "Create User",
			Operation:  audit.OperationCreate,
		})
		require.Error(t, err)
		assert.Nil(t, log)
		assert.Contains(t, err.Error(), "failed to create audit log")
	})
}

func TestAuditService_GetAuditLogsByUserID(t *testing.T) {
	t.Run("returns paginated logs", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		userID := int64(8)
		require.NoError(t, db.Create(&audit.AuditLog{UserID: &userID, Username: strPtr("u1"), ActionType: audit.ActionTypeUserAction, ActionName: "A1", Operation: audit.OperationCreate, Status: audit.StatusSuccess}).Error)
		require.NoError(t, db.Create(&audit.AuditLog{UserID: &userID, Username: strPtr("u1"), ActionType: audit.ActionTypeUserAction, ActionName: "A2", Operation: audit.OperationUpdate, Status: audit.StatusSuccess}).Error)

		result, err := svc.GetAuditLogsByUserID(userID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(2), result.Total)
		assert.Equal(t, 1, result.Current)
		assert.Equal(t, 10, result.Size)
		assert.Len(t, result.List.([]*audit.AuditLog), 2)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedAuditServiceRepoTest(t)

		result, err := svc.GetAuditLogsByUserID(8, 1, 10)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get audit logs")
	})

	t.Run("empty result preserves requested pagination", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)

		result, err := svc.GetAuditLogsByUserID(999, 2, 5)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(0), result.Total)
		assert.Equal(t, 2, result.Current)
		assert.Equal(t, 5, result.Size)
		assert.Empty(t, result.List.([]*audit.AuditLog))
	})
}

func TestAuditService_GetAuditLogByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Create(&audit.AuditLog{ID: 201, Username: strPtr("detail-user"), ActionType: audit.ActionTypeUserAction, ActionName: "Detail", Operation: audit.OperationCreate, Status: audit.StatusSuccess}).Error)

		log, err := svc.GetAuditLogByID(201)
		require.NoError(t, err)
		require.NotNil(t, log)
		assert.Equal(t, int64(201), log.ID)
		assert.Equal(t, "Detail", log.ActionName)
	})

	t.Run("not found", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)

		log, err := svc.GetAuditLogByID(999)
		require.Error(t, err)
		assert.Nil(t, log)
	})

	t.Run("repo error", func(t *testing.T) {
		svc := setupClosedAuditServiceRepoTest(t)

		log, err := svc.GetAuditLogByID(1)
		require.Error(t, err)
		assert.Nil(t, log)
	})
}

func TestAuditService_QueryAuditLogs(t *testing.T) {
	t.Run("defaults invalid pagination", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Create(&audit.AuditLog{Username: strPtr("query-user"), ActionType: audit.ActionTypeUserAction, ActionName: "Query", Operation: audit.OperationCreate, Status: audit.StatusSuccess}).Error)

		result, err := svc.QueryAuditLogs(&audit.AuditLogQuery{Page: 0, PageSize: 0})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.Current)
		assert.Equal(t, 20, result.Size)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.List.([]*audit.AuditLog), 1)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedAuditServiceRepoTest(t)

		result, err := svc.QueryAuditLogs(&audit.AuditLogQuery{Page: 1, PageSize: 10})
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to query audit logs")
	})

	t.Run("preserves explicit pagination", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Create(&audit.AuditLog{Username: strPtr("query-user-2"), ActionType: audit.ActionTypeUserAction, ActionName: "Query2", Operation: audit.OperationCreate, Status: audit.StatusSuccess}).Error)

		result, err := svc.QueryAuditLogs(&audit.AuditLogQuery{Page: 3, PageSize: 7})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 3, result.Current)
		assert.Equal(t, 7, result.Size)
		assert.Equal(t, int64(1), result.Total)
	})
}

func TestAuditService_RecordLoginFailure(t *testing.T) {
	t.Run("creates login failure", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)

		failure, err := svc.RecordLoginFailure(&audit.LoginFailureRequest{
			Username:      "bad-user",
			IPAddress:     strPtr("127.0.0.1"),
			FailureReason: strPtr("bad password"),
		})
		require.NoError(t, err)
		require.NotNil(t, failure)
		assert.NotZero(t, failure.ID)
		assert.Equal(t, "bad-user", failure.Username)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedAuditServiceRepoTest(t)

		failure, err := svc.RecordLoginFailure(&audit.LoginFailureRequest{Username: "bad-user"})
		require.Error(t, err)
		assert.Nil(t, failure)
		assert.Contains(t, err.Error(), "failed to record login failure")
	})
}

func TestAuditService_DeleteAuditLogsBeforeDate(t *testing.T) {
	t.Run("uses default retention when days invalid", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		oldTime := time.Now().AddDate(0, 0, -(DefaultRetentionDays + 5))
		recentTime := time.Now().AddDate(0, 0, -5)
		require.NoError(t, db.Create(&audit.AuditLog{Username: strPtr("old"), ActionType: audit.ActionTypeUserAction, ActionName: "Old", Operation: audit.OperationCreate, Status: audit.StatusSuccess, CreateTime: oldTime}).Error)
		require.NoError(t, db.Create(&audit.AuditLog{Username: strPtr("recent"), ActionType: audit.ActionTypeUserAction, ActionName: "Recent", Operation: audit.OperationCreate, Status: audit.StatusSuccess, CreateTime: recentTime}).Error)

		affected, err := svc.DeleteAuditLogsBeforeDate(0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
	})

	t.Run("deletes using custom retention", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		oldTime := time.Now().AddDate(0, 0, -10)
		require.NoError(t, db.Create(&audit.AuditLog{Username: strPtr("custom-old"), ActionType: audit.ActionTypeUserAction, ActionName: "Old", Operation: audit.OperationCreate, Status: audit.StatusSuccess, CreateTime: oldTime}).Error)

		affected, err := svc.DeleteAuditLogsBeforeDate(5)
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedAuditServiceRepoTest(t)

		affected, err := svc.DeleteAuditLogsBeforeDate(5)
		require.Error(t, err)
		assert.Zero(t, affected)
		assert.Contains(t, err.Error(), "failed to delete audit logs")
	})

	t.Run("no matches returns zero", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Create(&audit.AuditLog{Username: strPtr("recent-only"), ActionType: audit.ActionTypeUserAction, ActionName: "Recent", Operation: audit.OperationCreate, Status: audit.StatusSuccess, CreateTime: time.Now()}).Error)

		affected, err := svc.DeleteAuditLogsBeforeDate(5)
		require.NoError(t, err)
		assert.Zero(t, affected)
	})
}

func TestAuditService_CheckSuspiciousLoginActivity(t *testing.T) {
	t.Run("returns true when threshold met", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		now := time.Now()
		for i := 0; i < 3; i++ {
			require.NoError(t, db.Create(&audit.LoginFailure{Username: "suspicious", CreateTime: now.Add(-10 * time.Minute)}).Error)
		}

		suspicious, err := svc.CheckSuspiciousLoginActivity("suspicious", 3, time.Hour)
		require.NoError(t, err)
		assert.True(t, suspicious)
	})

	t.Run("returns false when threshold not met", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		now := time.Now()
		for i := 0; i < 2; i++ {
			require.NoError(t, db.Create(&audit.LoginFailure{Username: "normal", CreateTime: now.Add(-10 * time.Minute)}).Error)
		}

		suspicious, err := svc.CheckSuspiciousLoginActivity("normal", 3, time.Hour)
		require.NoError(t, err)
		assert.False(t, suspicious)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedAuditServiceRepoTest(t)

		suspicious, err := svc.CheckSuspiciousLoginActivity("suspicious", 3, time.Hour)
		require.Error(t, err)
		assert.False(t, suspicious)
	})
}

func TestAuditService_ExportAuditLogs(t *testing.T) {
	t.Run("returns error when no logs found", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)

		path, err := svc.ExportAuditLogs([]int64{999}, "csv")
		require.Error(t, err)
		assert.Empty(t, path)
		assert.Contains(t, err.Error(), "no audit logs found")
	})

	t.Run("exports csv file", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Create(&audit.AuditLog{ID: 101, Username: strPtr("csv-user"), ActionType: audit.ActionTypeUserAction, ActionName: "Export CSV", Operation: audit.OperationExport, Status: audit.StatusSuccess}).Error)

		path, err := svc.ExportAuditLogs([]int64{101}, "csv")
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(path, ".csv"))
		content, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		assert.Contains(t, string(content), "Action Type")
		assert.Contains(t, string(content), "Export CSV")
		require.NoError(t, os.Remove(path))
	})

	t.Run("exports json file", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Create(&audit.AuditLog{ID: 102, Username: strPtr("json-user"), ActionType: audit.ActionTypeUserAction, ActionName: "Export JSON", Operation: audit.OperationExport, Status: audit.StatusSuccess}).Error)

		path, err := svc.ExportAuditLogs([]int64{102}, "json")
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(path, ".json"))
		content, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		assert.Contains(t, string(content), "\n  {")
		assert.Contains(t, string(content), "Export JSON")
		require.NoError(t, os.Remove(path))
	})

	t.Run("defaults unknown format to csv", func(t *testing.T) {
		svc, db := setupAuditServiceRepoTest(t)
		require.NoError(t, db.Create(&audit.AuditLog{ID: 103, Username: strPtr("default-user"), ActionType: audit.ActionTypeUserAction, ActionName: "Default Export", Operation: audit.OperationExport, Status: audit.StatusSuccess}).Error)

		path, err := svc.ExportAuditLogs([]int64{103}, "xlsx")
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(path, ".csv"))
		require.NoError(t, os.Remove(path))
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedAuditServiceRepoTest(t)

		path, err := svc.ExportAuditLogs([]int64{101}, "csv")
		require.Error(t, err)
		assert.Empty(t, path)
		assert.Contains(t, err.Error(), "failed to get audit logs")
	})
}

func TestAuditService_ExportToCSV(t *testing.T) {
	t.Run("writes headers and record", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)
		userID := int64(7)
		resourceID := int64(9)
		filePath := filepath.Join(t.TempDir(), "audit.csv")

		err := svc.exportToCSV([]*audit.AuditLog{{
			ID:           1,
			UserID:       &userID,
			Username:     strPtr("csv-user"),
			ActionType:   audit.ActionTypeUserAction,
			ActionName:   "Write CSV",
			ResourceType: strPtr("USER"),
			ResourceID:   &resourceID,
			Operation:    audit.OperationCreate,
			Status:       audit.StatusSuccess,
			IPAddress:    strPtr("127.0.0.1"),
			CreateTime:   time.Unix(0, 0).UTC(),
		}}, filePath)
		require.NoError(t, err)

		content, readErr := os.ReadFile(filePath)
		require.NoError(t, readErr)
		assert.Contains(t, string(content), "ID,User ID,Username,Action Type")
		assert.Contains(t, string(content), "1,7,csv-user,USER_ACTION,Write CSV,USER,9,CREATE,SUCCESS,127.0.0.1")
	})

	t.Run("handles nil optional fields", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)
		filePath := filepath.Join(t.TempDir(), "audit-empty.csv")

		err := svc.exportToCSV([]*audit.AuditLog{{
			ID:         2,
			ActionType: audit.ActionTypeUserAction,
			ActionName: "Nil Fields",
			Operation:  audit.OperationCreate,
			Status:     audit.StatusSuccess,
			CreateTime: time.Unix(0, 0).UTC(),
		}}, filePath)
		require.NoError(t, err)

		content, readErr := os.ReadFile(filePath)
		require.NoError(t, readErr)
		assert.Contains(t, string(content), "2,,,USER_ACTION,Nil Fields,,,CREATE,SUCCESS,,")
		assert.Contains(t, string(content), "Nil Fields")
	})

	t.Run("returns create file error", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)
		badPath := filepath.Join(t.TempDir(), "missing-dir", "audit.csv")

		err := svc.exportToCSV([]*audit.AuditLog{{ID: 1, ActionType: audit.ActionTypeUserAction, ActionName: "bad", Operation: audit.OperationCreate, Status: audit.StatusSuccess, CreateTime: time.Now()}}, badPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create file")
	})
}

func TestAuditService_ExportToJSON(t *testing.T) {
	t.Run("writes indented json", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)
		filePath := filepath.Join(t.TempDir(), "audit.json")

		err := svc.exportToJSON([]*audit.AuditLog{{ID: 3, ActionType: audit.ActionTypeUserAction, ActionName: "Write JSON", Operation: audit.OperationCreate, Status: audit.StatusSuccess, CreateTime: time.Unix(0, 0).UTC()}}, filePath)
		require.NoError(t, err)

		content, readErr := os.ReadFile(filePath)
		require.NoError(t, readErr)
		assert.Contains(t, string(content), "\n  {")
		assert.Contains(t, string(content), "\"actionName\": \"Write JSON\"")
	})

	t.Run("returns write file error", func(t *testing.T) {
		svc, _ := setupAuditServiceRepoTest(t)
		badPath := filepath.Join(t.TempDir(), "missing-dir", "audit.json")

		err := svc.exportToJSON([]*audit.AuditLog{{ID: 4, ActionType: audit.ActionTypeUserAction, ActionName: "bad", Operation: audit.OperationCreate, Status: audit.StatusSuccess, CreateTime: time.Now()}}, badPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write file")
	})
}
