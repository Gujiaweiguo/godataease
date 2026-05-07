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
	defer recoverServicePanic(c)

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

func (h *ChartHandler) ViewOption(c *gin.Context) {
	defer recoverServicePanic(c)
	resourceId, err := strconv.ParseInt(c.Param("resourceId"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid resource ID")
		return
	}
	result, err := h.service.ViewOption(resourceId)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *ChartHandler) ChartBaseInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid chart ID")
		return
	}
	resourceTable := c.Param("resourceTable")
	result, err := h.service.ChartBaseInfo(id, resourceTable)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *ChartHandler) Data(c *gin.Context) {
	defer recoverServicePanic(c)

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

	sourceID, ok := parseIDParamMsg(c, "viewIdSource", "Invalid source chart ID")
	if !ok {
		return
	}
	targetID, ok := parseIDParamMsg(c, "viewIdTarget", "Invalid target chart ID")
	if !ok {
		return
	}
	var err error

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

	same := source.TableID != nil && target.TableID != nil && *source.TableID == *target.TableID
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

	datasetGroupID, ok := parseIDParamMsg(c, "id", "Invalid dataset ID")
	if !ok {
		return
	}
	chartID, ok := parseIDParamMsg(c, "chartId", "Invalid chart ID")
	if !ok {
		return
	}
	var err error

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

	id, ok := parseIDParamMsg(c, "id", "Invalid field ID")
	if !ok {
		return
	}
	chartID, ok := parseIDParamMsg(c, "chartId", "Invalid chart ID")
	if !ok {
		return
	}
	var err error
	if err = h.service.CopyField(id, chartID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteField handles POST /chart/deleteField/:id
func (h *ChartHandler) DeleteField(c *gin.Context) {
	defer recoverServicePanic(c)

	id, ok := parseIDParamMsg(c, "id", "Invalid field ID")
	if !ok {
		return
	}
	var err error
	if err = h.service.DeleteField(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteFieldByChart handles POST /chart/deleteFieldByChart/:chartId
func (h *ChartHandler) DeleteFieldByChart(c *gin.Context) {
	defer recoverServicePanic(c)
	chartID, ok := parseIDParamMsg(c, "chartId", "Invalid chart ID")
	if !ok {
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
	id, ok := parseIDParamMsg(c, "id", "Invalid chart ID")
	if !ok {
		return
	}
	var err error
	result, err := h.service.Query(&chart.ChartQueryRequest{ID: id})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *ChartHandler) GetDetail(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid chart ID")
	if !ok {
		return
	}
	var err error
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

	fieldID, ok := parseIDParamMsg(c, "fieldId", "Invalid field ID")
	if !ok {
		return
	}
	var err error
	if h.datasetService == nil {
		response.Success(c, []string{})
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

	fieldID, ok := parseIDParamMsg(c, "fieldId", "Invalid field ID")
	if !ok {
		return
	}
	var err error
	if h.datasetService == nil {
		response.Success(c, []string{})
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
	defer recoverServicePanic(c)

	h.InnerExportDetails(c)
}

// RegisterChartDataRoutes registers canonical chartData routes.
func RegisterChartDataRoutes(chartDataGroup *gin.RouterGroup, h *ChartHandler, permMiddleware *middleware.PermissionMiddleware) {
	if permMiddleware != nil {
		chartDataGroup.POST("/getData", permMiddleware.CheckChartDataView(), middleware.RowPermissionMiddleware(), h.Data)
	} else {
		chartDataGroup.POST("/getData", h.Data)
	}
	chartDataGroup.POST("/getFieldData/:fieldId/:fieldType", h.GetFieldData)
	chartDataGroup.POST("/getDrillFieldData/:fieldId", h.GetDrillFieldData)
	chartDataGroup.POST("/innerExportDetails", h.InnerExportDetails)
	chartDataGroup.POST("/innerExportDataSetDetails", h.InnerExportDataSetDetails)
}
