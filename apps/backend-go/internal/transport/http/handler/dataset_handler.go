package handler

import (
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"
	"errors"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type DatasetHandler struct {
	service            *service.DatasetService
	chartService       *service.ChartService
	chartExportService *service.ChartExportService
}

func NewDatasetHandler(dsSvc *service.DatasetService, chartSvc *service.ChartService) *DatasetHandler {
	return &DatasetHandler{service: dsSvc, chartService: chartSvc, chartExportService: service.NewChartExportService(chartSvc)}
}

func (h *DatasetHandler) Tree(c *gin.Context) {
	var req dataset.TreeRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Tree(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) Fields(c *gin.Context) {
	var req dataset.FieldsRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Fields(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) Preview(c *gin.Context) {
	var req dataset.PreviewRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Preview(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) PreviewWithPermission(c *gin.Context) {
	var req dataset.PreviewRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	userID := int64(middleware.GetUserID(c))
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	result, err := h.service.PreviewWithPermission(&req, userID)
	if err != nil {
		if errors.Is(err, service.ErrDatasetDatasourcePermissionDenied) {
			response.Forbidden(c, err.Error())
			return
		}
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) Save(c *gin.Context) {
	defer recoverServicePanic(c)
	req, ok := parseDatasetWriteRequest(c, true)
	if !ok {
		return
	}
	result, err := h.service.Save(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) Create(c *gin.Context) {
	defer recoverServicePanic(c)
	req, ok := parseDatasetWriteRequest(c, true)
	if !ok {
		return
	}
	result, err := h.service.Create(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) Rename(c *gin.Context) {
	defer recoverServicePanic(c)
	req, ok := parseDatasetWriteRequest(c, true)
	if !ok {
		return
	}
	if req.ID <= 0 {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	result, err := h.service.Rename(req.ID, req.Name)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) Move(c *gin.Context) {
	defer recoverServicePanic(c)
	req, ok := parseDatasetWriteRequest(c, false)
	if !ok {
		return
	}
	if req.ID <= 0 {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	pid := int64(0)
	if req.PID != nil {
		pid = *req.PID
	}
	result, err := h.service.Move(req.ID, pid)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) Delete(c *gin.Context) {
	defer recoverServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	if err = h.service.Delete(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DatasetHandler) PerDelete(c *gin.Context) {
	defer recoverServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	result, err := h.service.PerDelete(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) GetDetail(c *gin.Context) {
	defer recoverServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	result, err := buildDatasetDetail(h, id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) Details(c *gin.Context) {
	defer recoverServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	result, err := buildDatasetDetail(h, id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) DsDetails(c *gin.Context) {
	defer recoverServicePanic(c)
	ids, ok := parseDatasetIDs(c)
	if !ok {
		return
	}
	result := make([]gin.H, 0, len(ids))
	for _, id := range ids {
		detail, err := buildDatasetDetail(h, id)
		if err != nil {
			continue
		}
		result = append(result, detail)
	}
	response.Success(c, result)
}

func (h *DatasetHandler) GetSQLParams(c *gin.Context) {
	defer recoverServicePanic(c)
	ids, ok := parseDatasetIDs(c)
	if !ok {
		return
	}
	result, err := h.service.GetSQLParams(ids)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) BarInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	group, err := h.service.GetGroupByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed to get dataset: "+err.Error())
		return
	}
	barInfo := &dataset.BarInfo{
		ID: group.ID, Name: group.Name, CreateBy: group.CreateBy,
		CreateTime: group.CreateTime, UpdateBy: group.UpdateBy,
		LastUpdateTime: group.LastUpdateTime, IsCross: false,
	}
	if group.NodeType != nil {
		barInfo.NodeType = *group.NodeType
	}
	barInfo.Creator = h.service.ResolveUserName(group.CreateBy)
	barInfo.Updater = h.service.ResolveUserName(group.UpdateBy)
	response.Success(c, barInfo)
}

func (h *DatasetHandler) GetDatasetTotal(c *gin.Context) {
	defer recoverServicePanic(c)
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	id, ok := parseInt64Value(body["id"])
	if !ok {
		response.Success(c, int64(0))
		return
	}
	preview, err := h.service.Preview(&dataset.PreviewRequest{DatasetGroupID: id, Limit: 1})
	if err != nil {
		response.Success(c, int64(0))
		return
	}
	response.Success(c, preview.Total)
}

func (h *DatasetHandler) PreviewSQL(c *gin.Context) {
	defer recoverServicePanic(c)
	var req dataset.SQLPreviewRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.PreviewSQLWithUser(&req, int64(middleware.GetUserID(c)))
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) EnumValueObj(c *gin.Context) {
	defer recoverServicePanic(c)
	req, ok := parseEnumValueRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetFieldEnumObj(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) EnumValueDs(c *gin.Context) {
	defer recoverServicePanic(c)
	fieldID, ok := parseEnumFieldID(c)
	if !ok {
		return
	}
	result, err := h.service.GetFieldEnumDs(fieldID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) EnumValue(c *gin.Context) {
	defer recoverServicePanic(c)
	req, ok := parseMultFieldValuesRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetFieldEnum(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) ListByDatasetGroup(c *gin.Context) {
	defer recoverServicePanic(c)
	datasetID, err := strconv.ParseInt(c.Param("datasetId"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	if h == nil || h.chartService == nil {
		response.Success(c, []chart.ChartField{})
		return
	}
	userID := int64(middleware.GetUserID(c))
	var result *chart.ChartFieldListResponse
	if userID > 0 {
		result, err = h.chartService.ListByDQWithPermission(datasetID, 0, userID)
	} else {
		result, err = h.chartService.ListByDQ(datasetID, 0)
	}
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, flattenChartFieldList(result))
}

func (h *DatasetHandler) ListWithPermissions(c *gin.Context) {
	defer recoverServicePanic(c)
	datasetID, err := strconv.ParseInt(c.Param("datasetId"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	if h == nil || h.chartService == nil {
		response.Success(c, []chart.ChartField{})
		return
	}
	userID := int64(middleware.GetUserID(c))
	var result *chart.ChartFieldListResponse
	if userID > 0 {
		result, err = h.chartService.ListByDQWithPermission(datasetID, 0, userID)
	} else {
		result, err = h.chartService.ListByDQ(datasetID, 0)
	}
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, flattenChartFieldList(result))
}

func (h *DatasetHandler) SaveField(c *gin.Context) {
	defer recoverServicePanic(c)
	var field dataset.CoreDatasetTableField
	if err := c.ShouldBindBodyWith(&field, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if h == nil || h.service == nil {
		response.Error(c, "500000", "dataset service unavailable")
		return
	}
	result, err := h.service.SaveField(&field)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) GetFieldFunctions(c *gin.Context) {
	defer recoverServicePanic(c)
	if h == nil || h.service == nil {
		response.Success(c, []service.FunctionCategory{})
		return
	}
	result := h.service.GetFieldFunctions()
	response.Success(c, result)
}

func (h *DatasetHandler) MultFieldValuesForPermissions(c *gin.Context) {
	defer recoverServicePanic(c)
	req, ok := parseMultFieldValuesRequest(c)
	if !ok {
		return
	}
	if h == nil || h.service == nil {
		response.Success(c, []string{})
		return
	}
	result, err := h.service.GetFieldEnum(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) CopilotFields(c *gin.Context) {
	defer recoverServicePanic(c)
	datasetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	userID := int64(middleware.GetUserID(c))
	if userID <= 0 {
		response.Unauthorized(c, "authentication required")
		return
	}
	if h == nil || h.service == nil {
		response.Error(c, "500000", "dataset service unavailable")
		return
	}
	result, err := h.service.CopilotFields(datasetID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "dataset not found")
			return
		}
		if errors.Is(err, service.ErrDatasetViewPermissionDenied) {
			response.Forbidden(c, "insufficient permissions")
			return
		}
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) ListFieldsByDsIds(c *gin.Context) {
	defer recoverServicePanic(c)
	var req struct {
		DsIds []int64 `json:"dsIds"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if h == nil || h.service == nil {
		response.Success(c, []dataset.CoreDatasetTableField{})
		return
	}
	result, err := h.service.ListFieldsByDsIds(req.DsIds)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) DetailWithPerm(c *gin.Context) {
	defer recoverServicePanic(c)
	ids, ok := parseDatasetIDs(c)
	if !ok {
		return
	}
	userID := int64(middleware.GetUserID(c))
	result := make([]gin.H, 0, len(ids))
	for _, id := range ids {
		var (
			detail gin.H
			err    error
		)
		if userID > 0 {
			detail, err = buildDatasetDetailWithPermission(h, id, userID)
		} else {
			detail, err = buildDatasetDetail(h, id)
		}
		if err != nil {
			continue
		}
		result = append(result, detail)
	}
	response.Success(c, result)
}

func (h *DatasetHandler) ExportDataset(c *gin.Context) {
	defer recoverServicePanic(c)
	var req dataset.ExportDatasetRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if req.DataEaseBi {
		buf, err := h.chartExportService.InnerExportDetails(&service.ExportChartRequest{
			ViewName: req.ViewName,
			Header:   req.Header,
			Details:  req.Details,
		})
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
		return
	}

	result, err := h.service.ExportDataset(&req, int64(middleware.GetUserID(c)))
	if err != nil {
		response.Error(c, "500000", "Failed to export: "+err.Error())
		return
	}
	response.Success(c, result)
}

// GetFieldTree returns a tree structure of field values for the given field IDs.
func (h *DatasetHandler) GetFieldTree(c *gin.Context) {
	defer recoverServicePanic(c)
	req, ok := parseMultFieldValuesRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetFieldTree(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

// DeleteDatasetField handles POST /datasetField/delete/:id
func (h *DatasetHandler) DeleteDatasetField(c *gin.Context) {
	defer recoverServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid field ID")
		return
	}
	if h.service != nil {
		err = h.service.DeleteField(id)
	} else if h.chartService != nil {
		err = h.chartService.DeleteField(id)
	} else {
		response.Error(c, "500000", "service unavailable")
		return
	}
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteDatasetFieldByChart handles POST /datasetField/deleteByChartId/:id
func (h *DatasetHandler) DeleteDatasetFieldByChart(c *gin.Context) {
	defer recoverServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid chart ID")
		return
	}
	if h.service != nil {
		err = h.service.DeleteFieldByChart(id)
	} else if h.chartService != nil {
		err = h.chartService.DeleteFieldByChart(id)
	} else {
		response.Error(c, "500000", "service unavailable")
		return
	}
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

//nolint:dupl // route registration pattern is intentionally similar
func RegisterDatasetRoutes(r *gin.RouterGroup, h *DatasetHandler) {
	datasetGroup := r.Group("/dataset")
	{
		datasetGroup.POST("/tree", h.Tree)
		datasetGroup.POST("/fields", h.Fields)
		datasetGroup.POST("/preview", h.Preview)
		datasetGroup.POST("/previewWithPerm", h.PreviewWithPermission)
		datasetGroup.POST("/save", h.Save)
		datasetGroup.POST("/create", h.Create)
		datasetGroup.POST("/rename", h.Rename)
		datasetGroup.POST("/move", h.Move)
		datasetGroup.POST("/delete/:id", h.Delete)
		datasetGroup.POST("/perDelete/:id", h.PerDelete)
		datasetGroup.POST("/get/:id", h.GetDetail)
		datasetGroup.POST("/details/:id", h.Details)
		datasetGroup.POST("/dsDetails", h.DsDetails)
		datasetGroup.POST("/getSqlParams", h.GetSQLParams)
		datasetGroup.GET("/barInfo/:id", h.BarInfo)
		datasetGroup.POST("/getDatasetTotal", h.GetDatasetTotal)
		datasetGroup.POST("/previewSql", h.PreviewSQL)
		datasetGroup.POST("/enumValueObj", h.EnumValueObj)
		datasetGroup.POST("/enumValueDs", h.EnumValueDs)
		datasetGroup.POST("/enumValue", h.EnumValue)
		datasetGroup.POST("/exportDataset", h.ExportDataset)
		datasetGroup.POST("/detailWithPerm", h.DetailWithPerm)
		datasetGroup.POST("/fieldTree", h.GetFieldTree)
	}
}
