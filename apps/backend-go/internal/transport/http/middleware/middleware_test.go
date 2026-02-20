package middleware

import (
	"testing"
)

func TestConstants(t *testing.T) {
	if RowPermissionDatasetIDKey != "row_permission_dataset_id" {
		t.Errorf("Expected RowPermissionDatasetIDKey 'row_permission_dataset_id', got '%s'", RowPermissionDatasetIDKey)
	}
	if RowPermissionFilterKey != "row_permission_filter" {
		t.Errorf("Expected RowPermissionFilterKey 'row_permission_filter', got '%s'", RowPermissionFilterKey)
	}
}
