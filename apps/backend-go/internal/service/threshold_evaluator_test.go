package service

import (
	"encoding/json"
	"testing"
	"time"

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

func TestNormalizeTemplateStyles(t *testing.T) {
	t.Run("removes highlight styles from changeText spans", func(t *testing.T) {
		html := `<span id="changeText-1" style="background-color: #3370FF33">a</span><span id="changeText-2" style="color: #2b5fd9">b</span>`
		normalized := normalizeTemplateStyles(html)
		assert.NotContains(t, normalized, `background-color: #3370FF33`)
		assert.NotContains(t, normalized, `color: #2b5fd9`)
	})

	t.Run("keeps unrelated styles unchanged", func(t *testing.T) {
		html := `<span id="other" style="background-color: #3370FF33">a</span><span id="changeText-3">b</span>`
		assert.Equal(t, html, normalizeTemplateStyles(html))
	})
}

func TestThresholdResolveDynamicTimeValue(t *testing.T) {
	t.Run("returns original value for invalid payload", func(t *testing.T) {
		assert.Equal(t, "", resolveDynamicTimeValue(""))
		assert.Equal(t, "not-json", resolveDynamicTimeValue("not-json"))
		assert.Equal(t, "<nil>", resolveDynamicTimeValue(`{"timeFlag":1}`))
	})

	assertMatchesNowFormatted := func(t *testing.T, payload string, expected func(time.Time) string) {
		t.Helper()
		before := time.Now()
		actual := resolveDynamicTimeValue(payload)
		after := time.Now()
		assert.Contains(t, []string{expected(before), expected(after)}, actual)
	}

	t.Run("resolves common dynamic formats", func(t *testing.T) {
		assertMatchesNowFormatted(t, `{"format":"YYYY","timeFlag":1}`, func(now time.Time) string {
			return now.Format("2006")
		})
		assertMatchesNowFormatted(t, `{"format":"YYYY-MM","timeFlag":2}`, func(now time.Time) string {
			return applyDynamicOffset(now, 1, 2, 1).Format("2006-01")
		})
		assertMatchesNowFormatted(t, `{"format":"YYYY-MM","timeFlag":3}`, func(now time.Time) string {
			return applyDynamicOffset(now, 1, 2, 2).Format("2006-01")
		})
		assertMatchesNowFormatted(t, `{"format":"YYYY-MM","timeFlag":4}`, func(now time.Time) string {
			return time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()).Format("2006-01")
		})
		assertMatchesNowFormatted(t, `{"format":"YYYY-MM","timeFlag":5}`, func(now time.Time) string {
			return time.Date(now.Year(), time.December, 31, 0, 0, 0, 0, now.Location()).Format("2006-01")
		})
		assertMatchesNowFormatted(t, `{"format":"YYYY-MM-DD","timeFlag":4}`, func(now time.Time) string {
			return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		})
		assertMatchesNowFormatted(t, `{"format":"YYYY-MM-DD","timeFlag":5}`, func(now time.Time) string {
			monthStartNext := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
			return monthStartNext.AddDate(0, 0, -1).Format("2006-01-02")
		})
		assertMatchesNowFormatted(t, `{"format":"YYYY-MM-DD","timeFlag":9,"count":2,"unit":3,"suffix":0}`, func(now time.Time) string {
			return now.AddDate(0, 0, 2).Format("2006-01-02")
		})
		assertMatchesNowFormatted(t, `{"format":"YYYY-MM-DD HH","timeFlag":9,"count":1,"unit":2,"suffix":1,"time":"08:30:00"}`, func(now time.Time) string {
			return now.AddDate(0, -1, 0).Format("2006-01-02 15") + " 08:30:00"
		})
	})
}

func TestThresholdJavaTimeLayoutToGo(t *testing.T) {
	assert.Equal(t, "2006", javaTimeLayoutToGo("YYYY"))
	assert.Equal(t, "2006-01-02", javaTimeLayoutToGo("yyyy-MM-dd"))
	assert.Equal(t, "2006/01/02 15:04:05", javaTimeLayoutToGo("yyyy/MM/dd HH:mm:ss"))
}

func TestThresholdApplyDynamicOffset(t *testing.T) {
	base := time.Date(2026, time.March, 10, 9, 30, 0, 0, time.UTC)

	assert.Equal(t, base.AddDate(2, 0, 0), applyDynamicOffset(base, 2, 1, 0))
	assert.Equal(t, base.AddDate(0, 3, 0), applyDynamicOffset(base, 3, 2, 0))
	assert.Equal(t, base.AddDate(0, 0, 4), applyDynamicOffset(base, 4, 3, 0))
	assert.Equal(t, base.Add(-5*time.Hour), applyDynamicOffset(base, 5, 99, 1))
}

func TestThresholdPrimitiveHelpers(t *testing.T) {
	assert.Equal(t, 1, ternaryInt(true, 1, 2))
	assert.Equal(t, 2, ternaryInt(false, 1, 2))

	assert.Equal(t, "abc", thresholdToString("abc"))
	assert.Equal(t, "12", thresholdToString(json.Number("12")))
	assert.Equal(t, "true", thresholdToString(true))

	assert.Equal(t, 10, toInt(float64(10)))
	assert.Equal(t, 11, toInt(float32(11)))
	assert.Equal(t, 12, toInt(12))
	assert.Equal(t, 13, toInt(int64(13)))
	assert.Equal(t, 14, toInt(json.Number("14")))
	assert.Equal(t, 15, toInt(" 15 "))
	assert.Zero(t, toInt("bad"))
	assert.Zero(t, toInt(struct{}{}))
}

func TestThresholdResolveDynamicValues(t *testing.T) {
	rows := []map[string]any{
		{"C_123": float64(100), "C_789": "2024-01-02 03:04:05"},
		{"C_123": float64(300), "C_789": "2024-01-03 03:04:05"},
	}
	fieldMap := map[int64]FieldDTO{
		123: {ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat},
		789: {ID: 789, Name: "CreatedAt", DataeaseName: "C_789", DeType: deTypeTime},
	}
	tree := &thresholddomain.FilterTreeObj{Items: []thresholddomain.FilterTreeItem{
		{Type: itemType, FieldID: json.Number("123"), ValueType: dynamicValue, Value: "max"},
		{Type: treeType, SubTree: &thresholddomain.FilterTreeObj{Items: []thresholddomain.FilterTreeItem{{
			Type: itemType, FieldID: json.Number("789"), ValueType: dynamicValue, Value: `{"format":"YYYY-MM-DD","timeFlag":1}`,
		}}}},
	}}

	resolveDynamicValues(rows, tree, fieldMap)

	assert.Equal(t, "300", tree.Items[0].Value)
	assert.Equal(t, fieldMap[123], tree.Items[0].Field)
	assert.NotEqual(t, `{"format":"YYYY-MM-DD","timeFlag":1}`, tree.Items[1].SubTree.Items[0].Value)
	assert.Equal(t, fieldMap[789], tree.Items[1].SubTree.Items[0].Field)

	resolveDynamicValues(rows, nil, fieldMap)
}

func TestThresholdMatchesConditionItemAndLookupField(t *testing.T) {
	fieldMap := map[int64]FieldDTO{
		123: {ID: 123, Name: "Amount", DataeaseName: "C_123", DeType: deTypeFloat},
	}
	row := map[string]any{"C_123": float64(200)}

	assert.False(t, matchesConditionItem(row, nil, fieldMap))
	assert.False(t, matchesConditionItem(row, &thresholddomain.FilterTreeItem{Type: itemType, FieldID: json.Number("bad")}, fieldMap))
	assert.False(t, matchesConditionItem(row, &thresholddomain.FilterTreeItem{Type: itemType, FieldID: json.Number("999")}, fieldMap))

	item := &thresholddomain.FilterTreeItem{Type: itemType, FieldID: json.Number("123"), FilterType: logicFilter, Term: geTerm, Value: "200"}
	assert.True(t, matchesConditionItem(row, item, fieldMap))

	nested := &thresholddomain.FilterTreeItem{Type: treeType, SubTree: &thresholddomain.FilterTreeObj{Items: []thresholddomain.FilterTreeItem{*item}}}
	assert.True(t, matchesConditionItem(row, nested, fieldMap))

	field, ok := lookupField(item, fieldMap)
	assert.True(t, ok)
	assert.Equal(t, int64(123), field.ID)

	_, ok = lookupField(&thresholddomain.FilterTreeItem{FieldID: json.Number("999")}, fieldMap)
	assert.False(t, ok)
}

func TestThresholdFormattingHelpers(t *testing.T) {
	tree := &thresholddomain.FilterTreeObj{Items: []thresholddomain.FilterTreeItem{
		{Type: itemType, FieldID: json.Number("2")},
		{Type: treeType, SubTree: &thresholddomain.FilterTreeObj{Items: []thresholddomain.FilterTreeItem{{Type: itemType, FieldID: json.Number("1")}}}},
		{Type: itemType, FieldID: json.Number("2")},
	}}
	assert.Equal(t, []int64{1, 2}, collectThresholdFieldIDs(tree))
	assert.Nil(t, collectThresholdFieldIDs(nil))

	assert.Equal(t, "123", formatPreviewValue(123.0, deTypeFloat))
	assert.Equal(t, "12.5", formatPreviewValue(12.5, deTypeFloat))
	assert.Equal(t, "abc", formatPreviewValue("abc", deTypeString))
	assert.Equal(t, "", formatPreviewValue(nil, deTypeString))

	assert.Equal(t, `["a","b"]`, marshalStringList([]string{"a", "b"}))
}

func TestThresholdParserAndTextBranches(t *testing.T) {
	parsed, ok := parseFloat(" 12.5 ")
	assert.True(t, ok)
	assert.Equal(t, 12.5, parsed)
	_, ok = parseFloat("bad")
	assert.False(t, ok)

	digits, ok := parseDigitsInt64("2024-01-02 03:04:05")
	assert.True(t, ok)
	assert.Equal(t, int64(20240102030405), digits)
	_, ok = parseDigitsInt64("abc")
	assert.False(t, ok)

	assert.False(t, matchEnumFilter(nil, &thresholddomain.FilterTreeItem{EnumValue: []string{"x"}}))

	fieldMap := map[int64]FieldDTO{
		1: {ID: 1, Name: "Region", DataeaseName: "region", DeType: deTypeString},
		2: {ID: 2, Name: "Amount", DataeaseName: "amount", DeType: deTypeFloat},
	}
	tree := &thresholddomain.FilterTreeObj{Logic: "or", Items: []thresholddomain.FilterTreeItem{
		{Type: itemType, FieldID: json.Number("1"), FilterType: logicFilter, Term: eqTerm, Value: "East"},
		{Type: treeType, SubTree: &thresholddomain.FilterTreeObj{Logic: "and", Items: []thresholddomain.FilterTreeItem{{Type: itemType, FieldID: json.Number("2"), FilterType: logicFilter, Term: gtTerm, Value: "10"}}}},
	}}
	assert.Equal(t, "Region eq East OR Amount gt 10", ConvertRulesToText(tree, fieldMap))
}
