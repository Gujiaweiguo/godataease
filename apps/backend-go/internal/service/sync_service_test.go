package service

import (
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/syncmodule"
	"dataease/backend/internal/integration/seatunnel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncServiceHelpers_ParseStringID(t *testing.T) {
	id, err := parseStringID("42")
	require.NoError(t, err)
	assert.Equal(t, int64(42), id)

	_, err = parseStringID("")
	assert.EqualError(t, err, "id is required")

	_, err = parseStringID("abc")
	assert.EqualError(t, err, "invalid id")
}

func TestSyncServiceHelpers_ParseMillisAndSchedulerRate(t *testing.T) {
	assert.Equal(t, int64(123456), parseMillis("123456"))
	assert.Equal(t, int64(0), parseMillis("bad"))
	assert.Equal(t, "1", schedulerRate("CRON"))
	assert.Equal(t, "1", schedulerRate("fix_rate"))
	assert.Equal(t, "0", schedulerRate("none"))
}

func TestSyncServiceHelpers_SchedulerTypeAndFirstNonEmpty(t *testing.T) {
	assert.Equal(t, "CRON", schedulerType(auto.CoreDatasourceTask{Cron: "0 0 * * * ?"}))
	assert.Equal(t, "FIX_RATE", schedulerType(auto.CoreDatasourceTask{SyncRate: "1"}))
	assert.Equal(t, "NONE", schedulerType(auto.CoreDatasourceTask{}))
	assert.Equal(t, "b", firstNonEmpty("", "b", "c"))
	assert.Equal(t, "", firstNonEmpty("", " "))
}

func TestSyncServiceHelpers_ValueOrEmptyAndToSyncDatasourceDTO(t *testing.T) {
	assert.Equal(t, "", valueOrEmpty(nil))
	value := "ok"
	assert.Equal(t, "ok", valueOrEmpty(&value))

	desc := "desc"
	config := "{}"
	status := datasource.StatusSuccess
	createTime := int64(1)
	updateTime := int64(2)
	item := &datasource.CoreDatasource{
		ID:            7,
		Name:          "mysql-ds",
		Description:   &desc,
		Type:          "mysql",
		Configuration: &config,
		Status:        &status,
		CreateTime:    &createTime,
		UpdateTime:    &updateTime,
	}
	dto := toSyncDatasourceDTO(item)
	assert.Equal(t, "7", dto.ID)
	assert.Equal(t, "mysql-ds", dto.Name)
	assert.Equal(t, "desc", dto.Desc)
	assert.Equal(t, datasource.StatusSuccess, dto.Status)
}

func TestSyncServiceHelpers_TaskRowAndTaskInfoRoundTrip(t *testing.T) {
	row, err := toTaskRow(&syncmodule.TaskInfo{
		Name:            "sync-job",
		TaskKey:         "sync",
		SchedulerType:   "CRON",
		SchedulerConf:   "0 0 * * * ?",
		SchedulerOption: syncmodule.SchedulerOption{Interval: 5, Unit: "MINUTE"},
		Source:          syncmodule.Source{DatasourceID: "9", Type: "mysql"},
		Target:          syncmodule.Target{DatasourceID: "9", Type: "mysql", TableName: "orders"},
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(9), row.DsID)
	assert.Equal(t, seatunnel.StatusPending, row.TaskStatus)

	info := toTaskInfo(*row)
	assert.Equal(t, "sync-job", info.Name)
	assert.Equal(t, "9", info.Source.DatasourceID)
	assert.Equal(t, "orders", info.Target.TableName)

	_, err = toTaskRow(&syncmodule.TaskInfo{Name: "", Source: syncmodule.Source{DatasourceID: ""}}, nil)
	assert.Error(t, err)
}

func TestSyncServiceHelpers_TaskLogAndMarkTaskExecuted(t *testing.T) {
	logDTO := toTaskLog(auto.CoreDatasourceTaskLog{ID: 1, TaskID: 2, PhysicalTableName: "orders", Info: "done"})
	assert.Equal(t, "1", logDTO.ID)
	assert.Equal(t, "2", logDTO.JobID)
	assert.Equal(t, seatunnel.StatusPending, logDTO.Status)

	svc := &SyncService{}
	require.NoError(t, svc.markTaskExecuted(nil, nil, nil))
}

func TestSyncServiceHelpers_NilAndValidationBranches(t *testing.T) {
	svc := &SyncService{}
	fields, err := svc.GetDatasourceFields(nil)
	require.NoError(t, err)
	assert.NotNil(t, fields)

	err = svc.AddTask(nil)
	assert.EqualError(t, err, "task is required")

	err = svc.UpdateTask(nil)
	assert.EqualError(t, err, "task is required")

	err = svc.ClearTaskLog(&syncmodule.TaskLog{JobID: "bad"})
	assert.EqualError(t, err, "invalid id")
}
