package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/domain/visualization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ==================== Visualization Repo Round 5 ====================

func TestRound5Repo_Vis_QuerySizeCapAndPagination(t *testing.T) {
	repo, _ := setupVisualizationRepositoryTest(t)
	now := time.Now().UnixMilli()
	dvType := visualization.TypeDashboard

	for i := 0; i < 5; i++ {
		item := newVisualizationInfo(int64(5000+i), "paginated-item", dvType, now+int64(i))
		require.NoError(t, repo.Create(item))
	}

	list, total, err := repo.Query(&visualization.ListRequest{Current: 1, Size: 200})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	require.Len(t, list, 5)

	list, total, err = repo.Query(&visualization.ListRequest{Current: 99, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Empty(t, list)
}

func TestRound5Repo_Vis_ListAllByTypes_EmptyTypes(t *testing.T) {
	repo, _ := setupVisualizationRepositoryTest(t)
	now := time.Now().UnixMilli()
	require.NoError(t, repo.Create(newVisualizationInfo(5100, "any-type", visualization.TypeDashboard, now)))

	list, err := repo.ListAllByTypes([]string{})
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestRound5Repo_Vis_ListAllByTypesBatch_LimitZero(t *testing.T) {
	repo, _ := setupVisualizationRepositoryTest(t)
	now := time.Now().UnixMilli()
	dvType := visualization.TypeDashboard
	require.NoError(t, repo.Create(newVisualizationInfo(5200, "batch-limit", dvType, now)))
	require.NoError(t, repo.Create(newVisualizationInfo(5201, "batch-limit-2", dvType, now+1)))

	list, err := repo.ListAllByTypesBatch([]string{dvType}, 0, 0, nil)
	require.NoError(t, err)
	require.Len(t, list, 2)

	list, err = repo.ListAllByTypesBatch([]string{dvType}, 5200, 0, nil)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(5201), list[0].ID)
}

func TestRound5Repo_Vis_FindRecent_AscSort(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)
	status := 1
	leaf := visualization.NodeTypeLeaf
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{
		ID: 5300, Name: "Asc Test", NodeType: &leaf,
		Type: strPtrRound5("panel"), Status: &status,
	}).Error)
	require.NoError(t, db.Create(&visualizationCoreOptRecent{UID: 5301, ResourceID: 5300, Time: 100}).Error)

	results, err := repo.FindRecent(5301, &visualization.WorkbranchQueryRequest{Asc: true})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Asc Test", results[0].Name)
}

func TestRound5Repo_Vis_FindRecent_PanelFilter(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)
	status := 1
	leaf := visualization.NodeTypeLeaf
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{
		ID: 5310, Name: "Panel Only", NodeType: &leaf,
		Type: strPtrRound5("panel"), Status: &status,
	}).Error)
	require.NoError(t, db.Create(&visualizationCoreOptRecent{UID: 5311, ResourceID: 5310, Time: 50}).Error)

	results, err := repo.FindRecent(5311, &visualization.WorkbranchQueryRequest{Type: visualization.ResourceAliasPanel})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Panel Only", results[0].Name)
}

func TestRound5Repo_Vis_CopyChartViews_DefaultResourceTable(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)
	sourceSceneID := int64(5400)
	targetSceneID := int64(5401)
	copyID := int64(54000)

	require.NoError(t, db.Create([]*chart.CoreChartView{
		newChartView(5401, sourceSceneID, "default-copy-source"),
	}).Error)

	require.NoError(t, repo.CopyChartViews(sourceSceneID, targetSceneID, copyID, ""))

	mapping, err := repo.GetCopiedChartViewMapping(copyID)
	require.NoError(t, err)
	require.Len(t, mapping, 1)
	assert.Equal(t, int64(54001+5400), mapping[5401])

	views, err := repo.GetChartViewsBySceneID(targetSceneID)
	require.NoError(t, err)
	require.Len(t, views, 1)
}

func TestRound5Repo_Vis_FindLinkagesAndJumps_EmptyResults(t *testing.T) {
	repo, _ := setupVisualizationRepositoryTest(t)

	linkages, err := repo.FindLinkagesByDvID(99999)
	require.NoError(t, err)
	assert.Empty(t, linkages)

	fields, err := repo.FindLinkageFieldsByDvID(99999)
	require.NoError(t, err)
	assert.Empty(t, fields)

	jumps, err := repo.FindLinkJumpsByDvID(99999)
	require.NoError(t, err)
	assert.Empty(t, jumps)

	infos, err := repo.FindLinkJumpInfosByDvID(99999)
	require.NoError(t, err)
	assert.Empty(t, infos)

	targets, err := repo.FindLinkJumpTargetViewInfosByDvID(99999)
	require.NoError(t, err)
	assert.Empty(t, targets)
}

func TestRound5Repo_Vis_FindDatasetRelated_EmptyIDs(t *testing.T) {
	repo, _ := setupVisualizationRepositoryTest(t)

	groups, err := repo.FindDatasetGroupsByIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, groups)

	tables, err := repo.FindDatasetTablesByGroupIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, tables)

	fields, err := repo.FindDatasetTableFieldsByGroupIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, fields)

	dss, err := repo.FindDatasourcesByGroupIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, dss)

	tasks, err := repo.FindDatasourceTasksByGroupIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, tasks)

	charts, err := repo.FindChartViewsByIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, charts)
}

func TestRound5Repo_Vis_GetSnapshotChartViewsBySceneID_Empty(t *testing.T) {
	repo, _ := setupVisualizationRepositoryTest(t)
	views, err := repo.GetSnapshotChartViewsBySceneID(99999)
	require.NoError(t, err)
	assert.Empty(t, views)
}

func TestRound5Repo_Vis_GetChartViewsBySceneID_Empty(t *testing.T) {
	repo, _ := setupVisualizationRepositoryTest(t)
	views, err := repo.GetChartViewsBySceneID(99999)
	require.NoError(t, err)
	assert.Empty(t, views)
}

func TestRound5Repo_Vis_CountByNameAndPID_WithExclude(t *testing.T) {
	repo, _ := setupVisualizationRepositoryTest(t)
	now := time.Now().UnixMilli()
	require.NoError(t, repo.Create(newVisualizationInfo(5500, "unique-name", visualization.TypeDashboard, now)))

	pid := int64(0)
	count, err := repo.CountByNameAndPID("unique-name", &pid, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = repo.CountByNameAndPID("unique-name", &pid, int64PtrRound5(5500))
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// ==================== Resource Perm Repo Round 5 ====================

func TestRound5Repo_Perm_CheckPermissionConsistency_UserLimitSkip(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	for i := 0; i < 10001; i++ {
		require.NoError(t, repo.db.Create(&user.SysUser{
			UserID:   int64(20000 + i),
			Username: "bulk-user",
			Status:   user.StatusEnabled,
			DelFlag:  user.DelFlagNormal,
		}).Error)
	}

	result, err := repo.CheckPermissionConsistency()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 10001, result.UserCount)
	require.NotEmpty(t, result.Inconsistencies)
	assert.Contains(t, result.Inconsistencies[0].Description, "skipped")
}

func TestRound5Repo_Perm_CompareViews_BothDirections(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	userView := map[string]bool{
		"101:dataset:view": true,
		"202:screen:edit":  true,
	}
	resourceView := map[string]bool{
		"202:screen:edit":    true,
		"303:dashboard:view": true,
	}
	resourceMeta := map[string]resourceRow{
		"303:dashboard:view": {ResourceID: scopedResourceID(permission.ResourceTypeDashboard, 303), ResourceType: permission.ResourceTypeDashboard},
	}

	inconsistencies := repo.compareViews(userView, resourceView, resourceMeta)
	assert.Len(t, inconsistencies, 2)
}

func TestRound5Repo_Perm_CountActiveUsers_WithOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 6001, Username: "org-user-a", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 6002, Username: "org-user-b", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 6001, RoleID: 7001, OrgID: 42}).Error)

	count, err := repo.countActiveUsers(42)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = repo.countActiveUsers(99)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestRound5Repo_Perm_GetResourcePermIDs_Internal(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	require.NoError(t, repo.RegisterResource(7700, "test-res", permission.ResourceTypeDataset, nil))
	require.NoError(t, repo.ReplaceResourcePermissions(7700, permission.ResourceTypeDataset, []int64{301, 302}))

	scopedID := scopedResourceID(permission.ResourceTypeDataset, 7700)
	permIDs, err := repo.getResourcePermIDs(scopedID)
	require.NoError(t, err)
	assert.Len(t, permIDs, 2)
}

func TestRound5Repo_Perm_QueryDirectUserPerms(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	perm := &permission.SysPerm{PermName: "Test", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(perm).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 8001, Username: "direct-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 8001, PermID: perm.PermID, Status: 1, DelFlag: 0}).Error)

	rows, err := repo.queryDirectUserPerms([]int64{perm.PermID}, 0)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, int64(8001), rows[0].UserID)

	rows, err = repo.queryDirectUserPerms([]int64{perm.PermID}, 99)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound5Repo_Perm_QueryRoleUserPerms(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	perm := &permission.SysPerm{PermName: "Test2", PermKey: "dataset:edit", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(perm).Error)
	require.NoError(t, repo.db.Create(&role.SysRole{RoleID: 9001, RoleName: "TestRole", Status: role.StatusEnabled}).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 9002, Username: "role-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&permission.SysRolePerm{RoleID: 9001, PermID: perm.PermID}).Error)
	require.NoError(t, repo.db.Create(&user.SysUserRole{UserID: 9002, RoleID: 9001, OrgID: 5}).Error)

	rows, err := repo.queryRoleUserPerms([]int64{perm.PermID}, 0)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, int64(9002), rows[0].UserID)

	rows, err = repo.queryRoleUserPerms([]int64{perm.PermID}, 5)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	rows, err = repo.queryRoleUserPerms([]int64{perm.PermID}, 99)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound5Repo_Perm_RevokePermFromUser_DelegatesToInOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	require.NoError(t, repo.GrantPermToUser(10001, 10002, "admin"))

	hasPerm, err := repo.CheckUserPermission(10001, 10002)
	require.NoError(t, err)
	assert.True(t, hasPerm)

	require.NoError(t, repo.RevokePermFromUser(10001, 10002))
	hasPerm, err = repo.CheckUserPermission(10001, 10002)
	require.NoError(t, err)
	assert.False(t, hasPerm)
}

func TestRound5Repo_Perm_GetUserResourcesByOrg_NegativeOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	perm := &permission.SysPerm{PermName: "NegOrg", PermKey: "dataset:list", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(perm).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 11001, PermID: perm.PermID, Status: 1, DelFlag: 0}).Error)

	results, err := repo.GetUserResourcesByOrg(11001, permission.ResourceTypeDataset, -1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "direct", results[0].SourceType)
}

func TestRound5Repo_Perm_GetResourceUsersByOrg_NegativeOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	perm := &permission.SysPerm{PermName: "ResUser", PermKey: "dashboard:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(perm).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 11002, Username: "res-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 11002, PermID: perm.PermID, Status: 1, DelFlag: 0}).Error)
	require.NoError(t, repo.ReplaceResourcePermissions(11003, permission.ResourceTypeDashboard, []int64{perm.PermID}))

	results, err := repo.GetResourceUsersByOrg(11003, permission.ResourceTypeDashboard, -1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "direct", results[0].SourceType)
}

func TestRound5Repo_Perm_CheckPermissionConsistencyByOrg_NegativeOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	result, err := repo.CheckPermissionConsistencyByOrg(-1)
	require.NoError(t, err)
	assert.True(t, result.Consistent)
}

func TestRound5Repo_Perm_RegisterResource_EmptyName_WithPositiveParent(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	parentID := int64(10)
	require.NoError(t, repo.RegisterResource(12001, "", permission.ResourceTypeScreen, &parentID))

	permIDs, exists, err := repo.GetResourcePermissionIDs(12001, permission.ResourceTypeScreen)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Empty(t, permIDs)
}

func TestRound5Repo_Perm_RegisterResource_NilParentUpdate(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	pid := int64(55)
	require.NoError(t, repo.RegisterResource(12002, "Initial", permission.ResourceTypeDataset, &pid))
	require.NoError(t, repo.RegisterResource(12002, "Updated", permission.ResourceTypeDataset, nil))

	var resource permission.SysResource
	err := repo.db.Where("resource_id = ? AND resource_type = ?",
		scopedResourceID(permission.ResourceTypeDataset, 12002), permission.ResourceTypeDataset).First(&resource).Error
	require.NoError(t, err)
	assert.Equal(t, "Updated", resource.ResourceName)
}

func TestRound5Repo_Perm_ApplyGroupPermissions_NegativeInput(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	err := repo.ApplyGroupPermissions(-1, 1, permission.ResourceTypeDataset)
	require.NoError(t, err)

	err = repo.ApplyGroupPermissions(1, 0, permission.ResourceTypeDataset)
	require.NoError(t, err)

	err = repo.ApplyGroupPermissions(1, 1, "")
	require.NoError(t, err)
}

func TestRound5Repo_Perm_ExtractResourceID_EdgeCases(t *testing.T) {
	valid := extractResourceID(resourceRow{
		ResourceID:   scopedResourceID(permission.ResourceTypeDashboard, 42),
		ResourceType: permission.ResourceTypeDashboard,
	})
	assert.Equal(t, int64(42), valid)

	zero := extractResourceID(resourceRow{ResourceID: 0, ResourceType: permission.ResourceTypeDashboard})
	assert.Zero(t, zero)

	neg := extractResourceID(resourceRow{ResourceID: 5, ResourceType: permission.ResourceTypeDashboard})
	assert.Zero(t, neg)
}

// ==================== Dataset Repo Round 5 ====================

func TestRound5Repo_Ds_ListGroups_NoKeyword(t *testing.T) {
	repo, db := setupDatasetRepositoryTest(t)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 10001, Name: "Group A"}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 10002, Name: "Group B"}).Error)

	groups, err := repo.ListGroups(nil)
	require.NoError(t, err)
	require.Len(t, groups, 2)
}

func TestRound5Repo_Ds_ListGroupsBatch_NoFilters(t *testing.T) {
	repo, db := setupDatasetRepositoryTest(t)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 10010, Name: "Batch A"}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 10011, Name: "Batch B"}).Error)

	groups, err := repo.ListGroupsBatch(nil, 0, 0)
	require.NoError(t, err)
	require.Len(t, groups, 2)

	groups, err = repo.ListGroupsBatch(nil, 10010, 0)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, int64(10011), groups[0].ID)

	groups, err = repo.ListGroupsBatch(nil, 0, 1)
	require.NoError(t, err)
	require.Len(t, groups, 1)
}

func TestRound5Repo_Ds_PreviewSQL_LimitEdgeCases(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	rows, err := repo.PreviewSQL("SELECT id FROM preview_rows ORDER BY id ASC", 0)
	require.NoError(t, err)
	assert.Len(t, rows, 3)

	rows, err = repo.PreviewSQL("SELECT id FROM preview_rows ORDER BY id ASC", 999)
	require.NoError(t, err)
	assert.Len(t, rows, 3)
}

func TestRound5Repo_Ds_PreviewRows_LimitEdgeCases(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	rows, err := repo.PreviewRows("preview_rows", 0)
	require.NoError(t, err)
	assert.Len(t, rows, 3)

	rows, err = repo.PreviewRows("preview_rows", 999)
	require.NoError(t, err)
	assert.Len(t, rows, 3)
}

func TestRound5Repo_Ds_CountRows_InvalidTable(t *testing.T) {
	repo := &DatasetRepository{}
	_, err := repo.CountRows("invalid-table;drop")
	require.Error(t, err)
}

func TestRound5Repo_Ds_FindNearestGroupIDInWindow_InvalidInput(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	id, err := repo.FindNearestGroupIDInWindow(0, 100)
	require.NoError(t, err)
	assert.Nil(t, id)

	id, err = repo.FindNearestGroupIDInWindow(100, 0)
	require.NoError(t, err)
	assert.Nil(t, id)

	id, err = repo.FindNearestGroupIDInWindow(-1, -1)
	require.NoError(t, err)
	assert.Nil(t, id)
}

func TestRound5Repo_Ds_DeleteFieldByIDAndChartID_InvalidInput(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	deleted, err := repo.DeleteFieldByIDAndChartID(0, 1)
	require.Error(t, err)
	assert.False(t, deleted)

	deleted, err = repo.DeleteFieldByIDAndChartID(1, 0)
	require.Error(t, err)
	assert.False(t, deleted)
}

func TestRound5Repo_Ds_DeleteFieldsByChartID_InvalidInput(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	rows, err := repo.DeleteFieldsByChartID(0)
	require.Error(t, err)
	assert.Zero(t, rows)
}

func TestRound5Repo_Ds_QueryDistinctValues_InvalidTable(t *testing.T) {
	repo := &DatasetRepository{}
	_, err := repo.QueryDistinctValues("bad-table;", "col", nil, 0)
	require.Error(t, err)
}

func TestRound5Repo_Ds_QueryDistinctObjectValues_InvalidTable(t *testing.T) {
	repo := &DatasetRepository{}
	_, err := repo.QueryDistinctObjectValues("bad-table;", nil, nil, "", "", "", "", 0)
	require.Error(t, err)
}

func TestRound5Repo_Ds_QueryFieldTreeValues_InvalidTable(t *testing.T) {
	repo := &DatasetRepository{}
	_, err := repo.QueryFieldTreeValues("bad-table;", nil, nil, 0)
	require.Error(t, err)
}

func TestRound5Repo_Ds_PreviewRowsWithFilter_InvalidTable(t *testing.T) {
	repo := &DatasetRepository{}
	_, err := repo.PreviewRowsWithFilter("bad-table;", "", "", nil, 0)
	require.Error(t, err)
}

func TestRound5Repo_Ds_FindPrimaryTableName_NoTable(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)
	_, err := repo.FindPrimaryTableName(99999)
	require.Error(t, err)
}

func TestRound5Repo_Ds_ListFieldsByDsIds_Empty(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)
	fields, err := repo.ListFieldsByDsIds(nil)
	require.NoError(t, err)
	assert.Empty(t, fields)
}

func TestRound5Repo_Ds_ListChartViewsByDatasetGroupID_WithRealTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}))
	require.NoError(t, db.Exec("CREATE TABLE preview_rows (id INTEGER PRIMARY KEY, category TEXT, city TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO preview_rows (id, category, city) VALUES (1, 'A', 'Shanghai')").Error)
	require.NoError(t, db.Exec("CREATE TABLE core_chart_view (id INTEGER PRIMARY KEY, table_id INTEGER)").Error)

	repo := NewDatasetRepository(db)
	groupName := "ChartGroup"
	tableName := "preview_rows"
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 10501, Name: groupName}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 10502, DatasetGroupID: 10501, PhysicalTable: &tableName}).Error)
	require.NoError(t, db.Exec("INSERT INTO core_chart_view (id, table_id) VALUES (1051, 10502)").Error)

	views, err := repo.ListChartViewsByDatasetGroupID(10501)
	require.NoError(t, err)
	require.Len(t, views, 1)
}

func TestRound5Repo_Ds_NilRepoGuards(t *testing.T) {
	var repo *DatasetRepository

	groups, err := repo.ListGroups(nil)
	require.Error(t, err)
	assert.Nil(t, groups)

	groups, err = repo.ListGroupsBatch(nil, 0, 10)
	require.Error(t, err)
	assert.Nil(t, groups)
}

func TestRound5Repo_Ds_BuildSelectParts(t *testing.T) {
	parts, err := buildSelectParts([]dataset.EnumObjectColumn{
		{Column: "city", Alias: "label"},
		{Column: "amount", Alias: "value"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"`city` AS `label`", "`amount` AS `value`"}, parts)

	_, err = buildSelectParts([]dataset.EnumObjectColumn{{Column: "city", Alias: "   "}})
	require.Error(t, err)
}

func TestRound5Repo_Ds_BuildEnumWhereClauses_WithSearch(t *testing.T) {
	args := make([]interface{}, 0)
	clauses := buildEnumWhereClauses(
		[]dataset.EnumFilterClause{{Column: "category", Values: []string{"A"}}},
		"city",
		"Shang",
		&args,
	)
	require.Len(t, clauses, 2)
	assert.Len(t, args, 2)
}

func TestRound5Repo_Ds_BuildEnumWhereClauses_NoSearch(t *testing.T) {
	args := make([]interface{}, 0)
	clauses := buildEnumWhereClauses(nil, "", "", &args)
	assert.Empty(t, clauses)

	clauses = buildEnumWhereClauses(nil, "city", "", &args)
	assert.Empty(t, clauses)

	clauses = buildEnumWhereClauses(nil, "", "text", &args)
	assert.Empty(t, clauses)
}

func TestRound5Repo_Ds_BuildFilterWhereParts_EmptyFilter(t *testing.T) {
	args := make([]interface{}, 0)
	parts := buildFilterWhereParts([]dataset.EnumFilterClause{{Column: "", Values: nil}}, &args)
	assert.Empty(t, parts)
	assert.Empty(t, args)

	parts = buildFilterWhereParts([]dataset.EnumFilterClause{{Column: "city", Values: []string{}}}, &args)
	assert.Empty(t, parts)
}

// ==================== Visualization Repo missing-table errors ====================

func TestRound5Repo_Vis_FindRecent_MissingTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	repo := NewVisualizationRepository(db)

	_, err = repo.FindRecent(1, nil)
	require.Error(t, err)
}

func TestRound5Repo_Vis_CopyChartViews_MissingTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	repo := NewVisualizationRepository(db)

	err = repo.CopyChartViews(1, 2, 100, "")
	require.Error(t, err)

	err = repo.CopyLinkages(100)
	require.Error(t, err)

	err = repo.CopyLinkageFields(100)
	require.Error(t, err)

	err = repo.CopyLinkJumps(100)
	require.Error(t, err)

	err = repo.CopyLinkJumpInfos(100)
	require.Error(t, err)

	err = repo.CopyLinkJumpTargetInfos(100)
	require.Error(t, err)
}

func TestRound5Repo_Vis_GetCopiedChartViewMapping_MissingTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	repo := NewVisualizationRepository(db)

	_, err = repo.GetCopiedChartViewMapping(100)
	require.Error(t, err)
}

// ==================== Additional Resource Perm edge cases ====================

func TestRound5Repo_Perm_QueryUserView_WithOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	perm := &permission.SysPerm{PermName: "QUV", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(perm).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 13001, Username: "quv-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 13001, OrgID: ptrInt64Value(77), PermID: perm.PermID, Status: 1, DelFlag: 0}).Error)

	view, err := repo.queryUserView(77)
	require.NoError(t, err)
	assert.True(t, view["13001:dataset:view"])

	view, err = repo.queryUserView(88)
	require.NoError(t, err)
	assert.False(t, view["13001:dataset:view"])
}

func TestRound5Repo_Perm_QueryUserView_NoOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	perm := &permission.SysPerm{PermName: "QUV2", PermKey: "screen:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(perm).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 13002, Username: "quv-user2", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 13002, PermID: perm.PermID, Status: 1, DelFlag: 0}).Error)

	view, err := repo.queryUserView(0)
	require.NoError(t, err)
	assert.True(t, view["13002:screen:view"])
}

func TestRound5Repo_Perm_QueryResourceView_WithOrg(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	perm := &permission.SysPerm{PermName: "QRV", PermKey: "dataset:edit", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(perm).Error)
	require.NoError(t, repo.db.Create(&user.SysUser{UserID: 13003, Username: "qrv-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 13003, OrgID: ptrInt64Value(88), PermID: perm.PermID, Status: 1, DelFlag: 0}).Error)
	require.NoError(t, repo.ReplaceResourcePermissions(13004, permission.ResourceTypeDataset, []int64{perm.PermID}))

	view, meta, count, err := repo.queryResourceView(88)
	require.NoError(t, err)
	assert.True(t, view["13003:dataset:edit"])
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, meta)
}

func TestRound5Repo_Perm_GetUserResources_PrefixFiltering(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	datasetPerm := &permission.SysPerm{PermName: "DS View", PermKey: "dataset:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	screenPerm := &permission.SysPerm{PermName: "Screen View", PermKey: "screen:view", PermType: permission.PermTypeData, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal}
	require.NoError(t, repo.db.Create(datasetPerm).Error)
	require.NoError(t, repo.db.Create(screenPerm).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 14001, PermID: datasetPerm.PermID, Status: 1, DelFlag: 0}).Error)
	require.NoError(t, repo.db.Create(&permission.SysUserPerm{UserID: 14001, PermID: screenPerm.PermID, Status: 1, DelFlag: 0}).Error)

	results, err := repo.GetUserResources(14001, permission.ResourceTypeDataset)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "dataset:view", results[0].PermKey)

	results, err = repo.GetUserResources(14001, permission.ResourceTypeScreen)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "screen:view", results[0].PermKey)
}

func TestRound5Repo_Perm_GetResourceUsers_GovernedWithEmptyPerms(t *testing.T) {
	repo := setupResourcePermissionRepositoryTest(t)
	require.NoError(t, repo.RegisterResource(15001, "empty-perm-res", permission.ResourceTypeDataset, nil))

	results, err := repo.GetResourceUsers(15001, permission.ResourceTypeDataset)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestRound5Repo_Vis_FindDatasetTablesByGroupIDs_WithData(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)
	nodeType := dataset.NodeTypeDataset
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 20001, Name: "DsGroup", NodeType: &nodeType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 20002, DatasetGroupID: 20001}).Error)

	tables, err := repo.FindDatasetTablesByGroupIDs([]int64{20001})
	require.NoError(t, err)
	require.Len(t, tables, 1)
	assert.Equal(t, int64(20002), mapInt64Value(t, tables[0], "id"))
}

func TestRound5Repo_Vis_FindDatasetTableFieldsByGroupIDs_WithData(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)
	originName := "field1"
	fieldName := "Field1"
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 20003, DatasetGroupID: 20004, OriginName: &originName, Name: &fieldName}).Error)

	fields, err := repo.FindDatasetTableFieldsByGroupIDs([]int64{20004})
	require.NoError(t, err)
	require.Len(t, fields, 1)
	assert.Equal(t, int64(20003), mapInt64Value(t, fields[0], "id"))
}

func TestRound5Repo_Vis_FindDatasourcesByGroupIDs_WithData(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)
	dsStatus := "ready"
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 20005, Name: "DS5", Type: "mysql", Status: &dsStatus}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 20006, DatasetGroupID: 20007, DatasourceID: int64PtrRound5(20005)}).Error)

	dss, err := repo.FindDatasourcesByGroupIDs([]int64{20007})
	require.NoError(t, err)
	require.Len(t, dss, 1)
	assert.Equal(t, int64(20005), mapInt64Value(t, dss[0], "id"))
}

func TestRound5Repo_Vis_FindDatasourceTasksByGroupIDs_WithData(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)
	dsStatus := "ready"
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 20008, Name: "DS8", Type: "mysql", Status: &dsStatus}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 20009, DatasetGroupID: 20010, DatasourceID: int64PtrRound5(20008)}).Error)
	require.NoError(t, db.Create(&visualizationCoreDatasourceTask{ID: 20011, DsID: 20008, Name: "task-20011"}).Error)

	tasks, err := repo.FindDatasourceTasksByGroupIDs([]int64{20010})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, int64(20011), mapInt64Value(t, tasks[0], "id"))
}

func strPtrRound5(v string) *string { return &v }
func int64PtrRound5(v int64) *int64 { return &v }
