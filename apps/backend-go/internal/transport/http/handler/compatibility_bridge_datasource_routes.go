package handler

import (
	"errors"
	"io"
	"strconv"

	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func registerDatasourceCompatRoutes(
	r gin.IRouter,
	datasourceHandler *DatasourceHandler,
	permMiddleware *middleware.PermissionMiddleware,
	menuAuthMiddleware *middleware.MenuAuthMiddleware,
) {
	datasourceGroup := r.Group("/datasource")
	registerDatasourceCompatReadRoutes(datasourceGroup, datasourceHandler, permMiddleware, menuAuthMiddleware)
	registerDatasourceCompatUserRoutes(datasourceGroup, datasourceHandler)
	registerDatasourceCompatWriteRoutes(datasourceGroup, datasourceHandler)
	registerDatasourceCompatFileRoutes(datasourceGroup, datasourceHandler)
	registerDatasourceCompatDeleteRoutes(datasourceGroup, datasourceHandler)
}

func registerDatasourceCompatReadRoutes(r gin.IRouter, datasourceHandler *DatasourceHandler, permMiddleware *middleware.PermissionMiddleware, menuAuthMiddleware *middleware.MenuAuthMiddleware) {
	r.POST("/list", datasourceHandler.List)
	r.POST("/tree", datasourceHandler.treeCompat)
	if menuAuthMiddleware != nil {
		r.POST("/validate", menuAuthMiddleware.RequireMenuAuth(datasourceMenuPath), datasourceHandler.Validate)
	} else {
		r.POST("/validate", datasourceHandler.Validate)
	}
	if permMiddleware != nil {
		r.GET("/validate/:id", permMiddleware.CheckDatasourceView(), datasourceHandler.validateByIDCompat)
	} else {
		r.GET("/validate/:id", datasourceHandler.validateByIDCompat)
	}
	r.POST("/types", datasourceTypesCompat)
	r.POST("/getTables", datasourceHandler.getTablesCompat)
	r.POST("/getTableStatus", datasourceHandler.getTableStatusCompat)
	r.POST("/getSchema", datasourceHandler.getSchemaCompat)
	r.POST("/getTableField", datasourceHandler.getTableFieldCompat)
	r.POST("/previewData", datasourceHandler.previewDataCompat)
	r.GET("/get/:id", datasourceHandler.getCompat)
	r.GET("/hidePw/:id", datasourceHandler.hidePwCompat)
	r.GET("/getSimpleDs/:id", datasourceHandler.getSimpleDsCompat)
}

func registerDatasourceCompatUserRoutes(r gin.IRouter, datasourceHandler *DatasourceHandler) {
	r.GET("/showFinishPage", datasourceHandler.ShowFinishPage)
	r.POST("/setShowFinishPage", datasourceHandler.SetShowFinishPage)
	r.POST("/latestUse", datasourceHandler.LatestUse)
}

func registerDatasourceCompatWriteRoutes(r gin.IRouter, datasourceHandler *DatasourceHandler) {
	r.POST("/save", datasourceHandler.saveCompat)
	r.POST("/update", datasourceHandler.updateCompat)
	r.POST("/move", datasourceHandler.moveCompat)
	r.POST("/reName", datasourceHandler.renameCompat)
	r.POST("/createFolder", datasourceHandler.createFolderCompat)
	r.POST("/checkRepeat", datasourceHandler.checkRepeatCompat)
	r.POST("/checkApiDatasource", datasourceHandler.checkAPIDatasourceCompat)
}

func registerDatasourceCompatFileRoutes(r gin.IRouter, datasourceHandler *DatasourceHandler) {
	r.POST("/loadRemoteFile", datasourceHandler.loadRemoteFileCompat)
	r.POST("/syncApiTable", datasourceHandler.syncAPITableCompat)
	r.POST("/syncApiDs", datasourceHandler.syncAPIDsCompat)
	r.POST("/uploadFile", datasourceHandler.uploadFileCompat)
	r.POST("/listSyncRecord/:dsId/:page/:limit", datasourceHandler.listSyncRecordCompat)
}

func registerDatasourceCompatDeleteRoutes(r gin.IRouter, datasourceHandler *DatasourceHandler) {
	r.GET("/delete/:id", datasourceHandler.deleteCompat)
	r.POST("/delete/:id", datasourceHandler.deleteCompat)
	r.POST("/perDelete/:id", datasourceHandler.perDeleteCompat)
}

func (h *DatasourceHandler) treeCompat(c *gin.Context) {
	var req datasource.ListRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	list, err := h.service.Tree(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, buildDatasourceTreeResponse(list))
}

func (h *DatasourceHandler) validateByIDCompat(c *gin.Context) {
	id, ok := parseDatasourceIDParam(c)
	if !ok {
		return
	}
	result, err := h.service.ValidateByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func datasourceTypesCompat(c *gin.Context) {
	response.Success(c, []map[string]string{{"type": "MySQL", "name": "MySQL"}, {"type": "PostgreSQL", "name": "PostgreSQL"}, {"type": "SQLServer", "name": "SQL Server"}, {"type": "Oracle", "name": "Oracle"}, {"type": "Excel", "name": "Excel"}})
}

func (h *DatasourceHandler) getTablesCompat(c *gin.Context) {
	req, ok := parseTableRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetTables(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) getTableStatusCompat(c *gin.Context) {
	req, ok := parseTableRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetTableStatus(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) getSchemaCompat(c *gin.Context) {
	result, err := h.service.GetSchema()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) getTableFieldCompat(c *gin.Context) {
	req, ok := parseTableRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetTableField(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) previewDataCompat(c *gin.Context) {
	req, ok := parseTableRequest(c)
	if !ok {
		return
	}
	result, err := h.service.PreviewData(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) getCompat(c *gin.Context) {
	result, ok := h.getDatasourceByIDCompat(c)
	if !ok {
		return
	}
	response.Success(c, sanitizeDatasourceResponse(result, h.service))
}

func (h *DatasourceHandler) hidePwCompat(c *gin.Context) {
	h.getCompat(c)
}

func (h *DatasourceHandler) getSimpleDsCompat(c *gin.Context) {
	result, ok := h.getDatasourceByIDCompat(c)
	if !ok {
		return
	}
	response.Success(c, gin.H{"id": strconv.FormatInt(result.ID, 10), "name": result.Name, "type": result.Type})
}

func (h *DatasourceHandler) saveCompat(c *gin.Context) {
	req, ok := parseDatasourceWriteRequest(c, true)
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

func (h *DatasourceHandler) updateCompat(c *gin.Context) {
	req, ok := parseDatasourceWriteRequest(c, true)
	if !ok {
		return
	}
	result, err := h.service.Update(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) moveCompat(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
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
	result, err := h.service.Move(id, pid)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) renameCompat(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	id, ok := parseInt64Value(body["id"])
	if !ok || id <= 0 {
		response.Error(c, "500000", "Invalid datasource ID")
		return
	}
	name, _ := body["name"].(string)
	result, err := h.service.Rename(id, name)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) createFolderCompat(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	name, _ := body["name"].(string)
	pid, ok := parseInt64Value(body["pid"])
	if !ok {
		pid = 0
	}
	result, err := h.service.CreateFolder(name, pid)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) checkRepeatCompat(c *gin.Context) {
	req, ok := parseDatasourceWriteRequest(c, false)
	if !ok {
		return
	}
	result, err := h.service.CheckRepeat(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) checkAPIDatasourceCompat(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.CheckAPIDatasource(req)
	if err != nil {
		response.Error(c, "500000", "Failed to check api datasource: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) loadRemoteFileCompat(c *gin.Context) {
	var req struct {
		URL          string `json:"url"`
		UserName     string `json:"userName"`
		Password     string `json:"passwd"`
		DatasourceID int64  `json:"datasourceId"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.LoadRemoteFile(req.URL, req.UserName, req.Password, req.DatasourceID)
	if err != nil {
		response.Error(c, "500000", "Failed to load remote file: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) syncAPITableCompat(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.SyncAPITable(req)
	if err != nil {
		response.Error(c, "500000", "Failed to sync api table: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) syncAPIDsCompat(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.SyncAPIDs(req)
	if err != nil {
		response.Error(c, "500000", "Failed to sync api datasource: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) uploadFileCompat(c *gin.Context) {
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
	result, err := h.service.UploadFile(file, header, datasourceID, editType)
	if err != nil {
		response.Error(c, "500000", "Failed to process file: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) listSyncRecordCompat(c *gin.Context) {
	dsID, ok := parseDatasourceIDFromParam(c, "dsId")
	if !ok {
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
	result, err := h.service.ListSyncRecord(dsID, page, limit)
	if err != nil {
		response.Error(c, "500000", "Failed to list sync records: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) deleteCompat(c *gin.Context) {
	id, ok := parseDatasourceIDParam(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DatasourceHandler) perDeleteCompat(c *gin.Context) {
	id, ok := parseDatasourceIDParam(c)
	if !ok {
		return
	}
	result, err := h.service.PerDelete(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) getDatasourceByIDCompat(c *gin.Context) (*datasource.CoreDatasource, bool) {
	id, ok := parseDatasourceIDParam(c)
	if !ok {
		return nil, false
	}
	result, err := h.service.GetByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return nil, false
	}
	return result, true
}

func parseDatasourceIDParam(c *gin.Context) (int64, bool) {
	return parseDatasourceIDFromParam(c, "id")
}

func parseDatasourceIDFromParam(c *gin.Context, param string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid datasource ID")
		return 0, false
	}
	return id, true
}
