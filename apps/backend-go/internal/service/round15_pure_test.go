package service

import (
	"encoding/json"
	"testing"

	"dataease/backend/internal/domain/visualization"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// anyToString
// ---------------------------------------------------------------------------

func TestRound15_AnyToString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"nil returns empty", nil, ""},
		{"string trimmed", " hello ", "hello"},
		{"json Number", json.Number("42"), "42"},
		{"float64 integer", float64(3.0), "3"},
		{"float64 decimal", float64(3.14), "3.14"},
		{"float32 integer", float32(2.0), "2"},
		{"default int", 42, "42"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, anyToString(tc.input))
		})
	}
}

// ---------------------------------------------------------------------------
// toFloat64
// ---------------------------------------------------------------------------

func TestRound15_ToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		want   float64
		wantOK bool
	}{
		{"nil", nil, 0, false},
		{"float64", float64(3.14), 3.14, true},
		{"float32", float32(2.5), 2.5, true},
		{"json Number valid", json.Number("42.5"), 42.5, true},
		{"json Number invalid", json.Number("notanumber"), 0, false},
		{"string valid trimmed", " 3.14 ", 3.14, true},
		{"string invalid", "bad", 0, false},
		{"int via intLikeToFloat", int(42), 42, true},
		{"int64 via intLikeToFloat", int64(100), 100, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toFloat64(tc.input)
			assert.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// boolFromAnyMap
// ---------------------------------------------------------------------------

func TestRound15_BoolFromAnyMap(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"string true", "true", true},
		{"string TRUE", "TRUE", true},
		{"string 1", "1", true},
		{"string false", "false", false},
		{"string 0", "0", false},
		{"float64 1.0", float64(1.0), true},
		{"float64 0.0", float64(0.0), false},
		{"int 1", int(1), true},
		{"int 0", int(0), false},
		{"int64 1", int64(1), true},
		{"json Number 1", json.Number("1"), true},
		{"json Number true", json.Number("true"), true},
		{"nil default", nil, false},
		{"string trimmed 1", " 1 ", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, boolFromAnyMap(tc.input))
		})
	}
}

// ---------------------------------------------------------------------------
// int64FromVisualizationAny
// ---------------------------------------------------------------------------

func TestRound15_Int64FromVisualizationAny(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		want   int64
		wantOK bool
	}{
		{"int64", int64(42), 42, true},
		{"int", int(10), 10, true},
		{"float64 whole", float64(99.0), 99, true},
		{"json Number valid", json.Number("42"), 42, true},
		{"json Number invalid", json.Number("bad"), 0, false},
		{"string valid", "123", 123, true},
		{"string trimmed", " 456 ", 456, true},
		{"string invalid", "bad", 0, false},
		{"nil", nil, 0, false},
		{"bool default", true, 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := int64FromVisualizationAny(tc.input)
			assert.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// requireDatasetOrgValidator
// ---------------------------------------------------------------------------

func TestRound15_RequireDatasetOrgValidator(t *testing.T) {
	err := requireDatasetOrgValidator()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org-scoped dataset permission save is unsupported")
}

// ---------------------------------------------------------------------------
// resolveCopiedVisualizationStatus
// ---------------------------------------------------------------------------

func TestRound15_ResolveCopiedVisualizationStatus_FolderType(t *testing.T) {
	nodeType := visualizationNodeTypeFolder
	status := 0
	result := resolveCopiedVisualizationStatus(&status, &nodeType)
	assert.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

func TestRound15_ResolveCopiedVisualizationStatus_NonFolder(t *testing.T) {
	nodeType := "panel"
	status := 2
	result := resolveCopiedVisualizationStatus(&status, &nodeType)
	assert.Equal(t, &status, result)
}

func TestRound15_ResolveCopiedVisualizationStatus_NilNodeType(t *testing.T) {
	status := 3
	result := resolveCopiedVisualizationStatus(&status, nil)
	assert.Equal(t, &status, result)
}

// ---------------------------------------------------------------------------
// applySnapshotChartIntFields
// ---------------------------------------------------------------------------

func TestRound15_ApplySnapshotChartIntFields_SetsFields(t *testing.T) {
	view := &visualization.SnapshotCanvasChartView{}
	data := map[string]interface{}{
		"resultCount": 42,
		"refreshTime": 30,
	}
	applySnapshotChartIntFields(view, data)
	assert.NotNil(t, view.ResultCount)
	assert.Equal(t, 42, *view.ResultCount)
	assert.NotNil(t, view.RefreshTime)
	assert.Equal(t, 30, *view.RefreshTime)
}

func TestRound15_ApplySnapshotChartIntFields_MissingFields(t *testing.T) {
	view := &visualization.SnapshotCanvasChartView{}
	applySnapshotChartIntFields(view, map[string]interface{}{})
	assert.Nil(t, view.ResultCount)
	assert.Nil(t, view.RefreshTime)
}

// ---------------------------------------------------------------------------
// applySnapshotChartBoolFields
// ---------------------------------------------------------------------------

func TestRound15_ApplySnapshotChartBoolFields_SetsFields(t *testing.T) {
	view := &visualization.SnapshotCanvasChartView{}
	data := map[string]interface{}{
		"isPlugin":          true,
		"linkageActive":     "true",
		"jumpActive":        1,
		"aggregate":         0,
		"refreshViewEnable": "true",
	}
	applySnapshotChartBoolFields(view, data)
	assert.NotNil(t, view.IsPlugin)
	assert.True(t, *view.IsPlugin)
	assert.NotNil(t, view.LinkageActive)
	assert.True(t, *view.LinkageActive)
	assert.NotNil(t, view.JumpActive)
	assert.True(t, *view.JumpActive)
	assert.NotNil(t, view.Aggregate)
	assert.False(t, *view.Aggregate)
	assert.NotNil(t, view.RefreshViewEnable)
	assert.True(t, *view.RefreshViewEnable)
}

func TestRound15_ApplySnapshotChartBoolFields_MissingFields(t *testing.T) {
	view := &visualization.SnapshotCanvasChartView{}
	applySnapshotChartBoolFields(view, map[string]interface{}{})
	assert.Nil(t, view.IsPlugin)
}

// ---------------------------------------------------------------------------
// applySnapshotChartJSONFields
// ---------------------------------------------------------------------------

func TestRound15_ApplySnapshotChartJSONFields_SetsFields(t *testing.T) {
	view := &visualization.SnapshotCanvasChartView{}
	data := map[string]interface{}{
		"xAxis":  []string{"field1"},
		"yAxis":  "[]",
		"senior": map[string]interface{}{"key": "val"},
	}
	applySnapshotChartJSONFields(view, data)
	assert.NotNil(t, view.XAxis)
	assert.Contains(t, *view.XAxis, "field1")
	assert.NotNil(t, view.YAxis)
	assert.NotNil(t, view.Senior)
}

func TestRound15_ApplySnapshotChartJSONFields_MissingFields(t *testing.T) {
	view := &visualization.SnapshotCanvasChartView{}
	applySnapshotChartJSONFields(view, map[string]interface{}{})
	assert.Nil(t, view.XAxis)
	assert.Nil(t, view.YAxis)
}
