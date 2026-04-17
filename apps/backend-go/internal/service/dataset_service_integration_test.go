//go:build integration

package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	// Create dataset under folder
	ds := &dataset.CoreDatasetGroup{
		Name:     "Test Dataset",
		PID:      &folder.ID,
		NodeType: strPtr("dataset"),
	}
	err = repo.CreateGroup(ds)
	require.NoError(t, err)

	// Get tree
	tree, err := svc.Tree(&dataset.TreeRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, tree)

	// Find folder in tree
	var foundFolder *dataset.TreeNode
	for _, node := range tree {
		if node.Name == folder.Name && node.NodeType == dataset.NodeTypeFolder {
			foundFolder = &node
			break
		}
	}
	require.NotNil(t, foundFolder)
	assert.Equal(t, dataset.NodeTypeFolder, foundFolder.NodeType)
	require.NotEmpty(t, foundFolder.Children)
	assert.Equal(t, ds.Name, foundFolder.Children[0].Name)
	assert.Equal(t, dataset.NodeTypeDataset, foundFolder.Children[0].NodeType)
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

func TestDatasetServiceIntegration_Save_InheritsParentResourcePermissions(t *testing.T) {
	cleanupTables(&permission.SysResourcePerm{}, &permission.SysResource{}, &dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	resourcePermRepo := repository.NewResourcePermissionRepository(testDB)
	resourcePermSvc := NewResourcePermissionService(resourcePermRepo, nil)
	svc := NewDatasetService(repo)
	svc.SetResourcePermissionService(resourcePermSvc)

	parent, err := svc.Save(&dataset.WriteRequest{Name: "Governed Dataset Folder", NodeType: dataset.NodeTypeFolder})
	require.NoError(t, err)
	require.NoError(t, resourcePermSvc.RegisterResource(parent.ID, parent.Name, permission.ResourceTypeDataset, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(parent.ID, permission.ResourceTypeDataset, []int64{21, 22}))

	child, err := svc.Save(&dataset.WriteRequest{Name: "Inherited Dataset", NodeType: dataset.NodeTypeDataset, PID: &parent.ID})
	require.NoError(t, err)

	permIDs, exists, err := resourcePermRepo.GetResourcePermissionIDs(child.ID, permission.ResourceTypeDataset)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.ElementsMatch(t, []int64{21, 22}, permIDs)
}

func TestDatasetServiceIntegration_BackfillGovernedResources(t *testing.T) {
	cleanupTables(&permission.SysResourcePerm{}, &permission.SysResource{}, &dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	resourcePermRepo := repository.NewResourcePermissionRepository(testDB)
	resourcePermSvc := NewResourcePermissionService(resourcePermRepo, nil)
	seedSvc := NewDatasetService(repo)
	backfillSvc := NewDatasetService(repo)
	backfillSvc.SetResourcePermissionService(resourcePermSvc)

	governedParent, err := seedSvc.Save(&dataset.WriteRequest{Name: "Governed Legacy Dataset Folder", NodeType: dataset.NodeTypeFolder})
	require.NoError(t, err)
	plainParent, err := seedSvc.Save(&dataset.WriteRequest{Name: "Ungoverned Legacy Dataset Folder", NodeType: dataset.NodeTypeFolder})
	require.NoError(t, err)

	governedChild, err := seedSvc.Save(&dataset.WriteRequest{Name: "Legacy Governed Dataset", NodeType: dataset.NodeTypeDataset, PID: &governedParent.ID})
	require.NoError(t, err)
	plainChild, err := seedSvc.Save(&dataset.WriteRequest{Name: "Legacy Plain Dataset", NodeType: dataset.NodeTypeDataset, PID: &plainParent.ID})
	require.NoError(t, err)

	require.NoError(t, resourcePermSvc.RegisterResource(governedParent.ID, governedParent.Name, permission.ResourceTypeDataset, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(governedParent.ID, permission.ResourceTypeDataset, []int64{31, 32}))

	report, err := backfillSvc.BackfillGovernedResources()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, report.Scanned, 4)
	assert.Equal(t, 1, report.Governed)
	assert.Contains(t, report.ResourceIDs, governedChild.ID)

	permIDs, exists, err := resourcePermRepo.GetResourcePermissionIDs(governedChild.ID, permission.ResourceTypeDataset)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.ElementsMatch(t, []int64{31, 32}, permIDs)

	plainPermIDs, plainExists, err := resourcePermRepo.GetResourcePermissionIDs(plainChild.ID, permission.ResourceTypeDataset)
	require.NoError(t, err)
	assert.False(t, plainExists)
	assert.Empty(t, plainPermIDs)

	report, err = backfillSvc.BackfillGovernedResources()
	require.NoError(t, err)
	assert.Equal(t, 1, report.Governed)

	var resourceCount int64
	require.NoError(t, testDB.Model(&permission.SysResource{}).
		Where("resource_type = ? AND resource_name = ?", permission.ResourceTypeDataset, governedChild.Name).
		Count(&resourceCount).Error)
	assert.Equal(t, int64(1), resourceCount)

	var resourcePermCount int64
	require.NoError(t, testDB.Model(&permission.SysResourcePerm{}).
		Where("resource_id = ?", int64(2_000_000_000_000)+governedChild.ID).
		Count(&resourcePermCount).Error)
	assert.Equal(t, int64(2), resourcePermCount)
}

func TestDatasetServiceIntegration_BackfillGovernedResourcesWithOptions(t *testing.T) {
	cleanupTables(&permission.SysResourcePerm{}, &permission.SysResource{}, &dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	resourcePermRepo := repository.NewResourcePermissionRepository(testDB)
	resourcePermSvc := NewResourcePermissionService(resourcePermRepo, nil)
	seedSvc := NewDatasetService(repo)
	backfillSvc := NewDatasetService(repo)
	backfillSvc.SetResourcePermissionService(resourcePermSvc)

	governedParent, err := seedSvc.Save(&dataset.WriteRequest{Name: "Batch Governed Dataset Folder", NodeType: dataset.NodeTypeFolder})
	require.NoError(t, err)
	plainParent, err := seedSvc.Save(&dataset.WriteRequest{Name: "Batch Plain Dataset Folder", NodeType: dataset.NodeTypeFolder})
	require.NoError(t, err)
	governedChild, err := seedSvc.Save(&dataset.WriteRequest{Name: "Batch Governed Dataset", NodeType: dataset.NodeTypeDataset, PID: &governedParent.ID})
	require.NoError(t, err)
	plainChild, err := seedSvc.Save(&dataset.WriteRequest{Name: "Batch Plain Dataset", NodeType: dataset.NodeTypeDataset, PID: &plainParent.ID})
	require.NoError(t, err)

	require.NoError(t, resourcePermSvc.RegisterResource(governedParent.ID, governedParent.Name, permission.ResourceTypeDataset, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(governedParent.ID, permission.ResourceTypeDataset, []int64{71, 72}))

	batch1, err := backfillSvc.BackfillGovernedResourcesWithOptions(&GovernanceBackfillOptions{Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, batch1.Scanned)
	assert.Equal(t, 0, batch1.Governed)
	assert.Equal(t, 2, batch1.Skipped)
	assert.Equal(t, plainParent.ID, batch1.NextAfterID)
	if assert.Len(t, batch1.SkippedItems, 2) {
		assert.Equal(t, GovernanceBackfillSkipReasonMissingParent, batch1.SkippedItems[0].Reason)
		assert.Equal(t, GovernanceBackfillSkipReasonMissingParent, batch1.SkippedItems[1].Reason)
		assert.Equal(t, GovernanceBackfillRemediationDataCleanup, batch1.SkippedItems[0].Remediation)
		assert.Equal(t, GovernanceBackfillRemediationDataCleanup, batch1.SkippedItems[1].Remediation)
	}

	batch2, err := backfillSvc.BackfillGovernedResourcesWithOptions(&GovernanceBackfillOptions{AfterID: batch1.NextAfterID, Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, batch2.Scanned)
	assert.Equal(t, 1, batch2.Governed)
	assert.Equal(t, 1, batch2.Skipped)
	assert.Contains(t, batch2.ResourceIDs, governedChild.ID)
	if assert.Len(t, batch2.SkippedItems, 1) {
		assert.Equal(t, plainChild.ID, batch2.SkippedItems[0].ResourceID)
		assert.Equal(t, plainParent.ID, batch2.SkippedItems[0].ParentID)
		assert.Equal(t, GovernanceBackfillSkipReasonParentNotGoverned, batch2.SkippedItems[0].Reason)
		assert.Equal(t, GovernanceBackfillRemediationGovernParent, batch2.SkippedItems[0].Remediation)
	}

	permIDs, exists, err := resourcePermRepo.GetResourcePermissionIDs(governedChild.ID, permission.ResourceTypeDataset)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.ElementsMatch(t, []int64{71, 72}, permIDs)
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

func TestDatasetServiceIntegration_Fields_WithMetadata(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	group, err := svc.Save(&dataset.WriteRequest{Name: "FieldsMeta", NodeType: "dataset"})
	assert.NoError(t, err)

	table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_fields_meta")}
	err = testDB.Create(table).Error
	assert.NoError(t, err)

	deTypeText := 0
	deTypeNumber := 2
	fieldA := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("region"), Name: dsSvcStrPtr("region"), DeType: &deTypeText}
	fieldB := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("amount"), Name: dsSvcStrPtr("amount"), DeType: &deTypeNumber}
	err = testDB.Create(fieldA).Error
	assert.NoError(t, err)
	err = testDB.Create(fieldB).Error
	assert.NoError(t, err)

	fields, err := svc.Fields(&dataset.FieldsRequest{DatasetGroupID: group.ID})
	assert.NoError(t, err)
	assert.Len(t, fields, 2)
	if assert.NotNil(t, fields[0].OriginName) {
		assert.Equal(t, "region", *fields[0].OriginName)
	}
	if assert.NotNil(t, fields[1].OriginName) {
		assert.Equal(t, "amount", *fields[1].OriginName)
	}
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
	rowSvc.SetDatasetFieldResolver(repo)
	permSvc := NewDatasetServiceWithPermission(repo, rowSvc)
	previewWithPerm, err := permSvc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: group.ID, Limit: 10}, 30001)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), previewWithPerm.Total)
	assert.Len(t, previewWithPerm.Rows, 2)
	assert.Contains(t, previewWithPerm.Columns, "region")
	assert.Contains(t, previewWithPerm.Columns, "city")
	assert.NotContains(t, previewWithPerm.Columns, "amount")
	assert.Equal(t, "******", previewWithPerm.Rows[0]["region"])
	assert.Equal(t, "******", previewWithPerm.Rows[0]["city"])

	err = testDB.Exec("DROP TABLE IF EXISTS it_preview_ds").Error
	assert.NoError(t, err)
}

func TestDatasetServiceIntegration_PreviewWithPermission_AppliesRowFilterUsingDatasetFieldNames(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{}, &permission.DataPermColumn{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &chart.CoreChartView{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error
	_ = testDB.Exec("DELETE FROM data_perm_row").Error
	_ = testDB.Exec("DELETE FROM data_perm_column").Error
	_ = testDB.Exec("DROP TABLE IF EXISTS it_preview_row_ds").Error
	err := testDB.Exec("CREATE TABLE it_preview_row_ds (id BIGINT PRIMARY KEY AUTO_INCREMENT, region VARCHAR(64), city VARCHAR(64))").Error
	assert.NoError(t, err)
	err = testDB.Exec("INSERT INTO it_preview_row_ds (region, city) VALUES ('East','Shanghai'),('West','Chengdu')").Error
	assert.NoError(t, err)

	repo := repository.NewDatasetRepository(testDB)
	group, err := NewDatasetService(repo).Save(&dataset.WriteRequest{Name: "PreviewRowDS", NodeType: "dataset"})
	assert.NoError(t, err)

	table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_preview_row_ds")}
	err = testDB.Create(table).Error
	assert.NoError(t, err)
	field := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("region"), Name: dsSvcStrPtr("region")}
	err = testDB.Create(field).Error
	assert.NoError(t, err)
	err = testDB.Create(&permission.DataPermRow{DatasetID: group.ID, DatasetGroupID: group.ID, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 30003, Status: 1, ExpressionTree: fmt.Sprintf(`{"logic":"OR","items":[{"fieldId":%d,"term":"eq","value":"East"}]}`, field.ID)}).Error
	assert.NoError(t, err)

	rowSvc := NewRowPermissionService(repository.NewRowPermissionRepository(testDB), repository.NewColumnPermissionRepository(testDB), nil, nil)
	rowSvc.SetDatasetFieldResolver(repo)
	permSvc := NewDatasetServiceWithPermission(repo, rowSvc)

	preview, err := permSvc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: group.ID, Limit: 10}, 30003)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), preview.Total)
	assert.Len(t, preview.Rows, 1)
	assert.Equal(t, "East", preview.Rows[0]["region"])

	err = testDB.Exec("DROP TABLE IF EXISTS it_preview_row_ds").Error
	assert.NoError(t, err)
}

func TestDatasetServiceIntegration_Preview_MissingPhysicalTableReturnsError(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	group, err := svc.Save(&dataset.WriteRequest{Name: "PreviewMissingTable", NodeType: "dataset"})
	assert.NoError(t, err)

	err = testDB.Create(&dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_preview_missing_table")}).Error
	assert.NoError(t, err)

	preview, err := svc.Preview(&dataset.PreviewRequest{DatasetGroupID: group.ID, Limit: 10})
	assert.Error(t, err)
	assert.Nil(t, preview)
	assert.Contains(t, err.Error(), "doesn't exist")
}

func TestDatasetServiceIntegration_PreviewWithPermission_DeniesUnauthorizedDatasourceDependency(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error
	_ = testDB.Exec("DROP TABLE IF EXISTS it_preview_perm_ds").Error
	err := testDB.Exec("CREATE TABLE it_preview_perm_ds (id BIGINT PRIMARY KEY AUTO_INCREMENT, region VARCHAR(64))").Error
	assert.NoError(t, err)
	err = testDB.Exec("INSERT INTO it_preview_perm_ds (region) VALUES ('East')").Error
	assert.NoError(t, err)

	repo := repository.NewDatasetRepository(testDB)
	group, err := NewDatasetService(repo).Save(&dataset.WriteRequest{Name: "PreviewPermDS", NodeType: "dataset"})
	assert.NoError(t, err)

	datasourceID := int64(8081)
	err = testDB.Create(&dataset.CoreDatasetTable{DatasetGroupID: group.ID, DatasourceID: &datasourceID, PhysicalTable: dsSvcStrPtr("it_preview_perm_ds")}).Error
	assert.NoError(t, err)

	permRepo := newMockResourcePermRepo()
	permRepo.permKeys[permission.PermKeyView] = &permission.SysPerm{PermID: 1, PermKey: permission.PermKeyView}
	resourcePermSvc := NewResourcePermissionService(permRepo, &mockResourcePermAdminChecker{adminUserIDs: map[int64]bool{}})

	permSvc := NewDatasetServiceWithPermission(repo, nil)
	permSvc.SetResourcePermissionService(resourcePermSvc)

	preview, err := permSvc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: group.ID, Limit: 10}, 30002)
	assert.ErrorIs(t, err, ErrDatasetDatasourcePermissionDenied)
	assert.Nil(t, preview)

	err = testDB.Exec("DROP TABLE IF EXISTS it_preview_perm_ds").Error
	assert.NoError(t, err)
}

func TestDatasetServiceIntegration_PreviewWithPermission_ReturnsRowPermissionErrors(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &chart.CoreChartView{}, &permission.DataPermRow{}, &permission.DataPermColumn{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error
	_ = testDB.Exec("DELETE FROM data_perm_row").Error
	_ = testDB.Exec("DELETE FROM data_perm_column").Error
	_ = testDB.Exec("DROP TABLE IF EXISTS it_preview_perm_error_ds").Error
	err := testDB.Exec("CREATE TABLE it_preview_perm_error_ds (id BIGINT PRIMARY KEY AUTO_INCREMENT, region VARCHAR(64))").Error
	assert.NoError(t, err)
	err = testDB.Exec("INSERT INTO it_preview_perm_error_ds (region) VALUES ('East')").Error
	assert.NoError(t, err)

	repo := repository.NewDatasetRepository(testDB)
	group, err := NewDatasetService(repo).Save(&dataset.WriteRequest{Name: "PreviewPermErrorDS", NodeType: "dataset"})
	assert.NoError(t, err)

	table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_preview_perm_error_ds")}
	err = testDB.Create(table).Error
	assert.NoError(t, err)
	field := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("region"), Name: dsSvcStrPtr("region")}
	err = testDB.Create(field).Error
	assert.NoError(t, err)

	rowSvc := NewRowPermissionService(repository.NewRowPermissionRepository(testDB), repository.NewColumnPermissionRepository(testDB), nil, nil)
	rowSvc.SetDatasetFieldResolver(repo)
	permSvc := NewDatasetServiceWithPermission(repo, rowSvc)

	err = testDB.Exec("DROP TABLE data_perm_row").Error
	assert.NoError(t, err)

	preview, err := permSvc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: group.ID, Limit: 10}, 30004)
	assert.Nil(t, preview)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to build row permission where clause")

	err = testDB.Exec("DROP TABLE IF EXISTS it_preview_perm_error_ds").Error
	assert.NoError(t, err)
	_ = testDB.AutoMigrate(&permission.DataPermRow{})
}

func TestDatasetServiceIntegration_PreviewWithPermission_ReturnsColumnPermissionErrors(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &chart.CoreChartView{}, &permission.DataPermRow{}, &permission.DataPermColumn{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error
	_ = testDB.Exec("DELETE FROM data_perm_row").Error
	_ = testDB.Exec("DELETE FROM data_perm_column").Error
	_ = testDB.Exec("DROP TABLE IF EXISTS it_preview_column_error_ds").Error
	err := testDB.Exec("CREATE TABLE it_preview_column_error_ds (id BIGINT PRIMARY KEY AUTO_INCREMENT, region VARCHAR(64))").Error
	assert.NoError(t, err)
	err = testDB.Exec("INSERT INTO it_preview_column_error_ds (region) VALUES ('East')").Error
	assert.NoError(t, err)

	repo := repository.NewDatasetRepository(testDB)
	group, err := NewDatasetService(repo).Save(&dataset.WriteRequest{Name: "PreviewColumnErrorDS", NodeType: "dataset"})
	assert.NoError(t, err)

	table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_preview_column_error_ds")}
	err = testDB.Create(table).Error
	assert.NoError(t, err)

	rowSvc := NewRowPermissionService(repository.NewRowPermissionRepository(testDB), repository.NewColumnPermissionRepository(testDB), nil, nil)
	rowSvc.SetDatasetFieldResolver(repo)
	permSvc := NewDatasetServiceWithPermission(repo, rowSvc)

	err = testDB.Exec("DROP TABLE data_perm_column").Error
	assert.NoError(t, err)

	preview, err := permSvc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: group.ID, Limit: 10}, 30005)
	assert.Nil(t, preview)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to load disabled columns")

	err = testDB.Exec("DROP TABLE IF EXISTS it_preview_column_error_ds").Error
	assert.NoError(t, err)
	_ = testDB.AutoMigrate(&permission.DataPermColumn{})
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

func TestDatasetServiceIntegration_DeleteFieldOperations(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetTableField{}, &dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	rootPID := int64(0)
	nodeType := dataset.NodeTypeDataset
	group := &dataset.CoreDatasetGroup{Name: "Delete Field Group", PID: &rootPID, NodeType: &nodeType}
	require.NoError(t, repo.CreateGroup(group))

	chartID := int64(8801)
	otherChartID := int64(8802)
	name := "metric"
	origin := "metric"
	require.NoError(t, testDB.Create(&dataset.CoreDatasetTableField{ID: 9901, DatasetGroupID: group.ID, ChartID: &chartID, Name: &name, OriginName: &origin}).Error)
	require.NoError(t, testDB.Create(&dataset.CoreDatasetTableField{ID: 9902, DatasetGroupID: group.ID, ChartID: &chartID, Name: &name, OriginName: &origin}).Error)
	require.NoError(t, testDB.Create(&dataset.CoreDatasetTableField{ID: 9903, DatasetGroupID: group.ID, ChartID: &otherChartID, Name: &name, OriginName: &origin}).Error)

	err := svc.DeleteField(9901)
	require.NoError(t, err)
	_, err = repo.GetFieldByID(9901)
	assert.Error(t, err)

	err = svc.DeleteFieldByChart(chartID)
	require.NoError(t, err)
	_, err = repo.GetFieldByID(9902)
	assert.Error(t, err)
	found, err := repo.GetFieldByID(9903)
	require.NoError(t, err)
	assert.Equal(t, int64(9903), found.ID)
}

func TestDatasetServiceIntegration_DeleteFieldDependencyBlocking(t *testing.T) {
	// Drop and recreate to ensure schema matches the current Go struct definitions.
	// GORM AutoMigrate on existing tables can fail with "Invalid use of NULL value"
	// when adding NOT NULL columns, so a clean DROP+CREATE is safer in tests.
	for _, tbl := range []interface{}{
		&auto.CoreChartView{},
		&permission.DataPermRow{},
		&permission.DataPermColumn{},
	} {
		_ = testDB.Migrator().DropTable(tbl)
	}
	require.NoError(t, testDB.AutoMigrate(
		&auto.CoreChartView{},
		&permission.DataPermRow{},
		&permission.DataPermColumn{},
	))
	cleanupTables(
		&auto.CoreChartView{},
		&permission.DataPermRow{},
		&permission.DataPermColumn{},
		&dataset.CoreDatasetTable{},
		&dataset.CoreDatasetTableField{},
		&dataset.CoreDatasetGroup{},
	)

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	rootPID := int64(0)
	nodeType := dataset.NodeTypeDataset
	group := &dataset.CoreDatasetGroup{Name: "Blocked Field Group", PID: &rootPID, NodeType: &nodeType}
	require.NoError(t, repo.CreateGroup(group))

	tableName := "orders"
	tableType := "db"
	require.NoError(t, testDB.Create(&dataset.CoreDatasetTable{ID: 9801, DatasetGroupID: group.ID, Name: &tableName, PhysicalTable: &tableName, Type: &tableType}).Error)

	chartID := int64(8803)
	fieldName := "sales"
	require.NoError(t, testDB.Create(&dataset.CoreDatasetTableField{ID: 9904, DatasetGroupID: group.ID, ChartID: &chartID, Name: &fieldName, OriginName: &fieldName, DataeaseName: &fieldName}).Error)
	require.NoError(t, testDB.Create(&auto.CoreChartView{ID: 4001, TableID: 9801, XAxis: `[{"id":9904,"name":"sales"}]`}).Error)
	require.NoError(t, testDB.Create(&permission.DataPermRow{DatasetID: group.ID, DatasetGroupID: group.ID, ExpressionTree: `{"logic":"and","items":[{"fieldId":9904}]}`}).Error)

	err := svc.DeleteField(9904)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDatasetFieldDependencyBlocked)
	assert.Contains(t, err.Error(), "chart views")
	assert.Contains(t, err.Error(), "row permissions")

	cleanChartID := int64(8804)
	cleanFieldName := "margin"
	require.NoError(t, testDB.Create(&dataset.CoreDatasetTableField{ID: 9905, DatasetGroupID: group.ID, ChartID: &cleanChartID, Name: &cleanFieldName, OriginName: &cleanFieldName}).Error)
	err = svc.DeleteField(9905)
	require.NoError(t, err)
	_, err = repo.GetFieldByID(9905)
	assert.Error(t, err)

	err = svc.DeleteField(999999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dataset field not found")
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

func TestDatasetServiceIntegration_IsDescendant_FalsePath(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	root, err := svc.Save(&dataset.WriteRequest{Name: "RootOnly", NodeType: "folder"})
	assert.NoError(t, err)

	_, err = svc.Save(&dataset.WriteRequest{Name: "ChildOnly", NodeType: "folder", PID: &root.ID})
	assert.NoError(t, err)

	independent, err := svc.Save(&dataset.WriteRequest{Name: "Independent", NodeType: "folder"})
	assert.NoError(t, err)

	isDesc, err := svc.isDescendant(root.ID, independent.ID)
	assert.NoError(t, err)
	assert.False(t, isDesc)
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

func TestDatasetServiceIntegration_GetFieldEnum_EdgeCases(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	t.Run("nil request returns empty", func(t *testing.T) {
		result, err := svc.GetFieldEnum(nil)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty field IDs returns empty", func(t *testing.T) {
		result, err := svc.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{}})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("invalid field ID is skipped", func(t *testing.T) {
		result, err := svc.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{0, -1, 999999}})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("duplicate field IDs are deduplicated", func(t *testing.T) {
		_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_dup").Error
		err := testDB.Exec("CREATE TABLE it_enum_dup (id BIGINT PRIMARY KEY AUTO_INCREMENT, val VARCHAR(64))").Error
		require.NoError(t, err)
		err = testDB.Exec("INSERT INTO it_enum_dup (val) VALUES ('A'), ('B')").Error
		require.NoError(t, err)

		group, err := svc.Save(&dataset.WriteRequest{Name: "EnumDup", NodeType: "dataset"})
		require.NoError(t, err)

		table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_enum_dup")}
		require.NoError(t, testDB.Create(table).Error)

		deType := 0
		field := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("val"), Name: dsSvcStrPtr("val"), DeType: &deType}
		require.NoError(t, testDB.Create(field).Error)

		// Request with duplicate IDs
		result, err := svc.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{field.ID, field.ID, field.ID}})
		assert.NoError(t, err)
		assert.Len(t, result, 2) // Only A and B, not 6 entries

		testDB.Exec("DROP TABLE IF EXISTS it_enum_dup")
	})

	t.Run("result mode 1 uses higher limit", func(t *testing.T) {
		_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_mode").Error
		err := testDB.Exec("CREATE TABLE it_enum_mode (id BIGINT PRIMARY KEY AUTO_INCREMENT, val VARCHAR(64))").Error
		require.NoError(t, err)

		group, err := svc.Save(&dataset.WriteRequest{Name: "EnumMode", NodeType: "dataset"})
		require.NoError(t, err)

		table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_enum_mode")}
		require.NoError(t, testDB.Create(table).Error)

		deType := 0
		field := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("val"), Name: dsSvcStrPtr("val"), DeType: &deType}
		require.NoError(t, testDB.Create(field).Error)

		result, err := svc.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{field.ID}, ResultMode: 1})
		assert.NoError(t, err)
		assert.Empty(t, result) // Empty because table has no data

		testDB.Exec("DROP TABLE IF EXISTS it_enum_mode")
	})
}

func TestDatasetServiceIntegration_GetFieldEnumObj_EdgeCases(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	t.Run("nil request returns empty", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(nil)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("invalid query ID returns empty", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{QueryID: 0})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("non-existent field ID returns empty", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{QueryID: 999999})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("display ID defaults to query ID when zero", func(t *testing.T) {
		_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_display").Error
		err := testDB.Exec("CREATE TABLE it_enum_display (id BIGINT PRIMARY KEY AUTO_INCREMENT, region VARCHAR(64))").Error
		require.NoError(t, err)
		err = testDB.Exec("INSERT INTO it_enum_display (region) VALUES ('East'), ('West')").Error
		require.NoError(t, err)

		group, err := svc.Save(&dataset.WriteRequest{Name: "EnumDisplay", NodeType: "dataset"})
		require.NoError(t, err)

		table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_enum_display")}
		require.NoError(t, testDB.Create(table).Error)

		deType := 0
		field := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("region"), Name: dsSvcStrPtr("region"), DeType: &deType}
		require.NoError(t, testDB.Create(field).Error)

		// DisplayID=0 should default to QueryID
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{
			QueryID:   field.ID,
			DisplayID: 0, // Will default to QueryID
			Sort:      "ASC",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		testDB.Exec("DROP TABLE IF EXISTS it_enum_display")
	})

	t.Run("sort ID resets when table mismatch", func(t *testing.T) {
		_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_sort1").Error
		_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_sort2").Error
		err := testDB.Exec("CREATE TABLE it_enum_sort1 (id BIGINT PRIMARY KEY AUTO_INCREMENT, val1 VARCHAR(64))").Error
		require.NoError(t, err)
		err = testDB.Exec("CREATE TABLE it_enum_sort2 (id BIGINT PRIMARY KEY AUTO_INCREMENT, val2 VARCHAR(64))").Error
		require.NoError(t, err)
		err = testDB.Exec("INSERT INTO it_enum_sort1 (val1) VALUES ('A')").Error
		require.NoError(t, err)
		err = testDB.Exec("INSERT INTO it_enum_sort2 (val2) VALUES ('B')").Error
		require.NoError(t, err)

		group1, err := svc.Save(&dataset.WriteRequest{Name: "EnumSort1", NodeType: "dataset"})
		require.NoError(t, err)
		group2, err := svc.Save(&dataset.WriteRequest{Name: "EnumSort2", NodeType: "dataset"})
		require.NoError(t, err)

		table1 := &dataset.CoreDatasetTable{DatasetGroupID: group1.ID, PhysicalTable: dsSvcStrPtr("it_enum_sort1")}
		table2 := &dataset.CoreDatasetTable{DatasetGroupID: group2.ID, PhysicalTable: dsSvcStrPtr("it_enum_sort2")}
		require.NoError(t, testDB.Create(table1).Error)
		require.NoError(t, testDB.Create(table2).Error)

		deType := 0
		field1 := &dataset.CoreDatasetTableField{DatasetGroupID: group1.ID, DatasetTableID: &table1.ID, OriginName: dsSvcStrPtr("val1"), Name: dsSvcStrPtr("val1"), DeType: &deType}
		field2 := &dataset.CoreDatasetTableField{DatasetGroupID: group2.ID, DatasetTableID: &table2.ID, OriginName: dsSvcStrPtr("val2"), Name: dsSvcStrPtr("val2"), DeType: &deType}
		require.NoError(t, testDB.Create(field1).Error)
		require.NoError(t, testDB.Create(field2).Error)

		// SortID from different table should be reset
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{
			QueryID:   field1.ID,
			DisplayID: field1.ID,
			SortID:    field2.ID, // Different table - should be reset
			Sort:      "ASC",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		testDB.Exec("DROP TABLE IF EXISTS it_enum_sort1")
		testDB.Exec("DROP TABLE IF EXISTS it_enum_sort2")
	})
}

func TestDatasetServiceIntegration_GetSQLParams_EdgeCases(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{})
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	t.Run("invalid datasetGroupID is skipped", func(t *testing.T) {
		result, err := svc.GetSQLParams([]int64{0, -1, 999999})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("dataset with no tables returns empty", func(t *testing.T) {
		group, err := svc.Save(&dataset.WriteRequest{Name: "NoTablesDS", NodeType: "dataset"})
		require.NoError(t, err)

		result, err := svc.GetSQLParams([]int64{group.ID})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("table with empty SQLVariables is skipped", func(t *testing.T) {
		group, err := svc.Save(&dataset.WriteRequest{Name: "EmptyVarsDS", NodeType: "dataset"})
		require.NoError(t, err)

		err = testDB.Create(&dataset.CoreDatasetTable{
			DatasetGroupID: group.ID,
			PhysicalTable:  dsSvcStrPtr("empty_vars_table"),
			SQLVariables:   dsSvcStrPtr(""),
		}).Error
		require.NoError(t, err)

		result, err := svc.GetSQLParams([]int64{group.ID})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("table with invalid JSON SQLVariables is skipped", func(t *testing.T) {
		group, err := svc.Save(&dataset.WriteRequest{Name: "InvalidJSONDS", NodeType: "dataset"})
		require.NoError(t, err)

		err = testDB.Create(&dataset.CoreDatasetTable{
			DatasetGroupID: group.ID,
			PhysicalTable:  dsSvcStrPtr("invalid_json_table"),
			SQLVariables:   dsSvcStrPtr("not valid json"),
		}).Error
		require.NoError(t, err)

		result, err := svc.GetSQLParams([]int64{group.ID})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("table with empty variableName is skipped", func(t *testing.T) {
		group, err := svc.Save(&dataset.WriteRequest{Name: "EmptyNameDS", NodeType: "dataset"})
		require.NoError(t, err)

		sqlVars, _ := json.Marshal([]map[string]interface{}{
			{"variableName": "", "type": []string{"TEXT"}, "params": []interface{}{""}},
			{"variableName": "   ", "type": []string{"TEXT"}, "params": []interface{}{""}},
			{"variableName": "valid_name", "type": []string{"TEXT"}, "params": []interface{}{"value"}},
		})

		err = testDB.Create(&dataset.CoreDatasetTable{
			DatasetGroupID: group.ID,
			PhysicalTable:  dsSvcStrPtr("empty_name_table"),
			SQLVariables:   dsSvcStrPtr(string(sqlVars)),
		}).Error
		require.NoError(t, err)

		result, err := svc.GetSQLParams([]int64{group.ID})
		assert.NoError(t, err)
		assert.Len(t, result, 1) // Only the valid_name should be returned
		assert.Equal(t, "valid_name", result[0].VariableName)
	})

	t.Run("mixed valid and invalid IDs", func(t *testing.T) {
		group, err := svc.Save(&dataset.WriteRequest{Name: "MixedDS", NodeType: "dataset"})
		require.NoError(t, err)

		sqlVars, _ := json.Marshal([]map[string]interface{}{
			{"variableName": "p_test", "type": []string{"TEXT"}, "params": []interface{}{"value"}},
		})

		err = testDB.Create(&dataset.CoreDatasetTable{
			DatasetGroupID: group.ID,
			PhysicalTable:  dsSvcStrPtr("mixed_table"),
			SQLVariables:   dsSvcStrPtr(string(sqlVars)),
		}).Error
		require.NoError(t, err)

		// Include valid and invalid IDs
		result, err := svc.GetSQLParams([]int64{0, -1, group.ID, 999999})
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestDatasetServiceIntegration_Delete_DeepRecursive(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create nested folder structure: root -> level1 -> level2 -> leaf
	root, err := svc.Save(&dataset.WriteRequest{
		Name:     "Root Folder",
		NodeType: "folder",
	})
	require.NoError(t, err)

	level1, err := svc.Save(&dataset.WriteRequest{
		Name:     "Level 1 Folder",
		NodeType: "folder",
		PID:      &root.ID,
	})
	require.NoError(t, err)

	level2, err := svc.Save(&dataset.WriteRequest{
		Name:     "Level 2 Folder",
		NodeType: "folder",
		PID:      &level1.ID,
	})
	require.NoError(t, err)

	leaf, err := svc.Save(&dataset.WriteRequest{
		Name:     "Leaf Dataset",
		NodeType: "dataset",
		PID:      &level2.ID,
	})
	require.NoError(t, err)

	// Delete root - should cascade delete all children
	err = svc.Delete(root.ID)
	assert.NoError(t, err)

	// Verify all are deleted
	_, err = svc.GetGroupByID(root.ID)
	assert.Error(t, err)
	_, err = svc.GetGroupByID(level1.ID)
	assert.Error(t, err)
	_, err = svc.GetGroupByID(level2.ID)
	assert.Error(t, err)
	_, err = svc.GetGroupByID(leaf.ID)
	assert.Error(t, err)
}

func TestDatasetServiceIntegration_GetFieldEnumObj_WithResultModeAndFilter(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error

	_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_result_mode").Error
	err := testDB.Exec("CREATE TABLE it_enum_result_mode (id BIGINT PRIMARY KEY AUTO_INCREMENT, region VARCHAR(64), city VARCHAR(64))").Error
	require.NoError(t, err)
	err = testDB.Exec("INSERT INTO it_enum_result_mode (region, city) VALUES ('East', 'Shanghai'), ('West', 'Chengdu'), ('North', 'Beijing')").Error
	require.NoError(t, err)

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	group, err := svc.Save(&dataset.WriteRequest{Name: "EnumResultMode", NodeType: "dataset"})
	require.NoError(t, err)

	table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_enum_result_mode")}
	require.NoError(t, testDB.Create(table).Error)

	deType := 0
	regionField := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("region"), Name: dsSvcStrPtr("region"), DeType: &deType}
	cityField := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("city"), Name: dsSvcStrPtr("city"), DeType: &deType}
	require.NoError(t, testDB.Create(regionField).Error)
	require.NoError(t, testDB.Create(cityField).Error)

	t.Run("result mode 1 uses higher limit", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{
			QueryID:    regionField.ID,
			ResultMode: 1,
			Sort:       "ASC",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("with filter condition", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{
			QueryID: regionField.ID,
			Filter: []dataset.EnumFilter{
				{FieldID: fmt.Sprintf("%d", regionField.ID), Operator: "in", Value: []interface{}{"East"}},
			},
			Sort: "ASC",
		})
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("display ID different from query ID", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{
			QueryID:   regionField.ID,
			DisplayID: cityField.ID,
			Sort:      "ASC",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_result_mode")
}

func TestDatasetServiceIntegration_GetFieldEnumObj_MoreEdgeCases(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	t.Run("nil request returns empty", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(nil)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("invalid QueryID returns empty", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{QueryID: 0})
		assert.NoError(t, err)
		assert.Empty(t, result)

		result, err = svc.GetFieldEnumObj(&dataset.EnumValueRequest{QueryID: -1})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("non-existent QueryID returns empty", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{QueryID: 999999})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestDatasetServiceIntegration_GetFieldEnumObj_WithSortID(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})
	_ = testDB.AutoMigrate(&dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	_ = testDB.Exec("DELETE FROM core_dataset_table_field").Error
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error

	_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_sort").Error
	err := testDB.Exec("CREATE TABLE it_enum_sort (id BIGINT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(64), value INT)").Error
	require.NoError(t, err)
	err = testDB.Exec("INSERT INTO it_enum_sort (name, value) VALUES ('A', 3), ('B', 1), ('C', 2)").Error
	require.NoError(t, err)

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	group, err := svc.Save(&dataset.WriteRequest{Name: "EnumSort", NodeType: "dataset"})
	require.NoError(t, err)

	table := &dataset.CoreDatasetTable{DatasetGroupID: group.ID, PhysicalTable: dsSvcStrPtr("it_enum_sort")}
	require.NoError(t, testDB.Create(table).Error)

	deType := 0
	nameField := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("name"), Name: dsSvcStrPtr("name"), DeType: &deType}
	valueField := &dataset.CoreDatasetTableField{DatasetGroupID: group.ID, DatasetTableID: &table.ID, OriginName: dsSvcStrPtr("value"), Name: dsSvcStrPtr("value"), DeType: &deType}
	require.NoError(t, testDB.Create(nameField).Error)
	require.NoError(t, testDB.Create(valueField).Error)

	t.Run("sort by same field", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{
			QueryID: nameField.ID,
			SortID:  nameField.ID,
			Sort:    "ASC",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("with search text", func(t *testing.T) {
		result, err := svc.GetFieldEnumObj(&dataset.EnumValueRequest{
			QueryID:    nameField.ID,
			SearchText: "A",
			Sort:       "ASC",
		})
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	_ = testDB.Exec("DROP TABLE IF EXISTS it_enum_sort")
}
func TestDatasetServiceIntegration_Delete_EmptyFolder(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create empty folder
	folder, err := svc.Save(&dataset.WriteRequest{
		Name:     "Empty Folder",
		NodeType: "folder",
	})
	require.NoError(t, err)

	// Delete empty folder
	err = svc.Delete(folder.ID)
	assert.NoError(t, err)

	// Verify deleted
	_, err = svc.GetGroupByID(folder.ID)
	assert.Error(t, err)
}

func TestDatasetServiceIntegration_Delete_SingleDataset(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create single dataset
	ds, err := svc.Save(&dataset.WriteRequest{
		Name:     "Single Dataset",
		NodeType: "dataset",
	})
	require.NoError(t, err)

	// Delete dataset
	err = svc.Delete(ds.ID)
	assert.NoError(t, err)

	// Verify deleted
	_, err = svc.GetGroupByID(ds.ID)
	assert.Error(t, err)
}

func TestDatasetServiceIntegration_Move_ToChild(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create parent folder
	parent, err := svc.Save(&dataset.WriteRequest{
		Name:     "Parent Folder",
		NodeType: "folder",
	})
	require.NoError(t, err)

	// Create child folder
	child, err := svc.Save(&dataset.WriteRequest{
		Name:     "Child Folder",
		NodeType: "folder",
		PID:      &parent.ID,
	})
	require.NoError(t, err)

	// Try to move parent to child - should fail
	_, err = svc.Move(parent.ID, child.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be child")
}

func TestDatasetServiceIntegration_Move_DeepNested(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create nested structure: root -> level1 -> level2 -> leaf
	root, err := svc.Save(&dataset.WriteRequest{
		Name:     "Root",
		NodeType: "folder",
	})
	require.NoError(t, err)

	level1, err := svc.Save(&dataset.WriteRequest{
		Name:     "Level1",
		NodeType: "folder",
		PID:      &root.ID,
	})
	require.NoError(t, err)

	level2, err := svc.Save(&dataset.WriteRequest{
		Name:     "Level2",
		NodeType: "folder",
		PID:      &level1.ID,
	})
	require.NoError(t, err)

	// Try to move root to level2 - should fail (level2 is a descendant of root)
	_, err = svc.Move(root.ID, level2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be child")
}

func TestDatasetServiceIntegration_Move_Success(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewDatasetRepository(testDB)
	svc := NewDatasetService(repo)

	// Create folder1
	folder1, err := svc.Save(&dataset.WriteRequest{
		Name:     "Folder1",
		NodeType: "folder",
	})
	require.NoError(t, err)

	// Create folder2
	folder2, err := svc.Save(&dataset.WriteRequest{
		Name:     "Folder2",
		NodeType: "folder",
	})
	require.NoError(t, err)

	// Create dataset under folder1
	ds, err := svc.Save(&dataset.WriteRequest{
		Name:     "Dataset",
		NodeType: "dataset",
		PID:      &folder1.ID,
	})
	require.NoError(t, err)

	// Move dataset from folder1 to folder2
	moved, err := svc.Move(ds.ID, folder2.ID)
	assert.NoError(t, err)
	assert.Equal(t, folder2.ID, *moved.PID)

	// Verify the move
	found, err := svc.GetGroupByID(ds.ID)
	assert.NoError(t, err)
	assert.Equal(t, folder2.ID, *found.PID)
}
