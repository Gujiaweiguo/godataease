package handler

import (
	"strconv"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

func registerDatasetCompatRoutes(r gin.IRouter, datasetHandler *DatasetHandler, permMiddleware *middleware.PermissionMiddleware) {
	datasetTreeGroup := r.Group("/datasetTree")
	{
		datasetTreeGroup.POST("/tree", datasetHandler.Tree)
		datasetTreeGroup.POST("/get/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid dataset ID")
				return
			}
			result, err := buildDatasetDetail(datasetHandler, id)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetTreeGroup.POST("/details/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid dataset ID")
				return
			}
			result, err := buildDatasetDetail(datasetHandler, id)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetTreeGroup.POST("/dsDetails", func(c *gin.Context) {
			ids, ok := parseDatasetIDs(c)
			if !ok {
				return
			}
			result := make([]gin.H, 0, len(ids))
			for _, id := range ids {
				detail, err := buildDatasetDetail(datasetHandler, id)
				if err != nil {
					continue
				}
				result = append(result, detail)
			}
			response.Success(c, result)
		})
		if permMiddleware != nil {
			datasetTreeGroup.POST("/detailWithPerm", permMiddleware.CheckDatasetBatchView(), middleware.RowPermissionMiddleware(), datasetHandler.DetailWithPerm)
		} else {
			datasetTreeGroup.POST("/detailWithPerm", datasetHandler.DetailWithPerm)
		}
		datasetTreeGroup.POST("/getSqlParams", func(c *gin.Context) {
			ids, ok := parseDatasetIDs(c)
			if !ok {
				return
			}
			result, err := datasetHandler.service.GetSQLParams(ids)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetTreeGroup.POST("/save", func(c *gin.Context) {
			req, ok := parseDatasetWriteRequest(c, true)
			if !ok {
				return
			}
			result, err := datasetHandler.service.Save(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetTreeGroup.POST("/create", func(c *gin.Context) {
			req, ok := parseDatasetWriteRequest(c, true)
			if !ok {
				return
			}
			result, err := datasetHandler.service.Create(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetTreeGroup.POST("/rename", func(c *gin.Context) {
			req, ok := parseDatasetWriteRequest(c, true)
			if !ok {
				return
			}
			if req.ID <= 0 {
				response.Error(c, "500000", "Invalid dataset ID")
				return
			}
			result, err := datasetHandler.service.Rename(req.ID, req.Name)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetTreeGroup.POST("/move", func(c *gin.Context) {
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
			result, err := datasetHandler.service.Move(req.ID, pid)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetTreeGroup.POST("/delete/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid dataset ID")
				return
			}
			if err = datasetHandler.service.Delete(id); err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, nil)
		})
		datasetTreeGroup.POST("/perDelete/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid dataset ID")
				return
			}
			result, err := datasetHandler.service.PerDelete(id)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetTreeGroup.GET("/barInfo/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid dataset ID")
				return
			}
			group, err := datasetHandler.service.GetGroupByID(id)
			if err != nil {
				response.Error(c, "500000", "Failed to get dataset: "+err.Error())
				return
			}
			barInfo := &dataset.BarInfo{ID: group.ID, Name: group.Name, CreateBy: group.CreateBy, CreateTime: group.CreateTime, UpdateBy: group.UpdateBy, LastUpdateTime: group.LastUpdateTime, IsCross: false}
			if group.NodeType != nil {
				barInfo.NodeType = *group.NodeType
			}
			barInfo.Creator = datasetHandler.service.ResolveUserName(group.CreateBy)
			barInfo.Updater = datasetHandler.service.ResolveUserName(group.UpdateBy)
			response.Success(c, barInfo)
		})
		datasetTreeGroup.POST("/exportDataset", datasetHandler.ExportDataset)
	}

	datasetDataGroup := r.Group("/datasetData")
	{
		datasetDataGroup.POST("/tableField", datasetHandler.Fields)
		datasetDataGroup.POST("/previewData", datasetHandler.Preview)
		datasetDataGroup.POST("/getDatasetTotal", func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}
			id, ok := parseInt64Value(body["id"])
			if !ok {
				response.Success(c, int64(0))
				return
			}
			preview, err := datasetHandler.service.Preview(&dataset.PreviewRequest{DatasetGroupID: id, Limit: 1})
			if err != nil {
				response.Success(c, int64(0))
				return
			}
			response.Success(c, preview.Total)
		})
		datasetDataGroup.POST("/previewSql", func(c *gin.Context) {
			var req dataset.SQLPreviewRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}
			result, err := datasetHandler.service.PreviewSQLWithUser(&req, int64(middleware.GetUserID(c)))
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetDataGroup.POST("/enumValueObj", func(c *gin.Context) {
			req, ok := parseEnumValueRequest(c)
			if !ok {
				return
			}
			result, err := datasetHandler.service.GetFieldEnumObj(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetDataGroup.POST("/enumValueDs", func(c *gin.Context) {
			fieldID, ok := parseEnumFieldID(c)
			if !ok {
				return
			}
			result, err := datasetHandler.service.GetFieldEnumDs(fieldID)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasetDataGroup.POST("/enumValue", func(c *gin.Context) {
			req, ok := parseMultFieldValuesRequest(c)
			if !ok {
				return
			}
			result, err := datasetHandler.service.GetFieldEnum(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
	}
}
