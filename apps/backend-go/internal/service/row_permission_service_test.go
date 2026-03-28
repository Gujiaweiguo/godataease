package service

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockAdminChecker struct {
	adminUserIDs map[int64]bool
}

func (m *mockAdminChecker) IsAdmin(userID int64) bool {
	return m.adminUserIDs[userID]
}

type mockUserRoleRepo struct {
	roleIDs []int64
	err     error
}

func (m *mockUserRoleRepo) GetRoleIDsByUserID(_ int64) ([]int64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.roleIDs, nil
}

func setupRowPermissionServiceRepoTest(t *testing.T, userRoleRepo UserRoleRepositoryInterface, adminChecker AdminCheckerInterface) (*RowPermissionService, *repository.RowPermissionRepository, *repository.ColumnPermissionRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&permission.DataPermRow{}, &permission.DataPermColumn{}))

	rowRepo := repository.NewRowPermissionRepository(db)
	colRepo := repository.NewColumnPermissionRepository(db)
	return NewRowPermissionService(rowRepo, colRepo, userRoleRepo, adminChecker), rowRepo, colRepo, db
}

func TestBuildLogicCondition(t *testing.T) {
	svc := &RowPermissionService{
		adminChecker: &mockAdminChecker{adminUserIDs: map[int64]bool{1: true}},
	}

	tests := []struct {
		name     string
		field    string
		term     string
		value    string
		wantCond string
		wantArgs int
	}{
		{
			name:     "eq operator",
			field:    "`field1`",
			term:     permission.OperatorEq,
			value:    "test",
			wantCond: "`field1` = ?",
			wantArgs: 1,
		},
		{
			name:     "like operator",
			field:    "`field1`",
			term:     permission.OperatorLike,
			value:    "test",
			wantCond: "`field1` LIKE ?",
			wantArgs: 1,
		},
		{
			name:     "null operator",
			field:    "`field1`",
			term:     permission.OperatorNull,
			value:    "",
			wantCond: "`field1` IS NULL",
			wantArgs: 0,
		},
		{
			name:     "gt operator",
			field:    "`field1`",
			term:     permission.OperatorGt,
			value:    "100",
			wantCond: "`field1` > ?",
			wantArgs: 1,
		},
		{
			name:     "empty value with eq",
			field:    "`field1`",
			term:     permission.OperatorEq,
			value:    "",
			wantCond: "",
			wantArgs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, args := svc.buildLogicCondition(tt.field, tt.term, tt.value)
			if cond != tt.wantCond {
				t.Errorf("buildLogicCondition() condition = %v, want %v", cond, tt.wantCond)
			}
			if len(args) != tt.wantArgs {
				t.Errorf("buildLogicCondition() args count = %v, want %v", len(args), tt.wantArgs)
			}
		})
	}
}

func TestBuildEnumCondition(t *testing.T) {
	svc := &RowPermissionService{}

	tests := []struct {
		name     string
		field    string
		values   []string
		wantCond string
		wantArgs int
	}{
		{
			name:     "single value",
			field:    "`field1`",
			values:   []string{"a"},
			wantCond: "`field1` IN (?)",
			wantArgs: 1,
		},
		{
			name:     "multiple values",
			field:    "`field1`",
			values:   []string{"a", "b", "c"},
			wantCond: "`field1` IN (?, ?, ?)",
			wantArgs: 3,
		},
		{
			name:     "empty values",
			field:    "`field1`",
			values:   []string{},
			wantCond: "",
			wantArgs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, args := svc.buildEnumCondition(tt.field, tt.values)
			if cond != tt.wantCond {
				t.Errorf("buildEnumCondition() condition = %v, want %v", cond, tt.wantCond)
			}
			if len(args) != tt.wantArgs {
				t.Errorf("buildEnumCondition() args count = %v, want %v", len(args), tt.wantArgs)
			}
		})
	}
}

func TestParseTreeObj(t *testing.T) {
	svc := &RowPermissionService{}

	tests := []struct {
		name     string
		obj      *permission.DatasetRowPermissionsTreeObj
		wantCond string
	}{
		{
			name: "simple OR condition",
			obj: &permission.DatasetRowPermissionsTreeObj{
				Logic: "OR",
				Items: []permission.DatasetRowPermissionsTreeItem{
					{FieldID: 1, Term: permission.OperatorEq, Value: "a"},
					{FieldID: 2, Term: permission.OperatorEq, Value: "b"},
				},
			},
			wantCond: "(`1` = ? OR `2` = ?)",
		},
		{
			name: "simple AND condition",
			obj: &permission.DatasetRowPermissionsTreeObj{
				Logic: "AND",
				Items: []permission.DatasetRowPermissionsTreeItem{
					{FieldID: 1, Term: permission.OperatorEq, Value: "a"},
					{FieldID: 2, Term: permission.OperatorEq, Value: "b"},
				},
			},
			wantCond: "(`1` = ? AND `2` = ?)",
		},
		{
			name: "nested subtree",
			obj: &permission.DatasetRowPermissionsTreeObj{
				Logic: "OR",
				Items: []permission.DatasetRowPermissionsTreeItem{
					{FieldID: 1, Term: permission.OperatorEq, Value: "a"},
					{
						SubTree: &permission.DatasetRowPermissionsTreeObj{
							Logic: "AND",
							Items: []permission.DatasetRowPermissionsTreeItem{
								{FieldID: 2, Term: permission.OperatorEq, Value: "b"},
								{FieldID: 3, Term: permission.OperatorEq, Value: "c"},
							},
						},
					},
				},
			},
			wantCond: "(`1` = ? OR ((`2` = ? AND `3` = ?)))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, _ := svc.parseTreeObj(tt.obj)
			if cond != tt.wantCond {
				t.Errorf("parseTreeObj() condition = %v, want %v", cond, tt.wantCond)
			}
		})
	}
}

func TestEscapeSQL(t *testing.T) {
	svc := &RowPermissionService{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single quote escape",
			input: "test'value",
			want:  "test''value",
		},
		{
			name:  "backslash escape",
			input: "test\\value",
			want:  "test\\\\value",
		},
		{
			name:  "null byte removed",
			input: "test\x00value",
			want:  "testvalue",
		},
		{
			name:  "normal string unchanged",
			input: "normal_string",
			want:  "normal_string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.escapeSQL(tt.input)
			if got != tt.want {
				t.Errorf("escapeSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	// Test with nil admin checker
	svc := &RowPermissionService{adminChecker: nil}
	if svc.IsAdmin(1) {
		t.Error("Expected false when adminChecker is nil")
	}

	// Test with admin checker and admin user
	mockChecker := &mockAdminChecker{adminUserIDs: map[int64]bool{1: true}}
	svc = &RowPermissionService{adminChecker: mockChecker}
	if !svc.IsAdmin(1) {
		t.Error("Expected true for admin user")
	}

	// Test with admin checker and non-admin user
	if svc.IsAdmin(2) {
		t.Error("Expected false for non-admin user")
	}
}

func TestGetUserRoleIDs(t *testing.T) {
	// Test with nil repo
	svc := &RowPermissionService{userRoleRepo: nil}
	ids, err := svc.GetUserRoleIDs(1)
	if ids != nil || err != nil {
		t.Errorf("Expected nil, nil, got %v, %v", ids, err)
	}
}

func TestBuildItemCondition(t *testing.T) {
	svc := &RowPermissionService{}

	// Test with FieldID = 0
	cond, args := svc.buildItemCondition(&permission.DatasetRowPermissionsTreeItem{FieldID: 0})
	if cond != "" || args != nil {
		t.Errorf("Expected empty condition for FieldID=0, got %s, %v", cond, args)
	}

	// Test with enum filter
	cond, args = svc.buildItemCondition(&permission.DatasetRowPermissionsTreeItem{
		FieldID:    1,
		FilterType: "enum",
		EnumValue:  []string{"a", "b"},
	})
	if cond == "" {
		t.Error("Expected non-empty condition for enum filter")
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args for enum filter, got %d", len(args))
	}

	// Test with logic filter
	cond, args = svc.buildItemCondition(&permission.DatasetRowPermissionsTreeItem{
		FieldID: 1,
		Term:    permission.OperatorEq,
		Value:   "test",
	})
	if cond == "" {
		t.Error("Expected non-empty condition for logic filter")
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg for logic filter, got %d", len(args))
	}
}

func TestBuildLogicCondition_AllOperators(t *testing.T) {
	svc := &RowPermissionService{}

	tests := []struct {
		name     string
		term     string
		value    string
		wantCond string
		wantArgs int
	}{
		{"not_eq", "not_eq", "test", "`field` != ?", 1},
		{"not eq", "not eq", "test", "`field` != ?", 1},
		{"not_like", permission.OperatorNotLike, "test", "`field` NOT LIKE ?", 1},
		{"not_null", permission.OperatorNotNull, "", "`field` IS NOT NULL", 0},
		{"empty", permission.OperatorEmpty, "", "`field` = ''", 0},
		{"not_empty", permission.OperatorNotEmpty, "", "`field` != ''", 0},
		{"lt", permission.OperatorLt, "100", "`field` < ?", 1},
		{"ge", permission.OperatorGe, "100", "`field` >= ?", 1},
		{"le", permission.OperatorLe, "100", "`field` <= ?", 1},
		{"in", permission.OperatorIn, "test", "", 0},
		{"not_in", permission.OperatorNotIn, "test", "", 0},
		{"default", "unknown", "test", "`field` = ?", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, args := svc.buildLogicCondition("`field`", tt.term, tt.value)
			if cond != tt.wantCond {
				t.Errorf("condition = %v, want %v", cond, tt.wantCond)
			}
			if len(args) != tt.wantArgs {
				t.Errorf("args count = %v, want %v", len(args), tt.wantArgs)
			}
		})
	}
}

func TestParseTreeObj_EmptyAndNil(t *testing.T) {
	svc := &RowPermissionService{}

	// Test nil object
	cond, args := svc.parseTreeObj(nil)
	if cond != "" || args != nil {
		t.Errorf("Expected empty for nil, got %s, %v", cond, args)
	}

	// Test empty items
	cond, args = svc.parseTreeObj(&permission.DatasetRowPermissionsTreeObj{Items: []permission.DatasetRowPermissionsTreeItem{}})
	if cond != "" || args != nil {
		t.Errorf("Expected empty for empty items, got %s, %v", cond, args)
	}

	// Test default logic (empty string)
	cond, _ = svc.parseTreeObj(&permission.DatasetRowPermissionsTreeObj{
		Logic: "",
		Items: []permission.DatasetRowPermissionsTreeItem{
			{FieldID: 1, Term: permission.OperatorEq, Value: "a"},
		},
	})
	// Default should be OR
	if cond != "(`1` = ?)" {
		t.Errorf("Expected OR logic as default, got %s", cond)
	}
}

func TestBuildSelectColumns_Admin(t *testing.T) {
	svc := &RowPermissionService{
		adminChecker: &mockAdminChecker{adminUserIDs: map[int64]bool{1: true}},
	}

	// Test admin user returns *
	cols, err := svc.BuildSelectColumns(1, 1)
	if err != nil || cols != "*" {
		t.Errorf("Expected * for admin, got %s, %v", cols, err)
	}
}

func TestRowPermissionService_GetRowPermissionsTree(t *testing.T) {
	t.Run("admin returns nil", func(t *testing.T) {
		svc, _, _, _ := setupRowPermissionServiceRepoTest(t, nil, &mockAdminChecker{adminUserIDs: map[int64]bool{1: true}})

		perms, err := svc.GetRowPermissionsTree(10, 1)
		require.NoError(t, err)
		assert.Nil(t, perms)
	})

	t.Run("user only permissions", func(t *testing.T) {
		svc, rowRepo, _, _ := setupRowPermissionServiceRepoTest(t, nil, nil)
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 10, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: `{"logic":"OR","items":[{"fieldId":1,"term":"eq","value":"a"}]}`}))

		perms, err := svc.GetRowPermissionsTree(10, 7)
		require.NoError(t, err)
		assert.Len(t, perms, 1)
		assert.Equal(t, int64(7), perms[0].AuthTargetID)
	})

	t.Run("appends role permissions and tolerates role repo error", func(t *testing.T) {
		t.Run("appends role permissions", func(t *testing.T) {
			svc, rowRepo, _, _ := setupRowPermissionServiceRepoTest(t, &mockUserRoleRepo{roleIDs: []int64{9}}, nil)
			require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 10, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: `{"logic":"OR","items":[{"fieldId":1,"term":"eq","value":"a"}]}`}))
			require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 10, AuthTargetType: permission.AuthTargetTypeRole, AuthTargetID: 9, Status: permission.StatusEnabled, ExpressionTree: `{"logic":"OR","items":[{"fieldId":2,"term":"eq","value":"b"}]}`}))

			perms, err := svc.GetRowPermissionsTree(10, 7)
			require.NoError(t, err)
			assert.Len(t, perms, 2)
		})

		t.Run("role repo error does not fail", func(t *testing.T) {
			svc, rowRepo, _, _ := setupRowPermissionServiceRepoTest(t, &mockUserRoleRepo{err: errors.New("role repo failed")}, nil)
			require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 10, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: `{"logic":"OR","items":[{"fieldId":1,"term":"eq","value":"a"}]}`}))

			perms, err := svc.GetRowPermissionsTree(10, 7)
			require.NoError(t, err)
			assert.Len(t, perms, 1)
		})
	})

	t.Run("row repo error returns error", func(t *testing.T) {
		svc, _, _, db := setupRowPermissionServiceRepoTest(t, nil, nil)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())

		perms, treeErr := svc.GetRowPermissionsTree(10, 7)
		require.Error(t, treeErr)
		assert.Nil(t, perms)
	})
}

func TestRowPermissionService_BuildWhereClause(t *testing.T) {
	t.Run("admin returns nil", func(t *testing.T) {
		svc, _, _, _ := setupRowPermissionServiceRepoTest(t, nil, &mockAdminChecker{adminUserIDs: map[int64]bool{1: true}})

		result, err := svc.BuildWhereClause(10, 1)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("no permissions returns nil", func(t *testing.T) {
		svc, _, _, _ := setupRowPermissionServiceRepoTest(t, nil, nil)

		result, err := svc.BuildWhereClause(10, 7)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("skips disabled and empty expressions then combines valid with or", func(t *testing.T) {
		svc, rowRepo, _, db := setupRowPermissionServiceRepoTest(t, nil, nil)
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 10, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, ExpressionTree: `{"logic":"OR","items":[{"fieldId":1,"term":"eq","value":"ignored"}]}`}))
		require.NoError(t, db.Model(&permission.DataPermRow{}).Where("dataset_id = ? AND auth_target_id = ? AND expression_tree = ?", 10, 7, `{"logic":"OR","items":[{"fieldId":1,"term":"eq","value":"ignored"}]}`).Update("status", permission.StatusDisabled).Error)
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 10, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: ""}))
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 10, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: `{"logic":"OR","items":[{"fieldId":1,"term":"eq","value":"a"}]}`}))
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 10, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: `{"logic":"OR","items":[{"fieldId":2,"term":"eq","value":"b"}]}`}))

		result, err := svc.BuildWhereClause(10, 7)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "((`1` = ?) OR (`2` = ?))", result.Clause)
		assert.Equal(t, []interface{}{"a", "b"}, result.Args)
	})

	t.Run("invalid expression skipped when other perms valid", func(t *testing.T) {
		svc, rowRepo, _, _ := setupRowPermissionServiceRepoTest(t, nil, nil)
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 12, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: `not-json`}))
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 12, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: `{"logic":"OR","items":[{"fieldId":3,"term":"eq","value":"ok"}]}`}))

		result, err := svc.BuildWhereClause(12, 7)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "((`3` = ?))", result.Clause)
		assert.Equal(t, []interface{}{"ok"}, result.Args)
	})

	t.Run("all invalid or disabled returns nil", func(t *testing.T) {
		svc, rowRepo, _, db := setupRowPermissionServiceRepoTest(t, nil, nil)
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 13, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: ""}))
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 13, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, Status: permission.StatusEnabled, ExpressionTree: `not-json`}))
		require.NoError(t, rowRepo.Create(&permission.DataPermRow{DatasetID: 13, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 7, ExpressionTree: `{"logic":"OR","items":[{"fieldId":1,"term":"eq","value":"x"}]}`}))
		require.NoError(t, db.Model(&permission.DataPermRow{}).Where("dataset_id = ? AND auth_target_id = ? AND expression_tree = ?", 13, 7, `{"logic":"OR","items":[{"fieldId":1,"term":"eq","value":"x"}]}`).Update("status", permission.StatusDisabled).Error)

		result, err := svc.BuildWhereClause(13, 7)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("get tree error returns error", func(t *testing.T) {
		svc, _, _, db := setupRowPermissionServiceRepoTest(t, nil, nil)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())

		result, whereErr := svc.BuildWhereClause(10, 7)
		require.Error(t, whereErr)
		assert.Nil(t, result)
	})
}

func TestRowPermissionService_BuildSelectColumns(t *testing.T) {
	t.Run("no column permissions returns wildcard", func(t *testing.T) {
		svc, _, _, _ := setupRowPermissionServiceRepoTest(t, nil, nil)

		cols, err := svc.BuildSelectColumns(10, 7)
		require.NoError(t, err)
		assert.Equal(t, "*", cols)
	})

	t.Run("no disabled columns returns wildcard", func(t *testing.T) {
		svc, _, colRepo, _ := setupRowPermissionServiceRepoTest(t, nil, nil)
		require.NoError(t, colRepo.Create(&permission.DataPermColumn{DatasetID: 10, FieldName: "name", PermType: "mask", Status: permission.StatusEnabled}))

		cols, err := svc.BuildSelectColumns(10, 7)
		require.NoError(t, err)
		assert.Equal(t, "*", cols)
	})

	t.Run("excludes disabled columns and all excluded falls back to wildcard", func(t *testing.T) {
		t.Run("excludes disabled columns", func(t *testing.T) {
			svc, _, colRepo, _ := setupRowPermissionServiceRepoTest(t, nil, nil)
			require.NoError(t, colRepo.Create(&permission.DataPermColumn{DatasetID: 10, FieldName: "name", PermType: "disable", Status: permission.StatusEnabled}))
			require.NoError(t, colRepo.Create(&permission.DataPermColumn{DatasetID: 10, FieldName: "age", PermType: "disable", Status: permission.StatusEnabled}))
			require.NoError(t, colRepo.Create(&permission.DataPermColumn{DatasetID: 10, FieldName: "city", PermType: "mask", Status: permission.StatusEnabled}))

			cols, err := svc.BuildSelectColumns(10, 7)
			require.NoError(t, err)
			assert.Equal(t, "`city`", cols)
		})

		t.Run("all columns excluded falls back to wildcard", func(t *testing.T) {
			svc, _, colRepo, _ := setupRowPermissionServiceRepoTest(t, nil, nil)
			require.NoError(t, colRepo.Create(&permission.DataPermColumn{DatasetID: 11, FieldName: "only_col", PermType: "disable", Status: permission.StatusEnabled}))

			cols, err := svc.BuildSelectColumns(11, 7)
			require.NoError(t, err)
			assert.Equal(t, "*", cols)
		})
	})

	t.Run("list columns error returns wildcard", func(t *testing.T) {
		svc, _, colRepo, db := setupRowPermissionServiceRepoTest(t, nil, nil)
		require.NoError(t, colRepo.Create(&permission.DataPermColumn{DatasetID: 10, FieldName: "name", PermType: "disable", Status: permission.StatusEnabled}))
		require.NoError(t, db.Migrator().DropTable(&permission.DataPermColumn{}))

		cols, err := svc.BuildSelectColumns(10, 7)
		require.NoError(t, err)
		assert.Equal(t, "*", cols)
	})
}

func TestRowPermissionService_GetUserRoleIDs_AdditionalBranches(t *testing.T) {
	t.Run("nil repo returns nil", func(t *testing.T) {
		svc := &RowPermissionService{userRoleRepo: nil}
		ids, err := svc.GetUserRoleIDs(1)
		require.NoError(t, err)
		assert.Nil(t, ids)
	})

	t.Run("success delegates role ids", func(t *testing.T) {
		svc := &RowPermissionService{userRoleRepo: &mockUserRoleRepo{roleIDs: []int64{3, 4}}}
		ids, err := svc.GetUserRoleIDs(2)
		require.NoError(t, err)
		assert.Equal(t, []int64{3, 4}, ids)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		wantErr := errors.New("role ids failed")
		svc := &RowPermissionService{userRoleRepo: &mockUserRoleRepo{err: wantErr}}
		ids, err := svc.GetUserRoleIDs(2)
		assert.Nil(t, ids)
		assert.ErrorIs(t, err, wantErr)
	})
}
