package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/domain/visualization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRowPermissionRepoTest(t *testing.T) (*RowPermissionRepository, *ColumnPermissionRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&permission.DataPermRow{}, &permission.DataPermColumn{}))

	return NewRowPermissionRepository(db), NewColumnPermissionRepository(db), db
}

func newRound3RowPermission(datasetID int64, targetType string, targetID int64, expr string) *permission.DataPermRow {
	creator := "creator"
	createdAt := time.Unix(1710000300, 0)
	return &permission.DataPermRow{
		DatasetID:      datasetID,
		AuthTargetType: targetType,
		AuthTargetID:   targetID,
		ExpressionTree: expr,
		Status:         1,
		CreateBy:       &creator,
		CreateTime:     &createdAt,
	}
}

func newRound3ColumnPermission(datasetID int64, fieldName, permType, maskRule string) *permission.DataPermColumn {
	creator := "creator"
	createdAt := time.Unix(1710000400, 0)
	return &permission.DataPermColumn{
		DatasetID:  datasetID,
		FieldName:  fieldName,
		PermType:   permType,
		MaskRule:   maskRule,
		Status:     1,
		CreateBy:   &creator,
		CreateTime: &createdAt,
	}
}

func TestRound3_VisualizationRepo_SuccessAndErrorPaths(t *testing.T) {
	t.Run("success wrappers", func(t *testing.T) {
		repo, db := setupVisualizationRepositoryTest(t)
		require.Same(t, db, repo.db)

		now := time.Now().UnixMilli()
		dvType := visualization.TypeDashboard
		item := newVisualizationInfo(0, "round3-dashboard", dvType, now)
		require.NoError(t, repo.Create(item))

		loaded, err := repo.GetByID(item.ID)
		require.NoError(t, err)
		assert.Equal(t, "round3-dashboard", loaded.Name)

		item.Name = "round3-dashboard-updated"
		require.NoError(t, repo.Update(item))

		snapshot := newSnapshotChartView(77, item.ID, "snapshot-round3")
		require.NoError(t, repo.SaveSnapshotChartView(snapshot))

		chartA := newChartView(501, item.ID, "chart-a")
		chartB := newChartView(502, item.ID, "chart-b")
		require.NoError(t, db.Create([]*chart.CoreChartView{chartA, chartB}).Error)

		views, err := repo.GetChartViewsBySceneID(item.ID)
		require.NoError(t, err)
		require.Len(t, views, 2)

		snapshotViews, err := repo.GetSnapshotChartViewsBySceneID(item.ID)
		require.NoError(t, err)
		require.Len(t, snapshotViews, 1)

		rootCount, err := repo.CountByNameAndPID("round3-dashboard-updated", nil, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(1), rootCount)

		allItems, err := repo.ListAllByTypes(nil)
		require.NoError(t, err)
		require.NotEmpty(t, allItems)

		batchItems, err := repo.ListAllByTypesBatch([]string{dvType}, 0, 10, nil)
		require.NoError(t, err)
		require.Len(t, batchItems, 1)

		require.NoError(t, repo.DeleteLogic(item.ID, "round3"))
		_, err = repo.GetByID(item.ID)
		require.Error(t, err)
	})

	t.Run("missing table returns errors", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		repo := NewVisualizationRepository(db)

		err = repo.Create(&visualization.DataVisualizationInfo{Name: "missing"})
		require.Error(t, err)
		_, err = repo.GetByID(1)
		require.Error(t, err)
		err = repo.DeleteLogic(1, "nobody")
		require.Error(t, err)
		_, err = repo.ListAllByTypes(nil)
		require.Error(t, err)
		_, err = repo.ListAllByTypesBatch(nil, 0, 0, nil)
		require.Error(t, err)
		_, err = repo.CountByNameAndPID("missing", nil, nil)
		require.Error(t, err)
		_, err = repo.GetChartViewsBySceneID(1)
		require.Error(t, err)
		_, err = repo.GetSnapshotChartViewsBySceneID(1)
		require.Error(t, err)
	})
}

func TestRound3_UserRepo_SupplementalWrappers(t *testing.T) {
	t.Run("constructors and relation helpers", func(t *testing.T) {
		userRepo, userRoleRepo, db := setupUserRepositoryTest(t)
		permRepo := NewUserPermRepository(db)
		require.Same(t, db, userRepo.db)
		require.Same(t, db, userRoleRepo.db)
		require.Same(t, db, permRepo.db)

		require.NoError(t, db.Create(&role.SysRole{RoleID: 101, RoleName: "enabled", RoleCode: "enabled", Status: role.StatusEnabled}).Error)
		require.NoError(t, db.Create(&user.SysUser{UserID: 1001, Username: "alice-round3", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, db.Create(&user.SysUser{UserID: 1002, Username: "bob-round3", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 1001, RoleID: 101, OrgID: 9}))
		require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 1001, RoleID: 101, OrgID: 11}))

		inOrg, err := userRoleRepo.IsUserInOrg(1001, 9)
		require.NoError(t, err)
		assert.True(t, inOrg)

		roles, err := userRoleRepo.GetByUserID(1001)
		require.NoError(t, err)
		require.Len(t, roles, 2)

		roleIDs, err := userRoleRepo.GetRoleIDsByUserID(1001)
		require.NoError(t, err)
		assert.Equal(t, []int64{101}, roleIDs)

		require.NoError(t, userRoleRepo.SwitchOrgForUser(1001, 15))
		roles, err = userRoleRepo.GetByUserID(1001)
		require.NoError(t, err)
		require.Len(t, roles, 2)
		assert.Equal(t, int64(15), roles[0].OrgID)
		assert.Equal(t, int64(15), roles[1].OrgID)

		count, err := userRepo.CountByOrgID(15)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		require.NoError(t, permRepo.Create(&user.SysUserPerm{UserID: 1001, OrgID: 15, PermID: 201, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}))
		perms, err := permRepo.GetByUserID(1001)
		require.NoError(t, err)
		require.Len(t, perms, 1)
		require.NoError(t, permRepo.DeleteByUserID(1001))
		perms, err = permRepo.GetByUserID(1001)
		require.NoError(t, err)
		assert.Empty(t, perms)
	})

	t.Run("missing tables return errors", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		userRepo := NewUserRepository(db)
		userRoleRepo := NewUserRoleRepository(db)
		permRepo := NewUserPermRepository(db)

		_, err = userRepo.CountByOrgID(1)
		require.Error(t, err)
		err = userRoleRepo.SwitchOrgForUser(1, 2)
		require.Error(t, err)
		_, err = userRoleRepo.GetRoleIDsByUserID(1)
		require.Error(t, err)
		err = permRepo.Create(&user.SysUserPerm{UserID: 1, OrgID: 1, PermID: 1})
		require.Error(t, err)
	})
}

func TestRound3_RowPermissionRepo_RowPaths(t *testing.T) {
	rowRepo, _, _ := setupRowPermissionRepoTest(t)
	var nilRepo *RowPermissionRepository
	assert.Nil(t, nilRepo.DB())
	assert.NotNil(t, rowRepo.DB())

	perm1 := newRound3RowPermission(5001, permission.AuthTargetTypeUser, 9001, `{"logic":"and","items":[]}`)
	perm1.Status = 0
	perm1.CreateTime = nil
	perm1.DatasetGroupID = 0
	require.NoError(t, rowRepo.Create(perm1))
	assert.Equal(t, int64(5001), perm1.DatasetGroupID)
	assert.Equal(t, 1, perm1.Status)
	require.NotNil(t, perm1.CreateTime)

	perm2 := newRound3RowPermission(5001, permission.AuthTargetTypeRole, 9002, `{"logic":"or","items":[]}`)
	require.NoError(t, rowRepo.Create(perm2))
	perm3 := newRound3RowPermission(5001, permission.AuthTargetTypeRole, 9002, `{"logic":"and","items":[]}`)
	require.NoError(t, rowRepo.Create(perm3))

	list, err := rowRepo.ListByDatasetID(5001)
	require.NoError(t, err)
	require.Len(t, list, 3)

	paged, total, err := rowRepo.PagerByDatasetID(5001, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, paged, 3)
	assert.Equal(t, perm3.ID, paged[0].ID)

	filtered, filteredTotal, err := rowRepo.PagerByDatasetIDAndTarget(5001, permission.AuthTargetTypeRole, 9002, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), filteredTotal)
	require.Len(t, filtered, 1)
	assert.Equal(t, perm3.ID, filtered[0].ID)

	byID, err := rowRepo.GetByID(perm1.ID)
	require.NoError(t, err)
	assert.Equal(t, perm1.ExpressionTree, byID.ExpressionTree)

	perm1.ExpressionTree = `{"logic":"or","items":[]}`
	perm1.DatasetGroupID = 0
	require.NoError(t, rowRepo.Update(perm1))

	updated, err := rowRepo.GetByID(perm1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(5001), updated.DatasetGroupID)
	assert.Equal(t, `{"logic":"or","items":[]}`, updated.ExpressionTree)
	require.NotNil(t, updated.UpdateTime)

	userPerms, err := rowRepo.ListByDatasetIDAndUserID(5001, 9001)
	require.NoError(t, err)
	require.Len(t, userPerms, 1)

	rolePerms, err := rowRepo.ListByDatasetIDAndRoleIDs(5001, []int64{9002})
	require.NoError(t, err)
	require.Len(t, rolePerms, 2)

	emptyRolePerms, err := rowRepo.ListByDatasetIDAndRoleIDs(5001, nil)
	require.NoError(t, err)
	assert.Empty(t, emptyRolePerms)

	userOnly, err := rowRepo.GetUserPermissions(5001, 9001)
	require.NoError(t, err)
	require.Len(t, userOnly, 1)

	parsed, err := rowRepo.ParseExpressionTree(`{"logic":"and","items":[{"type":"item","fieldId":7,"fieldType":"string","filterType":"logic","term":"eq","value":"CN"}]}`)
	require.NoError(t, err)
	require.NotNil(t, parsed)
	assert.Equal(t, int64(7), parsed.Items[0].FieldID)

	nilParsed, err := rowRepo.ParseExpressionTree("")
	require.NoError(t, err)
	assert.Nil(t, nilParsed)

	invalidParsed, err := rowRepo.ParseExpressionTree("{bad json}")
	assert.Nil(t, invalidParsed)
	require.Error(t, err)

	require.NoError(t, rowRepo.Delete(perm2.ID))
	remaining, err := rowRepo.ListByDatasetID(5001)
	require.NoError(t, err)
	require.Len(t, remaining, 2)
}

func TestRound3_RowPermissionRepo_ColumnAndErrorPaths(t *testing.T) {
	_, columnRepo, _ := setupRowPermissionRepoTest(t)
	var nilRepo *ColumnPermissionRepository
	assert.Nil(t, nilRepo.DB())
	assert.NotNil(t, columnRepo.DB())

	perm1 := newRound3ColumnPermission(6001, "name", permission.PermTypeData, "")
	perm1.PermType = "disable"
	perm1.Status = 0
	perm1.CreateTime = nil
	perm1.DatasetGroupID = 0
	require.NoError(t, columnRepo.Create(perm1))
	assert.Equal(t, int64(6001), perm1.DatasetGroupID)
	assert.Equal(t, 1, perm1.Status)
	require.NotNil(t, perm1.CreateTime)

	perm2 := newRound3ColumnPermission(6001, "mobile", "mask", `{"builtInRule":"CompleteDesensitization"}`)
	require.NoError(t, columnRepo.Create(perm2))
	perm3 := newRound3ColumnPermission(6001, "mobile", "mask", `{"builtInRule":"custom"}`)
	require.NoError(t, columnRepo.Create(perm3))

	list, err := columnRepo.ListByDatasetID(6001)
	require.NoError(t, err)
	require.Len(t, list, 3)

	paged, total, err := columnRepo.PagerByDatasetID(6001, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, paged, 3)
	assert.Equal(t, perm3.ID, paged[0].ID)

	byID, err := columnRepo.GetByID(perm1.ID)
	require.NoError(t, err)
	assert.Equal(t, "name", byID.FieldName)

	perm1.FieldName = "username"
	perm1.PermType = "mask"
	perm1.MaskRule = `{"builtInRule":"Partial"}`
	perm1.DatasetGroupID = 0
	require.NoError(t, columnRepo.Update(perm1))

	updated, err := columnRepo.GetByID(perm1.ID)
	require.NoError(t, err)
	assert.Equal(t, "username", updated.FieldName)
	assert.Equal(t, int64(6001), updated.DatasetGroupID)
	require.NotNil(t, updated.UpdateTime)

	names, err := columnRepo.ListAllColumnNamesByDatasetID(6001)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"username", "mobile"}, names)

	require.NoError(t, columnRepo.Delete(perm2.ID))
	remaining, err := columnRepo.ListByDatasetID(6001)
	require.NoError(t, err)
	require.Len(t, remaining, 2)

	missingDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	missingRowRepo := NewRowPermissionRepository(missingDB)
	missingColumnRepo := NewColumnPermissionRepository(missingDB)

	_, err = missingRowRepo.GetByID(1)
	require.Error(t, err)
	_, _, err = missingRowRepo.PagerByDatasetID(1, 1, 1)
	require.Error(t, err)
	_, err = missingColumnRepo.GetByID(1)
	require.Error(t, err)
	_, err = missingColumnRepo.ListAllColumnNamesByDatasetID(1)
	require.Error(t, err)
	_, _, err = missingColumnRepo.PagerByDatasetID(1, 1, 1)
	require.Error(t, err)
	_, err = missingRowRepo.ListByDatasetID(1)
	require.Error(t, err)
	_, err = missingColumnRepo.ListByDatasetID(1)
	require.Error(t, err)
}
