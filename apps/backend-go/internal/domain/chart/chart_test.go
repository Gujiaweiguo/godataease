package chart

import (
	"testing"
)

func TestCoreChartView_TableName(t *testing.T) {
	chart := CoreChartView{}
	if chart.TableName() != "core_chart_view" {
		t.Errorf("Expected table name 'core_chart_view', got '%s'", chart.TableName())
	}
}

func TestCoreChartView_Fields(t *testing.T) {
	title := "Sales Chart"
	sceneID := int64(1)
	tableID := int64(2)
	chartType := "bar"
	render := "antv"
	resultCount := 100
	resultMode := "custom"
	xAxis := `["category"]`
	yAxis := `["value"]`
	createBy := "admin"

	chart := CoreChartView{
		ID:          1,
		Title:       &title,
		SceneID:     &sceneID,
		TableID:     &tableID,
		Type:        &chartType,
		Render:      &render,
		ResultCount: &resultCount,
		ResultMode:  &resultMode,
		XAxis:       &xAxis,
		YAxis:       &yAxis,
		CreateBy:    &createBy,
		CreateTime:  int64Ptr(1700000000),
		UpdateTime:  int64Ptr(1700001000),
	}

	if chart.ID != 1 {
		t.Errorf("Expected ID 1, got %d", chart.ID)
	}
	if *chart.Title != "Sales Chart" {
		t.Errorf("Expected Title 'Sales Chart', got '%s'", *chart.Title)
	}
	if *chart.Type != "bar" {
		t.Errorf("Expected Type 'bar', got '%s'", *chart.Type)
	}
}

func TestCoreChartView_NilFields(t *testing.T) {
	chart := CoreChartView{ID: 1}

	if chart.Title != nil {
		t.Error("Expected Title to be nil")
	}
	if chart.SceneID != nil {
		t.Error("Expected SceneID to be nil")
	}
	if chart.Type != nil {
		t.Error("Expected Type to be nil")
	}
}

func TestChartQueryRequest_Fields(t *testing.T) {
	req := ChartQueryRequest{ID: 1}

	if req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", req.ID)
	}
}

func TestChartDataRequest_Fields(t *testing.T) {
	resultCount := 50
	req := ChartDataRequest{
		ID:          1,
		ResultCount: &resultCount,
		ResultMode:  "custom",
	}

	if req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", req.ID)
	}
	if *req.ResultCount != 50 {
		t.Errorf("Expected ResultCount 50, got %d", *req.ResultCount)
	}
	if req.ResultMode != "custom" {
		t.Errorf("Expected ResultMode 'custom', got '%s'", req.ResultMode)
	}
}

func TestChartDataResponse_Fields(t *testing.T) {
	resp := ChartDataResponse{
		ChartID: 1,
		Columns: []string{"category", "value"},
		Rows: []map[string]interface{}{
			{"category": "A", "value": 100},
			{"category": "B", "value": 200},
		},
		Total: 2,
	}

	if resp.ChartID != 1 {
		t.Errorf("Expected ChartID 1, got %d", resp.ChartID)
	}
	if len(resp.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(resp.Columns))
	}
	if len(resp.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(resp.Rows))
	}
	if resp.Total != 2 {
		t.Errorf("Expected Total 2, got %d", resp.Total)
	}
}

func TestChartField_Fields(t *testing.T) {
	datasourceID := int64(1)
	datasetTableID := int64(2)
	chartID := int64(3)

	field := ChartField{
		ID:             1,
		DatasourceID:   &datasourceID,
		DatasetTableID: &datasetTableID,
		DatasetGroupID: 10,
		ChartID:        &chartID,
		OriginName:     "sales_amount",
		Name:           "Sales Amount",
		DataeaseName:   "sales_amount",
		FieldShortName: "sales",
		GroupType:      "q",
		Type:           "DECIMAL",
		DeType:         3,
		DeExtractType:  3,
		ExtField:       0,
		Checked:        true,
		Desensitized:   false,
		Summary:        "sum",
	}

	if field.ID != 1 {
		t.Errorf("Expected ID 1, got %d", field.ID)
	}
	if field.OriginName != "sales_amount" {
		t.Errorf("Expected OriginName 'sales_amount', got '%s'", field.OriginName)
	}
	if field.GroupType != "q" {
		t.Errorf("Expected GroupType 'q', got '%s'", field.GroupType)
	}
	if !field.Checked {
		t.Error("Expected Checked to be true")
	}
}

func TestChartField_NilPointers(t *testing.T) {
	field := ChartField{
		ID:             1,
		DatasetGroupID: 10,
		OriginName:     "field1",
		Name:           "Field 1",
		DataeaseName:   "field1",
		FieldShortName: "f1",
		GroupType:      "d",
		Type:           "VARCHAR",
		DeType:         0,
		DeExtractType:  0,
		ExtField:       0,
		Checked:        true,
		Desensitized:   false,
		Summary:        "",
	}

	if field.DatasourceID != nil {
		t.Error("Expected DatasourceID to be nil")
	}
	if field.DatasetTableID != nil {
		t.Error("Expected DatasetTableID to be nil")
	}
	if field.ChartID != nil {
		t.Error("Expected ChartID to be nil")
	}
}

func TestChartFieldListResponse_Fields(t *testing.T) {
	dimensions := []ChartField{
		{ID: 1, Name: "Category", GroupType: "d"},
	}
	quotas := []ChartField{
		{ID: 2, Name: "Value", GroupType: "q"},
	}

	resp := ChartFieldListResponse{
		DimensionList: dimensions,
		QuotaList:     quotas,
	}

	if len(resp.DimensionList) != 1 {
		t.Errorf("Expected 1 dimension, got %d", len(resp.DimensionList))
	}
	if len(resp.QuotaList) != 1 {
		t.Errorf("Expected 1 quota, got %d", len(resp.QuotaList))
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
