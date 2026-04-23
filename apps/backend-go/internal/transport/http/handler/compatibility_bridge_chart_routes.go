package handler

import (
	"errors"
	"net/url"
	"strconv"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

func registerChartCompatRoutes(r gin.IRouter, chartHandler *ChartHandler, datasetHandler *DatasetHandler, permMiddleware *middleware.PermissionMiddleware) {
	RegisterChartDataCompatRoutes(r.Group("/chartData"), chartHandler, datasetHandler, permMiddleware)
	registerChartGroupCompatRoutes(r.Group("/chart"), chartHandler, permMiddleware)
	registerDatasetFieldCompatRoutes(r.Group("/datasetField"), chartHandler, datasetHandler)
}

func RegisterChartDataCompatRoutes(chartDataGroup *gin.RouterGroup, chartHandler *ChartHandler, datasetHandler *DatasetHandler, permMiddleware *middleware.PermissionMiddleware) {
	if permMiddleware != nil {
		chartDataGroup.POST("/getData", permMiddleware.CheckChartDataView(), middleware.RowPermissionMiddleware(), chartHandler.Data)
	} else {
		chartDataGroup.POST("/getData", chartHandler.Data)
	}
	chartDataGroup.POST("/getFieldData/:fieldId/:fieldType", func(c *gin.Context) {
		fieldID, err := strconv.ParseInt(c.Param("fieldId"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid field ID")
			return
		}
		if datasetHandler == nil {
			response.Success(c, []string{})
			return
		}
		result, err := datasetHandler.service.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{fieldID}, ResultMode: 1})
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, result)
	})
	chartDataGroup.POST("/getDrillFieldData/:fieldId", func(c *gin.Context) {
		fieldID, err := strconv.ParseInt(c.Param("fieldId"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid field ID")
			return
		}
		if datasetHandler == nil {
			response.Success(c, []string{})
			return
		}
		result, err := datasetHandler.service.GetFieldEnumDs(fieldID)
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, result)
	})
	chartDataGroup.POST("/innerExportDetails", func(c *gin.Context) {
		var req service.ExportChartRequest
		if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
			response.Error(c, "500000", "Invalid request: "+err.Error())
			return
		}
		buf, err := chartHandler.exportService.InnerExportDetails(&req)
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
	})
	chartDataGroup.POST("/innerExportDataSetDetails", func(c *gin.Context) {
		var req service.ExportChartRequest
		if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
			response.Error(c, "500000", "Invalid request: "+err.Error())
			return
		}
		buf, err := chartHandler.exportService.InnerExportDetails(&req)
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
	})
}

func registerChartGroupCompatRoutes(chartGroup *gin.RouterGroup, chartHandler *ChartHandler, permMiddleware *middleware.PermissionMiddleware) {
	if permMiddleware != nil {
		chartGroup.POST("/getData", permMiddleware.CheckChartDataView(), middleware.RowPermissionMiddleware(), chartHandler.Data)
	} else {
		chartGroup.POST("/getData", chartHandler.Data)
	}
	chartGroup.POST("/getChart/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid chart ID")
			return
		}
		result, err := chartHandler.service.Query(&chart.ChartQueryRequest{ID: id})
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, result)
	})
	chartGroup.POST("/getDetail/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid chart ID")
			return
		}
		result, err := chartHandler.service.Query(&chart.ChartQueryRequest{ID: id})
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, result)
	})
	registerChartGroupQueryRoutes(chartGroup, chartHandler, permMiddleware)
	registerChartGroupFieldRoutes(chartGroup, chartHandler)
}

func registerChartGroupQueryRoutes(chartGroup *gin.RouterGroup, chartHandler *ChartHandler, permMiddleware *middleware.PermissionMiddleware) {
	chartGroup.GET("/checkSameDataSet/:viewIdSource/:viewIdTarget", makeCheckSameDataSetHandler(chartHandler))
	chartGroup.POST("/save", makeChartSaveHandler(chartHandler))
	listByDQHandler := makeListByDQHandler(chartHandler, permMiddleware)
	if permMiddleware != nil {
		chartGroup.POST("/listByDQ/:id/:chartId", permMiddleware.CheckDatasetView(), listByDQHandler)
	} else {
		chartGroup.POST("/listByDQ/:id/:chartId", listByDQHandler)
	}
}

func makeCheckSameDataSetHandler(chartHandler *ChartHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		source, err := chartHandler.service.Query(&chart.ChartQueryRequest{ID: sourceID})
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		target, err := chartHandler.service.Query(&chart.ChartQueryRequest{ID: targetID})
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
}

func makeChartSaveHandler(chartHandler *ChartHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
			response.Error(c, "500000", "Invalid request: "+err.Error())
			return
		}
		result, err := chartHandler.service.SaveFromMap(body)
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, result)
	}
}

func makeListByDQHandler(chartHandler *ChartHandler, permMiddleware *middleware.PermissionMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		if permMiddleware != nil {
			result, err = chartHandler.service.ListByDQWithPermission(datasetGroupID, chartID, userID)
		} else if userID > 0 {
			result, err = chartHandler.service.ListByDQWithPermission(datasetGroupID, chartID, userID)
		} else {
			result, err = chartHandler.service.ListByDQ(datasetGroupID, chartID)
		}
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, result)
	}
}

func registerChartGroupFieldRoutes(chartGroup *gin.RouterGroup, chartHandler *ChartHandler) {
	chartGroup.POST("/copyField/:id/:chartId", func(c *gin.Context) {
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
		if err = chartHandler.service.CopyField(id, chartID); err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, nil)
	})
	chartGroup.POST("/deleteField/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid field ID")
			return
		}
		if err = chartHandler.service.DeleteField(id); err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, nil)
	})
	chartGroup.POST("/deleteFieldByChart/:chartId", func(c *gin.Context) {
		chartID, err := strconv.ParseInt(c.Param("chartId"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid chart ID")
			return
		}
		if err = chartHandler.service.DeleteFieldByChart(chartID); err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, nil)
	})
}

func registerDatasetFieldCompatRoutes(datasetFieldGroup *gin.RouterGroup, chartHandler *ChartHandler, datasetHandler *DatasetHandler) {
	datasetFieldGroup.POST("/listByDatasetGroup/:datasetId", func(c *gin.Context) {
		datasetID, err := strconv.ParseInt(c.Param("datasetId"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid dataset ID")
			return
		}
		userID := int64(middleware.GetUserID(c))
		var result *chart.ChartFieldListResponse
		if userID > 0 {
			result, err = chartHandler.service.ListByDQWithPermission(datasetID, 0, userID)
		} else {
			result, err = chartHandler.service.ListByDQ(datasetID, 0)
		}
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, flattenChartFieldList(result))
	})
	datasetFieldGroup.GET("/listWithPermissions/:datasetId", func(c *gin.Context) {
		datasetID, err := strconv.ParseInt(c.Param("datasetId"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid dataset ID")
			return
		}
		userID := int64(middleware.GetUserID(c))
		var result *chart.ChartFieldListResponse
		if userID > 0 {
			result, err = chartHandler.service.ListByDQWithPermission(datasetID, 0, userID)
		} else {
			result, err = chartHandler.service.ListByDQ(datasetID, 0)
		}
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, flattenChartFieldList(result))
	})
	if datasetHandler != nil {
		RegisterDatasetFieldDeleteRoutes(datasetFieldGroup, datasetHandler, chartHandler)
	}
	datasetFieldGroup.POST("/save", func(c *gin.Context) {
		var field dataset.CoreDatasetTableField
		if err := c.ShouldBindBodyWith(&field, binding.JSON); err != nil {
			response.Error(c, "500000", "Invalid request: "+err.Error())
			return
		}
		if datasetHandler == nil {
			response.Error(c, "500000", "dataset service unavailable")
			return
		}
		result, err := datasetHandler.service.SaveField(&field)
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, result)
	})
	datasetFieldGroup.POST("/getFunction", func(c *gin.Context) {
		if datasetHandler == nil {
			response.Success(c, []service.FunctionCategory{})
			return
		}
		response.Success(c, datasetHandler.service.GetFieldFunctions())
	})
	registerDatasetFieldEnumRoutes(datasetFieldGroup, datasetHandler)
}

func registerDatasetFieldEnumRoutes(datasetFieldGroup *gin.RouterGroup, datasetHandler *DatasetHandler) {
	datasetFieldGroup.POST("/multFieldValuesForPermissions", func(c *gin.Context) {
		req, ok := parseMultFieldValuesRequest(c)
		if !ok {
			return
		}
		if datasetHandler == nil {
			response.Success(c, []string{})
			return
		}
		result, err := datasetHandler.service.GetFieldEnum(req)
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, result)
	})
	datasetFieldGroup.POST("/copilotFields/:id", func(c *gin.Context) {
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
		if datasetHandler == nil {
			response.Error(c, "500000", "dataset service unavailable")
			return
		}
		result, err := datasetHandler.service.CopilotFields(datasetID, userID)
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
	})
	datasetFieldGroup.POST("/listByDsIds", func(c *gin.Context) {
		var req struct {
			DsIds []int64 `json:"dsIds"`
		}
		if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
			response.Error(c, "500000", "Invalid request: "+err.Error())
			return
		}
		if datasetHandler == nil {
			response.Success(c, []dataset.CoreDatasetTableField{})
			return
		}
		result, err := datasetHandler.service.ListFieldsByDsIds(req.DsIds)
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, result)
	})
}

func RegisterDatasetFieldDeleteRoutes(r gin.IRouter, datasetHandler *DatasetHandler, chartHandler *ChartHandler) {
	r.POST("/delete/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid field ID")
			return
		}
		if datasetHandler != nil {
			err = datasetHandler.service.DeleteField(id)
		} else {
			err = chartHandler.service.DeleteField(id)
		}
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, nil)
	})
	r.POST("/deleteByChartId/:id", func(c *gin.Context) {
		chartID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, "500000", "Invalid chart ID")
			return
		}
		if datasetHandler != nil {
			err = datasetHandler.service.DeleteFieldByChart(chartID)
		} else {
			err = chartHandler.service.DeleteFieldByChart(chartID)
		}
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		response.Success(c, nil)
	})
}
