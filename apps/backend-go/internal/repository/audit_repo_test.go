package repository

import (
	"fmt"
	"testing"
	"time"

	auditdomain "dataease/backend/internal/domain/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuditRepositoryTest(t *testing.T) (*AuditLogRepository, *LoginFailureRepository, *AuditLogDetailRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auditdomain.AuditLog{}, &auditdomain.LoginFailure{}, &auditdomain.AuditLogDetail{}))

	return NewAuditLogRepository(db), NewLoginFailureRepository(db), NewAuditLogDetailRepository(db)
}

func auditInt64Ptr(v int64) *int64    { return &v }
func auditStringPtr(v string) *string { return &v }

func TestAuditLogRepository_CRUDBatchQueryAndDelete(t *testing.T) {
	repo, _, _ := setupAuditRepositoryTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	orgID := int64(9)
	resourceType := string(auditdomain.ResourceTypeDashboard)
	resourceName := "Sales Board"
	ip := "127.0.0.1"
	userAgent := "unit-test"
	beforeValue := "old"
	afterValue := "new"
	failureReason := "denied"

	created := &auditdomain.AuditLog{
		UserID:         auditInt64Ptr(100),
		Username:       auditStringPtr("alpha-user"),
		ActionType:     auditdomain.ActionTypeUserAction,
		ActionName:     "create dashboard",
		ResourceType:   &resourceType,
		ResourceID:     auditInt64Ptr(501),
		ResourceName:   &resourceName,
		Operation:      auditdomain.OperationCreate,
		Status:         auditdomain.StatusSuccess,
		IPAddress:      &ip,
		UserAgent:      &userAgent,
		BeforeValue:    &beforeValue,
		AfterValue:     &afterValue,
		OrganizationID: &orgID,
		CreateTime:     now.Add(-3 * time.Hour),
	}
	require.NoError(t, repo.Create(created))
	require.Positive(t, created.ID)

	batch := []*auditdomain.AuditLog{
		{
			UserID:         auditInt64Ptr(100),
			Username:       auditStringPtr("alpha-user"),
			ActionType:     auditdomain.ActionTypeUserAction,
			ActionName:     "update dashboard",
			ResourceType:   &resourceType,
			ResourceID:     auditInt64Ptr(502),
			Operation:      auditdomain.OperationUpdate,
			Status:         auditdomain.StatusSuccess,
			OrganizationID: &orgID,
			CreateTime:     now.Add(-2 * time.Hour),
		},
		{
			UserID:         auditInt64Ptr(200),
			Username:       auditStringPtr("beta-user"),
			ActionType:     auditdomain.ActionTypeDataAccess,
			ActionName:     "read dataset",
			ResourceType:   auditStringPtr(string(auditdomain.ResourceTypeDataset)),
			ResourceID:     auditInt64Ptr(601),
			Operation:      auditdomain.OperationExport,
			Status:         auditdomain.StatusFailed,
			FailureReason:  &failureReason,
			OrganizationID: auditInt64Ptr(10),
			CreateTime:     now.Add(-time.Hour),
		},
	}
	require.NoError(t, repo.CreateBatch(batch))

	got, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ActionName, got.ActionName)

	byUser, totalByUser, err := repo.GetByUserID(100, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), totalByUser)
	require.Len(t, byUser, 2)
	assert.Equal(t, "update dashboard", byUser[0].ActionName)

	username := "alpha"
	actionType := auditdomain.ActionTypeUserAction
	queryResourceType := auditdomain.ResourceTypeDashboard
	status := auditdomain.StatusSuccess
	startTime := now.Add(-4 * time.Hour)
	endTime := now.Add(-30 * time.Minute)
	query := &auditdomain.AuditLogQuery{
		UserID:         auditInt64Ptr(100),
		Username:       &username,
		ActionType:     &actionType,
		ResourceType:   &queryResourceType,
		OrganizationID: &orgID,
		StartTime:      &startTime,
		EndTime:        &endTime,
		Status:         &status,
		Page:           0,
		PageSize:       500,
	}
	rows, total, err := repo.Query(query)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, rows, 2)
	assert.Equal(t, "update dashboard", rows[0].ActionName)
	assert.Equal(t, "create dashboard", rows[1].ActionName)

	defaultPaged, defaultTotal, err := repo.Query(&auditdomain.AuditLogQuery{Page: -1, PageSize: -1})
	require.NoError(t, err)
	assert.Equal(t, int64(3), defaultTotal)
	require.Len(t, defaultPaged, 3)

	ids := []int64{created.ID, batch[1].ID}
	byIDs, err := repo.GetByIDs(ids)
	require.NoError(t, err)
	require.Len(t, byIDs, 2)

	deleted, err := repo.DeleteBeforeDate(now.Add(-90 * time.Minute))
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)

	remaining, remainingTotal, err := repo.Query(&auditdomain.AuditLogQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), remainingTotal)
	require.Len(t, remaining, 1)
	assert.Equal(t, "read dataset", remaining[0].ActionName)

	_, err = repo.GetByID(999999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestLoginFailureRepository_Queries(t *testing.T) {
	_, repo, _ := setupAuditRepositoryTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	for i, at := range []time.Time{now.Add(-2 * time.Hour), now.Add(-30 * time.Minute), now.Add(-10 * time.Minute)} {
		failure := &auditdomain.LoginFailure{
			Username:      "locked-user",
			IPAddress:     auditStringPtr(fmt.Sprintf("10.0.0.%d", i+1)),
			FailureReason: auditStringPtr("bad password"),
			UserAgent:     auditStringPtr("ua"),
			CreateTime:    at,
		}
		require.NoError(t, repo.Create(failure))
	}
	require.NoError(t, repo.Create(&auditdomain.LoginFailure{Username: "other-user", CreateTime: now.Add(-5 * time.Minute)}))

	latestTwo, err := repo.GetByUsername("locked-user", 2)
	require.NoError(t, err)
	require.Len(t, latestTwo, 2)
	assert.True(t, latestTwo[0].CreateTime.After(latestTwo[1].CreateTime) || latestTwo[0].CreateTime.Equal(latestTwo[1].CreateTime))

	count, err := repo.CountSinceTime("locked-user", now.Add(-45*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	listed, err := repo.ListSinceTime(now.Add(-45 * time.Minute))
	require.NoError(t, err)
	require.Len(t, listed, 3)
	assert.Equal(t, "other-user", listed[0].Username)
	assert.Equal(t, "locked-user", listed[1].Username)
	assert.Equal(t, "locked-user", listed[2].Username)
}

func TestAuditLogDetailRepository_CRUDBatchAndDelete(t *testing.T) {
	_, _, repo := setupAuditRepositoryTest(t)

	detailType := "field"
	keyOne := "status"
	valueOne := "enabled"
	created := &auditdomain.AuditLogDetail{
		AuditLogID:  300,
		DetailType:  &detailType,
		DetailKey:   &keyOne,
		DetailValue: &valueOne,
	}
	require.NoError(t, repo.Create(created))
	require.Positive(t, created.ID)

	batch := []*auditdomain.AuditLogDetail{
		{AuditLogID: 300, DetailType: &detailType, DetailKey: auditStringPtr("owner"), DetailValue: auditStringPtr("alice")},
		{AuditLogID: 301, DetailType: &detailType, DetailKey: auditStringPtr("owner"), DetailValue: auditStringPtr("bob")},
	}
	require.NoError(t, repo.CreateBatch(batch))

	byAuditLog, err := repo.GetByAuditLogID(300)
	require.NoError(t, err)
	require.Len(t, byAuditLog, 2)

	require.NoError(t, repo.DeleteByAuditLogID(300))
	afterDelete, err := repo.GetByAuditLogID(300)
	require.NoError(t, err)
	assert.Empty(t, afterDelete)

	remaining, err := repo.GetByAuditLogID(301)
	require.NoError(t, err)
	require.Len(t, remaining, 1)
	assert.Equal(t, int64(301), remaining[0].AuditLogID)
}
