//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColumnPermissionService_GetColumnPermissions(t *testing.T) {
	cleanupTables(&permission.DataPermColumn{})

	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	t.Run("get column permissions empty", func(t *testing.T) {
		result, err := svc.GetColumnPermissions(999)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("get column permissions with data", func(t *testing.T) {
		// Create test data
		perm := &permission.DataPermColumn{
			DatasetID:      1,
			DatasetGroupID: 1,
			FieldName:      "email",
			PermType:       permission.PermTypeMask,
			Status:         1,
		}
		require.NoError(t, testDB.Create(perm).Error)

		result, err := svc.GetColumnPermissions(1)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)
	})
}

func TestColumnPermissionService_GetDisabledColumns(t *testing.T) {
	cleanupTables(&permission.DataPermColumn{})

	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	t.Run("get disabled columns empty", func(t *testing.T) {
		result, err := svc.GetDisabledColumns(999)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("get disabled columns with data", func(t *testing.T) {
		// Create disabled column
		perm := &permission.DataPermColumn{
			DatasetID:      1,
			DatasetGroupID: 1,
			FieldName:      "secret_field",
			PermType:       permission.PermTypeDisable,
			Status:         1,
		}
		require.NoError(t, testDB.Create(perm).Error)

		// Create masked column (not disabled)
		perm2 := &permission.DataPermColumn{
			DatasetID:      1,
			DatasetGroupID: 1,
			FieldName:      "email",
			PermType:       permission.PermTypeMask,
			Status:         1,
		}
		require.NoError(t, testDB.Create(perm2).Error)

		result, err := svc.GetDisabledColumns(1)
		require.NoError(t, err)
		assert.True(t, result["secret_field"])
		assert.False(t, result["email"])
	})
}

func TestColumnPermissionService_GetMaskRules(t *testing.T) {
	cleanupTables(&permission.DataPermColumn{})

	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	t.Run("get mask rules empty", func(t *testing.T) {
		result, err := svc.GetMaskRules(999)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("get mask rules with data", func(t *testing.T) {
		// Create masked column with rule
		maskRule := `{"builtInRule":"` + permission.BuiltInRuleCompleteDesensitization + `"}`
		perm := &permission.DataPermColumn{
			DatasetID:      1,
			DatasetGroupID: 1,
			FieldName:      "phone",
			PermType:       permission.PermTypeMask,
			MaskRule:       maskRule,
			Status:         1,
		}
		require.NoError(t, testDB.Create(perm).Error)

		result, err := svc.GetMaskRules(1)
		require.NoError(t, err)
		assert.NotNil(t, result["phone"])
		assert.Equal(t, permission.BuiltInRuleCompleteDesensitization, result["phone"].BuiltInRule)
	})
}

func TestColumnPermissionService_ApplyMask(t *testing.T) {
	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	t.Run("apply mask with nil rule", func(t *testing.T) {
		result := svc.ApplyMask("test", nil)
		assert.Equal(t, "******", result)
	})

	t.Run("apply complete desensitization", func(t *testing.T) {
		rule := &permission.DesensitizationRule{
			BuiltInRule: permission.BuiltInRuleCompleteDesensitization,
		}
		result := svc.ApplyMask("sensitive", rule)
		assert.Equal(t, "******", result)
	})

	t.Run("apply keep first and last three", func(t *testing.T) {
		rule := &permission.DesensitizationRule{
			BuiltInRule: permission.BuiltInRuleKeepFirstAndLastThree,
		}
		result := svc.ApplyMask("1234567890", rule)
		assert.Equal(t, "123***890", result)
	})

	t.Run("apply keep first and last three short value", func(t *testing.T) {
		rule := &permission.DesensitizationRule{
			BuiltInRule: permission.BuiltInRuleKeepFirstAndLastThree,
		}
		result := svc.ApplyMask("abc", rule)
		assert.Equal(t, "XXX***XXX", result)
	})

	t.Run("apply keep middle three", func(t *testing.T) {
		rule := &permission.DesensitizationRule{
			BuiltInRule: permission.BuiltInRuleKeepMiddleThree,
		}
		result := svc.ApplyMask("12345678", rule)
		assert.Contains(t, result, "***")
		assert.Contains(t, result, "4")
	})

	t.Run("apply custom rule retain before m and after n", func(t *testing.T) {
		rule := &permission.DesensitizationRule{
			BuiltInRule:       permission.BuiltInRuleCustom,
			CustomBuiltInRule: permission.CustomRuleRetainBeforeMAndAfterN,
			M:                 2,
			N:                 2,
		}
		result := svc.ApplyMask("12345678", rule)
		assert.Equal(t, "12***78", result)
	})

	t.Run("apply custom rule retain m to n", func(t *testing.T) {
		rule := &permission.DesensitizationRule{
			BuiltInRule:       permission.BuiltInRuleCustom,
			CustomBuiltInRule: permission.CustomRuleRetainMToN,
			M:                 2,
			N:                 4,
		}
		result := svc.ApplyMask("12345678", rule)
		assert.Contains(t, result, "***")
	})

	t.Run("apply unknown built-in rule", func(t *testing.T) {
		rule := &permission.DesensitizationRule{
			BuiltInRule: "unknown",
		}
		result := svc.ApplyMask("test", rule)
		assert.Equal(t, "******", result)
	})
}

func TestColumnPermissionService_KeepFirstAndLastThree(t *testing.T) {
	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", "XXX***XXX"},
		{"short", "abc", "XXX***XXX"},
		{"exact 7", "1234567", "123***567"},
		{"long", "1234567890", "123***890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.keepFirstAndLastThree(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColumnPermissionService_KeepMiddleThree(t *testing.T) {
	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", "***XXX***"},
		{"short", "abc", "***XXX***"},
		{"4 chars", "abcd", "***bcd***"},
		{"8 chars", "12345678", "***456***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.keepMiddleThree(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColumnPermissionService_RetainBeforeMAndAfterN(t *testing.T) {
	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	tests := []struct {
		name     string
		input    string
		m        int
		n        int
		expected string
	}{
		{"both zero", "test", 0, 0, "******"},
		{"m only", "12345678", 3, 0, "123***"},
		{"n only", "12345678", 0, 3, "***678"},
		{"both positive", "12345678", 2, 2, "12***78"},
		{"value shorter than m+n", "abc", 2, 2, "XX***XX"},
		{"negative values", "test", -1, -1, "******"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.retainBeforeMAndAfterN(tt.input, tt.m, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColumnPermissionService_RetainMToN(t *testing.T) {
	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	tests := []struct {
		name     string
		input    string
		m        int
		n        int
		expected string
	}{
		{"both zero", "test", 0, 0, "******"},
		{"m=1, n=3", "12345678", 1, 3, "123***"},
		{"m=2, n=4", "12345678", 2, 4, "***234***"},
		{"n < m", "test", 3, 1, "*** ***"},
		{"empty input", "", 1, 3, "*** ***"},
		{"value shorter than m", "ab", 3, 5, "*** ***"},
		{"n exceeds length", "12345", 2, 10, "***2345***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.retainMToN(tt.input, tt.m, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColumnPermissionService_MaskRowData(t *testing.T) {
	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	t.Run("empty rules returns original", func(t *testing.T) {
		row := map[string]interface{}{"name": "John", "email": "john@example.com"}
		result := svc.MaskRowData(row, nil)
		assert.Equal(t, row, result)
	})

	t.Run("mask specified fields", func(t *testing.T) {
		row := map[string]interface{}{"name": "John", "email": "john@example.com"}
		rules := map[string]*permission.DesensitizationRule{
			"email": {BuiltInRule: permission.BuiltInRuleCompleteDesensitization},
		}
		result := svc.MaskRowData(row, rules)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, "******", result["email"])
	})

	t.Run("mask nil value", func(t *testing.T) {
		row := map[string]interface{}{"email": nil}
		rules := map[string]*permission.DesensitizationRule{
			"email": {BuiltInRule: permission.BuiltInRuleCompleteDesensitization},
		}
		result := svc.MaskRowData(row, rules)
		assert.Equal(t, "******", result["email"])
	})
}

func TestColumnPermissionService_FilterDisabledColumns(t *testing.T) {
	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	t.Run("empty disabled returns original", func(t *testing.T) {
		row := map[string]interface{}{"name": "John", "email": "john@example.com"}
		result := svc.FilterDisabledColumns(row, nil)
		assert.Equal(t, row, result)
	})

	t.Run("filter disabled columns", func(t *testing.T) {
		row := map[string]interface{}{
			"name":   "John",
			"email":  "john@example.com",
			"secret": "sensitive",
		}
		disabled := map[string]bool{
			"secret": true,
		}
		result := svc.FilterDisabledColumns(row, disabled)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, "john@example.com", result["email"])
		_, exists := result["secret"]
		assert.False(t, exists)
	})
}

func TestColumnPermissionService_ParseMaskRule(t *testing.T) {
	repo := repository.NewColumnPermissionRepository(testDB)
	svc := NewColumnPermissionService(repo)

	t.Run("parse valid rule", func(t *testing.T) {
		jsonRule := `{"builtInRule":"` + permission.BuiltInRuleCompleteDesensitization + `"}`
		result := svc.parseMaskRule(jsonRule)
		require.NotNil(t, result)
		assert.Equal(t, permission.BuiltInRuleCompleteDesensitization, result.BuiltInRule)
	})

	t.Run("parse empty rule", func(t *testing.T) {
		result := svc.parseMaskRule("")
		assert.Nil(t, result)
	})

	t.Run("parse invalid json", func(t *testing.T) {
		result := svc.parseMaskRule("invalid json")
		assert.Nil(t, result)
	})
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"bytes", []byte("bytes"), "bytes"},
		{"int", 123, ""},
		{"bool", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
