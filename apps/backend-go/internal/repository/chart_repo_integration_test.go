//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
)

func TestChartRepository_GetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewChartRepository(testDB)
	cleanupTables("core_chart_view")

	c := &chart.CoreChartView{
		Title: strPtr("Test Chart"),
		Type:  strPtr("bar"),
	}
	if err := testDB.Create(c).Error; err != nil {
		t.Fatalf("Failed to create chart: %v", err)
	}

	found, err := repo.GetByID(c.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if *found.Title != "Test Chart" {
		t.Errorf("Expected Title 'Test Chart', got '%s'", *found.Title)
	}
}

func TestChartRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewChartRepository(testDB)
	cleanupTables("core_chart_view")

	c := &chart.CoreChartView{
		Title: strPtr("Update Chart"),
		Type:  strPtr("bar"),
	}
	if err := testDB.Create(c).Error; err != nil {
		t.Fatalf("Failed to create chart: %v", err)
	}

	c.Title = strPtr("Updated Chart Title")
	err := repo.Update(c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(c.ID)
	if *found.Title != "Updated Chart Title" {
		t.Errorf("Expected Title 'Updated Chart Title', got '%s'", *found.Title)
	}
}

func TestChartRepository_CreateAndGetDatasetField(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewChartRepository(testDB)
	cleanupTables("core_dataset_table_field")

	field := &dataset.CoreDatasetTableField{
		DatasetGroupID: 1,
		OriginName:     strPtr("field1"),
		Name:           strPtr("Field 1"),
		DataeaseName:   strPtr("field1"),
		FieldShortName: strPtr("f1"),
		GroupType:      strPtr("d"),
		Type:           strPtr("VARCHAR"),
		DeType:         intPtr(0),
		DeExtractType:  intPtr(0),
		ExtField:       intPtr(0),
		Checked:        boolPtr(true),
	}

	err := repo.CreateDatasetField(field)
	if err != nil {
		t.Fatalf("CreateDatasetField failed: %v", err)
	}

	if field.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	found, err := repo.GetDatasetFieldByID(field.ID)
	if err != nil {
		t.Fatalf("GetDatasetFieldByID failed: %v", err)
	}

	if *found.Name != "Field 1" {
		t.Errorf("Expected Name 'Field 1', got '%s'", *found.Name)
	}
}

func TestChartRepository_ListDatasetFieldsByGroup(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewChartRepository(testDB)
	cleanupTables("core_dataset_table_field")

	for i := 1; i <= 3; i++ {
		field := &dataset.CoreDatasetTableField{
			DatasetGroupID: 100,
			OriginName:     strPtr("origin"),
			Name:           strPtr("Field"),
			DataeaseName:   strPtr("field"),
			FieldShortName: strPtr("f"),
			GroupType:      strPtr("d"),
			Type:           strPtr("VARCHAR"),
			DeType:         intPtr(0),
		}
		_ = repo.CreateDatasetField(field)
	}

	fields, err := repo.ListDatasetFieldsByGroup(100)
	if err != nil {
		t.Fatalf("ListDatasetFieldsByGroup failed: %v", err)
	}

	if len(fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(fields))
	}
}

func TestChartRepository_CountDatasetFieldName(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewChartRepository(testDB)
	cleanupTables("core_dataset_table_field")

	field := &dataset.CoreDatasetTableField{
		DatasetGroupID: 200,
		Name:           strPtr("UniqueField"),
	}
	_ = repo.CreateDatasetField(field)

	count, err := repo.CountDatasetFieldName(200, "UniqueField")
	if err != nil {
		t.Fatalf("CountDatasetFieldName failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, _ = repo.CountDatasetFieldName(200, "NotExisting")
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestChartRepository_UpdateDatasetFieldNames(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewChartRepository(testDB)
	cleanupTables("core_dataset_table_field")

	field := &dataset.CoreDatasetTableField{
		DatasetGroupID: 300,
		DataeaseName:   strPtr("old_name"),
		FieldShortName: strPtr("old_short"),
	}
	_ = repo.CreateDatasetField(field)

	err := repo.UpdateDatasetFieldNames(field.ID, "new_name", "new_short")
	if err != nil {
		t.Fatalf("UpdateDatasetFieldNames failed: %v", err)
	}

	found, _ := repo.GetDatasetFieldByID(field.ID)
	if *found.DataeaseName != "new_name" {
		t.Errorf("Expected DataeaseName 'new_name', got '%s'", *found.DataeaseName)
	}
}

func TestChartRepository_DeleteDatasetField(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewChartRepository(testDB)
	cleanupTables("core_dataset_table_field")

	field := &dataset.CoreDatasetTableField{
		DatasetGroupID: 400,
	}
	_ = repo.CreateDatasetField(field)

	err := repo.DeleteDatasetField(field.ID)
	if err != nil {
		t.Fatalf("DeleteDatasetField failed: %v", err)
	}

	_, err = repo.GetDatasetFieldByID(field.ID)
	if err == nil {
		t.Error("Expected error when getting deleted field")
	}
}

func intPtr(v int) *int {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}
