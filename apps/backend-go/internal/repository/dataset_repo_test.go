package repository

import "testing"

func TestPreviewRows_InvalidTableName(t *testing.T) {
	repo := &DatasetRepository{}
	_, err := repo.PreviewRows("core_dataset_table;drop table x", 10)
	if err == nil {
		t.Fatal("expected invalid table name error")
	}
}

func TestCountRows_InvalidTableName(t *testing.T) {
	repo := &DatasetRepository{}
	_, err := repo.CountRows("x` or 1=1 --")
	if err == nil {
		t.Fatal("expected invalid table name error")
	}
}
