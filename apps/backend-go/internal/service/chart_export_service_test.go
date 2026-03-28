package service

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestGenerateExcelFilename(t *testing.T) {
	t.Run("empty name uses export default", func(t *testing.T) {
		name := GenerateExcelFilename("")
		assert.Equal(t, "export_0.xlsx", name)
	})

	t.Run("replaces spaces and strips unsafe chars", func(t *testing.T) {
		name := GenerateExcelFilename("Report / 2026")
		assert.Equal(t, "Report__2026_0.xlsx", name)
	})

	t.Run("truncates long name", func(t *testing.T) {
		input := strings.Repeat("a", 120)
		name := GenerateExcelFilename(input)
		assert.True(t, strings.HasSuffix(name, "_0.xlsx"))
		assert.Len(t, name, len("_0.xlsx")+100)
	})

	t.Run("only unsafe chars falls back to timestamp suffix", func(t *testing.T) {
		name := GenerateExcelFilename("/?:*\\")
		assert.Equal(t, "_0.xlsx", name)
	})

	t.Run("preserves dash and underscore", func(t *testing.T) {
		name := GenerateExcelFilename("Report-Name_2026")
		assert.Equal(t, "Report-Name_2026_0.xlsx", name)
	})

	t.Run("exactly one hundred chars not over truncated", func(t *testing.T) {
		input := strings.Repeat("b", 100)
		name := GenerateExcelFilename(input)
		assert.True(t, strings.HasPrefix(name, input))
		assert.Len(t, name, 107)
	})
}

func TestCurrentTimeHelpers(t *testing.T) {
	assert.Equal(t, int64(0), currentTime())
	assert.Equal(t, int64(0), currentTimeSec())
	assert.Equal(t, int64(0), currentTimeNano())
	assert.Equal(t, int64(0), currentTimeMillis())
}

func openExportWorkbook(t *testing.T, buf *bytes.Buffer) *excelize.File {
	t.Helper()

	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	return f
}

func TestChartExportService_InnerExportDetails(t *testing.T) {
	t.Run("writes header and rows", func(t *testing.T) {
		svc := NewChartExportService(nil)

		buf, err := svc.InnerExportDetails(&ExportChartRequest{
			ViewName: "Test View",
			Header:   []string{"Name", "Value", "Count"},
			Details: [][]interface{}{
				{"Test1", int64(100), 1},
				{"Test2", int64(200), 2},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, buf)
		assert.Greater(t, buf.Len(), 0)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Equal(t, "Name", mustGetCellValue(t, f, "数据", "A1"))
		assert.Equal(t, "Value", mustGetCellValue(t, f, "数据", "B1"))
		assert.Equal(t, "Test1", mustGetCellValue(t, f, "数据", "A2"))
		assert.Equal(t, "100", mustGetCellValue(t, f, "数据", "B2"))
		assert.Equal(t, "2", mustGetCellValue(t, f, "数据", "C3"))
	})

	t.Run("handles nil and mixed value types", func(t *testing.T) {
		svc := NewChartExportService(nil)

		buf, err := svc.InnerExportDetails(&ExportChartRequest{
			ViewName: "Mixed Values Test",
			Header:   []string{"String", "Float", "Int", "Bool", "Nil", "Object"},
			Details: [][]interface{}{
				{"text", 45.67, 123, true, nil, map[string]string{"k": "v"}},
			},
		})
		require.NoError(t, err)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Equal(t, "text", mustGetCellValue(t, f, "数据", "A2"))
		assert.Equal(t, "45.67", mustGetCellValue(t, f, "数据", "B2"))
		assert.Equal(t, "123", mustGetCellValue(t, f, "数据", "C2"))
		assert.Equal(t, "TRUE", mustGetCellValue(t, f, "数据", "D2"))
		assert.Equal(t, "", mustGetCellValue(t, f, "数据", "E2"))
		assert.Contains(t, mustGetCellValue(t, f, "数据", "F2"), "map[k:v]")
	})

	t.Run("writes bool false and zero values", func(t *testing.T) {
		svc := NewChartExportService(nil)

		buf, err := svc.InnerExportDetails(&ExportChartRequest{
			ViewName: "Zero Values",
			Header:   []string{"Bool", "Int", "Float"},
			Details:  [][]interface{}{{false, 0, 0.0}},
		})
		require.NoError(t, err)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Equal(t, "FALSE", mustGetCellValue(t, f, "数据", "A2"))
		assert.Equal(t, "0", mustGetCellValue(t, f, "数据", "B2"))
		assert.Equal(t, "0", mustGetCellValue(t, f, "数据", "C2"))
	})

	t.Run("empty header and details still writes workbook", func(t *testing.T) {
		svc := NewChartExportService(nil)

		buf, err := svc.InnerExportDetails(&ExportChartRequest{ViewName: "Empty Test", Header: []string{}, Details: [][]interface{}{}})
		require.NoError(t, err)
		require.NotNil(t, buf)
		assert.Greater(t, buf.Len(), 0)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Contains(t, f.GetSheetList(), "数据")
	})

	t.Run("empty view name still builds workbook", func(t *testing.T) {
		svc := NewChartExportService(nil)

		buf, err := svc.InnerExportDetails(&ExportChartRequest{ViewName: "", Header: []string{"OnlyHeader"}, Details: [][]interface{}{}})
		require.NoError(t, err)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Equal(t, "OnlyHeader", mustGetCellValue(t, f, "数据", "A1"))
		assert.Equal(t, "", mustGetCellValue(t, f, "数据", "A2"))
	})

	t.Run("more detail columns than headers still writes cells", func(t *testing.T) {
		svc := NewChartExportService(nil)

		buf, err := svc.InnerExportDetails(&ExportChartRequest{ViewName: "Overflow", Header: []string{"A"}, Details: [][]interface{}{{"v1", "v2", "v3"}}})
		require.NoError(t, err)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Equal(t, "A", mustGetCellValue(t, f, "数据", "A1"))
		assert.Equal(t, "v1", mustGetCellValue(t, f, "数据", "A2"))
		assert.Equal(t, "v2", mustGetCellValue(t, f, "数据", "B2"))
		assert.Equal(t, "v3", mustGetCellValue(t, f, "数据", "C2"))
	})

	t.Run("empty header with data still writes cells", func(t *testing.T) {
		svc := NewChartExportService(nil)

		buf, err := svc.InnerExportDetails(&ExportChartRequest{ViewName: "No Header", Header: []string{}, Details: [][]interface{}{{"v1", "v2"}}})
		require.NoError(t, err)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Equal(t, "v1", mustGetCellValue(t, f, "数据", "A2"))
		assert.Equal(t, "v2", mustGetCellValue(t, f, "数据", "B2"))
	})
}

func TestChartExportService_InnerExportDetailsFromChart(t *testing.T) {
	t.Run("transforms columns and rows", func(t *testing.T) {
		repo := &fakeChartRepo{
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

		svc := NewChartExportService(NewChartService(repo))
		buf, err := svc.InnerExportDetailsFromChart(101, "V1")
		require.NoError(t, err)
		require.NotNil(t, buf)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Equal(t, "amount", mustGetCellValue(t, f, "数据", "A1"))
		assert.Equal(t, "region", mustGetCellValue(t, f, "数据", "B1"))
		assert.Equal(t, "10", mustGetCellValue(t, f, "数据", "A2"))
		assert.Equal(t, "east", mustGetCellValue(t, f, "数据", "B2"))
	})

	t.Run("multiple rows preserve column order", func(t *testing.T) {
		repo := &fakeChartRepo{
			data: map[int64]chartRegressionSample{
				104: {
					Name:    "ordered",
					ChartID: 104,
					Rows: []map[string]interface{}{
						{"region": "east", "amount": 10},
						{"region": "west", "amount": 20},
					},
					Total: 2,
				},
			},
		}

		buf, err := NewChartExportService(NewChartService(repo)).InnerExportDetailsFromChart(104, "Ordered")
		require.NoError(t, err)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Equal(t, "amount", mustGetCellValue(t, f, "数据", "A1"))
		assert.Equal(t, "region", mustGetCellValue(t, f, "数据", "B1"))
		assert.Equal(t, "10", mustGetCellValue(t, f, "数据", "A2"))
		assert.Equal(t, "east", mustGetCellValue(t, f, "数据", "B2"))
		assert.Equal(t, "20", mustGetCellValue(t, f, "数据", "A3"))
		assert.Equal(t, "west", mustGetCellValue(t, f, "数据", "B3"))
	})

	t.Run("query error", func(t *testing.T) {
		repo := &fakeChartRepo{data: map[int64]chartRegressionSample{}}
		svc := NewChartExportService(NewChartService(repo))

		buf, err := svc.InnerExportDetailsFromChart(999, "Missing")
		require.Error(t, err)
		assert.Nil(t, buf)
		assert.Contains(t, err.Error(), "failed to query chart data")
	})

	t.Run("empty rows still writes workbook", func(t *testing.T) {
		repo := &fakeChartRepo{data: map[int64]chartRegressionSample{102: {Name: "empty", ChartID: 102, Rows: []map[string]interface{}{}, Total: 0}}}
		svc := NewChartExportService(NewChartService(repo))

		buf, err := svc.InnerExportDetailsFromChart(102, "Empty")
		require.NoError(t, err)
		require.NotNil(t, buf)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Contains(t, f.GetSheetList(), "数据")
		assert.Equal(t, "", mustGetCellValue(t, f, "数据", "A1"))
	})

	t.Run("missing column value writes blank", func(t *testing.T) {
		repo := &fakeChartRepo{
			data: map[int64]chartRegressionSample{
				103: {
					Name:    "sparse",
					ChartID: 103,
					Rows: []map[string]interface{}{
						{"region": "east", "amount": 10},
						{"amount": 20},
					},
					Total: 2,
				},
			},
		}
		svc := NewChartExportService(NewChartService(repo))

		buf, err := svc.InnerExportDetailsFromChart(103, "Sparse")
		require.NoError(t, err)

		f := openExportWorkbook(t, buf)
		defer func() { _ = f.Close() }()
		assert.Equal(t, "amount", mustGetCellValue(t, f, "数据", "A1"))
		assert.Equal(t, "region", mustGetCellValue(t, f, "数据", "B1"))
		assert.Equal(t, "20", mustGetCellValue(t, f, "数据", "A3"))
		assert.Equal(t, "", mustGetCellValue(t, f, "数据", "B3"))
	})
}

func mustGetCellValue(t *testing.T, f *excelize.File, sheet, cell string) string {
	t.Helper()

	value, err := f.GetCellValue(sheet, cell)
	require.NoError(t, err)
	return value
}

func TestGenerateExcelFilename_MixedUnicodeAndUnsafeCharsStillReturnsSafeSuffix(t *testing.T) {
	name := GenerateExcelFilename("报告 Report / 2026")
	assert.Equal(t, "_Report__2026_0.xlsx", name)
}
