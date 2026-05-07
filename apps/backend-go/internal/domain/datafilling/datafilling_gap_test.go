package datafilling

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataFilling_TableNamesAndConstants(t *testing.T) {
	assert.Equal(t, "data_filling_forms", (DataFillingForm{}).TableName())
	assert.Equal(t, "df_commit_log", (DfCommitLog{}).TableName())
	assert.Equal(t, "data_filling_task", (DataFillingTask{}).TableName())
	assert.Equal(t, "data_filling_sub_task", (DataFillingSubTask{}).TableName())
	assert.Equal(t, "data_filling_sub_instance", (DataFillingSubInstance{}).TableName())

	assert.Equal(t, "folder", NodeTypeFolder)
	assert.Equal(t, "form", NodeTypeForm)
	assert.Equal(t, BaseType("nvarchar"), BaseTypeNvarchar)
	assert.Equal(t, BaseType("text"), BaseTypeText)
	assert.Equal(t, BaseType("number"), BaseTypeNumber)
	assert.Equal(t, BaseType("decimal"), BaseTypeDecimal)
	assert.Equal(t, BaseType("datetime"), BaseTypeDatetime)
}

func TestDataFilling_RequestResponseJSONShapes(t *testing.T) {
	keyword := "quarterly"
	taskID := int64(99)
	createRequest := CreateFormRequest{
		Name:           "Sales Form",
		PID:            8,
		NodeType:       NodeTypeForm,
		TableName:      "sales_form",
		DatasourceID:   12,
		Forms:          `[{"id":"f1"}]`,
		CreateIndex:    true,
		TableIndexes:   `[{"name":"idx_sales"}]`,
		UseExistsTable: true,
	}
	updateRequest := UpdateFormRequest{ID: 7, CreateFormRequest: createRequest}
	taskRequest := TaskPageRequest{TaskID: &taskID, Keyword: &keyword}
	treeRequest := TreeRequest{Keyword: &keyword}

	raw, err := json.Marshal(updateRequest)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"nodeType":"form"`)
	assert.Contains(t, string(raw), `"useExistsTable":true`)

	taskRaw, err := json.Marshal(taskRequest)
	require.NoError(t, err)
	assert.Contains(t, string(taskRaw), `"keyword":"quarterly"`)

	treeRequestRaw, err := json.Marshal(treeRequest)
	require.NoError(t, err)
	assert.Contains(t, string(treeRequestRaw), `"keyword":"quarterly"`)

	response := TaskPageResponse{
		Records: []*TaskInfoVO{{ID: 1, Name: "Q1 Collection", UIDList: []int64{2, 3}}},
		Total:   1,
		Current: 1,
		Size:    20,
	}
	tree := TreeResponse{{ID: 1, Name: "Root", NodeType: NodeTypeFolder, Children: []*TreeNode{{ID: 2, Name: "Leaf", NodeType: NodeTypeForm}}}}

	responseRaw, err := json.Marshal(response)
	require.NoError(t, err)
	assert.Contains(t, string(responseRaw), `"records"`)

	treeRaw, err := json.Marshal(tree)
	require.NoError(t, err)
	assert.Contains(t, string(treeRaw), `"children"`)
	assert.Contains(t, string(treeRaw), `"nodeType":"folder"`)
}
