package service

import (
	"context"
	"errors"
	"path"
	"testing"
	"time"

	"dataease/backend/internal/domain/permission"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPermissionCacheBackend struct {
	data map[string]string
}

func newMockPermissionCacheBackend() *mockPermissionCacheBackend {
	return &mockPermissionCacheBackend{data: make(map[string]string)}
}

func (m *mockPermissionCacheBackend) Get(ctx context.Context, key string) (string, error) {
	if value, ok := m.data[key]; ok {
		return value, nil
	}
	return "", errors.New("cache miss")
}

func (m *mockPermissionCacheBackend) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	switch typed := value.(type) {
	case string:
		m.data[key] = typed
	case []byte:
		m.data[key] = string(typed)
	default:
		m.data[key] = ""
	}
	return nil
}

func (m *mockPermissionCacheBackend) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *mockPermissionCacheBackend) DelByPattern(ctx context.Context, pattern string) error {
	for key := range m.data {
		matched, err := path.Match(pattern, key)
		if err != nil {
			return err
		}
		if matched {
			delete(m.data, key)
		}
	}
	return nil
}

func (m *mockPermissionCacheBackend) Exists(ctx context.Context, keys ...string) (int64, error) {
	var count int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			count++
		}
	}
	return count, nil
}

func newMockPermissionCacheService() (*permission.PermissionCacheService, *mockPermissionCacheBackend) {
	backend := newMockPermissionCacheBackend()
	return permission.NewPermissionCacheService(backend, 30*time.Minute), backend
}

func TestApplyMask_CompleteDesensitization(t *testing.T) {
	svc := &ColumnPermissionService{}
	rule := &permission.DesensitizationRule{
		BuiltInRule: permission.BuiltInRuleCompleteDesensitization,
	}

	result := svc.ApplyMask("sensitive_data", rule)
	if result != "******" {
		t.Errorf("Expected '******', got '%s'", result)
	}
}

func TestApplyMask_KeepFirstAndLastThree(t *testing.T) {
	svc := &ColumnPermissionService{}
	rule := &permission.DesensitizationRule{
		BuiltInRule: permission.BuiltInRuleKeepFirstAndLastThree,
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"abcdefghij", "abc***hij"},
		{"1234567", "123***567"},
		{"short", "XXX***XXX"},
		{"", "XXX***XXX"},
	}

	for _, tt := range tests {
		result := svc.ApplyMask(tt.input, rule)
		if result != tt.expected {
			t.Errorf("Input '%s': expected '%s', got '%s'", tt.input, tt.expected, result)
		}
	}
}

func TestApplyMask_KeepMiddleThree(t *testing.T) {
	svc := &ColumnPermissionService{}
	rule := &permission.DesensitizationRule{
		BuiltInRule: permission.BuiltInRuleKeepMiddleThree,
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"abcdefgh", "***def***"},
		{"123456", "***345***"},
		{"abc", "***XXX***"},
		{"", "***XXX***"},
	}

	for _, tt := range tests {
		result := svc.ApplyMask(tt.input, rule)
		if result != tt.expected {
			t.Errorf("Input '%s': expected '%s', got '%s'", tt.input, tt.expected, result)
		}
	}
}

func TestApplyMask_CustomRetainBeforeMAndAfterN(t *testing.T) {
	svc := &ColumnPermissionService{}
	rule := &permission.DesensitizationRule{
		BuiltInRule:       permission.BuiltInRuleCustom,
		CustomBuiltInRule: permission.CustomRuleRetainBeforeMAndAfterN,
		M:                 2,
		N:                 3,
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"abcdefghij", "ab***hij"},
		{"12345678", "12***678"},
		{"short", "sh***ort"},
		{"", "XX***XXX"},
	}

	for _, tt := range tests {
		result := svc.ApplyMask(tt.input, rule)
		if result != tt.expected {
			t.Errorf("Input '%s': expected '%s', got '%s'", tt.input, tt.expected, result)
		}
	}
}

func TestApplyMask_CustomRetainMToN(t *testing.T) {
	svc := &ColumnPermissionService{}
	rule := &permission.DesensitizationRule{
		BuiltInRule:       permission.BuiltInRuleCustom,
		CustomBuiltInRule: permission.CustomRuleRetainMToN,
		M:                 2,
		N:                 5,
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"abcdefghij", "***bcde***"},
		{"12345678", "***2345***"},
		{"", "*** ***"},
		{"a", "*** ***"},
	}

	for _, tt := range tests {
		result := svc.ApplyMask(tt.input, rule)
		if result != tt.expected {
			t.Errorf("Input '%s': expected '%s', got '%s'", tt.input, tt.expected, result)
		}
	}
}

func TestApplyMask_NilRule(t *testing.T) {
	svc := &ColumnPermissionService{}

	result := svc.ApplyMask("any_value", nil)
	if result != "******" {
		t.Errorf("Expected '******' for nil rule, got '%s'", result)
	}
}

func TestApplyMask_DefaultRule(t *testing.T) {
	svc := &ColumnPermissionService{}
	rule := &permission.DesensitizationRule{
		BuiltInRule: "unknown_rule",
	}

	result := svc.ApplyMask("test", rule)
	if result != "******" {
		t.Errorf("Expected '******' for unknown rule, got '%s'", result)
	}
}

func TestMaskRowData(t *testing.T) {
	svc := &ColumnPermissionService{}
	maskRules := map[string]*permission.DesensitizationRule{
		"phone": {BuiltInRule: permission.BuiltInRuleKeepFirstAndLastThree},
		"email": {BuiltInRule: permission.BuiltInRuleCompleteDesensitization},
	}

	row := map[string]interface{}{
		"phone":  "13812345678",
		"email":  "test@example.com",
		"name":   "John",
		"amount": 100,
	}

	result := svc.MaskRowData(row, maskRules)

	if result["phone"] != "138***678" {
		t.Errorf("Expected masked phone, got '%v'", result["phone"])
	}
	if result["email"] != "******" {
		t.Errorf("Expected masked email, got '%v'", result["email"])
	}
	if result["name"] != "John" {
		t.Errorf("Expected unchanged name, got '%v'", result["name"])
	}
	if result["amount"] != 100 {
		t.Errorf("Expected unchanged amount, got '%v'", result["amount"])
	}
}

func TestMaskRowData_EmptyRules(t *testing.T) {
	svc := &ColumnPermissionService{}
	row := map[string]interface{}{
		"name": "John",
	}

	result := svc.MaskRowData(row, nil)

	if result["name"] != "John" {
		t.Errorf("Expected unchanged data, got '%v'", result["name"])
	}
}

func TestFilterDisabledColumns(t *testing.T) {
	svc := &ColumnPermissionService{}
	disabledColumns := map[string]bool{
		"secret_field": true,
		"internal_id":  true,
	}

	row := map[string]interface{}{
		"name":         "John",
		"secret_field": "hidden",
		"internal_id":  12345,
		"status":       "active",
	}

	result := svc.FilterDisabledColumns(row, disabledColumns)

	if _, exists := result["secret_field"]; exists {
		t.Error("secret_field should be filtered out")
	}
	if _, exists := result["internal_id"]; exists {
		t.Error("internal_id should be filtered out")
	}
	if result["name"] != "John" {
		t.Errorf("Expected 'John', got '%v'", result["name"])
	}
	if result["status"] != "active" {
		t.Errorf("Expected 'active', got '%v'", result["status"])
	}
}

func TestFilterDisabledColumns_EmptyFilter(t *testing.T) {
	svc := &ColumnPermissionService{}
	row := map[string]interface{}{
		"name": "John",
	}

	result := svc.FilterDisabledColumns(row, nil)

	if len(result) != 1 {
		t.Errorf("Expected 1 field, got %d", len(result))
	}
}

func TestRetainMToN_MEquals1(t *testing.T) {
	svc := &ColumnPermissionService{}

	result := svc.retainMToN("abcdefghij", 1, 4)
	if result != "abcd***" {
		t.Errorf("Expected 'abcd***', got '%s'", result)
	}
}

func TestRetainBeforeMAndAfterN_ZeroValues(t *testing.T) {
	svc := &ColumnPermissionService{}

	result := svc.retainBeforeMAndAfterN("test", 0, 0)
	if result != "******" {
		t.Errorf("Expected '******', got '%s'", result)
	}
}

func TestApplyCustomRule_DefaultCase(t *testing.T) {
	svc := &ColumnPermissionService{}

	rule := &permission.DesensitizationRule{
		CustomBuiltInRule: "unknown_custom_rule",
	}

	result := svc.applyCustomRule("test", rule)
	if result != "******" {
		t.Errorf("Expected '******' for unknown custom rule, got '%s'", result)
	}
}

func TestApplyCustomRule_NilRule(t *testing.T) {
	svc := &ColumnPermissionService{}

	result := svc.applyCustomRule("test", nil)
	if result != "******" {
		t.Errorf("Expected '******' for nil rule, got '%s'", result)
	}
}

func TestRetainBeforeMAndAfterN_NegativeValues(t *testing.T) {
	svc := &ColumnPermissionService{}

	// Negative M and N should be treated as 0
	result := svc.retainBeforeMAndAfterN("test", -1, -1)
	if result != "******" {
		t.Errorf("Expected '******' for negative values, got '%s'", result)
	}

	// Negative M, positive N - m becomes 0, so retain 0 before, 2 after
	result = svc.retainBeforeMAndAfterN("test", -1, 2)
	if result != "***st" {
		t.Errorf("Expected '***st' for negative M, got '%s'", result)
	}

	// Positive M, negative N - n becomes 0, so retain 2 before, 0 after
	result = svc.retainBeforeMAndAfterN("test", 2, -1)
	if result != "te***" {
		t.Errorf("Expected 'te***' for negative N, got '%s'", result)
	}
}

func TestRetainMToN_EdgeCases(t *testing.T) {
	svc := &ColumnPermissionService{}

	// M >= N case
	result := svc.retainMToN("test", 3, 2)
	if result != "*** ***" {
		t.Errorf("Expected '*** ***' for M >= N, got '%s'", result)
	}

	// M and N out of bounds - returns as much as possible
	result = svc.retainMToN("ab", 0, 5)
	if result != "ab***" {
		t.Errorf("Expected 'ab***' for out of bounds, got '%s'", result)
	}
}

func TestColumnPermissionService_ParseMaskRule_Unit(t *testing.T) {
	svc := &ColumnPermissionService{}
	assert.Nil(t, svc.parseMaskRule(""))
	assert.Nil(t, svc.parseMaskRule("not-json"))

	rule := svc.parseMaskRule(`{"builtInRule":"complete_desensitization"}`)
	require.NotNil(t, rule)
	assert.Equal(t, "complete_desensitization", rule.BuiltInRule)
}

func TestColumnPermissionService_ToString(t *testing.T) {
	assert.Equal(t, "value", toString("value"))
	assert.Equal(t, "bytes", toString([]byte("bytes")))
	assert.Equal(t, "", toString(nil))
	assert.Equal(t, "", toString(123))
}

func TestColumnPermissionService_MaskRowData_NilMaskedValueUsesEmptyString(t *testing.T) {
	svc := &ColumnPermissionService{}
	maskRules := map[string]*permission.DesensitizationRule{
		"phone": {BuiltInRule: permission.BuiltInRuleKeepFirstAndLastThree},
	}

	result := svc.MaskRowData(map[string]interface{}{"phone": nil, "name": "John"}, maskRules)
	assert.Equal(t, "XXX***XXX", result["phone"])
	assert.Equal(t, "John", result["name"])
}

func TestColumnPermissionService_FilterDisabledColumns_AllColumnsDisabledReturnsEmptyMap(t *testing.T) {
	svc := &ColumnPermissionService{}
	result := svc.FilterDisabledColumns(map[string]interface{}{"a": 1, "b": 2}, map[string]bool{"a": true, "b": true})
	assert.Empty(t, result)
}

func TestColumnPermissionService_GetColumnPermissions_UsesCacheWhenAvailable(t *testing.T) {
	cacheSvc, _ := newMockPermissionCacheService()
	svc := NewColumnPermissionService(nil, cacheSvc)

	require.NoError(t, cacheSvc.SetColumnPermissions(context.Background(), 9, []*permission.DataPermColumn{{DatasetID: 9, FieldName: "cached", PermType: permission.PermTypeMask}}))

	perms, err := svc.GetColumnPermissions(9)
	require.NoError(t, err)
	require.Len(t, perms, 1)
	assert.Equal(t, "cached", perms[0].FieldName)
}
func TestColumnPermissionService_GetDerivedRules_ReuseCachedPermissions(t *testing.T) {
	cacheSvc, _ := newMockPermissionCacheService()
	svc := NewColumnPermissionService(nil, cacheSvc)
	require.NoError(t, cacheSvc.SetColumnPermissions(context.Background(), 9, []*permission.DataPermColumn{
		{DatasetID: 9, FieldName: "secret", PermType: permission.PermTypeDisable},
		{DatasetID: 9, FieldName: "phone", PermType: permission.PermTypeMask, MaskRule: `{"builtInRule":"` + permission.BuiltInRuleCompleteDesensitization + `"}`},
	}))

	disabled, err := svc.GetDisabledColumns(9)
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"secret": true}, disabled)

	rules, err := svc.GetMaskRules(9)
	require.NoError(t, err)
	require.Contains(t, rules, "phone")
	assert.Equal(t, permission.BuiltInRuleCompleteDesensitization, rules["phone"].BuiltInRule)
}

func TestColumnPermissionService_GetMaskRules_DefaultsEmptyMaskRuleFromCache(t *testing.T) {
	cacheSvc, _ := newMockPermissionCacheService()
	svc := NewColumnPermissionService(nil, cacheSvc)
	require.NoError(t, cacheSvc.SetColumnPermissions(context.Background(), 9, []*permission.DataPermColumn{{DatasetID: 9, FieldName: "region", PermType: permission.PermTypeMask, MaskRule: ""}}))

	rules, err := svc.GetMaskRules(9)
	require.NoError(t, err)
	require.Contains(t, rules, "region")
	require.NotNil(t, rules["region"])
	assert.Equal(t, permission.BuiltInRuleCompleteDesensitization, rules["region"].BuiltInRule)
}
