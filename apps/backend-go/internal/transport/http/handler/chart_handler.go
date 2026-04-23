package handler

import (
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type ChartHandler struct {
	service        *service.ChartService
	exportService  *service.ChartExportService
	datasetService *service.DatasetService
}

func NewChartHandler(svc *service.ChartService, dsSvc *service.DatasetService) *ChartHandler {
	return &ChartHandler{service: svc, exportService: service.NewChartExportService(svc), datasetService: dsSvc}
}

func (h *ChartHandler) Query(c *gin.Context) {
	var req chart.ChartQueryRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
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
	var reqMap map[string]interface{}
	if err := c.ShouldBindBodyWith(&reqMap, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	id, ok := chartDataIDFromMap(reqMap)
	if !ok || id <= 0 {
		response.Error(c, "500000", "Invalid request: chart id is required")
		return
	}
	req := &chart.ChartDataRequest{ID: id, Payload: reqMap}
	if resultCount, ok := chartDataResultCountFromMap(reqMap); ok {
		req.ResultCount = &resultCount
	}
	if resultMode, ok := reqMap["resultMode"].(string); ok {
		req.ResultMode = resultMode
	}

	userID := int64(middleware.GetUserID(c))
	var result *chart.ChartDataResponse
	var err error
	if userID > 0 {
		result, err = h.service.QueryDataWithPermission(req, userID)
	} else {
		result, err = h.service.QueryData(req)
	}
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	viewMap := make(map[string]interface{})
	view, viewErr := h.service.Query(&chart.ChartQueryRequest{ID: id})
	if viewErr == nil && view != nil {
		mergeChartViewIntoMap(viewMap, view)
	}
	for key, value := range reqMap {
		viewMap[key] = value
	}
	// Computed chart data goes as nested "data" inside the view config
	viewMap["data"] = result
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": viewMap,
	})
}

func chartDataIDFromMap(body map[string]interface{}) (int64, bool) {
	if body == nil {
		return 0, false
	}
	switch value := body["id"].(type) {
	case float64:
		return int64(value), true
	case int64:
		return value, true
	case int:
		return int64(value), true
	case string:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func chartDataResultCountFromMap(body map[string]interface{}) (int, bool) {
	if body == nil {
		return 0, false
	}
	switch value := body["resultCount"].(type) {
	case float64:
		return int(value), true
	case int:
		return value, true
	case int64:
		return int(value), true
	case string:
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func mergeChartViewIntoMap(target map[string]interface{}, view *chart.CoreChartView) {
	if target == nil || view == nil {
		return
	}
	raw, err := json.Marshal(view)
	if err != nil {
		return
	}
	var mapped map[string]interface{}
	if err = json.Unmarshal(raw, &mapped); err != nil {
		return
	}
	for key, value := range mapped {
		target[key] = decodeChartViewJSONField(key, value)
	}
}

func decodeChartViewJSONField(key string, value interface{}) interface{} {
	raw, ok := value.(string)
	if !ok {
		return value
	}
	if !isChartViewJSONField(key) {
		return value
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return value
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return value
	}
	return decoded
}

func isChartViewJSONField(key string) bool {
	switch key {
	case "xAxis", "yAxis", "customAttr", "customStyle", "customFilter", "xAxisExt", "yAxisExt", "extStack", "extBubble", "extLabel", "extTooltip", "customAttrMobile", "customStyleMobile", "drillFields", "senior", "snapshot", "viewFields", "extColor", "sortPriority":
		return true
	default:
		return false
	}
}

// CheckSameDataSet handles GET /chart/checkSameDataSet/:viewIdSource/:viewIdTarget
func (h *ChartHandler) CheckSameDataSet(c *gin.Context) {
	defer recoverServicePanic(c)

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
	defer recoverServicePanic(c)

	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
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
	defer recoverServicePanic(c)

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
	defer recoverServicePanic(c)

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
	defer recoverServicePanic(c)

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
	defer recoverServicePanic(c)
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
	defer recoverServicePanic(c)
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
	defer recoverServicePanic(c)
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
	defer recoverServicePanic(c)

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
	defer recoverServicePanic(c)

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
	defer recoverServicePanic(c)

	var req service.ExportChartRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
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
	h.InnerExportDetails(c)
}
