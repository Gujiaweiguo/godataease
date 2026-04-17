//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/chart"

	"github.com/stretchr/testify/assert"
)

func TestChartExportServiceIntegration_InnerExportDetails(t *testing.T) {
	svc := NewChartExportService(nil)

	req := &ExportChartRequest{
		ViewName: "Test View",
		Header:   []string{"Name", "Value", "Count"},
		Details: [][]interface{}{
			{"Test1", int64(100)},
			{"Test2", int64(200)},
			{"Test3", int64(300)},
		},
	}

	buf, err := svc.InnerExportDetails(req)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)
}

func TestChartExportServiceIntegration_InnerExportDetails_Empty(t *testing.T) {
	svc := NewChartExportService(nil)

	req := &ExportChartRequest{
		ViewName: "Empty Test",
		Header:   []string{},
		Details:  [][]interface{}{},
	}

	buf, err := svc.InnerExportDetails(req)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestChartExportServiceIntegration_InnerExportDetails_NilValues(t *testing.T) {
	svc := NewChartExportService(nil)

	req := &ExportChartRequest{
		ViewName: "Nil Values Test",
		Header:   []string{"Col1", "Col2", "Col3"},
		Details: [][]interface{}{
			{nil, nil, nil},
			{"a", "b", "c"},
		},
	}

	buf, err := svc.InnerExportDetails(req)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestChartExportServiceIntegration_InnerExportDetails_VariousTypes(t *testing.T) {
	svc := NewChartExportService(nil)

	req := &ExportChartRequest{
		ViewName: "Various Types Test",
		Header:   []string{"String", "Int", "Float", "Bool"},
		Details: [][]interface{}{
			{"text", 123, 45.67, true},
			{"", 0, 0.0, false},
		},
	}

	buf, err := svc.InnerExportDetails(req)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestChartExportServiceIntegration_InnerExportDetailsFromChart(t *testing.T) {
	repo := &fakeChartRepo{
		byID: map[int64]*chart.CoreChartView{
			101: {ID: 101},
		},
		data: map[int64]chartRegressionSample{
			101: {
				Name:        "export",
				ChartID:     101,
				ResultCount: 10,
				Rows: []map[string]interface{}{
					{"region": "east", "amount": 10},
				},
				Total: 1,
			},
		},
	}

	chartSvc := NewChartService(repo)
	svc := NewChartExportService(chartSvc)

	buf, err := svc.InnerExportDetailsFromChart(101, "V1")
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)
}

func TestChartExportServiceIntegration_GenerateExcelFilename(t *testing.T) {
	name := GenerateExcelFilename("Report / 2026")
	assert.Contains(t, name, "Report__2026_")
	assert.Contains(t, name, ".xlsx")

	empty := GenerateExcelFilename("")
	assert.Contains(t, empty, "export_")
	assert.Contains(t, empty, ".xlsx")

	ts := currentTimeMillis()
	assert.Greater(t, ts, int64(0), "currentTimeMillis should return a real timestamp")
}
