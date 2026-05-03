package service

import (
	"context"
	"fmt"
)

// GetChartDataForThreshold implements ThresholdChartDataAccessor for ChartService.
// It fetches chart rows and field metadata needed for threshold preview.
func (s *ChartService) GetChartDataForThreshold(ctx context.Context, chartID int64, resourceTable string) ([]map[string]any, []FieldDTO, error) {
	_ = ctx
	_ = resourceTable

	view, err := s.repo.GetByID(chartID)
	if err != nil {
		return nil, nil, fmt.Errorf("get chart view: %w", err)
	}

	limit := 100
	if view.ResultCount != nil && *view.ResultCount > 0 {
		limit = *view.ResultCount
	}

	rows, _, err := s.repo.QueryRows(chartID, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("query chart rows: %w", err)
	}

	fields, err := s.repo.ListDatasetFieldsByChart(chartID)
	if err != nil {
		return nil, nil, fmt.Errorf("list chart fields: %w", err)
	}

	fieldDTOs := make([]FieldDTO, 0, len(fields))
	for _, f := range fields {
		if f == nil || f.ID == 0 || f.DataeaseName == nil || f.Name == nil || f.DeType == nil {
			continue
		}
		fieldDTOs = append(fieldDTOs, FieldDTO{
			ID:           f.ID,
			Name:         *f.Name,
			DataeaseName: *f.DataeaseName,
			DeType:       *f.DeType,
		})
	}

	typedRows := make([]map[string]any, len(rows))
	copy(typedRows, rows)

	return typedRows, fieldDTOs, nil
}
