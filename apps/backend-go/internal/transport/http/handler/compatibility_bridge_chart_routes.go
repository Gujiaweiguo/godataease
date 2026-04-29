package handler

import (
	"strconv"

	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

func registerChartCompatRoutes(r gin.IRouter, chartHandler *ChartHandler, datasetHandler *DatasetHandler, permMiddleware *middleware.PermissionMiddleware) {
	RegisterChartDataCompatRoutes(r.Group("/chartData"), chartHandler, datasetHandler, permMiddleware)
	registerChartGroupCompatRoutes(r.Group("/chart"), chartHandler, permMiddleware)
	registerDatasetFieldCompatRoutes(r.Group("/datasetField"), chartHandler, datasetHandler)
}

func RegisterChartDataCompatRoutes(chartDataGroup *gin.RouterGroup, chartHandler *ChartHandler, _ *DatasetHandler, permMiddleware *middleware.PermissionMiddleware) {
	RegisterChartDataRoutes(chartDataGroup, chartHandler, permMiddleware)
}

func registerChartGroupCompatRoutes(chartGroup *gin.RouterGroup, chartHandler *ChartHandler, permMiddleware *middleware.PermissionMiddleware) {
	if permMiddleware != nil {
		chartGroup.POST("/getData", permMiddleware.CheckChartDataView(), middleware.RowPermissionMiddleware(), chartHandler.Data)
	} else {
		chartGroup.POST("/getData", chartHandler.Data)
	}
	chartGroup.POST("/getChart/:id", chartHandler.GetChart)
	chartGroup.POST("/getDetail/:id", chartHandler.GetDetail)
	registerChartGroupQueryRoutes(chartGroup, chartHandler, permMiddleware)
	registerChartGroupFieldRoutes(chartGroup, chartHandler)
}

func registerChartGroupQueryRoutes(chartGroup *gin.RouterGroup, chartHandler *ChartHandler, permMiddleware *middleware.PermissionMiddleware) {
	chartGroup.GET("/checkSameDataSet/:viewIdSource/:viewIdTarget", chartHandler.CheckSameDataSet)
	chartGroup.POST("/save", chartHandler.SaveFromMap)
	if permMiddleware != nil {
		chartGroup.POST("/listByDQ/:id/:chartId", permMiddleware.CheckDatasetView(), chartHandler.ListByDQ)
	} else {
		chartGroup.POST("/listByDQ/:id/:chartId", chartHandler.ListByDQ)
	}
}

func registerChartGroupFieldRoutes(chartGroup *gin.RouterGroup, chartHandler *ChartHandler) {
	chartGroup.POST("/copyField/:id/:chartId", chartHandler.CopyField)
	chartGroup.POST("/deleteField/:id", chartHandler.DeleteField)
	chartGroup.POST("/deleteFieldByChart/:chartId", chartHandler.DeleteFieldByChart)
}

func registerDatasetFieldCompatRoutes(datasetFieldGroup *gin.RouterGroup, chartHandler *ChartHandler, datasetHandler *DatasetHandler) {
	datasetCompatHandler := datasetHandler
	if datasetCompatHandler == nil {
		datasetCompatHandler = NewDatasetHandler(nil, chartHandler.service)
	} else if datasetCompatHandler.chartService == nil {
		datasetCompatHandler.chartService = chartHandler.service
	}
	datasetFieldGroup.POST("/listByDatasetGroup/:datasetId", datasetCompatHandler.ListByDatasetGroup)
	datasetFieldGroup.GET("/listWithPermissions/:datasetId", datasetCompatHandler.ListWithPermissions)
	if datasetHandler != nil {
		RegisterDatasetFieldDeleteRoutes(datasetFieldGroup, datasetHandler, chartHandler)
	}
	datasetFieldGroup.POST("/save", datasetCompatHandler.SaveField)
	datasetFieldGroup.POST("/getFunction", datasetCompatHandler.GetFieldFunctions)
	registerDatasetFieldEnumRoutes(datasetFieldGroup, datasetHandler)
}

func registerDatasetFieldEnumRoutes(datasetFieldGroup *gin.RouterGroup, datasetHandler *DatasetHandler) {
	datasetFieldGroup.POST("/multFieldValuesForPermissions", datasetHandler.MultFieldValuesForPermissions)
	datasetFieldGroup.POST("/copilotFields/:id", datasetHandler.CopilotFields)
	datasetFieldGroup.POST("/listByDsIds", datasetHandler.ListFieldsByDsIds)
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
