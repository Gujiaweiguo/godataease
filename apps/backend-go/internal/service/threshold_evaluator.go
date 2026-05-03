package service

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	thresholddomain "dataease/backend/internal/domain/threshold"
)

const (
	deTypeString = 0
	deTypeTime   = 1
	deTypeInt    = 2
	deTypeFloat  = 3
)

const (
	itemType     = "item"
	treeType     = "tree"
	logicFilter  = "logic"
	enumFilter   = "enum"
	fixedValue   = "fixed"
	dynamicValue = "dynamic"
	nullTerm     = "null"
	eqTerm       = "eq"
	notEqTerm    = "not_eq"
	inTerm       = "in"
	notInTerm    = "not_in"
	likeTerm     = "like"
	notLikeTerm  = "not_like"
	notNullTerm  = "not_null"
	emptyTerm    = "empty"
	notEmptyTerm = "not_empty"
	gtTerm       = "gt"
	geTerm       = "ge"
	ltTerm       = "lt"
	leTerm       = "le"
)

var nonDigitRegexp = regexp.MustCompile(`\D+`)

// FieldDTO is a minimal representation of DatasetTableFieldDTO used by the evaluator.
type FieldDTO struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	DataeaseName string `json:"dataeaseName"`
	DeType       int    `json:"deType"`
}

func FilterRows(rows []map[string]any, tree *thresholddomain.FilterTreeObj, fieldMap map[int64]FieldDTO) []map[string]any {
	resolveDynamicValues(rows, tree, fieldMap)
	if len(rows) == 0 {
		return []map[string]any{}
	}
	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if matchesConditionTree(row, tree, fieldMap) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func resolveDynamicValues(rows []map[string]any, tree *thresholddomain.FilterTreeObj, fieldMap map[int64]FieldDTO) {
	if tree == nil {
		return
	}
	for idx := range tree.Items {
		item := &tree.Items[idx]
		switch item.Type {
		case treeType:
			resolveDynamicValues(rows, item.SubTree, fieldMap)
		case itemType:
			if item.ValueType != dynamicValue {
				continue
			}
			field, ok := lookupField(item, fieldMap)
			if !ok {
				continue
			}
			item.Field = field
			switch field.DeType {
			case deTypeInt, deTypeFloat:
				item.Value = formatDynamicValue(rows, item, field)
			case deTypeTime:
				item.Value = resolveDynamicTimeValue(item.Value)
			}
		}
	}
}

func matchesConditionTree(row map[string]any, tree *thresholddomain.FilterTreeObj, fieldMap map[int64]FieldDTO) bool {
	if tree == nil || len(tree.Items) == 0 {
		return true
	}
	if strings.EqualFold(tree.Logic, "or") {
		for idx := range tree.Items {
			if matchesConditionItem(row, &tree.Items[idx], fieldMap) {
				return true
			}
		}
		return false
	}
	for idx := range tree.Items {
		if !matchesConditionItem(row, &tree.Items[idx], fieldMap) {
			return false
		}
	}
	return true
}

func matchesConditionItem(row map[string]any, item *thresholddomain.FilterTreeItem, fieldMap map[int64]FieldDTO) bool {
	if item == nil {
		return false
	}
	if item.Type == itemType {
		field, ok := lookupField(item, fieldMap)
		if !ok {
			return false
		}
		return rowMatch(row, item, field)
	}
	if item.Type == treeType && item.SubTree != nil {
		return matchesConditionTree(row, item.SubTree, fieldMap)
	}
	return false
}

func rowMatch(row map[string]any, item *thresholddomain.FilterTreeItem, field FieldDTO) bool {
	valueObj := row[field.DataeaseName]
	if item.FilterType == enumFilter {
		return matchEnumFilter(valueObj, item)
	}
	switch field.DeType {
	case deTypeString:
		return matchStringOperator(valueObj, item)
	case deTypeInt, deTypeFloat:
		return matchNumericOperator(valueObj, item)
	case deTypeTime:
		return matchTimeOperator(valueObj, item)
	default:
		return true
	}
}

func matchEnumFilter(valueObj any, item *thresholddomain.FilterTreeItem) bool {
	if valueObj == nil {
		return false
	}
	valueText := fmt.Sprint(valueObj)
	return slices.Contains(item.EnumValue, valueText)
}

func matchStringOperator(valueObj any, item *thresholddomain.FilterTreeItem) bool {
	term := item.Term
	if valueObj == nil {
		return term == nullTerm
	}
	valueText := fmt.Sprint(valueObj)
	switch term {
	case eqTerm:
		return item.Value == valueText
	case notEqTerm:
		return item.Value != valueText
	case inTerm:
		return containsString(splitCSV(item.Value), valueText)
	case notInTerm:
		return !containsString(splitCSV(item.Value), valueText)
	case likeTerm:
		return strings.Contains(item.Value, valueText)
	case notLikeTerm:
		return !strings.Contains(item.Value, valueText)
	case nullTerm:
		return false
	case notNullTerm:
		return true
	case emptyTerm:
		return strings.TrimSpace(valueText) == ""
	case notEmptyTerm:
		return strings.TrimSpace(valueText) != ""
	default:
		return item.Value == valueText
	}
}

func matchNumericOperator(valueObj any, item *thresholddomain.FilterTreeItem) bool {
	if valueObj == nil || strings.TrimSpace(item.Value) == "" {
		return false
	}
	targetVal, ok := parseFloat(item.Value)
	if !ok {
		return false
	}
	originVal, ok := parseFloat(fmt.Sprint(valueObj))
	if !ok {
		return false
	}
	switch item.Term {
	case eqTerm:
		return strconv.FormatFloat(originVal, 'f', -1, 64) == strconv.FormatFloat(targetVal, 'f', -1, 64)
	case notEqTerm:
		return strconv.FormatFloat(originVal, 'f', -1, 64) != strconv.FormatFloat(targetVal, 'f', -1, 64)
	case gtTerm:
		return targetVal < originVal
	case geTerm:
		return targetVal <= originVal
	case ltTerm:
		return targetVal > originVal
	case leTerm:
		return targetVal >= originVal
	default:
		return item.Value == fmt.Sprint(valueObj)
	}
}

func matchTimeOperator(valueObj any, item *thresholddomain.FilterTreeItem) bool {
	if valueObj == nil {
		return false
	}
	target, ok := parseDigitsInt64(item.Value)
	if !ok {
		return false
	}
	origin, ok := parseDigitsInt64(fmt.Sprint(valueObj))
	if !ok {
		return false
	}
	switch item.Term {
	case eqTerm:
		return origin == target
	case notEqTerm:
		return origin != target
	case gtTerm:
		return origin > target
	case geTerm:
		return origin >= target
	case ltTerm:
		return origin < target
	case leTerm:
		return origin <= target
	default:
		return origin == target
	}
}

func formatDynamicValue(rows []map[string]any, item *thresholddomain.FilterTreeItem, field FieldDTO) string {
	mode := item.Value
	var tempVal float64
	hasValue := false
	count := 0

	for _, row := range rows {
		valueObj := row[field.DataeaseName]
		if valueObj == nil {
			continue
		}
		value, ok := parseFloat(fmt.Sprint(valueObj))
		if !ok {
			continue
		}
		switch mode {
		case "min":
			if !hasValue || value < tempVal {
				tempVal = value
				hasValue = true
			}
		case "max":
			if !hasValue || value > tempVal {
				tempVal = value
				hasValue = true
			}
		case "average":
			tempVal += value
			count++
		}
	}

	if mode == "average" {
		if count == 0 {
			return "0f"
		}
		return strconv.FormatFloat(tempVal/float64(count), 'f', -1, 64)
	}
	if !hasValue {
		return "<nil>"
	}
	return strconv.FormatFloat(tempVal, 'f', -1, 64)
}

func ConvertRulesToText(tree *thresholddomain.FilterTreeObj, fieldMap map[int64]FieldDTO) string {
	if tree == nil || len(tree.Items) == 0 {
		return ""
	}
	joiner := " AND "
	if strings.EqualFold(tree.Logic, "or") {
		joiner = " OR "
	}
	parts := make([]string, 0, len(tree.Items))
	for idx := range tree.Items {
		item := &tree.Items[idx]
		if item.Type == treeType && item.SubTree != nil {
			child := ConvertRulesToText(item.SubTree, fieldMap)
			if child != "" {
				parts = append(parts, child)
			}
			continue
		}
		field, ok := lookupField(item, fieldMap)
		if !ok {
			continue
		}
		if item.FilterType == enumFilter {
			parts = append(parts, fmt.Sprintf("%s in ( %s )", field.Name, strings.Join(item.EnumValue, ",")))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %s %s", field.Name, item.Term, item.Value))
	}
	return strings.Join(parts, joiner)
}

func GeneratePreviewHTML(template string, rules *thresholddomain.FilterTreeObj, rows []map[string]any, fieldMap map[int64]FieldDTO, showFieldValue bool, thresholdLimit int) string {
	matchedRows := FilterRows(rows, rules, fieldMap)
	if len(matchedRows) == 0 {
		return ""
	}
	if thresholdLimit <= 0 {
		thresholdLimit = 5
	}
	if len(matchedRows) > thresholdLimit {
		matchedRows = matchedRows[:thresholdLimit]
	}

	result := strings.ReplaceAll(template, "[检测时间]", time.Now().Format("2006-01-02 15:04:05"))
	result = strings.ReplaceAll(result, "[触发告警]", html.EscapeString(ConvertRulesToText(rules, fieldMap)))

	spanRegexp := regexp.MustCompile(`<span[^>]*id="changeText-(-?\d+)"[^>]*>.*?</span>`)
	result = spanRegexp.ReplaceAllStringFunc(result, func(segment string) string {
		matches := spanRegexp.FindStringSubmatch(segment)
		if len(matches) != 2 {
			return segment
		}
		fieldID, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return segment
		}
		if fieldID == 2 && strings.Contains(segment, "[告警数据]") {
			return buildThresholdTableHTML(matchedRows, rules, fieldMap)
		}
		field, ok := fieldMap[fieldID]
		if !ok {
			return segment
		}
		replacement := field.Name
		if showFieldValue {
			values := make([]string, 0, len(matchedRows))
			for _, row := range matchedRows {
				values = append(values, formatPreviewValue(row[field.DataeaseName], field.DeType))
			}
			replacement = fmt.Sprintf("%s: %s", field.Name, marshalStringList(values))
		}
		return html.EscapeString(replacement)
	})

	if strings.Contains(result, "[告警数据]") {
		result = strings.ReplaceAll(result, "[告警数据]", buildThresholdTableHTML(matchedRows, rules, fieldMap))
	}

	return result
}

func lookupField(item *thresholddomain.FilterTreeItem, fieldMap map[int64]FieldDTO) (FieldDTO, bool) {
	if item == nil {
		return FieldDTO{}, false
	}
	id, err := item.FieldID.Int64()
	if err != nil {
		return FieldDTO{}, false
	}
	field, ok := fieldMap[id]
	return field, ok
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

func containsString(values []string, target string) bool {
	return slices.Contains(values, target)
}

func parseFloat(value string) (float64, bool) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func parseDigitsInt64(value string) (int64, bool) {
	digits := nonDigitRegexp.ReplaceAllString(value, "")
	if digits == "" {
		return 0, false
	}
	parsed, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func resolveDynamicTimeValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return value
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return value
	}
	format := strings.TrimSpace(thresholdToString(payload["format"]))
	if format == "" {
		return value
	}
	goLayout := javaTimeLayoutToGo(format)
	if goLayout == "" {
		return value
	}
	now := time.Now()
	timeFlag := toInt(payload["timeFlag"])
	if timeFlag == 9 {
		count := toInt(payload["count"])
		unit := toInt(payload["unit"])
		suffix := toInt(payload["suffix"])
		resolved := applyDynamicOffset(now, count, unit, suffix)
		if unit < 4 && strings.Contains(strings.ToUpper(format), "HH") {
			timePart := strings.TrimSpace(thresholdToString(payload["time"]))
			if timePart != "" {
				return resolved.Format(goLayout) + " " + timePart
			}
		}
		return resolved.Format(goLayout)
	}

	switch strings.ToUpper(format) {
	case "YYYY":
		return applyDynamicOffset(now, ternaryInt(timeFlag == 1, 0, 1), 1, timeFlag-1).Format(goLayout)
	case "YYYY-MM":
		if timeFlag == 4 {
			return time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()).Format(goLayout)
		}
		if timeFlag == 5 {
			return time.Date(now.Year(), time.December, 31, 0, 0, 0, 0, now.Location()).Format(goLayout)
		}
		return applyDynamicOffset(now, ternaryInt(timeFlag == 1, 0, 1), 2, timeFlag-1).Format(goLayout)
	default:
		if timeFlag == 4 {
			return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format(goLayout)
		}
		if timeFlag == 5 {
			monthStartNext := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
			return monthStartNext.AddDate(0, 0, -1).Format(goLayout)
		}
		return applyDynamicOffset(now, ternaryInt(timeFlag == 1, 0, 1), 3, timeFlag-1).Format(goLayout)
	}
}

func javaTimeLayoutToGo(format string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"yyyy", "2006",
		"MM", "01",
		"DD", "02",
		"dd", "02",
		"HH", "15",
		"mm", "04",
		"ss", "05",
	)
	return replacer.Replace(format)
}

func applyDynamicOffset(now time.Time, count, unit, suffix int) time.Time {
	if suffix == 1 {
		count = -count
	}
	switch unit {
	case 1:
		return now.AddDate(count, 0, 0)
	case 2:
		return now.AddDate(0, count, 0)
	case 3:
		return now.AddDate(0, 0, count)
	default:
		return now.Add(time.Duration(count) * time.Hour)
	}
}

func ternaryInt(condition bool, left, right int) int {
	if condition {
		return left
	}
	return right
}

func thresholdToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func toInt(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(v))
		return i
	default:
		return 0
	}
}

func buildThresholdTableHTML(rows []map[string]any, rules *thresholddomain.FilterTreeObj, fieldMap map[int64]FieldDTO) string {
	fieldIDs := collectThresholdFieldIDs(rules)
	if len(fieldIDs) == 0 {
		return ""
	}
	headers := []string{"NO"}
	for _, fieldID := range fieldIDs {
		if field, ok := fieldMap[fieldID]; ok {
			headers = append(headers, html.EscapeString(field.Name))
		}
	}
	var builder strings.Builder
	builder.WriteString(`<table style="min-width:35%;border-collapse:collapse;font-family:'Segoe UI',Arial,sans-serif;font-size:14px;border:1px solid;border-radius:8px;overflow:hidden;border-spacing:0">`)
	builder.WriteString(`<thead><tr style="border-bottom:2px double;border-color: inherit;">`)
	for _, header := range headers {
		builder.WriteString(`<th style="border: 1px dashed;border-color: inherit;padding:12px;text-align:left;font-weight:bold;letter-spacing:1px;text-transform:uppercase;">`)
		builder.WriteString(header)
		builder.WriteString(`</th>`)
	}
	builder.WriteString(`</tr></thead><tbody>`)
	for idx, row := range rows {
		builder.WriteString(`<tr style="border-bottom:1px dashed;border-color: inherit;">`)
		builder.WriteString(`<td style="border: 1px dashed;border-color: inherit;padding:12px">`)
		builder.WriteString(strconv.Itoa(idx + 1))
		builder.WriteString(`</td>`)
		for _, fieldID := range fieldIDs {
			field, ok := fieldMap[fieldID]
			if !ok {
				continue
			}
			builder.WriteString(`<td style="border: 1px dashed;border-color: inherit;padding:12px">`)
			builder.WriteString(html.EscapeString(formatPreviewValue(row[field.DataeaseName], field.DeType)))
			builder.WriteString(`</td>`)
		}
		builder.WriteString(`</tr>`)
	}
	builder.WriteString(`</tbody></table>`)
	return builder.String()
}

func collectThresholdFieldIDs(tree *thresholddomain.FilterTreeObj) []int64 {
	if tree == nil {
		return nil
	}
	set := make(map[int64]struct{})
	var walk func(*thresholddomain.FilterTreeObj)
	walk = func(node *thresholddomain.FilterTreeObj) {
		if node == nil {
			return
		}
		for idx := range node.Items {
			item := &node.Items[idx]
			if item.Type == treeType {
				walk(item.SubTree)
				continue
			}
			fieldID, err := item.FieldID.Int64()
			if err == nil {
				set[fieldID] = struct{}{}
			}
		}
	}
	walk(tree)
	ids := make([]int64, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

func formatPreviewValue(value any, deType int) string {
	if value == nil {
		return ""
	}
	if deType != deTypeFloat && deType != deTypeInt {
		return fmt.Sprint(value)
	}
	number, ok := parseFloat(fmt.Sprint(value))
	if !ok {
		return fmt.Sprint(value)
	}
	if math.Mod(number, 1) == 0 {
		return strconv.FormatFloat(number, 'f', 0, 64)
	}
	return strconv.FormatFloat(number, 'f', -1, 64)
}

func marshalStringList(values []string) string {
	data, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(data)
}
