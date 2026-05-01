package handler

import (
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
	_ = chartHandler
	r.POST("/delete/:id", datasetHandler.DeleteDatasetField)
	r.POST("/deleteByChartId/:id", datasetHandler.DeleteDatasetFieldByChart)
}
