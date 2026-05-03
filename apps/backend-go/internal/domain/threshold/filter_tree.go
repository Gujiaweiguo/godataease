package threshold

import "encoding/json"

// FilterTreeObj mirrors Java FilterTreeObj. Nesting is via FilterTreeItem.SubTree when Type=="tree".
type FilterTreeObj struct {
	Logic string           `json:"logic"`
	Items []FilterTreeItem `json:"items"`
}

// FilterTreeItem mirrors Java FilterTreeItem used by the threshold evaluator.
type FilterTreeItem struct {
	Type           string         `json:"type"`
	FieldID        json.Number    `json:"fieldId"`
	Field          any            `json:"field"`
	FilterType     string         `json:"filterType"`
	Term           string         `json:"term"`
	Value          string         `json:"value"`
	EnumValue      []string       `json:"enumValue"`
	ValueType      string         `json:"valueType"`
	FilterTypeTime string         `json:"filterTypeTime"`
	TimeType       string         `json:"timeType"`
	SubTree        *FilterTreeObj `json:"subTree"`
}
