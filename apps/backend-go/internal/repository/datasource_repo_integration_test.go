//go:build integration
// +build integration

package repository

import (
	"fmt"
	"testing"
	"time"

	"dataease/backend/internal/domain/datasource"
)

func TestDatasourceRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	ds := &datasource.CoreDatasource{
		Name:       "Test MySQL",
		Type:       "mysql",
		Status:     strPtr("Success"),
		CreateTime: int64Ptr(time.Now().Unix()),
	}

	err := repo.Create(ds)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if ds.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	found, err := repo.GetByID(ds.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Name != "Test MySQL" {
		t.Errorf("Expected Name 'Test MySQL', got '%s'", found.Name)
	}
}

func TestDatasourceRepository_Query(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	for i := 1; i <= 3; i++ {
		ds := &datasource.CoreDatasource{
			Name:       fmt.Sprintf("Query DB %d", i),
			Type:       "mysql",
			Status:     strPtr("Success"),
			CreateTime: int64Ptr(time.Now().Unix()),
		}
		_ = repo.Create(ds)
	}

	keyword := "Query"
	req := &datasource.ListRequest{
		Keyword: &keyword,
		Current: 1,
		Size:    10,
	}

	list, total, err := repo.Query(req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
	if len(list) != 3 {
		t.Errorf("Expected 3 items, got %d", len(list))
	}
}

func TestDatasourceRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	ds := &datasource.CoreDatasource{
		Name:       "Update DB",
		Type:       "mysql",
		Status:     strPtr("Success"),
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(ds)

	ds.Name = "Updated DB Name"
	err := repo.Update(ds)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(ds.ID)
	if found.Name != "Updated DB Name" {
		t.Errorf("Expected Name 'Updated DB Name', got '%s'", found.Name)
	}
}

func TestDatasourceRepository_SoftDelete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	ds := &datasource.CoreDatasource{
		Name:       "Delete DB",
		Type:       "mysql",
		Status:     strPtr("Success"),
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(ds)

	err := repo.SoftDelete(ds.ID)
	if err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	_, err = repo.GetByID(ds.ID)
	if err == nil {
		t.Error("Expected error when getting deleted datasource")
	}
}

func TestDatasourceRepository_ListChildren(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	parent := &datasource.CoreDatasource{
		Name:       "Parent Folder",
		Type:       "folder",
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(parent)

	for i := 1; i <= 2; i++ {
		child := &datasource.CoreDatasource{
			Name:       fmt.Sprintf("Child %d", i),
			Type:       "mysql",
			PID:        int64Ptr(parent.ID),
			CreateTime: int64Ptr(time.Now().Unix()),
		}
		_ = repo.Create(child)
	}

	children, err := repo.ListChildren(parent.ID)
	if err != nil {
		t.Fatalf("ListChildren failed: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestDatasourceRepository_CountByNameAndPID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	ds := &datasource.CoreDatasource{
		Name:       "Unique Name",
		Type:       "mysql",
		PID:        int64Ptr(0),
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(ds)

	count, err := repo.CountByNameAndPID("Unique Name", 0, nil)
	if err != nil {
		t.Fatalf("CountByNameAndPID failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, _ = repo.CountByNameAndPID("Not Exist", 0, nil)
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestDatasourceRepository_ListAll(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	for i := 1; i <= 3; i++ {
		ds := &datasource.CoreDatasource{
			Name:       fmt.Sprintf("ListAll DB %d", i),
			Type:       "mysql",
			CreateTime: int64Ptr(time.Now().Unix()),
		}
		_ = repo.Create(ds)
	}

	list, err := repo.ListAll(nil)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	if len(list) < 3 {
		t.Errorf("Expected at least 3 items, got %d", len(list))
	}
}

func TestDatasourceRepository_ListByType(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource")

	for i := 1; i <= 2; i++ {
		ds := &datasource.CoreDatasource{
			Name:       fmt.Sprintf("MySQL DB %d", i),
			Type:       "mysql",
			CreateTime: int64Ptr(time.Now().Unix()),
		}
		_ = repo.Create(ds)
	}

	pg := &datasource.CoreDatasource{
		Name:       "PostgreSQL DB",
		Type:       "postgresql",
		CreateTime: int64Ptr(time.Now().Unix()),
	}
	_ = repo.Create(pg)

	list, err := repo.ListByType("mysql", nil)
	if err != nil {
		t.Fatalf("ListByType failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 mysql datasources, got %d", len(list))
	}
}

func TestDatasourceRepository_ListSchemas(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewDatasourceRepository(testDB)

	schemas, err := repo.ListSchemas()
	if err != nil {
		t.Fatalf("ListSchemas failed: %v", err)
	}

	if len(schemas) == 0 {
		t.Error("Expected at least one schema")
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func strPtr(v string) *string {
	return &v
}

func TestDatasourceRepository_CreateSyncTaskLog(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource_task_log")

	now := time.Now().UnixMilli()
	record := &datasource.SyncRecord{
		DsID:        100,
		TaskID:      200,
		StartTime:   now,
		TaskStatus:  "running",
		TableName:   "test_table",
		CreateTime:  now,
		TriggerType: "table",
	}

	err := repo.CreateSyncTaskLog(record)
	if err != nil {
		t.Fatalf("CreateSyncTaskLog failed: %v", err)
	}
	if record.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}
}

func TestDatasourceRepository_ListSyncTaskLogs(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	repo := NewDatasourceRepository(testDB)
	cleanupTables("core_datasource_task_log")

	// 准备测试数据: 15条记录, 验证分页
	for i := 1; i <= 15; i++ {
		now := time.Now().UnixMilli() + int64(i)
		record := &datasource.SyncRecord{
			DsID:        100,
			TaskID:      int64(i),
			StartTime:   now,
			TaskStatus:  "success",
			TableName:   fmt.Sprintf("table_%d", i),
			CreateTime:  now,
			TriggerType: "table",
		}
		_ = repo.CreateSyncTaskLog(record)
	}

	// 第一页
	records, total, err := repo.ListSyncTaskLogs(100, 1, 10)
	if err != nil {
		t.Fatalf("ListSyncTaskLogs failed: %v", err)
	}
	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}
	if len(records) != 10 {
		t.Errorf("Expected 10 records on page 1, got %d", len(records))
	}

	// 第二页
	records2, _, err := repo.ListSyncTaskLogs(100, 2, 10)
	if err != nil {
		t.Fatalf("ListSyncTaskLogs page 2 failed: %v", err)
	}
	if len(records2) != 5 {
		t.Errorf("Expected 5 records on page 2, got %d", len(records2))
	}

	// 验证排序 (start_time DESC)
	if len(records) >= 2 {
		if records[0].StartTime < records[1].StartTime {
			t.Error("Expected records sorted by start_time DESC")
		}
	}

	// 验证过滤 (不同 dsID 返回空)
	otherRecords, otherTotal, _ := repo.ListSyncTaskLogs(999, 1, 10)
	if otherTotal != 0 || len(otherRecords) != 0 {
		t.Errorf("Expected empty result for different dsID, got total=%d, len=%d", otherTotal, len(otherRecords))
	}
}
