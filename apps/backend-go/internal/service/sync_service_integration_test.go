//go:build integration

package service

import (
	"strconv"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/syncmodule"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanupSyncTables() {
	testDB.Exec("DELETE FROM core_datasource_task_log")
	testDB.Exec("DELETE FROM core_datasource_task")
	testDB.Exec("DELETE FROM core_datasource")
	testDB.Exec("DELETE FROM core_ds_finish_page")
}

func TestSyncService_TaskLifecycleAndSummary(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	require.NoError(t, testDB.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasourceTask{}, &auto.CoreDatasourceTaskLog{}))
	cleanupSyncTables()

	dsRepo := repository.NewDatasourceRepository(testDB)
	syncRepo := repository.NewSyncRepository(testDB)
	dsSvc := NewDatasourceService(dsRepo)
	svc := NewSyncService(syncRepo, dsRepo, dsSvc)

	now := time.Now().UnixMilli()
	status := datasource.StatusSuccess
	ds := &datasource.CoreDatasource{
		Name:       "sync-source-ds",
		Type:       "mysql",
		Status:     &status,
		CreateTime: &now,
		UpdateTime: &now,
	}
	require.NoError(t, dsRepo.Create(ds))

	addReq := &syncmodule.TaskInfo{
		Name:          "nightly-sync",
		TaskKey:       "sync",
		SchedulerType: "NONE",
		Source: syncmodule.Source{
			DatasourceID: strconv.FormatInt(ds.ID, 10),
			Type:         "mysql",
		},
		Target: syncmodule.Target{
			DatasourceID: strconv.FormatInt(ds.ID, 10),
			Type:         "mysql",
			TableName:    "orders",
		},
	}
	require.NoError(t, svc.AddTask(addReq))

	page, err := svc.TaskPager(1, 10, &syncmodule.TaskGridRequest{Name: "nightly"})
	require.NoError(t, err)
	require.NotNil(t, page)
	require.Len(t, page.Records, 1)
	assert.Equal(t, "nightly-sync", page.Records[0].Name)

	id, err := strconv.ParseInt(page.Records[0].ID, 10, 64)
	require.NoError(t, err)

	detail, err := svc.GetTask(id)
	require.NoError(t, err)
	assert.Equal(t, "nightly-sync", detail.Name)
	assert.Equal(t, strconv.FormatInt(ds.ID, 10), detail.Source.DatasourceID)

	addr, cleanup := startSeatunnelServerForIntegration(t, false)
	defer cleanup()
	dsSvc.SetSeatunnelConfig(addr, 3*time.Second, 0)

	result, err := svc.ExecuteTask(id)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "99001", result["taskId"])

	logs, err := svc.TaskLogPager(1, 10, &syncmodule.TaskLogGridRequest{})
	require.NoError(t, err)
	require.NotNil(t, logs)
	assert.NotEmpty(t, logs.Records)

	counts, err := svc.ResourceCount()
	require.NoError(t, err)
	require.NotNil(t, counts)
	assert.Equal(t, int64(1), counts.JobCount)
	assert.Equal(t, int64(1), counts.DatasourceCount)
	assert.GreaterOrEqual(t, counts.JobLogCount, int64(1))

	chart, err := svc.LogChartData()
	require.NoError(t, err)
	assert.NotNil(t, chart)
	assert.Contains(t, chart, "values")
}
