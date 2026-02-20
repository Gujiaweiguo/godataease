package dataset

import (
	"testing"
)

func TestNodeTypeConstants(t *testing.T) {
	if NodeTypeFolder != "folder" {
		t.Errorf("Expected NodeTypeFolder 'folder', got '%s'", NodeTypeFolder)
	}
	if NodeTypeDataset != "dataset" {
		t.Errorf("Expected NodeTypeDataset 'dataset', got '%s'", NodeTypeDataset)
	}
}

func TestCoreDatasetGroup_TableName(t *testing.T) {
	dg := CoreDatasetGroup{}
	if dg.TableName() != "core_dataset_group" {
		t.Errorf("Expected table name 'core_dataset_group', got '%s'", dg.TableName())
	}
}

func TestCoreDatasetGroup_Fields(t *testing.T) {
	pid := int64(0)
	level := 1
	nodeType := "folder"
	dsType := "db"
	delFlag := 0

	dg := CoreDatasetGroup{
		ID:       1,
		Name:     "Sales Dataset",
		PID:      &pid,
		Level:    &level,
		NodeType: &nodeType,
		Type:     &dsType,
		DelFlag:  &delFlag,
	}

	if dg.ID != 1 {
		t.Errorf("Expected ID 1, got %d", dg.ID)
	}
	if dg.Name != "Sales Dataset" {
		t.Errorf("Expected Name 'Sales Dataset', got '%s'", dg.Name)
	}
	if *dg.PID != 0 {
		t.Errorf("Expected PID 0, got %d", *dg.PID)
	}
}

func TestCoreDatasetTable_TableName(t *testing.T) {
	dt := CoreDatasetTable{}
	if dt.TableName() != "core_dataset_table" {
		t.Errorf("Expected table name 'core_dataset_table', got '%s'", dt.TableName())
	}
}

func TestCoreDatasetTable_Fields(t *testing.T) {
	name := "orders"
	datasourceID := int64(1)
	tableName := "public.orders"
	dsType := "db"
	sqlVars := `{"vars":[]}`

	dt := CoreDatasetTable{
		ID:             1,
		Name:           &name,
		DatasourceID:   &datasourceID,
		DatasetGroupID: 10,
		PhysicalTable:  &tableName,
		Type:           &dsType,
		SQLVariables:   &sqlVars,
	}

	if dt.ID != 1 {
		t.Errorf("Expected ID 1, got %d", dt.ID)
	}
	if *dt.Name != "orders" {
		t.Errorf("Expected Name 'orders', got '%s'", *dt.Name)
	}
	if dt.DatasetGroupID != 10 {
		t.Errorf("Expected DatasetGroupID 10, got %d", dt.DatasetGroupID)
	}
}

func TestCoreDatasetTableField_TableName(t *testing.T) {
	df := CoreDatasetTableField{}
	if df.TableName() != "core_dataset_table_field" {
		t.Errorf("Expected table name 'core_dataset_table_field', got '%s'", df.TableName())
	}
}

func TestCoreDatasetTableField_Fields(t *testing.T) {
	datasourceID := int64(1)
	datasetTableID := int64(2)
	chartID := int64(3)
	originName := "order_id"
	name := "Order ID"
	dataeaseName := "order_id"
	fieldShortName := "oid"
	groupType := "d"
	fieldType := "BIGINT"
	deType := 2
	deExtractType := 2
	extField := 0
	checked := true
	params := "{}"

	df := CoreDatasetTableField{
		ID:             1,
		DatasourceID:   &datasourceID,
		DatasetTableID: &datasetTableID,
		DatasetGroupID: 10,
		ChartID:        &chartID,
		OriginName:     &originName,
		Name:           &name,
		DataeaseName:   &dataeaseName,
		FieldShortName: &fieldShortName,
		GroupType:      &groupType,
		Type:           &fieldType,
		DeType:         &deType,
		DeExtractType:  &deExtractType,
		ExtField:       &extField,
		Checked:        &checked,
		Params:         &params,
	}

	if df.ID != 1 {
		t.Errorf("Expected ID 1, got %d", df.ID)
	}
	if *df.OriginName != "order_id" {
		t.Errorf("Expected OriginName 'order_id', got '%s'", *df.OriginName)
	}
	if *df.GroupType != "d" {
		t.Errorf("Expected GroupType 'd', got '%s'", *df.GroupType)
	}
}

func TestTreeRequest_Fields(t *testing.T) {
	keyword := "sales"
	req := TreeRequest{Keyword: &keyword}

	if *req.Keyword != "sales" {
		t.Errorf("Expected Keyword 'sales', got '%s'", *req.Keyword)
	}
}

func TestTreeNode_Fields(t *testing.T) {
	children := []TreeNode{
		{ID: 2, Name: "Child", NodeType: "dataset"},
	}

	node := TreeNode{
		ID:       1,
		Name:     "Parent Folder",
		NodeType: "folder",
		Children: children,
	}

	if node.ID != 1 {
		t.Errorf("Expected ID 1, got %d", node.ID)
	}
	if node.Name != "Parent Folder" {
		t.Errorf("Expected Name 'Parent Folder', got '%s'", node.Name)
	}
	if node.NodeType != "folder" {
		t.Errorf("Expected NodeType 'folder', got '%s'", node.NodeType)
	}
	if len(node.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(node.Children))
	}
}

func TestTreeNode_NoChildren(t *testing.T) {
	node := TreeNode{
		ID:       1,
		Name:     "Leaf Node",
		NodeType: "dataset",
	}

	if node.Children != nil {
		t.Error("Expected Children to be nil")
	}
}

func TestFieldsRequest_Fields(t *testing.T) {
	req := FieldsRequest{DatasetGroupID: 1}

	if req.DatasetGroupID != 1 {
		t.Errorf("Expected DatasetGroupID 1, got %d", req.DatasetGroupID)
	}
}

func TestPreviewRequest_Fields(t *testing.T) {
	req := PreviewRequest{
		DatasetGroupID: 1,
		Limit:          100,
	}

	if req.DatasetGroupID != 1 {
		t.Errorf("Expected DatasetGroupID 1, got %d", req.DatasetGroupID)
	}
	if req.Limit != 100 {
		t.Errorf("Expected Limit 100, got %d", req.Limit)
	}
}

func TestPreviewResponse_Fields(t *testing.T) {
	resp := PreviewResponse{
		Columns: []string{"id", "name", "value"},
		Rows: []map[string]interface{}{
			{"id": 1, "name": "A", "value": 100},
		},
		Total: 1,
	}

	if len(resp.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(resp.Columns))
	}
	if len(resp.Rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(resp.Rows))
	}
}

func TestWriteRequest_Fields(t *testing.T) {
	pid := int64(0)
	dsType := "db"
	isCross := false

	req := WriteRequest{
		ID:       1,
		PID:      &pid,
		Name:     "New Dataset",
		NodeType: "dataset",
		Type:     &dsType,
		IsCross:  &isCross,
	}

	if req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", req.ID)
	}
	if req.Name != "New Dataset" {
		t.Errorf("Expected Name 'New Dataset', got '%s'", req.Name)
	}
	if req.NodeType != "dataset" {
		t.Errorf("Expected NodeType 'dataset', got '%s'", req.NodeType)
	}
}

func TestSQLPreviewRequest_Fields(t *testing.T) {
	req := SQLPreviewRequest{
		DatasourceID: 1,
		SQL:          "SELECT * FROM orders LIMIT 10",
		IsCross:      false,
	}

	if req.DatasourceID != 1 {
		t.Errorf("Expected DatasourceID 1, got %d", req.DatasourceID)
	}
	if req.SQL != "SELECT * FROM orders LIMIT 10" {
		t.Errorf("Expected SQL 'SELECT * FROM orders LIMIT 10', got '%s'", req.SQL)
	}
}

func TestSQLPreviewField_Fields(t *testing.T) {
	field := SQLPreviewField{
		OriginName: "order_id",
		DeType:     2,
	}

	if field.OriginName != "order_id" {
		t.Errorf("Expected OriginName 'order_id', got '%s'", field.OriginName)
	}
	if field.DeType != 2 {
		t.Errorf("Expected DeType 2, got %d", field.DeType)
	}
}

func TestSQLPreviewData_Fields(t *testing.T) {
	data := SQLPreviewData{
		Fields: []SQLPreviewField{
			{OriginName: "id", DeType: 2},
		},
		Data: []map[string]interface{}{
			{"id": 1},
		},
	}

	if len(data.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(data.Fields))
	}
	if len(data.Data) != 1 {
		t.Errorf("Expected 1 data row, got %d", len(data.Data))
	}
}

func TestSQLVariableDetails_Fields(t *testing.T) {
	vd := SQLVariableDetails{
		ID:              "var-1",
		VariableName:    "dateRange",
		Type:            []string{"daterange"},
		Params:          nil,
		DatasetGroupID:  1,
		DatasetTableID:  1,
		DatasetFullName: "Sales/Orders",
		DeType:          0,
	}

	if vd.ID != "var-1" {
		t.Errorf("Expected ID 'var-1', got '%s'", vd.ID)
	}
	if vd.VariableName != "dateRange" {
		t.Errorf("Expected VariableName 'dateRange', got '%s'", vd.VariableName)
	}
}

func TestEnumFilter_Fields(t *testing.T) {
	f := EnumFilter{
		FieldID:  "field-1",
		Operator: "in",
		Value:    []interface{}{"A", "B"},
	}

	if f.FieldID != "field-1" {
		t.Errorf("Expected FieldID 'field-1', got '%s'", f.FieldID)
	}
	if f.Operator != "in" {
		t.Errorf("Expected Operator 'in', got '%s'", f.Operator)
	}
}

func TestEnumValueRequest_Fields(t *testing.T) {
	req := EnumValueRequest{
		QueryID:    1,
		DisplayID:  2,
		SortID:     3,
		Sort:       "asc",
		SearchText: "test",
		Filter: []EnumFilter{
			{FieldID: "f1", Operator: "in", Value: []interface{}{"A"}},
		},
		ResultMode: 0,
	}

	if req.QueryID != 1 {
		t.Errorf("Expected QueryID 1, got %d", req.QueryID)
	}
	if len(req.Filter) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(req.Filter))
	}
}

func TestMultFieldValuesRequest_Fields(t *testing.T) {
	req := MultFieldValuesRequest{
		FieldIDs:   []int64{1, 2, 3},
		Filter:     nil,
		ResultMode: 0,
	}

	if len(req.FieldIDs) != 3 {
		t.Errorf("Expected 3 field IDs, got %d", len(req.FieldIDs))
	}
}

func TestEnumFilterClause_Fields(t *testing.T) {
	c := EnumFilterClause{
		Column: "category",
		Values: []string{"A", "B", "C"},
	}

	if c.Column != "category" {
		t.Errorf("Expected Column 'category', got '%s'", c.Column)
	}
	if len(c.Values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(c.Values))
	}
}

func TestEnumObjectColumn_Fields(t *testing.T) {
	c := EnumObjectColumn{
		Column: "value",
		Alias:  "val",
	}

	if c.Column != "value" {
		t.Errorf("Expected Column 'value', got '%s'", c.Column)
	}
	if c.Alias != "val" {
		t.Errorf("Expected Alias 'val', got '%s'", c.Alias)
	}
}
