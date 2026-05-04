package datafilling

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtTableFieldJSONRoundTrip(t *testing.T) {
	original := ExtTableField{
		Type:     "input",
		TypeName: "文本",
		Icon:     "edit",
		ID:       "field-1",
		Settings: ExtTableFieldSetting{
			Name:      "姓名",
			Required:  true,
			Unique:    true,
			InputType: "text",
			Mapping: ExtTableFieldMapping{
				ColumnName: "user_name",
				Type:       BaseTypeNvarchar,
				Size:       128,
			},
			Options: []FieldOption{{Name: "A", Value: "a"}},
		},
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ExtTableField
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, original, decoded)
	assert.Equal(t, "nvarchar", string(BaseTypeNvarchar))
	assert.Equal(t, "text", string(BaseTypeText))
	assert.Equal(t, "number", string(BaseTypeNumber))
	assert.Equal(t, "decimal", string(BaseTypeDecimal))
	assert.Equal(t, "datetime", string(BaseTypeDatetime))
}

func TestSearchParamAndTableDataResponseJSON(t *testing.T) {
	payload := TableDataRequest{
		ID:          1,
		CurrentPage: 2,
		PageSize:    10,
		SearchParams: []SearchParam{
			{Field: "name", Term: "eq", Value: "alice"},
			{Field: "status", Multiple: true, Values: []interface{}{"enabled", "disabled"}},
		},
	}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"searchParams"`)

	response := TableDataResponse{Data: []map[string]interface{}{{"id": "1"}}, Fields: `[]`, Total: 1, CurrentPage: 2, PageSize: 10, Key: "id"}
	raw, err = json.Marshal(response)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"currentPage":2`)

	assert.Equal(t, "df_commit_log", (DfCommitLog{}).TableName())
}
