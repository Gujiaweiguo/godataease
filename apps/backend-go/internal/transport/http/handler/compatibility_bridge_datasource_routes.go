package handler

import (
	"errors"
	"io"
	"strconv"

	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func registerDatasourceCompatRoutes(
	r gin.IRouter,
	datasourceHandler *DatasourceHandler,
	getCurrentUserID func(*gin.Context) int64,
	getCurrentUsername func(*gin.Context) string,
) {
	datasourceGroup := r.Group("/datasource")
	{
		datasourceGroup.POST("/list", datasourceHandler.List)
		datasourceGroup.POST("/tree", func(c *gin.Context) {
			var req datasource.ListRequest
			if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}

			list, err := datasourceHandler.service.Tree(&req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, buildDatasourceTreeResponse(list))
		})
		datasourceGroup.POST("/validate", datasourceHandler.Validate)
		datasourceGroup.GET("/validate/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid datasource ID")
				return
			}
			result, err := datasourceHandler.service.ValidateByID(id)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/types", func(c *gin.Context) {
			response.Success(c, []map[string]string{{"type": "MySQL", "name": "MySQL"}, {"type": "PostgreSQL", "name": "PostgreSQL"}, {"type": "SQLServer", "name": "SQL Server"}, {"type": "Oracle", "name": "Oracle"}, {"type": "Excel", "name": "Excel"}})
		})
		datasourceGroup.POST("/getTables", func(c *gin.Context) {
			req, ok := parseTableRequest(c)
			if !ok {
				return
			}
			result, err := datasourceHandler.service.GetTables(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/getTableStatus", func(c *gin.Context) {
			req, ok := parseTableRequest(c)
			if !ok {
				return
			}
			result, err := datasourceHandler.service.GetTableStatus(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/getSchema", func(c *gin.Context) {
			result, err := datasourceHandler.service.GetSchema()
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/getTableField", func(c *gin.Context) {
			req, ok := parseTableRequest(c)
			if !ok {
				return
			}
			result, err := datasourceHandler.service.GetTableField(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/previewData", func(c *gin.Context) {
			req, ok := parseTableRequest(c)
			if !ok {
				return
			}
			result, err := datasourceHandler.service.PreviewData(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.GET("/get/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid datasource ID")
				return
			}
			result, err := datasourceHandler.service.GetByID(id)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, sanitizeDatasourceResponse(result, datasourceHandler.service))
		})
		datasourceGroup.GET("/hidePw/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid datasource ID")
				return
			}
			result, err := datasourceHandler.service.GetByID(id)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, sanitizeDatasourceResponse(result, datasourceHandler.service))
		})
		datasourceGroup.GET("/getSimpleDs/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid datasource ID")
				return
			}
			result, err := datasourceHandler.service.GetByID(id)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, gin.H{"id": strconv.FormatInt(result.ID, 10), "name": result.Name, "type": result.Type})
		})
		datasourceGroup.GET("/showFinishPage", func(c *gin.Context) {
			result, err := datasourceHandler.service.ShowFinishPage(getCurrentUserID(c))
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/setShowFinishPage", func(c *gin.Context) {
			if err := datasourceHandler.service.SetShowFinishPage(getCurrentUserID(c)); err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, nil)
		})
		datasourceGroup.POST("/latestUse", func(c *gin.Context) {
			result, err := datasourceHandler.service.LatestTypes(getCurrentUsername(c))
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/save", func(c *gin.Context) {
			req, ok := parseDatasourceWriteRequest(c, true)
			if !ok {
				return
			}
			result, err := datasourceHandler.service.Save(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/update", func(c *gin.Context) {
			req, ok := parseDatasourceWriteRequest(c, true)
			if !ok {
				return
			}
			result, err := datasourceHandler.service.Update(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/move", func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}
			id, ok := parseInt64Value(body["id"])
			if !ok || id <= 0 {
				response.Error(c, "500000", "Invalid datasource ID")
				return
			}
			pid, ok := parseInt64Value(body["pid"])
			if !ok {
				pid = 0
			}
			result, err := datasourceHandler.service.Move(id, pid)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/reName", func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}
			id, ok := parseInt64Value(body["id"])
			if !ok || id <= 0 {
				response.Error(c, "500000", "Invalid datasource ID")
				return
			}
			name, _ := body["name"].(string)
			result, err := datasourceHandler.service.Rename(id, name)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/createFolder", func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}
			name, _ := body["name"].(string)
			pid, ok := parseInt64Value(body["pid"])
			if !ok {
				pid = 0
			}
			result, err := datasourceHandler.service.CreateFolder(name, pid)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/checkRepeat", func(c *gin.Context) {
			req, ok := parseDatasourceWriteRequest(c, false)
			if !ok {
				return
			}
			result, err := datasourceHandler.service.CheckRepeat(req)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/checkApiDatasource", func(c *gin.Context) {
			var req map[string]string
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}
			result, err := datasourceHandler.service.CheckAPIDatasource(req)
			if err != nil {
				response.Error(c, "500000", "Failed to check api datasource: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/loadRemoteFile", func(c *gin.Context) {
			var req struct {
				URL          string `json:"url"`
				UserName     string `json:"userName"`
				Password     string `json:"passwd"`
				DatasourceID int64  `json:"datasourceId"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}
			result, err := datasourceHandler.service.LoadRemoteFile(req.URL, req.UserName, req.Password, req.DatasourceID)
			if err != nil {
				response.Error(c, "500000", "Failed to load remote file: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/syncApiTable", func(c *gin.Context) {
			var req map[string]string
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}
			result, err := datasourceHandler.service.SyncAPITable(req)
			if err != nil {
				response.Error(c, "500000", "Failed to sync api table: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/syncApiDs", func(c *gin.Context) {
			var req map[string]string
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, "500000", "Invalid request: "+err.Error())
				return
			}
			result, err := datasourceHandler.service.SyncAPIDs(req)
			if err != nil {
				response.Error(c, "500000", "Failed to sync api datasource: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/uploadFile", func(c *gin.Context) {
			file, header, err := c.Request.FormFile("file")
			if err != nil {
				response.Error(c, "500000", "Failed to get uploaded file: "+err.Error())
				return
			}
			defer file.Close()
			var datasourceID int64
			if idStr := c.PostForm("id"); idStr != "" {
				datasourceID, _ = strconv.ParseInt(idStr, 10, 64)
			}
			var editType int
			if editTypeStr := c.PostForm("editType"); editTypeStr != "" {
				editType, _ = strconv.Atoi(editTypeStr)
			}
			result, err := datasourceHandler.service.UploadFile(file, header, datasourceID, editType)
			if err != nil {
				response.Error(c, "500000", "Failed to process file: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		datasourceGroup.POST("/listSyncRecord/:dsId/:page/:limit", func(c *gin.Context) {
			dsID, err := strconv.ParseInt(c.Param("dsId"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid datasource ID")
				return
			}
			page, _ := strconv.Atoi(c.Param("page"))
			if page < 1 {
				page = 1
			}
			limit, _ := strconv.Atoi(c.Param("limit"))
			if limit < 1 {
				limit = 10
			}
			result, err := datasourceHandler.service.ListSyncRecord(dsID, page, limit)
			if err != nil {
				response.Error(c, "500000", "Failed to list sync records: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		deleteDatasourceHandler := func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid datasource ID")
				return
			}
			if err = datasourceHandler.service.Delete(id); err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, nil)
		}
		datasourceGroup.GET("/delete/:id", deleteDatasourceHandler)
		datasourceGroup.POST("/delete/:id", deleteDatasourceHandler)
		datasourceGroup.POST("/perDelete/:id", func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, "500000", "Invalid datasource ID")
				return
			}
			result, err := datasourceHandler.service.PerDelete(id)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
	}
}
