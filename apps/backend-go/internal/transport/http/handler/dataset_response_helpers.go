package handler

import (
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"

	"github.com/gin-gonic/gin"
)

func flattenChartFieldList(result *chart.ChartFieldListResponse) []chart.ChartField {
	if result == nil {
		return []chart.ChartField{}
	}
	fields := make([]chart.ChartField, 0, len(result.DimensionList)+len(result.QuotaList))
	fields = append(fields, result.DimensionList...)
	fields = append(fields, result.QuotaList...)
	return fields
}

func buildDatasetDetail(h *DatasetHandler, datasetGroupID int64) (gin.H, error) {
	fields, err := h.service.Fields(&dataset.FieldsRequest{DatasetGroupID: datasetGroupID})
	if err != nil {
		return nil, err
	}

	previewData := make([]map[string]interface{}, 0)
	total := int64(0)
	preview, err := h.service.Preview(&dataset.PreviewRequest{DatasetGroupID: datasetGroupID, Limit: 100})
	if err == nil {
		previewData = preview.Rows
		total = preview.Total
	}

	return gin.H{
		"id":        datasetGroupID,
		"allFields": fields,
		"data": gin.H{
			"fields": fields,
			"data":   previewData,
		},
		"total":   total,
		"union":   []gin.H{},
		"isCross": false,
	}, nil
}

func buildDatasetDetailWithPermission(h *DatasetHandler, datasetGroupID int64, userID int64) (gin.H, error) {
	fields, err := h.service.FieldsWithPermission(datasetGroupID, userID)
	if err != nil {
		return nil, err
	}

	previewData := make([]map[string]interface{}, 0)
	total := int64(0)
	preview, err := h.service.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: datasetGroupID, Limit: 100}, userID)
	if err == nil {
		previewData = preview.Rows
		total = preview.Total
	}

	allFields := flattenChartFieldList(fields)

	return gin.H{
		"id":            datasetGroupID,
		"allFields":     allFields,
		"dimensionList": fields.DimensionList,
		"quotaList":     fields.QuotaList,
		"fields": gin.H{
			"dimensionList": fields.DimensionList,
			"quotaList":     fields.QuotaList,
		},
		"data": gin.H{
			"fields": allFields,
			"data":   previewData,
		},
		"total":   total,
		"union":   []gin.H{},
		"isCross": false,
	}, nil
}
