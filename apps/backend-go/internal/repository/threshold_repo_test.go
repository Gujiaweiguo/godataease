package repository

import (
	"context"
	"testing"

	"dataease/backend/internal/domain/auto"
	thresholddomain "dataease/backend/internal/domain/threshold"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupThresholdRepositoryTest(t *testing.T) (*ThresholdRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))

	return NewThresholdRepository(db), db
}

func TestThresholdRepository_CRUDAndSimpleOperations(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupThresholdRepositoryTest(t)

	infoA := &auto.XpackThresholdInfo{ID: 1, Name: "CPU Alert", Enable: true, RateType: 1, RateValue: "5m", ResourceType: "panel", ResourceID: 10, ChartType: "bar", ChartID: 100, MsgTitle: "CPU", MsgType: 1, RepeatSend: true, Status: true, Creator: 1, CreatorName: "alice", CreateTime: 1000, Oid: 1}
	infoB := &auto.XpackThresholdInfo{ID: 2, Name: "Memory Alert", Enable: false, RateType: 2, RateValue: "10m", ResourceType: "dashboard", ResourceID: 11, ChartType: "line", ChartID: 101, MsgTitle: "Memory", MsgType: 1, RepeatSend: false, Status: false, Creator: 2, CreatorName: "bob", CreateTime: 2000, Oid: 1}
	infoC := &auto.XpackThresholdInfo{ID: 3, Name: "Disk Alert", Enable: true, RateType: 1, RateValue: "15m", ResourceType: "panel", ResourceID: 12, ChartType: "pie", ChartID: 102, MsgTitle: "Disk", MsgType: 2, RepeatSend: true, Status: true, Creator: 3, CreatorName: "carol", CreateTime: 3000, Oid: 1}

	require.NoError(t, repo.Create(ctx, infoA))
	require.NoError(t, repo.Create(ctx, infoB))
	require.NoError(t, repo.Create(ctx, infoC))

	found, err := repo.GetByID(ctx, infoA.ID)
	require.NoError(t, err)
	assert.Equal(t, "CPU Alert", found.Name)

	infoA.Name = "CPU Alert Updated"
	require.NoError(t, repo.Update(ctx, infoA))
	updated, err := repo.GetByID(ctx, infoA.ID)
	require.NoError(t, err)
	assert.Equal(t, "CPU Alert Updated", updated.Name)

	require.NoError(t, repo.UpdateEnable(ctx, infoB.ID, true))
	enabled, err := repo.GetByID(ctx, infoB.ID)
	require.NoError(t, err)
	assert.True(t, enabled.Enable)

	require.NoError(t, repo.UpdateRecipients(ctx, []int64{infoA.ID, infoB.ID}, "u1", "r1", "e1", "lg1", "lsg1", "w1"))
	recipientUpdated, err := repo.GetByID(ctx, infoA.ID)
	require.NoError(t, err)
	assert.Equal(t, "u1", recipientUpdated.ReciUsers)
	assert.Equal(t, "w1", recipientUpdated.ReciWebhooks)
	require.NoError(t, repo.UpdateRecipients(ctx, nil, "", "", "", "", "", ""))

	exists, err := repo.ExistsByChartID(ctx, infoC.ChartID)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByChartID(ctx, 9999)
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, repo.DeleteByIDs(ctx, nil))
	require.NoError(t, repo.DeleteByIDs(ctx, []int64{infoC.ID}))
	_, err = repo.GetByID(ctx, infoC.ID)
	require.Error(t, err)

	require.NoError(t, repo.DeleteByChartID(ctx, infoB.ChartID))
	_, err = repo.GetByID(ctx, infoB.ID)
	require.Error(t, err)

	_, err = repo.GetByID(ctx, 999999)
	require.Error(t, err)
}

func TestThresholdRepository_PagerAndHelpers(t *testing.T) {
	ctx := context.Background()
	repo, db := setupThresholdRepositoryTest(t)

	rows := []*auto.XpackThresholdInfo{
		{ID: 1, Name: "CPU Warning", Enable: true, RateType: 1, RateValue: "5m", ResourceType: "panel", ResourceID: 10, ChartType: "bar", ChartID: 100, MsgTitle: "CPU", MsgType: 1, RepeatSend: true, Status: true, Creator: 1, CreatorName: "alice", CreateTime: 1000, Oid: 1},
		{ID: 2, Name: "Memory Warning", Enable: false, RateType: 1, RateValue: "10m", ResourceType: "dashboard", ResourceID: 11, ChartType: "line", ChartID: 101, MsgTitle: "Memory", MsgType: 1, RepeatSend: true, Status: false, Creator: 2, CreatorName: "bob", CreateTime: 2000, Oid: 1},
		{ID: 3, Name: "CPU Critical", Enable: true, RateType: 2, RateValue: "15m", ResourceType: "panel", ResourceID: 12, ChartType: "pie", ChartID: 102, MsgTitle: "Disk", MsgType: 2, RepeatSend: false, Status: true, Creator: 3, CreatorName: "carol", CreateTime: 3000, Oid: 1},
	}
	for _, row := range rows {
		require.NoError(t, db.Create(row).Error)
	}

	chartID := int64(102)
	gridReq := &thresholddomain.GridRequest{
		Keyword:          "CPU",
		ResourceTypeList: []string{"panel", "dashboard"},
		StatusList:       []int{1, 2, 1},
		EnableList:       []int{1, 3},
		ChartID:          &chartID,
	}
	list, total, err := repo.Pager(ctx, gridReq, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	assert.Equal(t, int64(3), list[0].ID)
	assert.Equal(t, "CPU Critical", list[0].Name)
	assert.Equal(t, "carol", list[0].CreateName)

	paged, total, err := repo.Pager(ctx, nil, 2, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, paged, 1)
	assert.Equal(t, int64(2), paged[0].ID)

	assert.Equal(t, []bool{true, false}, intFlagsToBools([]int{1, 0, 1, 2}))
	assert.Nil(t, intFlagsToBools(nil))
	page, size := normalizePage(0, 0)
	assert.Equal(t, 1, page)
	assert.Equal(t, 10, size)
}

func TestThresholdRepository_InstancePager(t *testing.T) {
	ctx := context.Background()
	repo, db := setupThresholdRepositoryTest(t)

	infos := []*auto.XpackThresholdInfo{
		{ID: 1, Name: "CPU Alert", Enable: true, RateType: 1, RateValue: "5m", ResourceType: "panel", ResourceID: 10, ChartType: "bar", ChartID: 100, MsgTitle: "CPU", MsgType: 1, RepeatSend: true, Status: true, Creator: 1, CreatorName: "alice", CreateTime: 1000, Oid: 1},
		{ID: 2, Name: "Memory Alert", Enable: true, RateType: 1, RateValue: "10m", ResourceType: "dashboard", ResourceID: 11, ChartType: "line", ChartID: 101, MsgTitle: "Memory", MsgType: 1, RepeatSend: true, Status: true, Creator: 2, CreatorName: "bob", CreateTime: 2000, Oid: 1},
	}
	for _, info := range infos {
		require.NoError(t, db.Create(info).Error)
	}

	instances := []*auto.XpackThresholdInstance{
		{ID: 1, TaskID: 1, ExecTime: 1000, Status: true, Content: "cpu high", Msg: "ok"},
		{ID: 2, TaskID: 1, ExecTime: 3000, Status: false, Content: "cpu critical", Msg: "failed"},
		{ID: 3, TaskID: 2, ExecTime: 2000, Status: true, Content: "memory high", Msg: "ok"},
	}
	for _, instance := range instances {
		require.NoError(t, db.Create(instance).Error)
	}

	thresholdID := int64(1)
	req := &thresholddomain.InstanceRequest{Keyword: "cpu", ThresholdID: &thresholdID}
	list, total, err := repo.InstancePager(ctx, req, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, list, 2)
	assert.Equal(t, int64(2), list[0].ID)
	assert.Equal(t, "CPU Alert", list[0].Name)
	assert.Equal(t, int64(1), list[1].TaskID)

	paged, total, err := repo.InstancePager(ctx, nil, 2, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, paged, 1)
	assert.Equal(t, int64(3), paged[0].ID)
	assert.Equal(t, "Memory Alert", paged[0].Name)

	missingReq := &thresholddomain.InstanceRequest{Keyword: "missing"}
	empty, total, err := repo.InstancePager(ctx, missingReq, 1, 10)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, empty)
}
