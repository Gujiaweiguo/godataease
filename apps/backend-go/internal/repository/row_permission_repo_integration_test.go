//go:build integration
// +build integration

package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/permission"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRowPermissionRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_row")

	repo := NewRowPermissionRepository(testDB)
	perm := newRowPermission(1001, permission.AuthTargetTypeUser, 2001, `{"logic":"and","items":[]}`)
	perm.DatasetGroupID = 3001

	require.NoError(t, repo.Create(perm))
	assert.NotZero(t, perm.ID)
	assert.Equal(t, 1, perm.Status)
	require.NotNil(t, perm.CreateTime)

	got, err := repo.GetByID(perm.ID)
	require.NoError(t, err)
	assert.Equal(t, perm.DatasetID, got.DatasetID)
	assert.Equal(t, perm.DatasetGroupID, got.DatasetGroupID)
	assert.Equal(t, perm.AuthTargetType, got.AuthTargetType)
	assert.Equal(t, perm.AuthTargetID, got.AuthTargetID)
	assert.Equal(t, perm.ExpressionTree, got.ExpressionTree)
	assert.Equal(t, perm.Status, got.Status)
	assert.Equal(t, derefString(t, perm.CreateBy), derefString(t, got.CreateBy))
}

func TestRowPermissionRepository_ListByDatasetID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_row")

	repo := NewRowPermissionRepository(testDB)
	perm1 := newRowPermission(1101, permission.AuthTargetTypeUser, 1, `{"logic":"and","items":[]}`)
	perm2 := newRowPermission(1101, permission.AuthTargetTypeRole, 2, `{"logic":"or","items":[]}`)
	perm3 := newRowPermission(1102, permission.AuthTargetTypeUser, 3, `{"logic":"and","items":[]}`)
	permDeleted := newRowPermission(1101, permission.AuthTargetTypeDept, 4, `{"logic":"and","items":[]}`)

	for _, perm := range []*permission.DataPermRow{perm1, perm2, perm3, permDeleted} {
		require.NoError(t, repo.Create(perm))
	}
	require.NoError(t, repo.Delete(permDeleted.ID))

	perms, err := repo.ListByDatasetID(1101)
	require.NoError(t, err)
	require.Len(t, perms, 2)
	assert.ElementsMatch(t, []int64{perm1.ID, perm2.ID}, []int64{perms[0].ID, perms[1].ID})
}

func TestRowPermissionRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_row")

	repo := NewRowPermissionRepository(testDB)
	perm := newRowPermission(1201, permission.AuthTargetTypeUser, 3001, `{"logic":"and","items":[]}`)
	require.NoError(t, repo.Create(perm))

	perm.DatasetGroupID = 4201
	perm.ExpressionTree = `{"logic":"or","items":[]}`
	updateBy := "updater"
	perm.UpdateBy = &updateBy
	require.NoError(t, repo.Update(perm))

	got, err := repo.GetByID(perm.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(4201), got.DatasetGroupID)
	assert.Equal(t, `{"logic":"or","items":[]}`, got.ExpressionTree)
	assert.Equal(t, "updater", derefString(t, got.UpdateBy))
	require.NotNil(t, got.UpdateTime)
}

func TestRowPermissionRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_row")

	repo := NewRowPermissionRepository(testDB)
	perm := newRowPermission(1301, permission.AuthTargetTypeUser, 4001, `{"logic":"and","items":[]}`)
	require.NoError(t, repo.Create(perm))

	require.NoError(t, repo.Delete(perm.ID))

	perms, err := repo.ListByDatasetID(1301)
	require.NoError(t, err)
	assert.Empty(t, perms)

	stored, err := repo.GetByID(perm.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, stored.Status)
	require.NotNil(t, stored.UpdateTime)
}

func TestRowPermissionRepository_PagerByDatasetID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_row")

	repo := NewRowPermissionRepository(testDB)
	perm1 := newRowPermission(1401, permission.AuthTargetTypeUser, 1, `{"logic":"and","items":[]}`)
	perm2 := newRowPermission(1401, permission.AuthTargetTypeRole, 2, `{"logic":"and","items":[]}`)
	perm3 := newRowPermission(1401, permission.AuthTargetTypeDept, 3, `{"logic":"and","items":[]}`)
	for _, perm := range []*permission.DataPermRow{perm1, perm2, perm3} {
		require.NoError(t, repo.Create(perm))
	}

	pageOne, total, err := repo.PagerByDatasetID(1401, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, pageOne, 2)
	assert.Equal(t, perm3.ID, pageOne[0].ID)
	assert.Equal(t, perm2.ID, pageOne[1].ID)

	pageTwo, secondTotal, err := repo.PagerByDatasetID(1401, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), secondTotal)
	require.Len(t, pageTwo, 1)
	assert.Equal(t, perm1.ID, pageTwo[0].ID)
}

func TestRowPermissionRepository_PagerByDatasetIDAndTarget(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_row")

	repo := NewRowPermissionRepository(testDB)
	match1 := newRowPermission(1501, permission.AuthTargetTypeRole, 5001, `{"logic":"and","items":[]}`)
	match2 := newRowPermission(1501, permission.AuthTargetTypeRole, 5001, `{"logic":"or","items":[]}`)
	otherTarget := newRowPermission(1501, permission.AuthTargetTypeRole, 5002, `{"logic":"and","items":[]}`)
	otherType := newRowPermission(1501, permission.AuthTargetTypeUser, 5001, `{"logic":"and","items":[]}`)
	for _, perm := range []*permission.DataPermRow{match1, match2, otherTarget, otherType} {
		require.NoError(t, repo.Create(perm))
	}

	perms, total, err := repo.PagerByDatasetIDAndTarget(1501, permission.AuthTargetTypeRole, 5001, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, perms, 2)
	assert.Equal(t, match2.ID, perms[0].ID)
	assert.Equal(t, match1.ID, perms[1].ID)
}

func TestRowPermissionRepository_ListByDatasetIDAndUserID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_row")

	repo := NewRowPermissionRepository(testDB)
	match := newRowPermission(1601, permission.AuthTargetTypeUser, 6001, `{"logic":"and","items":[]}`)
	otherUser := newRowPermission(1601, permission.AuthTargetTypeUser, 6002, `{"logic":"and","items":[]}`)
	otherRole := newRowPermission(1601, permission.AuthTargetTypeRole, 6001, `{"logic":"and","items":[]}`)
	for _, perm := range []*permission.DataPermRow{match, otherUser, otherRole} {
		require.NoError(t, repo.Create(perm))
	}

	perms, err := repo.ListByDatasetIDAndUserID(1601, 6001)
	require.NoError(t, err)
	require.Len(t, perms, 1)
	assert.Equal(t, match.ID, perms[0].ID)
}

func TestRowPermissionRepository_ListByDatasetIDAndRoleIDs(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_row")

	repo := NewRowPermissionRepository(testDB)
	match1 := newRowPermission(1701, permission.AuthTargetTypeRole, 7001, `{"logic":"and","items":[]}`)
	match2 := newRowPermission(1701, permission.AuthTargetTypeRole, 7002, `{"logic":"and","items":[]}`)
	other := newRowPermission(1701, permission.AuthTargetTypeRole, 7003, `{"logic":"and","items":[]}`)
	for _, perm := range []*permission.DataPermRow{match1, match2, other} {
		require.NoError(t, repo.Create(perm))
	}

	perms, err := repo.ListByDatasetIDAndRoleIDs(1701, []int64{7001, 7002})
	require.NoError(t, err)
	require.Len(t, perms, 2)
	assert.ElementsMatch(t, []int64{match1.ID, match2.ID}, []int64{perms[0].ID, perms[1].ID})

	empty, err := repo.ListByDatasetIDAndRoleIDs(1701, nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestRowPermissionRepository_GetUserPermissions(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_row")

	repo := NewRowPermissionRepository(testDB)
	match := newRowPermission(1801, permission.AuthTargetTypeUser, 8001, `{"logic":"and","items":[]}`)
	other := newRowPermission(1801, permission.AuthTargetTypeUser, 8002, `{"logic":"and","items":[]}`)
	for _, perm := range []*permission.DataPermRow{match, other} {
		require.NoError(t, repo.Create(perm))
	}

	perms, err := repo.GetUserPermissions(1801, 8001)
	require.NoError(t, err)
	require.Len(t, perms, 1)
	assert.Equal(t, match.ID, perms[0].ID)
}

func TestRowPermissionRepository_ParseExpressionTree(t *testing.T) {
	repo := NewRowPermissionRepository(testDB)

	parsed, err := repo.ParseExpressionTree(`{"logic":"and","items":[{"type":"item","fieldId":1,"fieldType":"string","filterType":"logic","term":"eq","value":"CN"}]}`)
	require.NoError(t, err)
	require.NotNil(t, parsed)
	assert.Equal(t, "and", parsed.Logic)
	require.Len(t, parsed.Items, 1)
	assert.Equal(t, int64(1), parsed.Items[0].FieldID)
	assert.Equal(t, "CN", parsed.Items[0].Value)

	empty, err := repo.ParseExpressionTree("")
	require.NoError(t, err)
	assert.Nil(t, empty)

	invalid, err := repo.ParseExpressionTree("{invalid json}")
	assert.Nil(t, invalid)
	assert.Error(t, err)
}

func TestColumnPermissionRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_column")

	repo := NewColumnPermissionRepository(testDB)
	perm := newColumnPermission(1901, "mobile", permission.PermTypeMask, `{"builtInRule":"CompleteDesensitization"}`)
	perm.DatasetGroupID = 2901

	require.NoError(t, repo.Create(perm))
	assert.NotZero(t, perm.ID)
	assert.Equal(t, 1, perm.Status)
	require.NotNil(t, perm.CreateTime)

	got, err := repo.GetByID(perm.ID)
	require.NoError(t, err)
	assert.Equal(t, perm.DatasetID, got.DatasetID)
	assert.Equal(t, perm.DatasetGroupID, got.DatasetGroupID)
	assert.Equal(t, perm.FieldName, got.FieldName)
	assert.Equal(t, perm.PermType, got.PermType)
	assert.Equal(t, perm.MaskRule, got.MaskRule)
	assert.Equal(t, derefString(t, perm.CreateBy), derefString(t, got.CreateBy))
}

func TestColumnPermissionRepository_ListByDatasetID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_column")

	repo := NewColumnPermissionRepository(testDB)
	perm1 := newColumnPermission(2001, "name", permission.PermTypeDisable, "")
	perm2 := newColumnPermission(2001, "mobile", permission.PermTypeMask, `{"builtInRule":"CompleteDesensitization"}`)
	perm3 := newColumnPermission(2002, "email", permission.PermTypeMask, `{"builtInRule":"custom"}`)
	deleted := newColumnPermission(2001, "address", permission.PermTypeDisable, "")
	for _, perm := range []*permission.DataPermColumn{perm1, perm2, perm3, deleted} {
		require.NoError(t, repo.Create(perm))
	}
	require.NoError(t, repo.Delete(deleted.ID))

	perms, err := repo.ListByDatasetID(2001)
	require.NoError(t, err)
	require.Len(t, perms, 2)
	assert.ElementsMatch(t, []int64{perm1.ID, perm2.ID}, []int64{perms[0].ID, perms[1].ID})
}

func TestColumnPermissionRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_column")

	repo := NewColumnPermissionRepository(testDB)
	perm := newColumnPermission(2101, "mobile", permission.PermTypeMask, `{"builtInRule":"CompleteDesensitization"}`)
	require.NoError(t, repo.Create(perm))

	perm.FieldName = "phone"
	perm.DatasetGroupID = 3101
	perm.PermType = permission.PermTypeDisable
	perm.MaskRule = ""
	updateBy := "column-updater"
	perm.UpdateBy = &updateBy
	require.NoError(t, repo.Update(perm))

	got, err := repo.GetByID(perm.ID)
	require.NoError(t, err)
	assert.Equal(t, "phone", got.FieldName)
	assert.Equal(t, int64(3101), got.DatasetGroupID)
	assert.Equal(t, permission.PermTypeDisable, got.PermType)
	assert.Equal(t, "", got.MaskRule)
	assert.Equal(t, "column-updater", derefString(t, got.UpdateBy))
	require.NotNil(t, got.UpdateTime)
}

func TestColumnPermissionRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_column")

	repo := NewColumnPermissionRepository(testDB)
	perm := newColumnPermission(2201, "mobile", permission.PermTypeMask, `{"builtInRule":"CompleteDesensitization"}`)
	require.NoError(t, repo.Create(perm))

	require.NoError(t, repo.Delete(perm.ID))

	perms, err := repo.ListByDatasetID(2201)
	require.NoError(t, err)
	assert.Empty(t, perms)

	stored, err := repo.GetByID(perm.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, stored.Status)
	require.NotNil(t, stored.UpdateTime)
}

func TestColumnPermissionRepository_PagerByDatasetID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_column")

	repo := NewColumnPermissionRepository(testDB)
	perm1 := newColumnPermission(2301, "name", permission.PermTypeDisable, "")
	perm2 := newColumnPermission(2301, "mobile", permission.PermTypeMask, `{"builtInRule":"CompleteDesensitization"}`)
	perm3 := newColumnPermission(2301, "email", permission.PermTypeMask, `{"builtInRule":"custom"}`)
	for _, perm := range []*permission.DataPermColumn{perm1, perm2, perm3} {
		require.NoError(t, repo.Create(perm))
	}

	pageOne, total, err := repo.PagerByDatasetID(2301, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, pageOne, 2)
	assert.Equal(t, perm3.ID, pageOne[0].ID)
	assert.Equal(t, perm2.ID, pageOne[1].ID)

	pageTwo, secondTotal, err := repo.PagerByDatasetID(2301, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), secondTotal)
	require.Len(t, pageTwo, 1)
	assert.Equal(t, perm1.ID, pageTwo[0].ID)
}

func TestColumnPermissionRepository_ListAllColumnNamesByDatasetID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_perm_column")

	repo := NewColumnPermissionRepository(testDB)
	perm1 := newColumnPermission(2401, "name", permission.PermTypeDisable, "")
	perm2 := newColumnPermission(2401, "name", permission.PermTypeMask, `{"builtInRule":"CompleteDesensitization"}`)
	perm3 := newColumnPermission(2401, "mobile", permission.PermTypeMask, `{"builtInRule":"custom"}`)
	otherDataset := newColumnPermission(2402, "email", permission.PermTypeMask, `{"builtInRule":"custom"}`)
	for _, perm := range []*permission.DataPermColumn{perm1, perm2, perm3, otherDataset} {
		require.NoError(t, repo.Create(perm))
	}

	names, err := repo.ListAllColumnNamesByDatasetID(2401)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"name", "mobile"}, names)
}

func newRowPermission(datasetID int64, targetType string, targetID int64, expressionTree string) *permission.DataPermRow {
	creator := "creator"
	createdAt := time.Unix(1710000100, 0)
	return &permission.DataPermRow{
		DatasetID:      datasetID,
		AuthTargetType: targetType,
		AuthTargetID:   targetID,
		ExpressionTree: expressionTree,
		Status:         1,
		CreateBy:       &creator,
		CreateTime:     &createdAt,
	}
}

func newColumnPermission(datasetID int64, fieldName, permType, maskRule string) *permission.DataPermColumn {
	creator := "creator"
	createdAt := time.Unix(1710000200, 0)
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

func derefString(t *testing.T, value *string) string {
	t.Helper()
	if value == nil {
		return ""
	}
	return *value
}
