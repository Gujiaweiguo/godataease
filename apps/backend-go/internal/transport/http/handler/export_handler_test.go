package handler

import (
	"testing"

	"dataease/backend/internal/domain/permission"
)

func TestNormalizeExportResourceType(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		{name: "dataset", raw: permission.ResourceTypeDataset, expected: permission.ResourceTypeDataset},
		{name: "dashboard", raw: permission.ResourceTypeDashboard, expected: permission.ResourceTypeDashboard},
		{name: "screen", raw: permission.ResourceTypeScreen, expected: permission.ResourceTypeScreen},
		{name: "datasource", raw: permission.ResourceTypeDatasource, expected: permission.ResourceTypeDatasource},
		{name: "unknown", raw: "report", expected: ""},
		{name: "empty", raw: "", expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeExportResourceType(tc.raw)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
