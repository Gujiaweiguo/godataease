package handler

import (
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChartHandler struct {
	service        *service.ChartService
	exportService  *service.ChartExportService
	datasetService *service.DatasetService
}

func NewChartHandler(svc *service.ChartService, dsSvc *service.DatasetService) *ChartHandler {
	return &ChartHandler{service: svc, exportService: service.NewChartExportService(svc), datasetService: dsSvc}
}

func recoverChartServicePanic(c *gin.Context) {
	if r := recover(); r != nil {
		response.Error(c, "500000", "Service unavailable")
	}
}

func (h *ChartHandler) Query(c *gin.Context) {
	var req chart.ChartQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Query(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *ChartHandler) Data(c *gin.Context) {
	var req chart.ChartDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	userID := int64(middleware.GetUserID(c))
	var result *chart.ChartDataResponse
	var err error
	if userID > 0 {
		result, err = h.service.QueryDataWithPermission(&req, userID)
	} else {
		result, err = h.service.QueryData(&req)
	}
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

// CheckSameDataSet handles GET /chart/checkSameDataSet/:viewIdSource/:viewIdTarget
func (h *ChartHandler) CheckSameDataSet(c *gin.Context) {
	defer recoverChartServicePanic(c)

	sourceID, err := strconv.ParseInt(c.Param("viewIdSource"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid source chart ID")
		return
	}
	targetID, err := strconv.ParseInt(c.Param("viewIdTarget"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid target chart ID")
		return
	}

	source, err := h.service.Query(&chart.ChartQueryRequest{ID: sourceID})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	target, err := h.service.Query(&chart.ChartQueryRequest{ID: targetID})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	same := false
	if source.TableID != nil && target.TableID != nil && *source.TableID == *target.TableID {
		same = true
	}
	response.Success(c, same)
}

// SaveFromMap handles POST /chart/save
func (h *ChartHandler) SaveFromMap(c *gin.Context) {
	defer recoverChartServicePanic(c)

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.SaveFromMap(body)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

// ListByDQ handles POST /chart/listByDQ/:id/:chartId
func (h *ChartHandler) ListByDQ(c *gin.Context) {
	defer recoverChartServicePanic(c)

	datasetGroupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	chartID, err := strconv.ParseInt(c.Param("chartId"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid chart ID")
		return
	}

	userID := int64(middleware.GetUserID(c))
	var result *chart.ChartFieldListResponse
	if userID > 0 {
		result, err = h.service.ListByDQWithPermission(datasetGroupID, chartID, userID)
	} else {
		result, err = h.service.ListByDQ(datasetGroupID, chartID)
	}
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

// CopyField handles POST /chart/copyField/:id/:chartId
func (h *ChartHandler) CopyField(c *gin.Context) {
	defer recoverChartServicePanic(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid field ID")
		return
	}
	chartID, err := strconv.ParseInt(c.Param("chartId"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid chart ID")
		return
	}
	if err = h.service.CopyField(id, chartID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteField handles POST /chart/deleteField/:id
func (h *ChartHandler) DeleteField(c *gin.Context) {
	defer recoverChartServicePanic(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid field ID")
		return
	}
	if err = h.service.DeleteField(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteFieldByChart handles POST /chart/deleteFieldByChart/:chartId
func (h *ChartHandler) DeleteFieldByChart(c *gin.Context) {
	defer recoverChartServicePanic(c)
	chartID, err := strconv.ParseInt(c.Param("chartId"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid chart ID")
		return
	}
	if err := h.service.DeleteFieldByChart(chartID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *ChartHandler) GetChart(c *gin.Context) {
	defer recoverChartServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid chart ID")
		return
	}
	result, err := h.service.Query(&chart.ChartQueryRequest{ID: id})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *ChartHandler) GetDetail(c *gin.Context) {
	defer recoverChartServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid chart ID")
		return
	}
	result, err := h.service.Query(&chart.ChartQueryRequest{ID: id})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

// GetFieldData handles POST /chartData/getFieldData/:fieldId/:fieldType
func (h *ChartHandler) GetFieldData(c *gin.Context) {
	defer recoverDatasetServicePanic(c)

	fieldID, err := strconv.ParseInt(c.Param("fieldId"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid field ID")
		return
	}
	result, err := h.datasetService.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{fieldID}, ResultMode: 1})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

// GetDrillFieldData handles POST /chartData/getDrillFieldData/:fieldId
func (h *ChartHandler) GetDrillFieldData(c *gin.Context) {
	defer recoverDatasetServicePanic(c)

	fieldID, err := strconv.ParseInt(c.Param("fieldId"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid field ID")
		return
	}
	result, err := h.datasetService.GetFieldEnumDs(fieldID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

// InnerExportDetails handles POST /chartData/innerExportDetails
func (h *ChartHandler) InnerExportDetails(c *gin.Context) {
	defer recoverChartServicePanic(c)

	var req service.ExportChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	buf, err := h.exportService.InnerExportDetails(&req)
	if err != nil {
		response.Error(c, "500000", "Failed to export: "+err.Error())
		return
	}

	filename := service.GenerateExcelFilename(req.ViewName)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// InnerExportDataSetDetails handles POST /chartData/innerExportDataSetDetails
func (h *ChartHandler) InnerExportDataSetDetails(c *gin.Context) {
	defer recoverChartServicePanic(c)

	var req service.ExportChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	buf, err := h.exportService.InnerExportDetails(&req)
	if err != nil {
		response.Error(c, "500000", "Failed to export: "+err.Error())
		return
	}

	filename := service.GenerateExcelFilename(req.ViewName)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// RegisterChartRoutes is deprecated — registration is done via router.go registerChartRoutes().
// Kept for backward compatibility.
func RegisterChartRoutes(r *gin.RouterGroup, h *ChartHandler) {
	chartGroup := r.Group("/chart")
	{
		chartGroup.POST("/query", h.Query)
		chartGroup.POST("/data", h.Data)
	}
}
