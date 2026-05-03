//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"dataease/backend/internal/domain/auto"
	thresholddomain "dataease/backend/internal/domain/threshold"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestThresholdRepository_CreateAndGet(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	repo := NewThresholdRepository(testDB)
	info := newThresholdInfo(1001, 2001, "threshold-create")
	require.NoError(t, repo.Create(context.Background(), info))

	got, err := repo.GetByID(context.Background(), info.ID)
	require.NoError(t, err)
	assert.Equal(t, info.Name, got.Name)
	assert.Equal(t, info.ChartID, got.ChartID)
	assert.Equal(t, info.ReciUsers, got.ReciUsers)
}

func TestThresholdRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	repo := NewThresholdRepository(testDB)
	info := newThresholdInfo(1002, 2002, "threshold-update")
	require.NoError(t, repo.Create(context.Background(), info))

	info.Name = "threshold-updated"
	require.NoError(t, repo.Update(context.Background(), info))

	got, err := repo.GetByID(context.Background(), info.ID)
	require.NoError(t, err)
	assert.Equal(t, "threshold-updated", got.Name)
}

func TestThresholdRepository_DeleteByIDs(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	repo := NewThresholdRepository(testDB)
	info := newThresholdInfo(1003, 2003, "threshold-delete-ids")
	require.NoError(t, repo.Create(context.Background(), info))

	require.NoError(t, repo.DeleteByIDs(context.Background(), []int64{info.ID}))
	_, err := repo.GetByID(context.Background(), info.ID)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestThresholdRepository_DeleteByChartID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	repo := NewThresholdRepository(testDB)
	info := newThresholdInfo(1004, 2999, "threshold-delete-chart")
	require.NoError(t, repo.Create(context.Background(), info))

	require.NoError(t, repo.DeleteByChartID(context.Background(), 2999))
	_, err := repo.GetByID(context.Background(), info.ID)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestThresholdRepository_Pager(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	repo := NewThresholdRepository(testDB)
	first := newThresholdInfo(1005, 2005, "sales-threshold")
	second := newThresholdInfo(1006, 2006, "profit-threshold")
	second.Enable = false
	require.NoError(t, repo.Create(context.Background(), first))
	require.NoError(t, repo.Create(context.Background(), second))

	rows, total, err := repo.Pager(context.Background(), &thresholddomain.GridRequest{
		Keyword:    "sales",
		EnableList: []int{1},
	}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	assert.Equal(t, first.ID, rows[0].ID)
	assert.Equal(t, first.CreatorName, rows[0].CreateName)
}

func TestThresholdRepository_ExistsByChartID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	repo := NewThresholdRepository(testDB)
	info := newThresholdInfo(1007, 3007, "threshold-exists")
	require.NoError(t, repo.Create(context.Background(), info))

	exists, err := repo.ExistsByChartID(context.Background(), 3007)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByChartID(context.Background(), 3999)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestThresholdRepository_InstancePager(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	repo := NewThresholdRepository(testDB)
	info := newThresholdInfo(1008, 3008, "instance-source-threshold")
	require.NoError(t, repo.Create(context.Background(), info))
	require.NoError(t, testDB.Create(&auto.XpackThresholdInstance{
		ID:       5001,
		TaskID:   info.ID,
		ExecTime: 1700000001,
		Status:   true,
		Content:  "alert content",
		Msg:      "execution message",
	}).Error)

	rows, total, err := repo.InstancePager(context.Background(), &thresholddomain.InstanceRequest{
		Keyword:     "alert",
		ThresholdID: &info.ID,
	}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	assert.Equal(t, info.Name, rows[0].Name)
	assert.Equal(t, "alert content", rows[0].Content)
}

func TestThresholdRepository_UpdateEnable(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	repo := NewThresholdRepository(testDB)
	info := newThresholdInfo(1009, 3009, "threshold-enable")
	require.NoError(t, repo.Create(context.Background(), info))

	require.NoError(t, repo.UpdateEnable(context.Background(), info.ID, false))
	got, err := repo.GetByID(context.Background(), info.ID)
	require.NoError(t, err)
	assert.False(t, got.Enable)
}

func newThresholdInfo(id, chartID int64, name string) *auto.XpackThresholdInfo {
	return &auto.XpackThresholdInfo{
		ID:                  id,
		Name:                name,
		Enable:              true,
		RateType:            1,
		RateValue:           "5",
		ResourceType:        "chart",
		ResourceID:          8001,
		ChartType:           "bar",
		ChartID:             chartID,
		ThresholdRules:      `{"logic":"and","items":[],"children":[]}`,
		Recisetting:         "[1]",
		ReciUsers:           `["user-1"]`,
		ReciRoles:           `["role-1"]`,
		ReciEmails:          `["user@example.com"]`,
		ReciLarkGroups:      `["lark-group"]`,
		ReciLarksuiteGroups: `["suite-group"]`,
		ReciWebhooks:        `["https://example.com/hook"]`,
		MsgTitle:            "threshold title",
		MsgType:             0,
		MsgContent:          "threshold content",
		RepeatSend:          true,
		Status:              true,
		Creator:             9001,
		CreatorName:         "tester",
		CreateTime:          1700000000,
		Oid:                 1,
	}
}
