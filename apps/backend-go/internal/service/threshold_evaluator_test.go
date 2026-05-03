package service

import (
	"encoding/json"
	"testing"

	thresholddomain "dataease/backend/internal/domain/threshold"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThresholdFilterRows(t *testing.T) {
	rows := []map[string]any{
		{"C_123": float64(100), "C_456": "apple"},
		{"C_123": float64(200), "C_456": "banana"},
		{"C_123": float64(300), "C_456": "cherry"},
	}
	fieldMap := map[int64]FieldDTO{
		123: {ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat},
		456: {ID: 456, Name: "Fruit", DataeaseName: "C_456", DeType: deTypeString},
	}

	andTree := &thresholddomain.FilterTreeObj{
		Logic: "and",
		Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("123"), FilterType: "logic", Term: "gt", Value: "150"},
			{Type: "item", FieldID: json.Number("456"), FilterType: "logic", Term: "in", Value: "banana,cherry"},
		},
	}
	orTree := &thresholddomain.FilterTreeObj{
		Logic: "or",
		Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("123"), FilterType: "logic", Term: "eq", Value: "100"},
			{Type: "item", FieldID: json.Number("456"), FilterType: "logic", Term: "eq", Value: "cherry"},
		},
	}

	assert.Len(t, FilterRows(rows, andTree, fieldMap), 2)
	filtered := FilterRows(rows, orTree, fieldMap)
	assert.Len(t, filtered, 2)
	assert.Equal(t, "apple", filtered[0]["C_456"])
	assert.Equal(t, "cherry", filtered[1]["C_456"])
}

func TestThresholdRowMatch_StringOperators(t *testing.T) {
	field := FieldDTO{ID: 456, Name: "Fruit", DataeaseName: "C_456", DeType: deTypeString}
	tests := []struct {
		name string
		row  map[string]any
		item thresholddomain.FilterTreeItem
		want bool
	}{
		{name: "eq", row: map[string]any{"C_456": "apple"}, item: thresholddomain.FilterTreeItem{Term: "eq", Value: "apple"}, want: true},
		{name: "not_eq", row: map[string]any{"C_456": "apple"}, item: thresholddomain.FilterTreeItem{Term: "not_eq", Value: "banana"}, want: true},
		{name: "in", row: map[string]any{"C_456": "banana"}, item: thresholddomain.FilterTreeItem{Term: "in", Value: "apple,banana"}, want: true},
		{name: "not_in", row: map[string]any{"C_456": "cherry"}, item: thresholddomain.FilterTreeItem{Term: "not_in", Value: "apple,banana"}, want: true},
		{name: "like", row: map[string]any{"C_456": "app"}, item: thresholddomain.FilterTreeItem{Term: "like", Value: "pineapple"}, want: true},
		{name: "not_like", row: map[string]any{"C_456": "pear"}, item: thresholddomain.FilterTreeItem{Term: "not_like", Value: "pineapple"}, want: true},
		{name: "null", row: map[string]any{"C_456": nil}, item: thresholddomain.FilterTreeItem{Term: "null"}, want: true},
		{name: "not_null", row: map[string]any{"C_456": "apple"}, item: thresholddomain.FilterTreeItem{Term: "not_null"}, want: true},
		{name: "empty", row: map[string]any{"C_456": "   "}, item: thresholddomain.FilterTreeItem{Term: "empty"}, want: true},
		{name: "not_empty", row: map[string]any{"C_456": "apple"}, item: thresholddomain.FilterTreeItem{Term: "not_empty"}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.item.FilterType = "logic"
			assert.Equal(t, tt.want, rowMatch(tt.row, &tt.item, field))
		})
	}
}

func TestThresholdRowMatch_NumericOperators(t *testing.T) {
	field := FieldDTO{ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat}
	tests := []struct {
		name     string
		term     string
		value    string
		rowValue any
		want     bool
	}{
		{name: "eq", term: "eq", value: "100", rowValue: float64(100), want: true},
		{name: "not_eq", term: "not_eq", value: "100", rowValue: float64(200), want: true},
		{name: "gt", term: "gt", value: "100", rowValue: float64(101), want: true},
		{name: "ge", term: "ge", value: "100", rowValue: float64(100), want: true},
		{name: "lt", term: "lt", value: "100", rowValue: float64(99), want: true},
		{name: "le", term: "le", value: "100", rowValue: float64(100), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := thresholddomain.FilterTreeItem{FilterType: "logic", Term: tt.term, Value: tt.value}
			assert.Equal(t, tt.want, rowMatch(map[string]any{"C_123": tt.rowValue}, &item, field))
		})
	}
}

func TestThresholdRowMatch_TimeOperators(t *testing.T) {
	field := FieldDTO{ID: 789, Name: "CreatedAt", DataeaseName: "C_789", DeType: deTypeTime}
	tests := []struct {
		name     string
		term     string
		value    string
		rowValue any
		want     bool
	}{
		{name: "eq", term: "eq", value: "2024-01-02 03:04:05", rowValue: "2024/01/02 03:04:05", want: true},
		{name: "not_eq", term: "not_eq", value: "2024-01-02 03:04:05", rowValue: "2024-01-03 03:04:05", want: true},
		{name: "gt", term: "gt", value: "2024-01-02 03:04:05", rowValue: "2024-01-03 03:04:05", want: true},
		{name: "ge", term: "ge", value: "2024-01-02 03:04:05", rowValue: "2024-01-02 03:04:05", want: true},
		{name: "lt", term: "lt", value: "2024-01-02 03:04:05", rowValue: "2024-01-01 03:04:05", want: true},
		{name: "le", term: "le", value: "2024-01-02 03:04:05", rowValue: "2024-01-02 03:04:05", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := thresholddomain.FilterTreeItem{FilterType: "logic", Term: tt.term, Value: tt.value}
			assert.Equal(t, tt.want, rowMatch(map[string]any{"C_789": tt.rowValue}, &item, field))
		})
	}
}

func TestThresholdRowMatch_EnumFilter(t *testing.T) {
	field := FieldDTO{ID: 456, Name: "Fruit", DataeaseName: "C_456", DeType: deTypeString}
	item := thresholddomain.FilterTreeItem{FilterType: "enum", EnumValue: []string{"apple", "banana"}}
	assert.True(t, rowMatch(map[string]any{"C_456": "banana"}, &item, field))
	assert.False(t, rowMatch(map[string]any{"C_456": "cherry"}, &item, field))
}

func TestThresholdMatchesConditionTree_AND(t *testing.T) {
	row := map[string]any{"C_123": float64(200), "C_456": "banana"}
	fieldMap := map[int64]FieldDTO{
		123: {ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat},
		456: {ID: 456, Name: "Fruit", DataeaseName: "C_456", DeType: deTypeString},
	}
	tree := &thresholddomain.FilterTreeObj{Logic: "and", Items: []thresholddomain.FilterTreeItem{
		{Type: "item", FieldID: json.Number("123"), FilterType: "logic", Term: "gt", Value: "150"},
		{Type: "item", FieldID: json.Number("456"), FilterType: "logic", Term: "eq", Value: "banana"},
	}}
	assert.True(t, matchesConditionTree(row, tree, fieldMap))
}

func TestThresholdMatchesConditionTree_OR(t *testing.T) {
	row := map[string]any{"C_123": float64(200), "C_456": "banana"}
	fieldMap := map[int64]FieldDTO{
		123: {ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat},
		456: {ID: 456, Name: "Fruit", DataeaseName: "C_456", DeType: deTypeString},
	}
	tree := &thresholddomain.FilterTreeObj{Logic: "or", Items: []thresholddomain.FilterTreeItem{
		{Type: "item", FieldID: json.Number("123"), FilterType: "logic", Term: "gt", Value: "250"},
		{Type: "item", FieldID: json.Number("456"), FilterType: "logic", Term: "eq", Value: "banana"},
	}}
	assert.True(t, matchesConditionTree(row, tree, fieldMap))
}

func TestThresholdMatchesConditionTree_NestedTree(t *testing.T) {
	row := map[string]any{"C_123": float64(200), "C_456": "banana", "C_789": "2024-01-02 00:00:00"}
	fieldMap := map[int64]FieldDTO{
		123: {ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat},
		456: {ID: 456, Name: "Fruit", DataeaseName: "C_456", DeType: deTypeString},
		789: {ID: 789, Name: "CreatedAt", DataeaseName: "C_789", DeType: deTypeTime},
	}
	tree := &thresholddomain.FilterTreeObj{Logic: "and", Items: []thresholddomain.FilterTreeItem{
		{Type: "item", FieldID: json.Number("123"), FilterType: "logic", Term: "gt", Value: "150"},
		{Type: "tree", SubTree: &thresholddomain.FilterTreeObj{Logic: "or", Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("456"), FilterType: "logic", Term: "eq", Value: "banana"},
			{Type: "item", FieldID: json.Number("789"), FilterType: "logic", Term: "lt", Value: "2024-01-01 00:00:00"},
		}}},
	}}
	assert.True(t, matchesConditionTree(row, tree, fieldMap))
}

func TestThresholdMatchesConditionTree_EmptyTree(t *testing.T) {
	assert.True(t, matchesConditionTree(map[string]any{"x": 1}, nil, nil))
	assert.True(t, matchesConditionTree(map[string]any{"x": 1}, &thresholddomain.FilterTreeObj{}, nil))
}

func TestThresholdFormatDynamicValue(t *testing.T) {
	rows := []map[string]any{{"C_123": float64(100)}, {"C_123": float64(200)}, {"C_123": float64(300)}}
	field := FieldDTO{ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat}
	assert.Equal(t, "100", formatDynamicValue(rows, &thresholddomain.FilterTreeItem{Value: "min"}, field))
	assert.Equal(t, "300", formatDynamicValue(rows, &thresholddomain.FilterTreeItem{Value: "max"}, field))
	assert.Equal(t, "200", formatDynamicValue(rows, &thresholddomain.FilterTreeItem{Value: "average"}, field))
}

func TestThresholdConvertRulesToText(t *testing.T) {
	fieldMap := map[int64]FieldDTO{
		123: {ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat},
		456: {ID: 456, Name: "Fruit", DataeaseName: "C_456", DeType: deTypeString},
	}
	tree := &thresholddomain.FilterTreeObj{Logic: "and", Items: []thresholddomain.FilterTreeItem{
		{Type: "item", FieldID: json.Number("123"), FilterType: "logic", Term: "gt", Value: "100"},
		{Type: "item", FieldID: json.Number("456"), FilterType: "enum", EnumValue: []string{"apple", "banana"}},
	}}
	assert.Equal(t, "Amount gt 100 AND Fruit in ( apple,banana )", ConvertRulesToText(tree, fieldMap))
}

func TestThresholdGeneratePreviewHTML(t *testing.T) {
	rows := []map[string]any{
		{"C_123": float64(200), "C_456": "banana"},
		{"C_123": float64(300), "C_456": "cherry"},
	}
	fieldMap := map[int64]FieldDTO{
		123: {ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat},
		456: {ID: 456, Name: "Fruit", DataeaseName: "C_456", DeType: deTypeString},
	}
	rules := &thresholddomain.FilterTreeObj{Logic: "and", Items: []thresholddomain.FilterTreeItem{
		{Type: "item", FieldID: json.Number("123"), FilterType: "logic", Term: "gt", Value: "150"},
		{Type: "item", FieldID: json.Number("456"), FilterType: "logic", Term: "not_empty", Value: ""},
	}}
	template := `[检测时间] [触发告警] <span id="changeText-123">old</span> <span id="changeText-2"><span data-mce-content="[告警数据]">[告警数据]</span></span>`
	html := GeneratePreviewHTML(template, rules, rows, fieldMap, true, 1)
	require.NotEmpty(t, html)
	assert.Contains(t, html, `Amount gt 150`)
	assert.Contains(t, html, `Amount: [&#34;200&#34;]`)
	assert.Contains(t, html, `<table`)
	assert.Contains(t, html, `banana`)
	assert.NotContains(t, html, `cherry`)
}
