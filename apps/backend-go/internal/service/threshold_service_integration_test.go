//go:build integration

package service

import (
	"context"
	"encoding/json"
	"testing"

	"dataease/backend/internal/domain/auto"
	thresholddomain "dataease/backend/internal/domain/threshold"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThresholdServiceIntegration_CreateAndFormInfo(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	svc := NewThresholdService(repository.NewThresholdRepository(testDB))
	req := thresholdIntegrationRequest("threshold-create-form", 4101, 5101)

	created, err := svc.Create(context.Background(), req, 101, "integration-user", 201)
	require.NoError(t, err)

	form, err := svc.FormInfo(context.Background(), created.ID, "core")
	require.NoError(t, err)
	assert.Equal(t, created.ID, form.ID)
	assert.Equal(t, req.Name, form.Name)
	assert.Equal(t, req.UIDList, form.UIDList)
	assert.Equal(t, req.RIDList, form.RIDList)
	assert.Equal(t, req.EmailList, form.EmailList)
	assert.Equal(t, req.LarkGroupList, form.LarkGroupList)
	assert.Equal(t, req.LarksuiteGroupList, form.LarksuiteGroupList)
	assert.Equal(t, req.WebhookList, form.WebhookList)
	assert.Equal(t, req.ReciFlagList, form.ReciFlagList)
	assert.Equal(t, "core", form.ResourceTable)
}

func TestThresholdServiceIntegration_PagerFiltering(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	svc := NewThresholdService(repository.NewThresholdRepository(testDB))
	for idx, name := range []string{"ops cpu threshold", "finance revenue threshold", "ops memory threshold"} {
		_, err := svc.Create(context.Background(), thresholdIntegrationRequest(name, int64(4201+idx), int64(5201+idx)), 1, "pager-user", 1)
		require.NoError(t, err)
	}

	page, err := svc.Pager(context.Background(), &thresholddomain.GridRequest{Keyword: "ops"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), page.Total)
	rows, ok := page.List.([]*thresholddomain.GridVO)
	require.True(t, ok)
	require.Len(t, rows, 2)
	assert.Contains(t, rows[0].Name, "ops")
	assert.Contains(t, rows[1].Name, "ops")
}

func TestThresholdServiceIntegration_EnableSwitch(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	svc := NewThresholdService(repository.NewThresholdRepository(testDB))
	created, err := svc.Create(context.Background(), thresholdIntegrationRequest("threshold-switch", 4301, 5301), 1, "switch-user", 1)
	require.NoError(t, err)
	assert.True(t, created.Enable)

	disabled := false
	err = svc.SwitchEnable(context.Background(), &thresholddomain.SwitchRequest{ID: created.ID, Enable: &disabled, ResourceTable: "core"})
	require.NoError(t, err)

	form, err := svc.FormInfo(context.Background(), created.ID, "core")
	require.NoError(t, err)
	require.NotNil(t, form.Enable)
	assert.False(t, *form.Enable)
}

func TestThresholdServiceIntegration_BatchReci(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	svc := NewThresholdService(repository.NewThresholdRepository(testDB))
	first, err := svc.Create(context.Background(), thresholdIntegrationRequest("threshold-batch-one", 4401, 5401), 1, "batch-user", 1)
	require.NoError(t, err)
	second, err := svc.Create(context.Background(), thresholdIntegrationRequest("threshold-batch-two", 4402, 5402), 1, "batch-user", 1)
	require.NoError(t, err)

	err = svc.BatchReci(context.Background(), &thresholddomain.BatchReciRequest{
		BaseReciDTO: thresholddomain.BaseReciDTO{
			UIDList:            []string{"updated-user"},
			RIDList:            []string{"updated-role"},
			EmailList:          []string{"updated@example.com"},
			LarkGroupList:      []string{"updated-lark"},
			LarksuiteGroupList: []string{"updated-suite"},
			WebhookList:        []string{"https://example.com/updated-hook"},
		},
		IDList: []int64{first.ID, second.ID},
	})
	require.NoError(t, err)

	for _, id := range []int64{first.ID, second.ID} {
		form, formErr := svc.FormInfo(context.Background(), id, "core")
		require.NoError(t, formErr)
		assert.Equal(t, []string{"updated-user"}, form.UIDList)
		assert.Equal(t, []string{"updated-role"}, form.RIDList)
		assert.Equal(t, []string{"updated@example.com"}, form.EmailList)
		assert.Equal(t, []string{"updated-lark"}, form.LarkGroupList)
		assert.Equal(t, []string{"updated-suite"}, form.LarksuiteGroupList)
		assert.Equal(t, []string{"https://example.com/updated-hook"}, form.WebhookList)
	}
}

func TestThresholdServiceIntegration_DeleteByChart(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	svc := NewThresholdService(repository.NewThresholdRepository(testDB))
	created, err := svc.Create(context.Background(), thresholdIntegrationRequest("threshold-delete-chart", 4501, 5501), 1, "delete-user", 1)
	require.NoError(t, err)

	err = svc.DeleteWithChart(context.Background(), created.ChartID, "core")
	require.NoError(t, err)

	exists, err := svc.AnyThreshold(context.Background(), created.ChartID, "core")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestThresholdServiceIntegration_InstancePager(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	svc := NewThresholdService(repository.NewThresholdRepository(testDB))
	created, err := svc.Create(context.Background(), thresholdIntegrationRequest("threshold-instance", 4601, 5601), 1, "instance-user", 1)
	require.NoError(t, err)
	require.NoError(t, testDB.Create([]*auto.XpackThresholdInstance{
		{ID: 9001, TaskID: created.ID, ExecTime: 1700000001, Status: true, Content: "cpu reached", Msg: ""},
		{ID: 9002, TaskID: created.ID, ExecTime: 1700000002, Status: false, Content: "memory normal", Msg: "send failed"},
	}).Error)

	page, err := svc.InstancePager(context.Background(), &thresholddomain.InstanceRequest{Keyword: "failed", ThresholdID: &created.ID}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), page.Total)
	rows, ok := page.List.([]*thresholddomain.InstanceVO)
	require.True(t, ok)
	require.Len(t, rows, 1)
	assert.Equal(t, int64(9002), rows[0].ID)
	assert.Equal(t, created.ID, rows[0].TaskID)
	assert.Equal(t, "threshold-instance", rows[0].Name)
	assert.Equal(t, "send failed", rows[0].Msg)
}

func TestThresholdServiceIntegration_Edit(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	svc := NewThresholdService(repository.NewThresholdRepository(testDB))
	created, err := svc.Create(context.Background(), thresholdIntegrationRequest("threshold-edit-before", 4701, 5701), 1, "edit-user", 1)
	require.NoError(t, err)

	editReq := thresholdIntegrationRequest("threshold-edit-after", 4701, 5701)
	editReq.ID = created.ID
	editReq.ThresholdRules = `{"logic":"or","items":[]}`
	edited, err := svc.Edit(context.Background(), editReq)
	require.NoError(t, err)
	assert.Equal(t, "threshold-edit-after", edited.Name)
	assert.Equal(t, `{"logic":"or","items":[]}`, edited.ThresholdRules)
}

func TestThresholdServiceIntegration_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	svc := NewThresholdService(repository.NewThresholdRepository(testDB))
	first, err := svc.Create(context.Background(), thresholdIntegrationRequest("threshold-del-a", 4801, 5801), 1, "del-user", 1)
	require.NoError(t, err)
	second, err := svc.Create(context.Background(), thresholdIntegrationRequest("threshold-del-b", 4802, 5802), 1, "del-user", 1)
	require.NoError(t, err)

	err = svc.Delete(context.Background(), []int64{first.ID, second.ID}, "core")
	require.NoError(t, err)

	exists, err := svc.AnyThreshold(context.Background(), first.ChartID, "core")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestThresholdServiceIntegration_AnyThreshold(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&auto.XpackThresholdInfo{}, &auto.XpackThresholdInstance{}))
	cleanupTables("xpack_threshold_instance", "xpack_threshold_info")

	svc := NewThresholdService(repository.NewThresholdRepository(testDB))
	_, err := svc.Create(context.Background(), thresholdIntegrationRequest("threshold-exists", 4901, 5901), 1, "exists-user", 1)
	require.NoError(t, err)

	exists, err := svc.AnyThreshold(context.Background(), 4901, "core")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = svc.AnyThreshold(context.Background(), 9999, "core")
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = svc.AnyThreshold(context.Background(), 4901, "snapshot")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestThresholdServiceIntegration_EvaluatorCoverage(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	rows := []map[string]any{
		{"C_100": float64(50), "C_200": "alpha", "C_300": "2024-01-01 00:00:00"},
		{"C_100": float64(150), "C_200": "beta", "C_300": "2024-06-01 00:00:00"},
		{"C_100": float64(250), "C_200": "gamma", "C_300": "2024-12-01 00:00:00"},
	}
	fieldMap := map[int64]FieldDTO{
		100: {ID: 100, Name: "Value", DataeaseName: "C_100", DeType: deTypeFloat},
		200: {ID: 200, Name: "Label", DataeaseName: "C_200", DeType: deTypeString},
		300: {ID: 300, Name: "Date", DataeaseName: "C_300", DeType: deTypeTime},
	}

	t.Run("FilterRows_AND", func(t *testing.T) {
		tree := &thresholddomain.FilterTreeObj{Logic: "and", Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("100"), FilterType: "logic", Term: "gt", Value: "100"},
			{Type: "item", FieldID: json.Number("200"), FilterType: "logic", Term: "not_empty", Value: ""},
		}}
		filtered := FilterRows(rows, tree, fieldMap)
		assert.Len(t, filtered, 2)
	})

	t.Run("FilterRows_OR", func(t *testing.T) {
		tree := &thresholddomain.FilterTreeObj{Logic: "or", Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("100"), FilterType: "logic", Term: "eq", Value: "50"},
			{Type: "item", FieldID: json.Number("200"), FilterType: "logic", Term: "eq", Value: "gamma"},
		}}
		filtered := FilterRows(rows, tree, fieldMap)
		assert.Len(t, filtered, 2)
	})

	t.Run("StringOperators", func(t *testing.T) {
		field := fieldMap[200]
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "eq", Value: "alpha"}, field))
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "not_eq", Value: "beta"}, field))
		assert.True(t, rowMatch(rows[1], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "in", Value: "alpha,beta"}, field))
		assert.True(t, rowMatch(rows[2], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "not_in", Value: "alpha,beta"}, field))
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "like", Value: "alphabet"}, field))
		assert.True(t, rowMatch(rows[2], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "not_like", Value: "alphabet"}, field))
		assert.True(t, rowMatch(map[string]any{"C_200": nil}, &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "null"}, field))
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "not_null"}, field))
		assert.True(t, rowMatch(map[string]any{"C_200": "  "}, &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "empty"}, field))
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "not_empty"}, field))
	})

	t.Run("NumericOperators", func(t *testing.T) {
		field := fieldMap[100]
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "eq", Value: "50"}, field))
		assert.True(t, rowMatch(rows[1], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "gt", Value: "100"}, field))
		assert.True(t, rowMatch(rows[1], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "ge", Value: "150"}, field))
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "lt", Value: "100"}, field))
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "le", Value: "50"}, field))
	})

	t.Run("TimeOperators", func(t *testing.T) {
		field := fieldMap[300]
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "eq", Value: "2024-01-01 00:00:00"}, field))
		assert.True(t, rowMatch(rows[2], &thresholddomain.FilterTreeItem{FilterType: "logic", Term: "gt", Value: "2024-06-01 00:00:00"}, field))
	})

	t.Run("EnumFilter", func(t *testing.T) {
		field := fieldMap[200]
		assert.True(t, rowMatch(rows[0], &thresholddomain.FilterTreeItem{FilterType: "enum", EnumValue: []string{"alpha", "beta"}}, field))
		assert.False(t, rowMatch(rows[2], &thresholddomain.FilterTreeItem{FilterType: "enum", EnumValue: []string{"alpha", "beta"}}, field))
	})

	t.Run("NestedTree", func(t *testing.T) {
		tree := &thresholddomain.FilterTreeObj{Logic: "and", Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("100"), FilterType: "logic", Term: "gt", Value: "0"},
			{Type: "tree", SubTree: &thresholddomain.FilterTreeObj{Logic: "or", Items: []thresholddomain.FilterTreeItem{
				{Type: "item", FieldID: json.Number("200"), FilterType: "logic", Term: "eq", Value: "gamma"},
				{Type: "item", FieldID: json.Number("300"), FilterType: "logic", Term: "gt", Value: "2024-05-01 00:00:00"},
			}}},
		}}
		assert.True(t, matchesConditionTree(rows[2], tree, fieldMap))
		assert.False(t, matchesConditionTree(rows[0], tree, fieldMap))
	})

	t.Run("EmptyTree", func(t *testing.T) {
		assert.True(t, matchesConditionTree(rows[0], nil, fieldMap))
		assert.True(t, matchesConditionTree(rows[0], &thresholddomain.FilterTreeObj{}, fieldMap))
	})

	t.Run("ConvertRulesToText", func(t *testing.T) {
		tree := &thresholddomain.FilterTreeObj{Logic: "and", Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("100"), FilterType: "logic", Term: "gt", Value: "100"},
			{Type: "item", FieldID: json.Number("200"), FilterType: "enum", EnumValue: []string{"alpha", "beta"}},
		}}
		text := ConvertRulesToText(tree, fieldMap)
		assert.Contains(t, text, "Value")
		assert.Contains(t, text, "gt")
		assert.Contains(t, text, "Label")
	})

	t.Run("GeneratePreviewHTML", func(t *testing.T) {
		tree := &thresholddomain.FilterTreeObj{Logic: "and", Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("100"), FilterType: "logic", Term: "gt", Value: "0"},
		}}
		html := GeneratePreviewHTML("[检测时间] [触发告警] <span id=\"changeText-100\">x</span> <span id=\"changeText-2\"><span data-mce-content=\"[告警数据]\">[告警数据]</span></span>", tree, rows, fieldMap, true, 2)
		assert.NotEmpty(t, html)
		assert.Contains(t, html, "Value")
		assert.Contains(t, html, "<table")
	})

	t.Run("DynamicValues", func(t *testing.T) {
		dynamicTree := &thresholddomain.FilterTreeObj{Logic: "or", Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("100"), FilterType: "logic", Term: "gt", Value: "max", ValueType: "dynamic"},
		}}
		resolveDynamicValues(rows, dynamicTree, fieldMap)
		assert.Equal(t, "250", dynamicTree.Items[0].Value)

		minTree := &thresholddomain.FilterTreeObj{Logic: "or", Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("100"), FilterType: "logic", Term: "lt", Value: "min", ValueType: "dynamic"},
		}}
		resolveDynamicValues(rows, minTree, fieldMap)
		assert.Equal(t, "50", minTree.Items[0].Value)

		avgTree := &thresholddomain.FilterTreeObj{Logic: "or", Items: []thresholddomain.FilterTreeItem{
			{Type: "item", FieldID: json.Number("100"), FilterType: "logic", Term: "eq", Value: "average", ValueType: "dynamic"},
		}}
		resolveDynamicValues(rows, avgTree, fieldMap)
		assert.Equal(t, "150", avgTree.Items[0].Value)
	})

	t.Run("FormatDynamicValue", func(t *testing.T) {
		field := fieldMap[100]
		assert.Equal(t, "50", formatDynamicValue(rows, &thresholddomain.FilterTreeItem{Value: "min"}, field))
		assert.Equal(t, "250", formatDynamicValue(rows, &thresholddomain.FilterTreeItem{Value: "max"}, field))
		assert.Equal(t, "150", formatDynamicValue(rows, &thresholddomain.FilterTreeItem{Value: "average"}, field))
	})
}

func TestThresholdServiceIntegration_Preview_Stub(t *testing.T) {
	if testDB == nil {
		t.Skip("test database not available")
	}
	svc := NewThresholdService(nil)
	_, err := svc.Preview(context.Background(), &thresholddomain.PreviewRequest{ChartID: 1, ThresholdRules: "{}", MsgContent: "test"})
	assert.Error(t, err)
}

func thresholdIntegrationRequest(name string, chartID, resourceID int64) *thresholddomain.CreateRequest {
	req := sampleThresholdRequest()
	req.Name = name
	req.ChartID = chartID
	req.ResourceID = resourceID
	return req
}
