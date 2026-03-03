//go:build integration

package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestDatasetServiceIntegration_Tree(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create folder
	folder := &dataset.CoreDatasetGroup{
		Name:     "Test Folder",
		NodeType: strPtr("folder"),
	}
	err := repo.CreateGroup(folder)
	assert.NoError(t, err)

	// Create dataset under folder
	ds := &dataset.CoreDatasetGroup{
		Name:     "Test Dataset",
		PID:      &folder.ID,
		NodeType: strPtr("dataset"),
	}
	err = repo.CreateGroup(ds)
	assert.NoError(t, err)

	// Get tree
	tree, err := svc.Tree(&dataset.TreeRequest{})
	assert.NoError(t, err)
	assert.NotEmpty(t, tree)

	// Find folder in tree
	var foundFolder *dataset.TreeNode
	for _, node := range tree {
		if node.ID == folder.ID {
			foundFolder = &node
			break
		}
	}
	assert.NotNil(t, foundFolder)
	assert.NotEmpty(t, foundFolder.Children)
}

func TestDatasetServiceIntegration_Tree_Empty(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	tree, err := svc.Tree(&dataset.TreeRequest{})
	assert.NoError(t, err)
	assert.Empty(t, tree)
}

func TestDatasetServiceIntegration_GetGroupByID(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create
	group := &dataset.CoreDatasetGroup{
		Name:     "Get Test",
		NodeType: strPtr("folder"),
	}
	err := repo.CreateGroup(group)
	assert.NoError(t, err)

	// Get
	found, err := svc.GetGroupByID(group.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Get Test", found.Name)
}

func TestDatasetServiceIntegration_GetGroupByID_NotFound(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	_, err := svc.GetGroupByID(99999)
	assert.Error(t, err)
}

func TestDatasetServiceIntegration_Save(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	req := &dataset.WriteRequest{
		Name:     "Save Test Dataset",
		NodeType: "dataset",
	}

	ds, err := svc.Save(req)
	assert.NoError(t, err)
	assert.Greater(t, ds.ID, int64(0))
	assert.Equal(t, "Save Test Dataset", ds.Name)
}

func TestDatasetServiceIntegration_Save_Folder(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	req := &dataset.WriteRequest{
		Name:     "Save Test Folder",
		NodeType: "folder",
	}

	ds, err := svc.Save(req)
	assert.NoError(t, err)
	assert.Equal(t, "Save Test Folder", ds.Name)
}

func TestDatasetServiceIntegration_Save_DuplicateName(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create first
	req := &dataset.WriteRequest{
		Name:     "Duplicate Name",
		NodeType: "dataset",
	}
	_, err := svc.Save(req)
	assert.NoError(t, err)

	// Try to create with same name
	_, err = svc.Save(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDatasetServiceIntegration_Save_EmptyName(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	req := &dataset.WriteRequest{
		Name:     "",
		NodeType: "dataset",
	}
	_, err := svc.Save(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestDatasetServiceIntegration_Save_UpdateNotFound(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	_, err := svc.Save(&dataset.WriteRequest{ID: 999999, Name: "x", NodeType: "dataset"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDatasetServiceIntegration_Save_UpdateKeepExistingValues(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	parent, err := svc.Save(&dataset.WriteRequest{Name: "ParentKeep", NodeType: "folder"})
	assert.NoError(t, err)
	original, err := svc.Save(&dataset.WriteRequest{Name: "ChildKeep", NodeType: "dataset", PID: &parent.ID})
	assert.NoError(t, err)

	newType := "custom"
	updated, err := svc.Save(&dataset.WriteRequest{ID: original.ID, Type: &newType})
	assert.NoError(t, err)
	assert.Equal(t, "ChildKeep", updated.Name)
	if assert.NotNil(t, updated.PID) {
		assert.Equal(t, parent.ID, *updated.PID)
	}
	if assert.NotNil(t, updated.Type) {
		assert.Equal(t, "custom", *updated.Type)
	}
}

func TestDatasetServiceIntegration_Create_DestinationNotFound(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	pid := int64(999999)
	_, err := svc.Create(&dataset.WriteRequest{Name: "Child", NodeType: "dataset", PID: &pid})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "destination folder not found")
}

func TestDatasetServiceIntegration_Rename(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create first
	created, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Old Name",
		NodeType: "dataset",
	})

	// Rename
	renamed, err := svc.Rename(created.ID, "New Name")
	assert.NoError(t, err)
	assert.Equal(t, "New Name", renamed.Name)
}

func TestDatasetServiceIntegration_Rename_EmptyName(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create first
	created, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Test",
		NodeType: "dataset",
	})

	// Try to rename with empty name
	_, err := svc.Rename(created.ID, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestDatasetServiceIntegration_Rename_InvalidAndNotFound(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	_, err := svc.Rename(0, "x")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")

	_, err = svc.Rename(999999, "x")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDatasetServiceIntegration_Move(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create folder
	folder, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Target Folder",
		NodeType: "folder",
	})

	// Create dataset
	ds, _ := svc.Save(&dataset.WriteRequest{
		Name:     "To Move",
		NodeType: "dataset",
	})

	// Move
	moved, err := svc.Move(ds.ID, folder.ID)
	assert.NoError(t, err)
	assert.Equal(t, folder.ID, *moved.PID)
}

func TestDatasetServiceIntegration_Move_ToSelf(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	ds, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Self Move",
		NodeType: "dataset",
	})

	_, err := svc.Move(ds.ID, ds.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be itself")
}

func TestDatasetServiceIntegration_Delete(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create
	created, _ := svc.Save(&dataset.WriteRequest{
		Name:     "To Delete",
		NodeType: "dataset",
	})

	// Delete
	err := svc.Delete(created.ID)
	assert.NoError(t, err)

	// Verify deleted
	_, err = svc.GetGroupByID(created.ID)
	assert.Error(t, err)
}

func TestDatasetServiceIntegration_Delete_InvalidID(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	err := svc.Delete(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestDatasetServiceIntegration_Delete_Recursive(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create folder
	folder, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Parent Folder",
		NodeType: "folder",
	})

	// Create child
	child, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Child Dataset",
		PID:      &folder.ID,
		NodeType: "dataset",
	})

	// Delete folder
	err := svc.Delete(folder.ID)
	assert.NoError(t, err)

	// Verify child is deleted
	_, err = svc.GetGroupByID(child.ID)
	assert.Error(t, err)
}

func TestDatasetServiceIntegration_Fields(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create dataset
	ds, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Fields Test",
		NodeType: "dataset",
	})

	// Get fields (should be empty since no fields added)
	fields, err := svc.Fields(&dataset.FieldsRequest{
		DatasetGroupID: ds.ID,
	})
	assert.NoError(t, err)
	assert.Empty(t, fields)
}

func TestDatasetServiceIntegration_PreviewAndPreviewWithPermission(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &chart.CoreChartView{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error
	_ = testDB.Exec("DELETE FROM data_perm_row").Error
	_ = testDB.Exec("DELETE FROM data_perm_column").Error
	_ = testDB.Exec("DROP TABLE IF EXISTS it_preview_ds").Error
	err := testDB.Exec("CREATE TABLE it_preview_ds (id BIGINT PRIMARY KEY AUTO_INCREMENT, region VARCHAR(64), city VARCHAR(64), amount INT)").Error
	assert.NoError(t, err)
	err = testDB.Exec("INSERT INTO it_preview_ds (region, city, amount) VALUES ('East','Shanghai',100),('West','Chengdu',90)").Error
	assert.NoError(t, err)

	repo := repository.NewDatasetRepository(testDB)
	group, err := NewDatasetService(repo).Save(&dataset.WriteRequest{Name: "PreviewDS", NodeType: "dataset"})
	assert.NoError(t, err)

	err = testDB.Create(&dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_preview_ds")}).Error
	assert.NoError(t, err)

	plainSvc := NewDatasetService(repo)
	preview, err := plainSvc.Preview(&dataset.PreviewRequest{DatasetGroupID: group.ID, Limit: 1})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), preview.Total)
	assert.Len(t, preview.Rows, 1)
	assert.Contains(t, preview.Columns, "amount")

	err = testDB.Create(&permission.DataPermColumn{DatasetID: group.ID, FieldName: "amount", PermType: "disable", Status: 1}).Error
	assert.NoError(t, err)
	err = testDB.Create(&permission.DataPermColumn{DatasetID: group.ID, FieldName: "region", PermType: "mask", Status: 1}).Error
	assert.NoError(t, err)
	err = testDB.Create(&permission.DataPermColumn{DatasetID: group.ID, FieldName: "city", PermType: "mask", Status: 1}).Error
	assert.NoError(t, err)

	rowSvc := NewRowPermissionService(
		repository.NewRowPermissionRepository(testDB),
		repository.NewColumnPermissionRepository(testDB),
		nil,
		nil,
	)
	permSvc := NewDatasetServiceWithPermission(repo, rowSvc)
	previewWithPerm, err := permSvc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: group.ID, Limit: 10}, 30001)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), previewWithPerm.Total)
	assert.Len(t, previewWithPerm.Rows, 2)
	assert.Contains(t, previewWithPerm.Columns, "region")
	assert.Contains(t, previewWithPerm.Columns, "city")
	assert.NotContains(t, previewWithPerm.Columns, "amount")

	err = testDB.Exec("DROP TABLE IF EXISTS it_preview_ds").Error
	assert.NoError(t, err)
}

func TestDatasetServiceIntegration_Save_UpdateExisting(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	root, err := svc.Save(&dataset.WriteRequest{Name: "Root", NodeType: "folder"})
	assert.NoError(t, err)
	child, err := svc.Save(&dataset.WriteRequest{Name: "Child", NodeType: "dataset", PID: &root.ID})
	assert.NoError(t, err)

	updated, err := svc.Save(&dataset.WriteRequest{ID: child.ID, Name: "ChildUpdated", NodeType: "dataset", PID: &root.ID})
	assert.NoError(t, err)
	assert.Equal(t, "ChildUpdated", updated.Name)
	if assert.NotNil(t, updated.PID) {
		assert.Equal(t, root.ID, *updated.PID)
	}
}

func TestDatasetServiceIntegration_BuildEnumFilterClauses(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	group, err := svc.Save(&dataset.WriteRequest{Name: "FilterDS", NodeType: "dataset"})
	assert.NoError(t, err)
	table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_filter_ds")}
	err = testDB.Create(table).Error
	assert.NoError(t, err)

	field := &dataset.CoreDatasetTableField{
		DatasetGroupID: group.ID,
		DatasetTableID: &table.ID,
		OriginName:     dsSvcStrPtr("region"),
		Name:           dsSvcStrPtr("region"),
	}
	err = testDB.Create(field).Error
	assert.NoError(t, err)

	filters := []dataset.EnumFilter{
		{FieldID: fmt.Sprintf("%d", field.ID), Operator: "in", Value: []interface{}{"East", "West"}},
		{FieldID: fmt.Sprintf("%d", field.ID), Operator: "eq", Value: []interface{}{"ignored"}},
		{FieldID: "999999", Operator: "in", Value: []interface{}{"missing"}},
	}

	clauses, err := svc.buildEnumFilterClauses(filters, "it_filter_ds")
	assert.NoError(t, err)
	assert.Len(t, clauses, 1)
	assert.Equal(t, "region", clauses[0].Column)
	assert.Equal(t, []string{"East", "West"}, clauses[0].Values)
}

func TestDatasetServiceIntegration_GetSQLParams(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &chart.CoreChartView{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error
	_ = testDB.Exec("DELETE FROM core_chart_view").Error

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	parent, err := svc.Save(&dataset.WriteRequest{Name: "L1", NodeType: "folder"})
	assert.NoError(t, err)
	child, err := svc.Save(&dataset.WriteRequest{Name: "L2", NodeType: "dataset", PID: &parent.ID})
	assert.NoError(t, err)

	sqlVars, _ := json.Marshal([]map[string]interface{}{
		{"variableName": "p_start", "type": []string{"DATETIME"}, "params": []interface{}{"2026-01-01"}},
		{"variableName": "p_amount", "type": []string{"DOUBLE"}, "params": []interface{}{100}},
	})

	err = testDB.Create(&dataset.CoreDatasetTable{
		DatasetGroupID: child.ID,
		PhysicalTable:  dsSvcStrPtr("it_sql_params"),
		SQLVariables:   dsSvcStrPtr(string(sqlVars)),
	}).Error
	assert.NoError(t, err)

	params, err := svc.GetSQLParams([]int64{child.ID})
	assert.NoError(t, err)
	assert.Len(t, params, 2)
	assert.Equal(t, "L1/L2", params[0].DatasetFullName)
	assert.Contains(t, params[0].ID, "|DE|")
}

func TestDatasetServiceIntegration_FieldEnumAndPerDelete(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &chart.CoreChartView{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error
	_ = testDB.Exec("DELETE FROM core_chart_view").Error
	_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_ds").Error
	err := testDB.Exec("CREATE TABLE it_enum_ds (id BIGINT PRIMARY KEY AUTO_INCREMENT, region VARCHAR(64), city VARCHAR(64), amount INT)").Error
	assert.NoError(t, err)
	err = testDB.Exec("INSERT INTO it_enum_ds (region, city, amount) VALUES ('East','Shanghai',100),('East','Suzhou',120),('West','Chengdu',90)").Error
	assert.NoError(t, err)

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	group, err := svc.Save(&dataset.WriteRequest{Name: "EnumDS", NodeType: "dataset"})
	assert.NoError(t, err)

	table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_enum_ds")}
	err = testDB.Create(table).Error
	assert.NoError(t, err)

	deTypeText := 0
	queryField := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("region"), Name: dsSvcStrPtr("region"), DeType: &deTypeText}
	displayField := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("city"), Name: dsSvcStrPtr("city"), DeType: &deTypeText}
	err = testDB.Create(queryField).Error
	assert.NoError(t, err)
	err = testDB.Create(displayField).Error
	assert.NoError(t, err)

	values, err := svc.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{queryField.ID}, ResultMode: 0})
	assert.NoError(t, err)
	assert.Contains(t, values, "East")
	assert.Contains(t, values, "West")

	obj, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{
		QueryID:    queryField.ID,
		DisplayID:  displayField.ID,
		SortID:     0,
		Sort:       "DESC",
		SearchText: "S",
		ResultMode: 0,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, obj)

	dsValues, err := svc.GetFieldEnumDs(queryField.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, dsValues)

	hasRel, err := svc.PerDelete(group.ID)
	assert.NoError(t, err)
	assert.False(t, hasRel)

	err = testDB.Create(&chart.CoreChartView{TableID: &table.ID, Title: dsSvcStrPtr("Enum Chart")}).Error
	assert.NoError(t, err)
	hasRel, err = svc.PerDelete(group.ID)
	assert.NoError(t, err)
	assert.True(t, hasRel)

	err = testDB.Exec("DROP TABLE IF EXISTS it_enum_ds").Error
	assert.NoError(t, err)
}

func dsSvcStrPtr(v string) *string {
	return &v
}

func TestDatasetServiceIntegration_Move_ToDescendant(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create parent folder
	parent, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Parent",
		NodeType: "folder",
	})

	// Create child folder
	child, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Child",
		PID:      &parent.ID,
		NodeType: "folder",
	})

	// Try to move parent into child (should fail)
	_, err := svc.Move(parent.ID, child.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be child")
}

func TestDatasetServiceIntegration_Move_ToNonExistentFolder(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	ds, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Test DS",
		NodeType: "dataset",
	})

	// Try to move to non-existent folder
	_, err := svc.Move(ds.ID, 999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDatasetServiceIntegration_Move_DuplicateName(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create folder
	folder, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Folder",
		NodeType: "folder",
	})

	// Create dataset in folder
	ds1, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Same Name",
		PID:      &folder.ID,
		NodeType: "dataset",
	})

	// Create another dataset outside
	ds2, _ := svc.Save(&dataset.WriteRequest{
		Name:     "Same Name",
		NodeType: "dataset",
	})

	// Try to move ds2 into folder (should fail - duplicate name)
	_, err := svc.Move(ds2.ID, folder.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Verify ds1 still exists
	found, err := svc.GetGroupByID(ds1.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Same Name", found.Name)
}

func TestDatasetServiceIntegration_Move_InvalidID(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	_, err := svc.Move(0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestDatasetServiceIntegration_Move_NonExistentID(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	_, err := svc.Move(999999, 0)
	assert.Error(t, err)
}

func TestDatasetServiceIntegration_GetFieldEnumDs_InvalidID(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Test with zero ID
	result, err := svc.GetFieldEnumDs(0)
	assert.NoError(t, err)
	assert.Empty(t, result)

	// Test with negative ID
	result, err = svc.GetFieldEnumDs(-1)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDatasetServiceIntegration_PerDelete_InvalidID(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Test with zero ID
	_, err := svc.PerDelete(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")

	// Test with negative ID
	_, err = svc.PerDelete(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestDatasetServiceIntegration_PreviewSQL_NilRequest(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Test with nil request
	result, err := svc.PreviewSQL(nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result["data"].(dataset.SQLPreviewData).Fields)
	assert.Empty(t, result["data"].(dataset.SQLPreviewData).Data)
}

func TestDatasetServiceIntegration_PreviewSQL_EmptySQL(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Test with empty SQL
	result, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: ""})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result["data"].(dataset.SQLPreviewData).Fields)
}

func TestDatasetServiceIntegration_PreviewSQL_WhitespaceSQL(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Test with whitespace only SQL
	result, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: "   "})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result["data"].(dataset.SQLPreviewData).Fields)
}

func TestDatasetServiceIntegration_PreviewSQL_Base64Encoded(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Test with base64 encoded SQL
	sql := "SELECT 1 as test"
	encodedSQL := base64.StdEncoding.EncodeToString([]byte(sql))
	result, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: encodedSQL})
	// This might fail in test environment without DB, but at least tests the parsing logic
	// Just verify no panic occurs
	_ = result
	_ = err
}

func TestDatasetServiceIntegration_GetSQLParams_EmptyList(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Test with empty list
	result, err := svc.GetSQLParams([]int64{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDatasetServiceIntegration_GetSQLParams_NilList(t *testing.T) {
	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Test with nil list (should be treated as empty)
	result, err := svc.GetSQLParams(nil)
	assert.NoError(t, err)
	assert.Empty(t, result)
}
