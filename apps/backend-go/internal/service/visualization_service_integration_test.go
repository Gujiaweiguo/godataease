//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisualizationServiceIntegration_Save(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Test save panel
	nodeType := "panel"
	req := &visualization.SaveRequest{
		Name:            "Test Dashboard",
		NodeType:        &nodeType,
		CanvasStyleData: strPtr("{\"style\":\"dark\"}"),
		ComponentData:   strPtr("{\"components\":[]}"),
	}

	id, err := svc.Save(req, "tester")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify saved
	detail, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "Test Dashboard", detail.Name)
}

func TestVisualizationServiceIntegration_Save_Folder(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Test save folder
	nodeType := "folder"
	req := &visualization.SaveRequest{
		Name:     "Test Folder",
		NodeType: &nodeType,
	}

	id, err := svc.Save(req, "tester")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify folder status is 1 (active)
	detail, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "Test Folder", detail.Name)
	if detail.Status != nil {
		assert.Equal(t, 1, *detail.Status)
	}
}

func TestVisualizationServiceIntegration_Copy(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	nodeType := "panel"
	visType := "dashboard"
	mobileLayout := true
	createReq := &visualization.SaveRequest{
		Name:            "Copy Source",
		NodeType:        &nodeType,
		Type:            &visType,
		CanvasStyleData: strPtr("{\"style\":\"source\"}"),
		ComponentData:   strPtr("{\"components\":[1]}"),
		MobileLayout:    &mobileLayout,
		ContentID:       strPtr("content-1"),
		CheckVersion:    strPtr("ver-1"),
	}
	sourceID, err := svc.Save(createReq, "creator")
	require.NoError(t, err)

	newPID := int64(100)
	copyID, err := svc.Copy(&visualization.CopyRequest{
		ID:           sourceID,
		Name:         "Copy Target",
		PID:          &newPID,
		Type:         &visType,
		NodeType:     &nodeType,
		MobileLayout: &mobileLayout,
	}, "copier")
	require.NoError(t, err)
	assert.NotEqual(t, sourceID, copyID)

	copied, err := repo.GetByID(copyID)
	require.NoError(t, err)
	assert.Equal(t, "Copy Target", copied.Name)
	assert.Equal(t, newPID, *copied.PID)
	assert.Equal(t, "dashboard", *copied.Type)
	assert.Equal(t, "panel", *copied.NodeType)
	assert.Equal(t, "{\"style\":\"source\"}", *copied.CanvasStyleData)
	assert.Equal(t, "{\"components\":[1]}", *copied.ComponentData)
	assert.Equal(t, "content-1", *copied.ContentID)
	assert.Equal(t, "ver-1", *copied.CheckVersion)
	assert.Equal(t, "copier", *copied.CreateBy)
}

func TestVisualizationServiceIntegration_Save_ScreenInheritsParentResourcePermissions(t *testing.T) {
	cleanupTables(&permission.SysResourcePerm{}, &permission.SysResource{}, &visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	resourcePermRepo := repository.NewResourcePermissionRepository(testDB)
	resourcePermSvc := NewResourcePermissionService(resourcePermRepo, nil)
	svc := NewVisualizationService(repo)
	svc.SetResourcePermissionService(resourcePermSvc)

	folderNodeType := "folder"
	screenType := "dataV"
	parentID, err := svc.Save(&visualization.SaveRequest{
		Name:     "Governed Screen Folder",
		NodeType: &folderNodeType,
		Type:     &screenType,
	}, "tester")
	require.NoError(t, err)
	require.NoError(t, resourcePermSvc.RegisterResource(parentID, "Governed Screen Folder", permission.ResourceTypeScreen, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(parentID, permission.ResourceTypeScreen, []int64{31, 32}))

	panelNodeType := "panel"
	childID, err := svc.Save(&visualization.SaveRequest{
		Name:     "Inherited Screen",
		NodeType: &panelNodeType,
		Type:     &screenType,
		PID:      &parentID,
	}, "tester")
	require.NoError(t, err)

	permIDs, exists, err := resourcePermRepo.GetResourcePermissionIDs(childID, permission.ResourceTypeScreen)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.ElementsMatch(t, []int64{31, 32}, permIDs)
}

func TestVisualizationServiceIntegration_BackfillGovernedResources(t *testing.T) {
	cleanupTables(&permission.SysResourcePerm{}, &permission.SysResource{}, &visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	resourcePermRepo := repository.NewResourcePermissionRepository(testDB)
	resourcePermSvc := NewResourcePermissionService(resourcePermRepo, nil)
	seedSvc := NewVisualizationService(repo)
	backfillSvc := NewVisualizationService(repo)
	backfillSvc.SetResourcePermissionService(resourcePermSvc)

	folderNodeType := "folder"
	panelNodeType := "panel"
	dashboardType := "dashboard"
	screenType := "dataV"

	dashboardParentID, err := seedSvc.Save(&visualization.SaveRequest{
		Name:     "Governed Dashboard Folder",
		NodeType: &folderNodeType,
		Type:     &dashboardType,
	}, "tester")
	require.NoError(t, err)
	screenParentID, err := seedSvc.Save(&visualization.SaveRequest{
		Name:     "Governed Screen Folder",
		NodeType: &folderNodeType,
		Type:     &screenType,
	}, "tester")
	require.NoError(t, err)
	plainParentID, err := seedSvc.Save(&visualization.SaveRequest{
		Name:     "Ungoverned Dashboard Folder",
		NodeType: &folderNodeType,
		Type:     &dashboardType,
	}, "tester")
	require.NoError(t, err)

	dashboardChildID, err := seedSvc.Save(&visualization.SaveRequest{
		Name:     "Legacy Dashboard",
		NodeType: &panelNodeType,
		Type:     &dashboardType,
		PID:      &dashboardParentID,
	}, "tester")
	require.NoError(t, err)
	screenChildID, err := seedSvc.Save(&visualization.SaveRequest{
		Name:     "Legacy Screen",
		NodeType: &panelNodeType,
		Type:     &screenType,
		PID:      &screenParentID,
	}, "tester")
	require.NoError(t, err)
	plainChildID, err := seedSvc.Save(&visualization.SaveRequest{
		Name:     "Legacy Plain Dashboard",
		NodeType: &panelNodeType,
		Type:     &dashboardType,
		PID:      &plainParentID,
	}, "tester")
	require.NoError(t, err)

	require.NoError(t, resourcePermSvc.RegisterResource(dashboardParentID, "Governed Dashboard Folder", permission.ResourceTypeDashboard, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(dashboardParentID, permission.ResourceTypeDashboard, []int64{41, 42}))
	require.NoError(t, resourcePermSvc.RegisterResource(screenParentID, "Governed Screen Folder", permission.ResourceTypeScreen, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(screenParentID, permission.ResourceTypeScreen, []int64{51, 52}))

	report, err := backfillSvc.BackfillGovernedResources()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, report.Scanned, 6)
	assert.Equal(t, 2, report.Governed)
	assert.Contains(t, report.ResourceIDs, dashboardChildID)
	assert.Contains(t, report.ResourceIDs, screenChildID)

	dashboardPermIDs, dashboardExists, err := resourcePermRepo.GetResourcePermissionIDs(dashboardChildID, permission.ResourceTypeDashboard)
	require.NoError(t, err)
	assert.True(t, dashboardExists)
	assert.ElementsMatch(t, []int64{41, 42}, dashboardPermIDs)

	screenPermIDs, screenExists, err := resourcePermRepo.GetResourcePermissionIDs(screenChildID, permission.ResourceTypeScreen)
	require.NoError(t, err)
	assert.True(t, screenExists)
	assert.ElementsMatch(t, []int64{51, 52}, screenPermIDs)

	plainPermIDs, plainExists, err := resourcePermRepo.GetResourcePermissionIDs(plainChildID, permission.ResourceTypeDashboard)
	require.NoError(t, err)
	assert.False(t, plainExists)
	assert.Empty(t, plainPermIDs)

	report, err = backfillSvc.BackfillGovernedResources()
	require.NoError(t, err)
	assert.Equal(t, 2, report.Governed)

	var dashboardResourceCount int64
	require.NoError(t, testDB.Model(&permission.SysResource{}).
		Where("resource_type = ? AND resource_name = ?", permission.ResourceTypeDashboard, "Legacy Dashboard").
		Count(&dashboardResourceCount).Error)
	assert.Equal(t, int64(1), dashboardResourceCount)

	var screenResourceCount int64
	require.NoError(t, testDB.Model(&permission.SysResource{}).
		Where("resource_type = ? AND resource_name = ?", permission.ResourceTypeScreen, "Legacy Screen").
		Count(&screenResourceCount).Error)
	assert.Equal(t, int64(1), screenResourceCount)
}

func TestVisualizationServiceIntegration_BackfillGovernedResourcesWithOptions(t *testing.T) {
	cleanupTables(&permission.SysResourcePerm{}, &permission.SysResource{}, &visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	resourcePermRepo := repository.NewResourcePermissionRepository(testDB)
	resourcePermSvc := NewResourcePermissionService(resourcePermRepo, nil)
	seedSvc := NewVisualizationService(repo)
	backfillSvc := NewVisualizationService(repo)
	backfillSvc.SetResourcePermissionService(resourcePermSvc)

	folderNodeType := "folder"
	panelNodeType := "panel"
	dashboardType := "dashboard"
	screenType := "dataV"

	dashboardParentID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Batch Governed Dashboard Folder", NodeType: &folderNodeType, Type: &dashboardType}, "tester")
	require.NoError(t, err)
	plainParentID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Batch Plain Dashboard Folder", NodeType: &folderNodeType, Type: &dashboardType}, "tester")
	require.NoError(t, err)
	screenParentID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Batch Governed Screen Folder", NodeType: &folderNodeType, Type: &screenType}, "tester")
	require.NoError(t, err)
	dashboardChildID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Batch Legacy Dashboard", NodeType: &panelNodeType, Type: &dashboardType, PID: &dashboardParentID}, "tester")
	require.NoError(t, err)
	plainChildID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Batch Legacy Plain Dashboard", NodeType: &panelNodeType, Type: &dashboardType, PID: &plainParentID}, "tester")
	require.NoError(t, err)
	screenChildID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Batch Legacy Screen", NodeType: &panelNodeType, Type: &screenType, PID: &screenParentID}, "tester")
	require.NoError(t, err)

	require.NoError(t, resourcePermSvc.RegisterResource(dashboardParentID, "Batch Governed Dashboard Folder", permission.ResourceTypeDashboard, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(dashboardParentID, permission.ResourceTypeDashboard, []int64{81, 82}))
	require.NoError(t, resourcePermSvc.RegisterResource(screenParentID, "Batch Governed Screen Folder", permission.ResourceTypeScreen, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(screenParentID, permission.ResourceTypeScreen, []int64{91, 92}))

	batch1, err := backfillSvc.BackfillGovernedVisualizationResourcesWithOptions(&GovernanceBackfillOptions{Limit: 3})
	require.NoError(t, err)
	assert.Equal(t, 3, batch1.Scanned)
	assert.Equal(t, 0, batch1.Governed)
	assert.Equal(t, 3, batch1.Skipped)
	assert.Equal(t, screenParentID, batch1.NextAfterID)
	if assert.Len(t, batch1.SkippedItems, 3) {
		assert.Equal(t, GovernanceBackfillSkipReasonMissingParent, batch1.SkippedItems[0].Reason)
		assert.Equal(t, GovernanceBackfillSkipReasonMissingParent, batch1.SkippedItems[1].Reason)
		assert.Equal(t, GovernanceBackfillSkipReasonMissingParent, batch1.SkippedItems[2].Reason)
		assert.Equal(t, GovernanceBackfillRemediationDataCleanup, batch1.SkippedItems[0].Remediation)
		assert.Equal(t, GovernanceBackfillRemediationDataCleanup, batch1.SkippedItems[1].Remediation)
		assert.Equal(t, GovernanceBackfillRemediationDataCleanup, batch1.SkippedItems[2].Remediation)
	}

	batch2, err := backfillSvc.BackfillGovernedVisualizationResourcesWithOptions(&GovernanceBackfillOptions{AfterID: batch1.NextAfterID, Limit: 3})
	require.NoError(t, err)
	assert.Equal(t, 3, batch2.Scanned)
	assert.Equal(t, 2, batch2.Governed)
	assert.Equal(t, 1, batch2.Skipped)
	assert.Contains(t, batch2.ResourceIDs, dashboardChildID)
	assert.Contains(t, batch2.ResourceIDs, screenChildID)
	if assert.Len(t, batch2.SkippedItems, 1) {
		assert.Equal(t, plainChildID, batch2.SkippedItems[0].ResourceID)
		assert.Equal(t, permission.ResourceTypeDashboard, batch2.SkippedItems[0].ResourceType)
		assert.Equal(t, GovernanceBackfillSkipReasonParentNotGoverned, batch2.SkippedItems[0].Reason)
		assert.Equal(t, GovernanceBackfillRemediationGovernParent, batch2.SkippedItems[0].Remediation)
	}

	dashboardPermIDs, dashboardExists, err := resourcePermRepo.GetResourcePermissionIDs(dashboardChildID, permission.ResourceTypeDashboard)
	require.NoError(t, err)
	assert.True(t, dashboardExists)
	assert.ElementsMatch(t, []int64{81, 82}, dashboardPermIDs)

	screenPermIDs, screenExists, err := resourcePermRepo.GetResourcePermissionIDs(screenChildID, permission.ResourceTypeScreen)
	require.NoError(t, err)
	assert.True(t, screenExists)
	assert.ElementsMatch(t, []int64{91, 92}, screenPermIDs)
}

func TestVisualizationServiceIntegration_AppCanvasNameCheck(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{})

	repo := repository.NewVisualizationRepository(testDB)
	datasetRepo := repository.NewDatasetRepository(testDB)
	svc := NewVisualizationService(repo)
	svc.SetDatasetRepository(datasetRepo)

	pid := int64(10)
	nodeType := dataset.NodeTypeFolder
	require.NoError(t, testDB.Create(&dataset.CoreDatasetGroup{
		ID:             101,
		Name:           "Folder A",
		PID:            &pid,
		NodeType:       &nodeType,
		CreateBy:       "tester",
		CreateTime:     1,
		UpdateBy:       "tester",
		LastUpdateTime: 1,
	}).Error)

	result, err := svc.AppCanvasNameCheck(&visualization.AppCanvasNameCheckRequest{
		DatasetFolderPid:  &pid,
		DatasetFolderName: "Folder A",
	})
	require.NoError(t, err)
	assert.Equal(t, "repeat", result)

	result, err = svc.AppCanvasNameCheck(&visualization.AppCanvasNameCheckRequest{
		DatasetFolderPid:  &pid,
		DatasetFolderName: "Folder B",
	})
	require.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestVisualizationServiceIntegration_RecordExportLog(t *testing.T) {
	cleanupTables(&audit.AuditLog{})

	repo := repository.NewVisualizationRepository(testDB)
	auditSvc := NewAuditService(
		repository.NewAuditLogRepository(testDB),
		repository.NewLoginFailureRepository(testDB),
		repository.NewAuditLogDetailRepository(testDB),
	)
	svc := NewVisualizationService(repo)
	svc.SetAuditService(auditSvc)

	id := int64(123)
	userID := int64(7)
	username := "tester"
	ipAddress := "127.0.0.1"
	userAgent := "integration-test"
	require.NoError(t, svc.RecordExportLog(&visualization.ExportLogRequest{ID: &id, Type: "screen"}, &userID, &username, &ipAddress, &userAgent, "app"))

	var logs []audit.AuditLog
	require.NoError(t, testDB.Order("id ASC").Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.Equal(t, audit.ActionTypeDataAccess, logs[0].ActionType)
	assert.Equal(t, audit.OperationExport, logs[0].Operation)
	assert.Equal(t, "导出应用模板", logs[0].ActionName)
	require.NotNil(t, logs[0].ResourceType)
	assert.Equal(t, "SCREEN", *logs[0].ResourceType)
}

func ensureExport2AppCheckTables(t *testing.T) {
	t.Helper()
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS core_chart_view (
		id BIGINT PRIMARY KEY,
		title VARCHAR(255),
		scene_id BIGINT,
		table_id BIGINT,
		type VARCHAR(64)
	)`).Error)
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS core_dataset_table (
		id BIGINT PRIMARY KEY,
		name VARCHAR(255),
		datasource_id BIGINT,
		dataset_group_id BIGINT,
		table_name VARCHAR(255),
		type VARCHAR(64)
	)`).Error)
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS core_dataset_table_field (
		id BIGINT PRIMARY KEY,
		datasource_id BIGINT,
		dataset_table_id BIGINT,
		dataset_group_id BIGINT,
		chart_id BIGINT,
		origin_name VARCHAR(255),
		name VARCHAR(255),
		group_type VARCHAR(64),
		type VARCHAR(64),
		de_type INT
	)`).Error)
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS visualization_linkage (
		id BIGINT PRIMARY KEY,
		dv_id BIGINT,
		source_view_id BIGINT,
		target_view_id BIGINT,
		linkage_active TINYINT
	)`).Error)
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS visualization_linkage_field (
		id BIGINT PRIMARY KEY,
		linkage_id BIGINT,
		source_field BIGINT,
		target_field BIGINT
	)`).Error)
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS visualization_link_jump (
		id BIGINT PRIMARY KEY,
		source_dv_id BIGINT,
		source_view_id BIGINT,
		checked TINYINT
	)`).Error)
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS visualization_link_jump_info (
		id BIGINT PRIMARY KEY,
		link_jump_id BIGINT,
		link_type VARCHAR(64),
		jump_type VARCHAR(64),
		source_field_id BIGINT,
		checked TINYINT
	)`).Error)
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS visualization_link_jump_target_view_info (
		target_id BIGINT PRIMARY KEY,
		link_jump_info_id BIGINT,
		source_field_active_id BIGINT,
		target_view_id VARCHAR(64),
		target_field_id VARCHAR(64)
	)`).Error)
	require.NoError(t, testDB.Exec(`DELETE FROM visualization_link_jump_target_view_info`).Error)
	require.NoError(t, testDB.Exec(`DELETE FROM visualization_link_jump_info`).Error)
	require.NoError(t, testDB.Exec(`DELETE FROM visualization_link_jump`).Error)
	require.NoError(t, testDB.Exec(`DELETE FROM visualization_linkage_field`).Error)
	require.NoError(t, testDB.Exec(`DELETE FROM visualization_linkage`).Error)
	require.NoError(t, testDB.Exec(`DELETE FROM core_dataset_table_field`).Error)
	require.NoError(t, testDB.Exec(`DELETE FROM core_chart_view`).Error)
	require.NoError(t, testDB.Exec(`DELETE FROM core_dataset_table`).Error)
}

func TestVisualizationServiceIntegration_Export2AppCheck(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{}, &auto.CoreDatasourceTask{}, &auto.CoreDatasource{})
	ensureExport2AppCheckTables(t)

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	groupPID := int64(0)
	nodeType := dataset.NodeTypeDataset
	require.NoError(t, testDB.Create(&dataset.CoreDatasetGroup{ID: 100, Name: "grp1", PID: &groupPID, NodeType: &nodeType, CreateBy: "tester", CreateTime: 1, UpdateBy: "tester", LastUpdateTime: 1}).Error)
	require.NoError(t, testDB.Create(&auto.CoreDatasource{ID: 1, Name: "ds1", Type: "MySQL"}).Error)
	require.NoError(t, testDB.Create(&auto.CoreDatasourceTask{ID: 2, DsID: 1, Name: "task1", UpdateType: "all", SyncRate: "0"}).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO core_dataset_table (id, name, datasource_id, dataset_group_id, table_name, type) VALUES (200, 't1', 1, 100, 'table_a', 'db')`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO core_dataset_table_field (id, datasource_id, dataset_table_id, dataset_group_id, chart_id, origin_name, name, group_type, type, de_type) VALUES (300, 1, 200, 100, 400, 'f1', 'field1', 'd', 'varchar', 0)`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO core_chart_view (id, title, scene_id, table_id, type) VALUES (400, 'chart1', 500, 100, 'bar')`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO visualization_linkage (id, dv_id, source_view_id, target_view_id, linkage_active) VALUES (700, 500, 400, 401, 1)`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO visualization_linkage_field (id, linkage_id, source_field, target_field) VALUES (800, 700, 300, 301)`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO visualization_link_jump (id, source_dv_id, source_view_id, checked) VALUES (900, 500, 400, 1)`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO visualization_link_jump_info (id, link_jump_id, link_type, jump_type, source_field_id, checked) VALUES (1000, 900, 'inner', '_blank', 300, 1)`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO visualization_link_jump_target_view_info (target_id, link_jump_info_id, source_field_active_id, target_view_id, target_field_id) VALUES (1100, 1000, 300, '999', '888')`).Error)

	resp, err := svc.Export2AppCheck(&visualization.Export2AppCheckRequest{DvID: 500, ViewIDs: []int64{400}, DsIDs: []int64{100}})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.CheckStatus)
	assert.Equal(t, "success", resp.CheckMes)
	assert.Len(t, resp.ChartViewsInfo, 1)
	assert.Len(t, resp.DatasourceInfo, 1)
	assert.Len(t, resp.LinkJumps, 1)
}

func TestVisualizationServiceIntegration_Export2AppCheck_RejectsAPIDataSource(t *testing.T) {
	cleanupTables(&dataset.CoreDatasetGroup{}, &auto.CoreDatasource{})
	ensureExport2AppCheckTables(t)

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	groupPID := int64(0)
	nodeType := dataset.NodeTypeDataset
	require.NoError(t, testDB.Create(&dataset.CoreDatasetGroup{ID: 101, Name: "grp-api", PID: &groupPID, NodeType: &nodeType, CreateBy: "tester", CreateTime: 1, UpdateBy: "tester", LastUpdateTime: 1}).Error)
	require.NoError(t, testDB.Create(&auto.CoreDatasource{ID: 11, Name: "api-ds", Type: "API"}).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO core_dataset_table (id, name, datasource_id, dataset_group_id, table_name, type) VALUES (201, 't2', 11, 101, 'table_b', 'db')`).Error)

	resp, err := svc.Export2AppCheck(&visualization.Export2AppCheckRequest{DvID: 0, DsIDs: []int64{101}})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "API")
}

func TestVisualizationServiceIntegration_HelperCoverage(t *testing.T) {
	cleanupTables()
	ensureExport2AppCheckTables(t)

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)
	svc.SetTemplateService(NewTemplateService(repository.NewTemplateRepository(testDB)))
	svc.SetTemplateExtendDataRepo(repository.NewTemplateExtendDataRepository(testDB))

	require.NoError(t, testDB.Exec(`INSERT INTO core_chart_view (id, title, scene_id, table_id, type) VALUES (401, 'chart-helper', 501, 100, 'line')`).Error)
	rows, err := svc.ViewDetailList(501)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	converted := stringifyExportIDs([]map[string]interface{}{{
		"id":        int64(1),
		"scene_id":  int32(2),
		"target_id": uint64(3),
		"plain":     "keep",
	}})
	assert.Equal(t, "1", converted[0]["id"])
	assert.Equal(t, "2", converted[0]["scene_id"])
	assert.Equal(t, "3", converted[0]["target_id"])
	assert.Equal(t, "keep", converted[0]["plain"])

	appData := `{"visualizationInfo":{"id":11}}`
	assert.Contains(t, processAppData(appData, 22), `"id":22`)
	assert.Equal(t, `not-json`, processAppData(`not-json`, 22))

	value, ok := extractInt64Value(int64(7))
	assert.True(t, ok)
	assert.Equal(t, int64(7), value)
	_, ok = extractInt64Value("bad")
	assert.False(t, ok)
}

func TestVisualizationServiceIntegration_BackfillGovernedResourcesWithOptions_FiltersByOrg(t *testing.T) {
	cleanupTables(&permission.SysResourcePerm{}, &permission.SysResource{}, &visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	resourcePermRepo := repository.NewResourcePermissionRepository(testDB)
	resourcePermSvc := NewResourcePermissionService(resourcePermRepo, nil)
	seedSvc := NewVisualizationService(repo)
	backfillSvc := NewVisualizationService(repo)
	backfillSvc.SetResourcePermissionService(resourcePermSvc)

	folderNodeType := "folder"
	panelNodeType := "panel"
	dashboardType := "dashboard"
	orgA := int64(101)
	orgB := int64(202)

	governedParentAID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Org A Governed Folder", NodeType: &folderNodeType, Type: &dashboardType}, "tester")
	require.NoError(t, err)
	governedParentBID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Org B Governed Folder", NodeType: &folderNodeType, Type: &dashboardType}, "tester")
	require.NoError(t, err)
	childAID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Org A Legacy Dashboard", NodeType: &panelNodeType, Type: &dashboardType, PID: &governedParentAID}, "tester")
	require.NoError(t, err)
	childBID, err := seedSvc.Save(&visualization.SaveRequest{Name: "Org B Legacy Dashboard", NodeType: &panelNodeType, Type: &dashboardType, PID: &governedParentBID}, "tester")
	require.NoError(t, err)

	require.NoError(t, testDB.Model(&visualization.DataVisualizationInfo{}).Where("id IN ?", []int64{governedParentAID, childAID}).Update("org_id", orgA).Error)
	require.NoError(t, testDB.Model(&visualization.DataVisualizationInfo{}).Where("id IN ?", []int64{governedParentBID, childBID}).Update("org_id", orgB).Error)

	require.NoError(t, resourcePermSvc.RegisterResource(governedParentAID, "Org A Governed Folder", permission.ResourceTypeDashboard, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(governedParentAID, permission.ResourceTypeDashboard, []int64{111, 112}))
	require.NoError(t, resourcePermSvc.RegisterResource(governedParentBID, "Org B Governed Folder", permission.ResourceTypeDashboard, nil))
	require.NoError(t, resourcePermSvc.ReplaceResourcePermissions(governedParentBID, permission.ResourceTypeDashboard, []int64{211, 212}))

	report, err := backfillSvc.BackfillGovernedVisualizationResourcesWithOptions(&GovernanceBackfillOptions{OrgID: &orgA, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, &orgA, report.OrgID)
	assert.Equal(t, 2, report.Scanned)
	assert.Equal(t, 1, report.Governed)
	assert.Contains(t, report.ResourceIDs, childAID)
	assert.NotContains(t, report.ResourceIDs, childBID)
	assert.Equal(t, "current_request_batch", report.RollbackBoundary)
	assert.Equal(t, "idempotent_recompute", report.RerunStrategy)

	permA, existsA, err := resourcePermRepo.GetResourcePermissionIDs(childAID, permission.ResourceTypeDashboard)
	require.NoError(t, err)
	assert.True(t, existsA)
	assert.ElementsMatch(t, []int64{111, 112}, permA)

	permB, existsB, err := resourcePermRepo.GetResourcePermissionIDs(childBID, permission.ResourceTypeDashboard)
	require.NoError(t, err)
	assert.False(t, existsB)
	assert.Empty(t, permB)
}

func TestVisualizationServiceIntegration_Update(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create first
	nodeType := "panel"
	createReq := &visualization.SaveRequest{
		Name:     "Original Name",
		NodeType: &nodeType,
	}
	id, err := svc.Save(createReq, "creator")
	assert.NoError(t, err)

	// Update
	updateReq := &visualization.UpdateRequest{
		ID:   id,
		Name: strPtr("Updated Name"),
	}
	err = svc.Update(updateReq, "updater")
	assert.NoError(t, err)

	// Verify updated
	detail, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", detail.Name)
}

func TestVisualizationServiceIntegration_Update_NotFound(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	updateReq := &visualization.UpdateRequest{
		ID:   99999,
		Name: strPtr("Updated Name"),
	}
	err := svc.Update(updateReq, "updater")
	assert.Error(t, err)
}

func TestVisualizationServiceIntegration_Detail(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create first
	nodeType := "panel"
	createReq := &visualization.SaveRequest{
		Name:     "Detail Test",
		NodeType: &nodeType,
	}
	id, err := svc.Save(createReq, "tester")
	assert.NoError(t, err)

	// Get detail
	req := &visualization.DetailRequest{ID: visualization.FlexInt(id)}
	detail, err := svc.Detail(req)
	assert.NoError(t, err)
	assert.Equal(t, "Detail Test", detail.Name)
	assert.Equal(t, id, detail.ID)
}

func TestVisualizationServiceIntegration_Detail_Completeness(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	nodeType := "panel"
	visType := "dashboard"
	mobileLayout := true
	createReq := &visualization.SaveRequest{
		Name:            "Detail Complete",
		NodeType:        &nodeType,
		Type:            &visType,
		CanvasStyleData: strPtr("{\"theme\":\"dark\"}"),
		ComponentData:   strPtr("{\"views\":[1,2]}"),
		MobileLayout:    &mobileLayout,
		ContentID:       strPtr("content-complete"),
		CheckVersion:    strPtr("v-check-1"),
	}
	id, err := svc.Save(createReq, "tester")
	assert.NoError(t, err)

	detail, err := svc.Detail(&visualization.DetailRequest{ID: visualization.FlexInt(id)})
	assert.NoError(t, err)
	assert.Equal(t, "Detail Complete", detail.Name)
	if assert.NotNil(t, detail.NodeType) {
		assert.Equal(t, "panel", *detail.NodeType)
	}
	if assert.NotNil(t, detail.Type) {
		assert.Equal(t, "dashboard", *detail.Type)
	}
	if assert.NotNil(t, detail.CanvasStyleData) {
		assert.Equal(t, "{\"theme\":\"dark\"}", *detail.CanvasStyleData)
	}
	if assert.NotNil(t, detail.ComponentData) {
		assert.Equal(t, "{\"views\":[1,2]}", *detail.ComponentData)
	}
	if assert.NotNil(t, detail.MobileLayout) {
		assert.True(t, *detail.MobileLayout)
	}
	if assert.NotNil(t, detail.ContentID) {
		assert.Equal(t, "content-complete", *detail.ContentID)
	}
	if assert.NotNil(t, detail.CheckVersion) {
		assert.Equal(t, "v-check-1", *detail.CheckVersion)
	}
}

func TestVisualizationServiceIntegration_Detail_NotFound(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	req := &visualization.DetailRequest{ID: visualization.FlexInt(99999)}
	_, err := svc.Detail(req)
	assert.Error(t, err)
}

func TestVisualizationServiceIntegration_List(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create multiple items
	nodeType := "panel"
	for i := 0; i < 5; i++ {
		_, err := svc.Save(&visualization.SaveRequest{
			Name:     "List Test " + string(rune('A'+i)),
			NodeType: &nodeType,
		}, "tester")
		assert.NoError(t, err)
	}

	// List with pagination
	req := &visualization.ListRequest{
		Current: 1,
		Size:    10,
	}
	resp, err := svc.List(req)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, resp.Total, int64(5))
	assert.Len(t, resp.List, 5)
}

func TestVisualizationServiceIntegration_List_WithKeyword(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create items
	nodeType := "panel"
	svc.Save(&visualization.SaveRequest{Name: "Alpha Dashboard", NodeType: &nodeType}, "tester")
	svc.Save(&visualization.SaveRequest{Name: "Beta Dashboard", NodeType: &nodeType}, "tester")
	svc.Save(&visualization.SaveRequest{Name: "Gamma Report", NodeType: &nodeType}, "tester")

	// Search with keyword
	keyword := "Dashboard"
	req := &visualization.ListRequest{
		Keyword: &keyword,
		Current: 1,
		Size:    10,
	}
	resp, err := svc.List(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)
}

func TestVisualizationServiceIntegration_DeleteLogic(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create first
	nodeType := "panel"
	createReq := &visualization.SaveRequest{
		Name:     "To Delete",
		NodeType: &nodeType,
	}
	id, err := svc.Save(createReq, "tester")
	assert.NoError(t, err)

	// Delete (logic)
	err = svc.DeleteLogic(id, "deleter")
	assert.NoError(t, err)

	// Verify deleted (should not be found)
	_, err = repo.GetByID(id)
	assert.Error(t, err) // Should be record not found
}

func TestVisualizationServiceIntegration_DeleteLogic_NotFound(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Delete non-existent should not error (just updates 0 rows)
	err := svc.DeleteLogic(99999, "deleter")
	assert.NoError(t, err)
}

func TestVisualizationServiceIntegration_Update_AllFields(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create first
	nodeType := "panel"
	createReq := &visualization.SaveRequest{
		Name:            "Original Name",
		NodeType:        &nodeType,
		CanvasStyleData: strPtr("{\"style\":\"original\"}"),
		ComponentData:   strPtr("{\"components\":[]}"),
	}
	id, err := svc.Save(createReq, "creator")
	require.NoError(t, err)

	// Update all fields
	newPID := int64(0)
	newType := "dashboard"
	newStatus := 0
	mobileLayout := true
	updateReq := &visualization.UpdateRequest{
		ID:              id,
		Name:            strPtr("Updated Name"),
		PID:             &newPID,
		Type:            &newType,
		CanvasStyleData: strPtr("{\"style\":\"updated\"}"),
		ComponentData:   strPtr("{\"components\":[1,2,3]}"),
		MobileLayout:    &mobileLayout,
		Status:          &newStatus,
	}
	err = svc.Update(updateReq, "updater")
	require.NoError(t, err)

	// Verify updated
	detail, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", detail.Name)
	require.NotNil(t, detail.Type)
	assert.Equal(t, newType, *detail.Type)
	require.NotNil(t, detail.Status)
	assert.Equal(t, newStatus, *detail.Status)
}

func TestVisualizationServiceIntegration_List_WithPaging(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create multiple items
	nodeType := "panel"
	for i := 0; i < 15; i++ {
		_, err := svc.Save(&visualization.SaveRequest{
			Name:     "Dashboard " + string(rune('A'+i)),
			NodeType: &nodeType,
		}, "tester")
		require.NoError(t, err)
	}

	// Test pagination - page 1
	current := 1
	size := 10
	resp, err := svc.List(&visualization.ListRequest{Current: current, Size: size})
	require.NoError(t, err)
	assert.Equal(t, int64(15), resp.Total)
	assert.Equal(t, 10, len(resp.List))
	assert.Equal(t, 1, resp.Current)
	assert.Equal(t, 10, resp.Size)

	// Test pagination - page 2
	current = 2
	resp, err = svc.List(&visualization.ListRequest{Current: current, Size: size})
	require.NoError(t, err)
	assert.Equal(t, 5, len(resp.List))
}

func TestVisualizationServiceIntegration_List_EdgeCases(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create some items
	nodeType := "panel"
	for i := 0; i < 5; i++ {
		_, err := svc.Save(&visualization.SaveRequest{
			Name:     "Dashboard " + string(rune('A'+i)),
			NodeType: &nodeType,
		}, "tester")
		require.NoError(t, err)
	}

	t.Run("list with zero current", func(t *testing.T) {
		resp, err := svc.List(&visualization.ListRequest{Current: 0, Size: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Current) // Should default to 1
	})

	t.Run("list with negative current", func(t *testing.T) {
		resp, err := svc.List(&visualization.ListRequest{Current: -1, Size: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Current) // Should default to 1
	})

	t.Run("list with zero size", func(t *testing.T) {
		resp, err := svc.List(&visualization.ListRequest{Current: 1, Size: 0})
		require.NoError(t, err)
		assert.Equal(t, 10, resp.Size) // Should default to 10
	})

	t.Run("list with negative size", func(t *testing.T) {
		resp, err := svc.List(&visualization.ListRequest{Current: 1, Size: -5})
		require.NoError(t, err)
		assert.Equal(t, 10, resp.Size) // Should default to 10
	})
}

func TestVisualizationServiceIntegration_InteractiveTree(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	folderType := "folder"
	panelType := "panel"
	dashboardType := "dashboard"
	dataVType := "dataV"
	mobileLayout := true

	dashboardFolderID, err := svc.Save(&visualization.SaveRequest{
		Name:     "Dashboard Root",
		NodeType: &folderType,
		Type:     &dashboardType,
	}, "tester")
	require.NoError(t, err)

	dashboardPanelID, err := svc.Save(&visualization.SaveRequest{
		Name:         "Dashboard Child",
		PID:          &dashboardFolderID,
		NodeType:     &panelType,
		Type:         &dashboardType,
		MobileLayout: &mobileLayout,
	}, "tester")
	require.NoError(t, err)

	dataVID, err := svc.Save(&visualization.SaveRequest{
		Name:     "Big Screen A",
		NodeType: &panelType,
		Type:     &dataVType,
	}, "tester")
	require.NoError(t, err)

	published := 1
	err = svc.Update(&visualization.UpdateRequest{ID: dashboardPanelID, Status: &published}, "tester")
	require.NoError(t, err)
	err = svc.Update(&visualization.UpdateRequest{ID: dataVID, Status: &published}, "tester")
	require.NoError(t, err)

	deleted := true
	err = repo.Create(&visualization.DataVisualizationInfo{Name: "Deleted Dashboard", NodeType: &panelType, Type: &dashboardType, DeleteFlag: &deleted})
	require.NoError(t, err)

	dashboardItems, err := svc.InteractiveTree("dashboard")
	require.NoError(t, err)
	assert.Len(t, dashboardItems, 2)
	for _, item := range dashboardItems {
		if assert.NotNil(t, item.Type) {
			assert.Equal(t, "dashboard", *item.Type)
		}
		assert.NotEqual(t, "Deleted Dashboard", item.Name)
	}

	dataVItems, err := svc.InteractiveTree("dataV")
	require.NoError(t, err)
	assert.Len(t, dataVItems, 1)
	assert.Equal(t, "Big Screen A", dataVItems[0].Name)
	if assert.NotNil(t, dataVItems[0].Type) {
		assert.Equal(t, "dataV", *dataVItems[0].Type)
	}
}
