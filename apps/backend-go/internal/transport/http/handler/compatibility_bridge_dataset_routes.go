package handler

import (
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

func registerDatasetCompatRoutes(r gin.IRouter, datasetHandler *DatasetHandler, permMiddleware *middleware.PermissionMiddleware) {
	registerDatasetTreeCompatRoutes(r.Group("/datasetTree"), datasetHandler, permMiddleware)
	registerDatasetDataCompatRoutes(r.Group("/datasetData"), datasetHandler)
}

func registerDatasetTreeCompatRoutes(g *gin.RouterGroup, datasetHandler *DatasetHandler, permMiddleware *middleware.PermissionMiddleware) {
	registerDatasetTreeQueryRoutes(g, datasetHandler, permMiddleware)
	registerDatasetTreeWriteRoutes(g, datasetHandler)
}

func registerDatasetTreeQueryRoutes(g *gin.RouterGroup, datasetHandler *DatasetHandler, permMiddleware *middleware.PermissionMiddleware) {
	g.POST("/tree", datasetHandler.Tree)
	g.POST("/get/:id", datasetHandler.GetDetail)
	g.POST("/details/:id", datasetHandler.Details)
	g.POST("/dsDetails", datasetHandler.DsDetails)
	if permMiddleware != nil {
		g.POST("/detailWithPerm", permMiddleware.CheckDatasetBatchView(), middleware.RowPermissionMiddleware(), datasetHandler.DetailWithPerm)
	} else {
		g.POST("/detailWithPerm", datasetHandler.DetailWithPerm)
	}
	g.POST("/getSqlParams", datasetHandler.GetSQLParams)
	g.GET("/barInfo/:id", datasetHandler.BarInfo)
	g.POST("/exportDataset", datasetHandler.ExportDataset)
	g.POST("/delete/:id", datasetHandler.Delete)
	g.POST("/perDelete/:id", datasetHandler.PerDelete)
}

func registerDatasetTreeWriteRoutes(g *gin.RouterGroup, datasetHandler *DatasetHandler) {
	g.POST("/save", datasetHandler.Save)
	g.POST("/create", datasetHandler.Create)
	g.POST("/rename", datasetHandler.Rename)
	g.POST("/move", datasetHandler.Move)
}

func registerDatasetDataCompatRoutes(g *gin.RouterGroup, datasetHandler *DatasetHandler) {
	g.POST("/tableField", datasetHandler.Fields)
	g.POST("/previewData", datasetHandler.Preview)
	g.POST("/getDatasetTotal", datasetHandler.GetDatasetTotal)
	g.POST("/previewSql", datasetHandler.PreviewSQL)
	g.POST("/enumValueObj", datasetHandler.EnumValueObj)
	g.POST("/enumValueDs", datasetHandler.EnumValueDs)
	g.POST("/enumValue", datasetHandler.EnumValue)
	g.POST("/getFieldTree", datasetHandler.GetFieldTree)
}
