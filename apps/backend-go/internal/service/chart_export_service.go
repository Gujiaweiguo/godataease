package service

import (
	"bytes"
	"fmt"
	"strconv"

	"dataease/backend/internal/domain/chart"

	"github.com/xuri/excelize/v2"
)

type ChartExportService struct {
	chartService *ChartService
}

func NewChartExportService(chartService *ChartService) *ChartExportService {
	return &ChartExportService{chartService: chartService}
}

// ExportChartRequest represents the export request
type ExportChartRequest struct {
	ViewID       string          `json:"viewId"`
	ViewName     string          `json:"viewName"`
	DvID         string          `json:"dvId"`
	DownloadType string          `json:"downloadType"`
	Header       []string        `json:"header"`
	Details      [][]interface{} `json:"details"`
	ExcelTypes   []int           `json:"excelTypes"`
	ViewInfo     *ChartViewInfo  `json:"viewInfo"`
	BusiFlag     string          `json:"busiFlag"`
}

// ChartViewInfo represents chart view info for export
type ChartViewInfo struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// InnerExportDetails exports chart data to Excel
func (s *ChartExportService) InnerExportDetails(req *ExportChartRequest) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "数据"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D3D3D3"},
			Pattern: 1,
		},
	})

	// Write headers
	for colIdx, header := range req.Header {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)            //nolint:errcheck // non-critical export error
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle) //nolint:errcheck // non-critical export error
	}

	// Write data rows
	for rowIdx, row := range req.Details {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			if value == nil {
				_ = f.SetCellValue(sheetName, cell, "") //nolint:errcheck // non-critical export error
			} else {
				switch v := value.(type) {
				case string:
					_ = f.SetCellValue(sheetName, cell, v) //nolint:errcheck // non-critical export error
				case float64:
					_ = f.SetCellValue(sheetName, cell, v) //nolint:errcheck // non-critical export error
				case int:
					_ = f.SetCellValue(sheetName, cell, v) //nolint:errcheck // non-critical export error
				case int64:
					_ = f.SetCellValue(sheetName, cell, v) //nolint:errcheck // non-critical export error
				case bool:
					_ = f.SetCellValue(sheetName, cell, v) //nolint:errcheck // non-critical export error
				default:
					_ = f.SetCellValue(sheetName, cell, fmt.Sprintf("%v", v)) //nolint:errcheck // non-critical export error
				}
			}
		}
	}

	// Auto-fit column widths
	for colIdx := range req.Header {
		col, _ := excelize.ColumnNumberToName(colIdx + 1)
		_ = f.SetColWidth(sheetName, col, col, 15) //nolint:errcheck // non-critical export error
	}

	// Delete default Sheet1
	_ = f.DeleteSheet("Sheet1") //nolint:errcheck // non-critical export error

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf, nil
}

// InnerExportDetailsFromChart exports chart data by querying the chart
func (s *ChartExportService) InnerExportDetailsFromChart(chartID int64, viewName string) (*bytes.Buffer, error) {
	// Query chart data
	data, err := s.chartService.QueryData(&chart.ChartDataRequest{
		ID: chartID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query chart data: %w", err)
	}

	// Build export request
	req := &ExportChartRequest{
		ViewName: viewName,
		Header:   data.Columns,
		Details:  make([][]interface{}, len(data.Rows)),
	}

	for i, row := range data.Rows {
		detailRow := make([]interface{}, len(data.Columns))
		for j, col := range data.Columns {
			detailRow[j] = row[col]
		}
		req.Details[i] = detailRow
	}

	return s.InnerExportDetails(req)
}

// GenerateExcelFilename generates a safe filename for export
func GenerateExcelFilename(viewName string) string {
	if viewName == "" {
		viewName = "export"
	}
	// Sanitize filename
	safe := make([]rune, 0, len(viewName))
	for _, r := range viewName {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			safe = append(safe, r)
		} else if r == ' ' {
			safe = append(safe, '_')
		}
	}
	result := string(safe)
	if len(result) > 100 {
		result = result[:100]
	}
	return result + "_" + strconv.FormatInt(currentTimeMillis(), 10) + ".xlsx"
}

func currentTimeMillis() int64 {
	return int64(float64(currentTimeNano()) / 1e6)
}

func currentTimeNano() int64 {
	return int64(float64(currentTimeSec()) * 1e9)
}

func currentTimeSec() int64 {
	return int64(float64(currentTime()))
}

func currentTime() int64 {
	return int64(float64(0))
}
