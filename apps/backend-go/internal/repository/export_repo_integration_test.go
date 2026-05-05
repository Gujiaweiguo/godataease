//go:build integration
// +build integration

package repository

import (
	"testing"

	exportdomain "dataease/backend/internal/domain/export"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestExportRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_export_task")

	repo := NewExportRepository(testDB)
	task := &exportdomain.ExportTask{
		ID:                "export-create-1",
		UserID:            1001,
		FileName:          "sales-report.xlsx",
		FileSize:          12.5,
		FileSizeUnit:      "MB",
		ExportFrom:        2001,
		ExportStatus:      "SUCCESS",
		Msg:               "done",
		ExportFromType:    "dashboard",
		ExportTime:        1710000001,
		ExportProgress:    "100",
		ExportMachineName: "worker-1",
	}

	require.NoError(t, repo.Create(task))

	got, err := repo.GetByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, got.ID)
	assert.Equal(t, task.UserID, got.UserID)
	assert.Equal(t, task.FileName, got.FileName)
	assert.Equal(t, task.FileSize, got.FileSize)
	assert.Equal(t, task.FileSizeUnit, got.FileSizeUnit)
	assert.Equal(t, task.ExportFrom, got.ExportFrom)
	assert.Equal(t, task.ExportStatus, got.ExportStatus)
	assert.Equal(t, task.Msg, got.Msg)
	assert.Equal(t, task.ExportFromType, got.ExportFromType)
	assert.Equal(t, task.ExportTime, got.ExportTime)
	assert.Equal(t, task.ExportProgress, got.ExportProgress)
	assert.Equal(t, task.ExportMachineName, got.ExportMachineName)

	var record coreExportTask
	require.NoError(t, testDB.Where("id = ?", task.ID).First(&record).Error)
	assert.Equal(t, "{}", record.Params)
}

func TestExportRepository_List(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_export_task")

	repo := NewExportRepository(testDB)
	tasks := []*exportdomain.ExportTask{
		{ID: "export-list-1", UserID: 1, FileName: "one.xlsx", FileSize: 1, FileSizeUnit: "MB", ExportFrom: 11, ExportStatus: "PENDING", Msg: "queued", ExportFromType: "dashboard", ExportTime: 1710000001, ExportProgress: "0", ExportMachineName: "worker-a"},
		{ID: "export-list-2", UserID: 2, FileName: "two.xlsx", FileSize: 2, FileSizeUnit: "MB", ExportFrom: 12, ExportStatus: "SUCCESS", Msg: "done", ExportFromType: "dashboard", ExportTime: 1710000002, ExportProgress: "100", ExportMachineName: "worker-b"},
		{ID: "export-list-3", UserID: 3, FileName: "three.xlsx", FileSize: 3, FileSizeUnit: "MB", ExportFrom: 13, ExportStatus: "PENDING", Msg: "running", ExportFromType: "dataset", ExportTime: 1710000003, ExportProgress: "50", ExportMachineName: "worker-c"},
	}
	for _, task := range tasks {
		require.NoError(t, repo.Create(task))
	}

	filtered, total, err := repo.List(1, 10, "PENDING")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, filtered, 2)
	assert.Equal(t, "export-list-3", filtered[0].ID)
	assert.Equal(t, "export-list-1", filtered[1].ID)

	paged, totalAll, err := repo.List(1, 2, "all")
	require.NoError(t, err)
	assert.Equal(t, int64(3), totalAll)
	require.Len(t, paged, 2)
	assert.Equal(t, "export-list-3", paged[0].ID)
	assert.Equal(t, "export-list-2", paged[1].ID)

	secondPage, secondTotal, err := repo.List(2, 2, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), secondTotal)
	require.Len(t, secondPage, 1)
	assert.Equal(t, "export-list-1", secondPage[0].ID)
}

func TestExportRepository_UpdateStatus(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_export_task")

	repo := NewExportRepository(testDB)
	task := &exportdomain.ExportTask{
		ID:                "export-update-status-1",
		UserID:            1002,
		FileName:          "update.xlsx",
		FileSize:          9,
		FileSizeUnit:      "MB",
		ExportFrom:        2100,
		ExportStatus:      "PENDING",
		Msg:               "waiting",
		ExportFromType:    "dashboard",
		ExportTime:        1710000010,
		ExportProgress:    "0",
		ExportMachineName: "worker-2",
	}
	require.NoError(t, repo.Create(task))

	require.NoError(t, repo.UpdateStatus(task.ID, "FAILED"))

	got, err := repo.GetByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "FAILED", got.ExportStatus)
}

func TestExportRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_export_task")

	repo := NewExportRepository(testDB)
	task := &exportdomain.ExportTask{ID: "export-delete-1", UserID: 1, FileName: "delete.xlsx", FileSize: 1, FileSizeUnit: "MB", ExportFrom: 1, ExportStatus: "SUCCESS", Msg: "done", ExportFromType: "dashboard", ExportTime: 1710000020, ExportProgress: "100", ExportMachineName: "worker-3"}
	require.NoError(t, repo.Create(task))

	require.NoError(t, repo.Delete(task.ID))

	_, err := repo.GetByID(task.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestExportRepository_DeleteBatch(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_export_task")

	repo := NewExportRepository(testDB)
	tasks := []*exportdomain.ExportTask{
		{ID: "export-batch-1", UserID: 1, FileName: "one.xlsx", FileSize: 1, FileSizeUnit: "MB", ExportFrom: 1, ExportStatus: "SUCCESS", Msg: "done", ExportFromType: "dashboard", ExportTime: 1710000031, ExportProgress: "100", ExportMachineName: "worker-1"},
		{ID: "export-batch-2", UserID: 2, FileName: "two.xlsx", FileSize: 2, FileSizeUnit: "MB", ExportFrom: 2, ExportStatus: "FAILED", Msg: "fail", ExportFromType: "dashboard", ExportTime: 1710000032, ExportProgress: "80", ExportMachineName: "worker-2"},
		{ID: "export-batch-3", UserID: 3, FileName: "three.xlsx", FileSize: 3, FileSizeUnit: "MB", ExportFrom: 3, ExportStatus: "PENDING", Msg: "wait", ExportFromType: "dataset", ExportTime: 1710000033, ExportProgress: "10", ExportMachineName: "worker-3"},
	}
	for _, task := range tasks {
		require.NoError(t, repo.Create(task))
	}

	require.NoError(t, repo.DeleteBatch([]string{"export-batch-1", "export-batch-2"}))
	require.NoError(t, repo.DeleteBatch(nil))

	_, err := repo.GetByID("export-batch-1")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = repo.GetByID("export-batch-2")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	remaining, err := repo.GetByID("export-batch-3")
	require.NoError(t, err)
	assert.Equal(t, "export-batch-3", remaining.ID)
}

func TestExportRepository_DeleteAllByType(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_export_task")

	repo := NewExportRepository(testDB)
	tasks := []*exportdomain.ExportTask{
		{ID: "export-type-1", UserID: 1, FileName: "dashboard-1.xlsx", FileSize: 1, FileSizeUnit: "MB", ExportFrom: 1, ExportStatus: "SUCCESS", Msg: "done", ExportFromType: "dashboard", ExportTime: 1710000041, ExportProgress: "100", ExportMachineName: "worker-1"},
		{ID: "export-type-2", UserID: 2, FileName: "dataset-1.xlsx", FileSize: 2, FileSizeUnit: "MB", ExportFrom: 2, ExportStatus: "SUCCESS", Msg: "done", ExportFromType: "dataset", ExportTime: 1710000042, ExportProgress: "100", ExportMachineName: "worker-2"},
		{ID: "export-type-3", UserID: 3, FileName: "dashboard-2.xlsx", FileSize: 3, FileSizeUnit: "MB", ExportFrom: 3, ExportStatus: "FAILED", Msg: "fail", ExportFromType: "dashboard", ExportTime: 1710000043, ExportProgress: "20", ExportMachineName: "worker-3"},
	}
	for _, task := range tasks {
		require.NoError(t, repo.Create(task))
	}

	require.NoError(t, repo.DeleteAllByType("dashboard"))

	_, err := repo.GetByID("export-type-1")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = repo.GetByID("export-type-3")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	remaining, err := repo.GetByID("export-type-2")
	require.NoError(t, err)
	assert.Equal(t, "dataset", remaining.ExportFromType)
}

func TestExportRepository_CountByStatus(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_export_task")

	repo := NewExportRepository(testDB)
	tasks := []*exportdomain.ExportTask{
		{ID: "export-count-1", UserID: 1, FileName: "one.xlsx", FileSize: 1, FileSizeUnit: "MB", ExportFrom: 1, ExportStatus: "SUCCESS", Msg: "done", ExportFromType: "dashboard", ExportTime: 1710000051, ExportProgress: "100", ExportMachineName: "worker-1"},
		{ID: "export-count-2", UserID: 2, FileName: "two.xlsx", FileSize: 2, FileSizeUnit: "MB", ExportFrom: 2, ExportStatus: "SUCCESS", Msg: "done", ExportFromType: "dashboard", ExportTime: 1710000052, ExportProgress: "100", ExportMachineName: "worker-2"},
		{ID: "export-count-3", UserID: 3, FileName: "three.xlsx", FileSize: 3, FileSizeUnit: "MB", ExportFrom: 3, ExportStatus: "FAILED", Msg: "fail", ExportFromType: "dataset", ExportTime: 1710000053, ExportProgress: "0", ExportMachineName: "worker-3"},
		{ID: "export-count-4", UserID: 4, FileName: "four.xlsx", FileSize: 4, FileSizeUnit: "MB", ExportFrom: 4, ExportStatus: "PENDING", Msg: "wait", ExportFromType: "dataset", ExportTime: 1710000054, ExportProgress: "25", ExportMachineName: "worker-4"},
	}
	for _, task := range tasks {
		require.NoError(t, repo.Create(task))
	}

	counts, err := repo.CountByStatus()
	require.NoError(t, err)
	assert.Equal(t, int64(2), counts["SUCCESS"])
	assert.Equal(t, int64(1), counts["FAILED"])
	assert.Equal(t, int64(1), counts["PENDING"])
}
