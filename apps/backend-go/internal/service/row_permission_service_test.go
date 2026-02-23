package service

import (
	"testing"

	"dataease/backend/internal/domain/permission"
)

type mockAdminChecker struct {
	adminUserIDs map[int64]bool
}

func (m *mockAdminChecker) IsAdmin(userID int64) bool {
	return m.adminUserIDs[userID]
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
