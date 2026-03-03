//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

type rowPermRoleRepoStub struct {
	roleIDs []int64
	err     error
}

func (s *rowPermRoleRepoStub) GetRoleIDsByUserID(userID int64) ([]int64, error) {
	return s.roleIDs, s.err
}

type rowPermAdminCheckerStub struct {
	adminUsers map[int64]bool
}

func (s *rowPermAdminCheckerStub) IsAdmin(userID int64) bool {
	if s.adminUsers == nil {
		return false
	}
	return s.adminUsers[userID]
}

func TestRowPermissionServiceIntegration_BuildWhereClause_MergeUserAndRole(t *testing.T) {
	cleanupTables(&permission.DataPermRow{}, &permission.DataPermColumn{})

	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(
		rowRepo,
		colRepo,
		&rowPermRoleRepoStub{roleIDs: []int64{3001}},
		&rowPermAdminCheckerStub{},
	)

	err := testDB.Create(&permission.DataPermRow{
		DatasetID:      1001,
		AuthTargetType: permission.AuthTargetTypeUser,
		AuthTargetID:   2001,
		Status:         1,
		ExpressionTree: `{"logic":"OR","items":[{"fieldId":1,"term":"eq","value":"east"}]}`,
	}).Error
	assert.NoError(t, err)

	err = testDB.Create(&permission.DataPermRow{
		DatasetID:      1001,
		AuthTargetType: permission.AuthTargetTypeRole,
		AuthTargetID:   3001,
		Status:         1,
		ExpressionTree: `{"logic":"AND","items":[{"fieldId":2,"term":"gt","value":"10"}]}`,
	}).Error
	assert.NoError(t, err)

	result, err := svc.BuildWhereClause(1001, 2001)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Clause, "`1` = ?")
	assert.Contains(t, result.Clause, "`2` > ?")
	assert.Contains(t, result.Clause, " OR ")
	assert.Len(t, result.Args, 2)
	assert.Equal(t, "east", result.Args[0])
	assert.Equal(t, "10", result.Args[1])
}

func TestRowPermissionServiceIntegration_BuildWhereClause_InvalidRulesIgnored(t *testing.T) {
	cleanupTables(&permission.DataPermRow{}, &permission.DataPermColumn{})

	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(rowRepo, colRepo, nil, &rowPermAdminCheckerStub{})

	err := testDB.Create(&permission.DataPermRow{
		DatasetID:      1002,
		AuthTargetType: permission.AuthTargetTypeUser,
		AuthTargetID:   2002,
		Status:         1,
		ExpressionTree: "{bad json}",
	}).Error
	assert.NoError(t, err)

	err = testDB.Create(&permission.DataPermRow{
		DatasetID:      1002,
		AuthTargetType: permission.AuthTargetTypeUser,
		AuthTargetID:   2002,
		Status:         1,
		ExpressionTree: "",
	}).Error
	assert.NoError(t, err)

	result, err := svc.BuildWhereClause(1002, 2002)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestRowPermissionServiceIntegration_BuildWhereClause_AdminBypass(t *testing.T) {
	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(
		rowRepo,
		colRepo,
		nil,
		&rowPermAdminCheckerStub{adminUsers: map[int64]bool{1: true}},
	)

	result, err := svc.BuildWhereClause(1003, 1)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestRowPermissionServiceIntegration_BuildSelectColumns_ExcludeDisabled(t *testing.T) {
	cleanupTables(&permission.DataPermRow{}, &permission.DataPermColumn{})

	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(rowRepo, colRepo, nil, &rowPermAdminCheckerStub{})

	err := testDB.Create(&permission.DataPermColumn{DatasetID: 1004, FieldName: "region", PermType: "mask", Status: 1}).Error
	assert.NoError(t, err)
	err = testDB.Create(&permission.DataPermColumn{DatasetID: 1004, FieldName: "amount", PermType: "disable", Status: 1}).Error
	assert.NoError(t, err)
	err = testDB.Create(&permission.DataPermColumn{DatasetID: 1004, FieldName: "city", PermType: "mask", Status: 1}).Error
	assert.NoError(t, err)

	columns, err := svc.BuildSelectColumns(1004, 2004)
	assert.NoError(t, err)
	assert.NotEqual(t, "*", columns)
	assert.Contains(t, columns, "`region`")
	assert.Contains(t, columns, "`city`")
	assert.NotContains(t, columns, "`amount`")
}

func TestRowPermissionServiceIntegration_BuildSelectColumns_AllDisabledFallbackWildcard(t *testing.T) {
	cleanupTables(&permission.DataPermRow{}, &permission.DataPermColumn{})

	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(rowRepo, colRepo, nil, &rowPermAdminCheckerStub{})

	err := testDB.Create(&permission.DataPermColumn{DatasetID: 1005, FieldName: "secret", PermType: "disable", Status: 1}).Error
	assert.NoError(t, err)

	columns, err := svc.BuildSelectColumns(1005, 2005)
	assert.NoError(t, err)
	assert.Equal(t, "*", columns)
}

func TestRowPermissionServiceIntegration_GetRowPermissionsTree_RoleLookupErrorIgnored(t *testing.T) {
	cleanupTables(&permission.DataPermRow{}, &permission.DataPermColumn{})

	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(
		rowRepo,
		colRepo,
		&rowPermRoleRepoStub{err: assert.AnError},
		&rowPermAdminCheckerStub{},
	)

	err := testDB.Create(&permission.DataPermRow{
		DatasetID:      1006,
		AuthTargetType: permission.AuthTargetTypeUser,
		AuthTargetID:   2006,
		Status:         1,
		ExpressionTree: `{"logic":"OR","items":[{"fieldId":5,"term":"eq","value":"ok"}]}`,
	}).Error
	assert.NoError(t, err)

	list, queryErr := svc.GetRowPermissionsTree(1006, 2006)
	assert.NoError(t, queryErr)
	assert.Len(t, list, 1)
}

func TestRowPermissionServiceIntegration_IsAdminAndGetRoleIDs(t *testing.T) {
	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(
		rowRepo,
		colRepo,
		&rowPermRoleRepoStub{roleIDs: []int64{10, 20}},
		&rowPermAdminCheckerStub{adminUsers: map[int64]bool{999: true}},
	)

	assert.True(t, svc.IsAdmin(999))
	assert.False(t, svc.IsAdmin(1000))

	roles, err := svc.GetUserRoleIDs(100)
	assert.NoError(t, err)
	assert.Equal(t, []int64{10, 20}, roles)

	svcWithoutRoleRepo := NewRowPermissionService(rowRepo, colRepo, nil, &rowPermAdminCheckerStub{})
	roles, err = svcWithoutRoleRepo.GetUserRoleIDs(100)
	assert.NoError(t, err)
	assert.Nil(t, roles)
}


// TestBuildLogicCondition tests the buildLogicCondition method with various operators
func TestRowPermissionServiceIntegration_BuildLogicCondition_AllOperators(t *testing.T) {
	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(rowRepo, colRepo, nil, nil)

	tests := []struct {
		name      string
		term      string
		value     string
		wantClause string
		wantArgs   int
	}{
		{"eq operator", "eq", "test", "`field` = ?", 1},
		{"not_eq operator", "not_eq", "val", "`field` != ?", 1},
		{"not eq operator (space)", "not eq", "val", "`field` != ?", 1},
		{"like operator", "like", "pattern", "`field` LIKE ?", 1},
		{"not_like operator", "not_like", "pattern", "`field` NOT LIKE ?", 1},
		{"null operator", "null", "", "`field` IS NULL", 0},
		{"not_null operator", "not_null", "", "`field` IS NOT NULL", 0},
		{"empty operator", "empty", "", "`field` = ''", 0},
		{"not_empty operator", "not_empty", "", "`field` != ''", 0},
		{"gt operator", "gt", "100", "`field` > ?", 1},
		{"lt operator", "lt", "50", "`field` < ?", 1},
		{"ge operator", "ge", "100", "`field` >= ?", 1},
		{"le operator", "le", "50", "`field` <= ?", 1},
		{"in operator (returns empty)", "in", "a,b", "", 0},
		{"not_in operator (returns empty)", "not_in", "a,b", "", 0},
		{"unknown operator defaults to eq", "unknown", "val", "`field` = ?", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clause, args := svc.buildLogicCondition("`field`", tt.term, tt.value)
			if tt.wantClause != "" {
				assert.Equal(t, tt.wantClause, clause)
			} else {
				assert.Empty(t, clause)
			}
			if tt.wantArgs > 0 {
				assert.Len(t, args, tt.wantArgs)
			} else {
				assert.Nil(t, args)
			}
		})
	}
}

// TestBuildLogicCondition_EmptyValue tests that empty value returns empty for non-null operators
func TestRowPermissionServiceIntegration_BuildLogicCondition_EmptyValue(t *testing.T) {
	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(rowRepo, colRepo, nil, nil)

	// Empty value with non-null operator should return empty
	clause, args := svc.buildLogicCondition("`field`", "eq", "")
	assert.Empty(t, clause)
	assert.Nil(t, args)

	// Empty value with null operator should still work
	clause, args = svc.buildLogicCondition("`field`", "null", "")
	assert.Equal(t, "`field` IS NULL", clause)
	assert.Nil(t, args)
}

// TestBuildEnumCondition tests the buildEnumCondition method
func TestRowPermissionServiceIntegration_BuildEnumCondition(t *testing.T) {
	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(rowRepo, colRepo, nil, nil)

	// Empty values should return empty
	clause, args := svc.buildEnumCondition("`field`", []string{})
	assert.Empty(t, clause)
	assert.Nil(t, args)

	// Single value
	clause, args = svc.buildEnumCondition("`field`", []string{"value1"})
	assert.Equal(t, "`field` IN (?)", clause)
	assert.Len(t, args, 1)
	assert.Equal(t, "value1", args[0])

	// Multiple values
	clause, args = svc.buildEnumCondition("`field`", []string{"a", "b", "c"})
	assert.Equal(t, "`field` IN (?, ?, ?)", clause)
	assert.Len(t, args, 3)
}

// TestEscapeSQL tests SQL escaping
func TestRowPermissionServiceIntegration_EscapeSQL(t *testing.T) {
	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(rowRepo, colRepo, nil, nil)

	tests := []struct {
		input  string
		expect string
	}{
		{"normal", "normal"},
		{"it's", "it''s"},
		{"back\\slash", "back\\\\slash"},
		{"null\x00byte", "nullbyte"},
		{"new\nline", "newline"},
		{"carriage\rreturn", "carriagereturn"},
		{"ctrl\x1az", "ctrlz"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := svc.escapeSQL(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

// TestBuildItemCondition tests buildItemCondition with fieldID=0 case
func TestRowPermissionServiceIntegration_BuildItemCondition_ZeroFieldID(t *testing.T) {
	rowRepo := repository.NewRowPermissionRepository(testDB)
	colRepo := repository.NewColumnPermissionRepository(testDB)
	svc := NewRowPermissionService(rowRepo, colRepo, nil, nil)

	// fieldID=0 should return empty
	item := &permission.DatasetRowPermissionsTreeItem{FieldID: 0}
	clause, args := svc.buildItemCondition(item)
	assert.Empty(t, clause)
	assert.Nil(t, args)
}