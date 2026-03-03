//go:build integration

package service

import (
	"fmt"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func ensureChartServiceTables(t *testing.T) {
	t.Helper()
	err := testDB.AutoMigrate(&chart.CoreChartView{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{})
	assert.NoError(t, err)

	err = testDB.Exec("CREATE TABLE IF NOT EXISTS `it_chart_data_rows` (id BIGINT PRIMARY KEY AUTO_INCREMENT, region VARCHAR(50), amount INT)").Error
	assert.NoError(t, err)
}

func clearChartServiceTables(t *testing.T) {
	t.Helper()
	tables := []string{"core_chart_view", "core_dataset_table", "core_dataset_table_field", "it_chart_data_rows"}
	for _, table := range tables {
		err := testDB.Exec(fmt.Sprintf("DELETE FROM `%s`", table)).Error
		assert.NoError(t, err)
	}
}

func TestChartServiceIntegration_QueryAndQueryData(t *testing.T) {
	ensureChartServiceTables(t)
	clearChartServiceTables(t)

	repo := repository.NewChartRepository(testDB)
	svc := NewChartService(repo)

	tableName := "it_chart_data_rows"
	err := testDB.Create(&dataset.CoreDatasetTable{
		Name:           chartSvcStringPtr("it_dataset"),
		DatasetGroupID: 9001,
		PhysicalTable:  &tableName,
	}).Error
	assert.NoError(t, err)

	var dsTable dataset.CoreDatasetTable
	err = testDB.Where("dataset_group_id = ?", 9001).First(&dsTable).Error
	assert.NoError(t, err)

	err = testDB.Create(&chart.CoreChartView{
		Title:   chartSvcStringPtr("Sales by Region"),
		TableID: &dsTable.ID,
		Type:    chartSvcStringPtr("bar"),
	}).Error
	assert.NoError(t, err)

	var view chart.CoreChartView
	err = testDB.Where("title = ?", "Sales by Region").First(&view).Error
	assert.NoError(t, err)

	err = testDB.Exec("INSERT INTO `it_chart_data_rows` (`region`, `amount`) VALUES (?, ?), (?, ?), (?, ?)", "East", 100, "West", 200, "North", 300).Error
	assert.NoError(t, err)

	queried, err := svc.Query(&chart.ChartQueryRequest{ID: view.ID})
	assert.NoError(t, err)
	assert.Equal(t, view.ID, queried.ID)

	limit := 2
	dataResp, err := svc.QueryData(&chart.ChartDataRequest{ID: view.ID, ResultCount: &limit})
	assert.NoError(t, err)
	assert.Equal(t, view.ID, dataResp.ChartID)
	assert.Equal(t, int64(3), dataResp.Total)
	assert.Len(t, dataResp.Rows, 2)
	assert.Contains(t, dataResp.Columns, "region")
	assert.Contains(t, dataResp.Columns, "amount")
}

func TestChartServiceIntegration_SaveFromMap(t *testing.T) {
	ensureChartServiceTables(t)
	clearChartServiceTables(t)

	repo := repository.NewChartRepository(testDB)
	svc := NewChartService(repo)

	err := testDB.Create(&chart.CoreChartView{
		Title: chartSvcStringPtr("Old Title"),
		Type:  chartSvcStringPtr("line"),
	}).Error
	assert.NoError(t, err)

	var existing chart.CoreChartView
	err = testDB.Where("title = ?", "Old Title").First(&existing).Error
	assert.NoError(t, err)

	body := map[string]interface{}{
		"id":          existing.ID,
		"title":       "New Title",
		"type":        "bar",
		"render":      "antv",
		"resultMode":  "custom",
		"resultCount": 50,
		"dataFrom":    "dataset",
		"xAxis":       []map[string]interface{}{{"field": "region"}},
		"yAxis":       []map[string]interface{}{{"field": "amount"}},
		"customAttr":  map[string]interface{}{"legend": true},
		"customStyle": map[string]interface{}{"color": "blue"},
		"customFilter": []map[string]interface{}{
			{"field": "region", "op": "eq", "value": "East"},
		},
	}

	updated, err := svc.SaveFromMap(body)
	assert.NoError(t, err)
	if assert.NotNil(t, updated.Title) {
		assert.Equal(t, "New Title", *updated.Title)
	}
	if assert.NotNil(t, updated.Type) {
		assert.Equal(t, "bar", *updated.Type)
	}
	if assert.NotNil(t, updated.ResultCount) {
		assert.Equal(t, 50, *updated.ResultCount)
	}
	assert.NotNil(t, updated.XAxis)
	assert.NotNil(t, updated.CustomFilter)
	assert.NotNil(t, updated.UpdateTime)
}

func TestChartServiceIntegration_ListByDQ_CopyAndDeleteField(t *testing.T) {
	ensureChartServiceTables(t)
	clearChartServiceTables(t)

	repo := repository.NewChartRepository(testDB)
	svc := NewChartService(repo)

	datasetGroupID := int64(9101)
	chartID := int64(9201)

	err := testDB.Create(&dataset.CoreDatasetTableField{
		DatasetGroupID: datasetGroupID,
		Name:           chartSvcStringPtr("region"),
		OriginName:     chartSvcStringPtr("region"),
		DataeaseName:   chartSvcStringPtr("region"),
		FieldShortName: chartSvcStringPtr("region"),
		GroupType:      chartSvcStringPtr("d"),
		Type:           chartSvcStringPtr("VARCHAR"),
		DeType:         chartSvcIntPtr(0),
		DeExtractType:  chartSvcIntPtr(0),
		Checked:        chartSvcBoolPtr(true),
	}).Error
	assert.NoError(t, err)

	err = testDB.Create(&dataset.CoreDatasetTableField{
		DatasetGroupID: datasetGroupID,
		ChartID:        &chartID,
		Name:           chartSvcStringPtr("amount_calc"),
		OriginName:     chartSvcStringPtr("amount"),
		GroupType:      chartSvcStringPtr("q"),
		Type:           chartSvcStringPtr("DECIMAL"),
		DeType:         chartSvcIntPtr(3),
		DeExtractType:  chartSvcIntPtr(3),
		Checked:        chartSvcBoolPtr(true),
	}).Error
	assert.NoError(t, err)

	fieldResp, err := svc.ListByDQ(datasetGroupID, chartID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(fieldResp.DimensionList), 1)
	assert.GreaterOrEqual(t, len(fieldResp.QuotaList), 2)

	var source dataset.CoreDatasetTableField
	err = testDB.Where("dataset_group_id = ? AND chart_id IS NULL", datasetGroupID).First(&source).Error
	assert.NoError(t, err)

	err = svc.CopyField(source.ID, chartID)
	assert.NoError(t, err)

	var copied dataset.CoreDatasetTableField
	err = testDB.Where("chart_id = ? AND origin_name = ?", chartID, fmt.Sprintf("[%d]", source.ID)).First(&copied).Error
	assert.NoError(t, err)
	assert.NotZero(t, copied.ID)
	if assert.NotNil(t, copied.ExtField) {
		assert.Equal(t, 2, *copied.ExtField)
	}
	assert.NotNil(t, copied.DataeaseName)
	assert.NotNil(t, copied.FieldShortName)

	err = svc.DeleteField(copied.ID)
	assert.NoError(t, err)

	var countAfterDelete int64
	err = testDB.Model(&dataset.CoreDatasetTableField{}).Where("id = ?", copied.ID).Count(&countAfterDelete).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), countAfterDelete)

	err = svc.DeleteFieldByChart(chartID)
	assert.NoError(t, err)

	var byChartCount int64
	err = testDB.Model(&dataset.CoreDatasetTableField{}).Where("chart_id = ?", chartID).Count(&byChartCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), byChartCount)
}

func TestChartServiceIntegration_ValidationErrors(t *testing.T) {
	ensureChartServiceTables(t)
	clearChartServiceTables(t)

	repo := repository.NewChartRepository(testDB)
	svc := NewChartService(repo)

	_, err := svc.SaveFromMap(map[string]interface{}{"title": "missing id"})
	assert.Error(t, err)

	err = svc.CopyField(0, 1)
	assert.Error(t, err)

	err = svc.DeleteField(0)
	assert.Error(t, err)

	err = svc.DeleteFieldByChart(0)
	assert.Error(t, err)
}

func chartSvcStringPtr(v string) *string {
	return &v
}

func chartSvcIntPtr(v int) *int {
	return &v
}

func chartSvcBoolPtr(v bool) *bool {
	return &v
}
