package service

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
)

type chartRegressionSample struct {
	Name            string                   `json:"name"`
	ChartID         int64                    `json:"chartId"`
	ResultCount     int                      `json:"resultCount"`
	ExpectedColumns []string                 `json:"expectedColumns"`
	Rows            []map[string]interface{} `json:"rows"`
	Total           int64                    `json:"total"`
}

type chartRegressionSet struct {
	Samples []chartRegressionSample `json:"samples"`
}

type fakeChartRepo struct {
	byID               map[int64]*chart.CoreChartView
	data               map[int64]chartRegressionSample
	dsFieldsByGroup    map[int64][]*dataset.CoreDatasetTableField
	chartFieldsByChart map[int64][]*dataset.CoreDatasetTableField
	fieldsByID         map[int64]*dataset.CoreDatasetTableField
	nextID             int64
}

func (r *fakeChartRepo) GetByID(id int64) (*chart.CoreChartView, error) {
	v, ok := r.byID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (r *fakeChartRepo) QueryRows(chartID int64, limit int) ([]map[string]interface{}, int64, error) {
	s, ok := r.data[chartID]
	if !ok {
		return nil, 0, errors.New("not found")
	}
	if limit < 1 {
		limit = 100
	}
	if limit > len(s.Rows) {
		limit = len(s.Rows)
	}
	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		rowCopy := make(map[string]interface{}, len(s.Rows[i]))
		for k, v := range s.Rows[i] {
			rowCopy[k] = v
		}
		result = append(result, rowCopy)
	}
	return result, s.Total, nil
}

func (r *fakeChartRepo) Update(view *chart.CoreChartView) error { return nil }

func (r *fakeChartRepo) ListDatasetFieldsByGroup(datasetGroupID int64) ([]*dataset.CoreDatasetTableField, error) {
	if r.dsFieldsByGroup == nil {
		return []*dataset.CoreDatasetTableField{}, nil
	}
	list := r.dsFieldsByGroup[datasetGroupID]
	result := make([]*dataset.CoreDatasetTableField, 0, len(list))
	for _, f := range list {
		result = append(result, cloneDatasetField(f))
	}
	return result, nil
}

func (r *fakeChartRepo) ListDatasetFieldsByChart(chartID int64) ([]*dataset.CoreDatasetTableField, error) {
	if r.chartFieldsByChart == nil {
		return []*dataset.CoreDatasetTableField{}, nil
	}
	list := r.chartFieldsByChart[chartID]
	result := make([]*dataset.CoreDatasetTableField, 0, len(list))
	for _, f := range list {
		result = append(result, cloneDatasetField(f))
	}
	return result, nil
}

func (r *fakeChartRepo) GetDatasetFieldByID(id int64) (*dataset.CoreDatasetTableField, error) {
	if r.fieldsByID == nil {
		return nil, errors.New("not found")
	}
	f, ok := r.fieldsByID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return cloneDatasetField(f), nil
}

func (r *fakeChartRepo) CountDatasetFieldName(datasetGroupID int64, name string) (int64, error) {
	if r.fieldsByID == nil {
		return 0, nil
	}
	var count int64
	for _, f := range r.fieldsByID {
		if f == nil || f.Name == nil {
			continue
		}
		if f.DatasetGroupID == datasetGroupID && strings.EqualFold(*f.Name, name) {
			count++
		}
	}
	return count, nil
}

func (r *fakeChartRepo) CreateDatasetField(field *dataset.CoreDatasetTableField) error {
	if r.fieldsByID == nil {
		r.fieldsByID = make(map[int64]*dataset.CoreDatasetTableField)
	}
	if r.chartFieldsByChart == nil {
		r.chartFieldsByChart = make(map[int64][]*dataset.CoreDatasetTableField)
	}
	if field.ID <= 0 {
		if r.nextID <= 0 {
			r.nextID = 1000
		}
		field.ID = r.nextID
		r.nextID++
	}
	cloned := cloneDatasetField(field)
	r.fieldsByID[cloned.ID] = cloned
	if cloned.ChartID != nil {
		r.chartFieldsByChart[*cloned.ChartID] = append(r.chartFieldsByChart[*cloned.ChartID], cloneDatasetField(cloned))
	}
	return nil
}

func (r *fakeChartRepo) UpdateDatasetFieldNames(id int64, dataeaseName string, fieldShortName string) error {
	if r.fieldsByID == nil {
		return nil
	}
	f, ok := r.fieldsByID[id]
	if !ok || f == nil {
		return nil
	}
	f.DataeaseName = &dataeaseName
	f.FieldShortName = &fieldShortName
	if f.ChartID != nil && r.chartFieldsByChart != nil {
		fields := r.chartFieldsByChart[*f.ChartID]
		for _, item := range fields {
			if item == nil || item.ID != id {
				continue
			}
			item.DataeaseName = &dataeaseName
			item.FieldShortName = &fieldShortName
		}
	}
	return nil
}

func (r *fakeChartRepo) DeleteDatasetField(id int64) error {
	if r.fieldsByID != nil {
		delete(r.fieldsByID, id)
	}
	if r.chartFieldsByChart != nil {
		for chartID, fields := range r.chartFieldsByChart {
			filtered := make([]*dataset.CoreDatasetTableField, 0, len(fields))
			for _, f := range fields {
				if f == nil || f.ID == id {
					continue
				}
				filtered = append(filtered, f)
			}
			r.chartFieldsByChart[chartID] = filtered
		}
	}
	return nil
}

func (r *fakeChartRepo) DeleteDatasetFieldsByChart(chartID int64) error {
	if r.chartFieldsByChart == nil {
		return nil
	}
	for _, f := range r.chartFieldsByChart[chartID] {
		if f == nil {
			continue
		}
		if r.fieldsByID != nil {
			delete(r.fieldsByID, f.ID)
		}
	}
	delete(r.chartFieldsByChart, chartID)
	return nil
}

func cloneDatasetField(src *dataset.CoreDatasetTableField) *dataset.CoreDatasetTableField {
	if src == nil {
		return nil
	}
	cloned := *src
	return &cloned
}

func TestChartQueryData_RegressionSamples(t *testing.T) {
	raw, err := os.ReadFile("testdata/chart_consistency_samples.json")
	if err != nil {
		t.Fatalf("read regression samples failed: %v", err)
	}

	var set chartRegressionSet
	if err = json.Unmarshal(raw, &set); err != nil {
		t.Fatalf("parse regression samples failed: %v", err)
	}
	if len(set.Samples) == 0 {
		t.Fatal("regression sample set is empty")
	}

	repo := &fakeChartRepo{
		byID: make(map[int64]*chart.CoreChartView),
		data: make(map[int64]chartRegressionSample),
	}
	for _, sample := range set.Samples {
		repo.byID[sample.ChartID] = &chart.CoreChartView{ID: sample.ChartID}
		repo.data[sample.ChartID] = sample
	}

	svc := NewChartService(repo)

	for _, sample := range set.Samples {
		sample := sample
		t.Run(sample.Name, func(t *testing.T) {
			resultCount := sample.ResultCount
			resp, err := svc.QueryData(&chart.ChartDataRequest{ID: sample.ChartID, ResultCount: &resultCount})
			if err != nil {
				t.Fatalf("query data failed: %v", err)
			}

			if resp.ChartID != sample.ChartID {
				t.Fatalf("unexpected chart id: %d", resp.ChartID)
			}
			if resp.Total != sample.Total {
				t.Fatalf("unexpected total: %d", resp.Total)
			}

			expectedColumns := append([]string(nil), sample.ExpectedColumns...)
			sort.Strings(expectedColumns)
			if !reflect.DeepEqual(resp.Columns, expectedColumns) {
				t.Fatalf("unexpected columns: got=%v want=%v", resp.Columns, expectedColumns)
			}

			expectedRows := sample.Rows
			if sample.ResultCount < len(expectedRows) {
				expectedRows = expectedRows[:sample.ResultCount]
			}
			if !reflect.DeepEqual(resp.Rows, expectedRows) {
				t.Fatalf("unexpected rows: got=%v want=%v", resp.Rows, expectedRows)
			}
		})
	}
}

func TestChartListByDQ_SplitsAndCount(t *testing.T) {
	nameD := "region"
	originD := "region"
	dataeaseD := "region"
	groupD := "d"
	typeD := "VARCHAR"
	deTypeD := 0
	checked := true

	nameQ := "amount"
	originQ := "amount"
	dataeaseQ := "amount"
	groupQ := "q"
	typeQ := "DECIMAL"
	deTypeQ := 3

	repo := &fakeChartRepo{
		byID:               map[int64]*chart.CoreChartView{},
		data:               map[int64]chartRegressionSample{},
		dsFieldsByGroup:    map[int64][]*dataset.CoreDatasetTableField{},
		chartFieldsByChart: map[int64][]*dataset.CoreDatasetTableField{},
		fieldsByID:         map[int64]*dataset.CoreDatasetTableField{},
	}
	repo.dsFieldsByGroup[11] = []*dataset.CoreDatasetTableField{
		{ID: 1, DatasetGroupID: 11, Name: &nameD, OriginName: &originD, DataeaseName: &dataeaseD, GroupType: &groupD, Type: &typeD, DeType: &deTypeD, Checked: &checked},
		{ID: 2, DatasetGroupID: 11, Name: &nameQ, OriginName: &originQ, DataeaseName: &dataeaseQ, GroupType: &groupQ, Type: &typeQ, DeType: &deTypeQ, Checked: &checked},
	}

	svc := NewChartService(repo)
	result, err := svc.ListByDQ(11, 99)
	if err != nil {
		t.Fatalf("ListByDQ failed: %v", err)
	}
	if len(result.DimensionList) != 1 {
		t.Fatalf("expected 1 dimension field, got %d", len(result.DimensionList))
	}
	if len(result.QuotaList) != 2 {
		t.Fatalf("expected 2 quota fields (including count), got %d", len(result.QuotaList))
	}
}

func TestChartCopyAndDeleteField(t *testing.T) {
	name := "sales"
	origin := "sales"
	dataease := "sales"
	group := "q"
	typeName := "DECIMAL"
	deType := 3
	checked := true
	repo := &fakeChartRepo{
		byID:               map[int64]*chart.CoreChartView{},
		data:               map[int64]chartRegressionSample{},
		dsFieldsByGroup:    map[int64][]*dataset.CoreDatasetTableField{},
		chartFieldsByChart: map[int64][]*dataset.CoreDatasetTableField{},
		fieldsByID:         map[int64]*dataset.CoreDatasetTableField{},
		nextID:             2000,
	}
	repo.fieldsByID[10] = &dataset.CoreDatasetTableField{ID: 10, DatasetGroupID: 11, Name: &name, OriginName: &origin, DataeaseName: &dataease, GroupType: &group, Type: &typeName, DeType: &deType, Checked: &checked}

	svc := NewChartService(repo)
	if err := svc.CopyField(10, 99); err != nil {
		t.Fatalf("CopyField failed: %v", err)
	}
	if len(repo.chartFieldsByChart[99]) != 1 {
		t.Fatalf("expected 1 copied field, got %d", len(repo.chartFieldsByChart[99]))
	}
	copiedID := repo.chartFieldsByChart[99][0].ID
	if copiedID == 0 {
		t.Fatal("expected copied field id assigned")
	}
	if repo.chartFieldsByChart[99][0].DataeaseName == nil {
		t.Fatal("expected copied field dataeaseName generated")
	}
	if !strings.HasPrefix(*repo.chartFieldsByChart[99][0].DataeaseName, "f_") {
		t.Fatalf("expected dataeaseName prefixed with f_, got %s", *repo.chartFieldsByChart[99][0].DataeaseName)
	}
	if len(*repo.chartFieldsByChart[99][0].DataeaseName) != 18 {
		t.Fatalf("expected dataeaseName length 18, got %d", len(*repo.chartFieldsByChart[99][0].DataeaseName))
	}

	if err := svc.DeleteField(copiedID); err != nil {
		t.Fatalf("DeleteField failed: %v", err)
	}
	if len(repo.chartFieldsByChart[99]) != 0 {
		t.Fatalf("expected copied field deleted, got %d", len(repo.chartFieldsByChart[99]))
	}
}

// Test helper functions
func TestChart_marshalJSONField(t *testing.T) {
	tests := []struct {
		name     string
		body     map[string]interface{}
		key      string
		wantStr  string
		wantBool bool
	}{
		{"key exists", map[string]interface{}{"field": "value"}, "field", "\"value\"", true},
		{"key not exists", map[string]interface{}{"field": "value"}, "other", "", false},
		{"nil body", nil, "field", "", false},
	}

	// Test object value separately
	objBody := map[string]interface{}{"field": map[string]interface{}{"a": 1}}
	gotStr, gotBool := marshalJSONField(objBody, "field")
	if !gotBool {
		t.Error("marshalJSONField object value should return true")
	}
	if gotStr != `{"a":1}` {
		t.Errorf("marshalJSONField object value = %v, want {\"a\":1}", gotStr)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotBool := marshalJSONField(tt.body, tt.key)
			if gotBool != tt.wantBool {
				t.Errorf("marshalJSONField() bool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotBool && gotStr != tt.wantStr {
				t.Errorf("marshalJSONField() str = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestChart_stringFromAny(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantStr  string
		wantBool bool
	}{
		{"valid string", "hello", "hello", true},
		{"string with spaces", "  hello  ", "hello", true},
		{"empty string", "", "", false},
		{"whitespace only", "   ", "", false},
		{"not a string", 123, "", false},
		{"nil", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotBool := stringFromAny(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("stringFromAny() bool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotStr != tt.wantStr {
				t.Errorf("stringFromAny() str = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestChart_int64FromAny(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantVal  int64
		wantBool bool
	}{
		{"int64", int64(100), 100, true},
		{"int", int(50), 50, true},
		{"float64", float64(123.45), 123, true},
		{"json.Number", json.Number("999"), 999, true},
		{"string number", "456", 456, true},
		{"string with spaces", " 789 ", 789, true},
		{"invalid string", "abc", 0, false},
		{"nil", nil, 0, false},
		{"other type", []int{}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotBool := int64FromAny(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("int64FromAny() bool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotVal != tt.wantVal {
				t.Errorf("int64FromAny() val = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestChart_intFromAny(t *testing.T) {
	val, ok := intFromAny(int64(100))
	if !ok || val != 100 {
		t.Errorf("intFromAny() = %d, %v, want 100, true", val, ok)
	}

	_, ok = intFromAny("invalid")
	if ok {
		t.Error("intFromAny() should return false for invalid input")
	}
}

func TestChart_stringValue(t *testing.T) {
	s := "  hello  "
	if got := stringValue(&s); got != "hello" {
		t.Errorf("stringValue() = %v, want hello", got)
	}

	if got := stringValue(nil); got != "" {
		t.Errorf("stringValue(nil) = %v, want empty", got)
	}
}

func TestChart_intPointerValue(t *testing.T) {
	i := 42
	if got := intPointerValue(&i); got != 42 {
		t.Errorf("intPointerValue() = %v, want 42", got)
	}

	if got := intPointerValue(nil); got != 0 {
		t.Errorf("intPointerValue(nil) = %v, want 0", got)
	}
}

func TestChart_boolPointerValue(t *testing.T) {
	b := true
	if got := boolPointerValue(&b); got != true {
		t.Errorf("boolPointerValue() = %v, want true", got)
	}

	if got := boolPointerValue(nil); got != false {
		t.Errorf("boolPointerValue(nil) = %v, want false", got)
	}
}

func TestChart_fieldNameShort(t *testing.T) {
	result := fieldNameShort("test")
	if !strings.HasPrefix(result, "f_") {
		t.Errorf("fieldNameShort() = %v, want f_ prefix", result)
	}
	if len(result) != 18 { // f_ + 16 hex chars (truncated from md5)
		t.Errorf("fieldNameShort() length = %d, want 18", len(result))
	}
}

func TestChart_convertToChartField(t *testing.T) {
	tests := []struct {
		name      string
		field     *dataset.CoreDatasetTableField
		wantGroup string
		wantSumm  string
	}{
		{
			name: "text field with default group",
			field: &dataset.CoreDatasetTableField{
				ID:        1,
				DeType:    intPtrForChart(0),
				GroupType: strPtr(""),
			},
			wantGroup: "d",
			wantSumm:  "count",
		},
		{
			name: "numeric field with default group",
			field: &dataset.CoreDatasetTableField{
				ID:        2,
				DeType:    intPtrForChart(2),
				GroupType: strPtr(""),
			},
			wantGroup: "q",
			wantSumm:  "sum",
		},
		{
			name: "double field with default group",
			field: &dataset.CoreDatasetTableField{
				ID:        3,
				DeType:    intPtrForChart(3),
				GroupType: strPtr(""),
			},
			wantGroup: "q",
			wantSumm:  "sum",
		},
		{
			name: "datetime field with count summary",
			field: &dataset.CoreDatasetTableField{
				ID:        4,
				DeType:    intPtrForChart(1),
				GroupType: strPtr(""),
			},
			wantGroup: "d",
			wantSumm:  "count",
		},
		{
			name: "field with ID -1 uses count summary",
			field: &dataset.CoreDatasetTableField{
				ID:        -1,
				DeType:    intPtrForChart(2),
				GroupType: strPtr(""),
			},
			wantGroup: "q",
			wantSumm:  "count",
		},
		{
			name: "field with explicit group type",
			field: &dataset.CoreDatasetTableField{
				ID:        5,
				DeType:    intPtrForChart(0),
				GroupType: strPtr("q"),
			},
			wantGroup: "q",
			wantSumm:  "count", // deType=0 uses count summary
		},
		{
			name: "field with deType 7 uses count summary",
			field: &dataset.CoreDatasetTableField{
				ID:        6,
				DeType:    intPtrForChart(7),
				GroupType: strPtr(""),
			},
			wantGroup: "d",
			wantSumm:  "count",
		},
		{
			name: "field with nil DeType uses 0",
			field: &dataset.CoreDatasetTableField{
				ID:            7,
				DeType:        nil,
				DeExtractType: intPtrForChart(2),
				GroupType:     strPtr(""),
			},
			wantGroup: "d",     // nil DeType defaults to 0
			wantSumm:  "count", // deType=0 uses count summary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToChartField(tt.field)
			if result.GroupType != tt.wantGroup {
				t.Errorf("GroupType = %v, want %v", result.GroupType, tt.wantGroup)
			}
			if result.Summary != tt.wantSumm {
				t.Errorf("Summary = %v, want %v", result.Summary, tt.wantSumm)
			}
		})
	}
}

func intPtrForChart(i int) *int { return &i }
