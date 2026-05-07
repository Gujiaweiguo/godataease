package handler

import (
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
	r.POST("/tree", datasourceHandler.Tree)
	if menuAuthMiddleware != nil {
		r.POST("/validate", menuAuthMiddleware.RequireMenuAuth(datasourceMenuPath), datasourceHandler.Validate)
	} else {
		r.POST("/validate", datasourceHandler.Validate)
	}
	if permMiddleware != nil {
		r.GET("/validate/:id", permMiddleware.CheckDatasourceView(), datasourceHandler.ValidateByID)
	} else {
		r.GET("/validate/:id", datasourceHandler.ValidateByID)
	}
	r.POST("/types", datasourceHandler.Types)
	r.POST("/getTables", datasourceHandler.Tables)
	r.POST("/getTableStatus", datasourceHandler.TableStatus)
	r.POST("/getSchema", datasourceHandler.Schema)
	r.POST("/getTableField", datasourceHandler.TableField)
	r.POST("/previewData", datasourceHandler.PreviewData)
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
	r.POST("/save", datasourceHandler.Save)
	r.POST("/update", datasourceHandler.Update)
	r.POST("/move", datasourceHandler.moveCompat)
	r.POST("/reName", datasourceHandler.renameCompat)
	r.POST("/createFolder", datasourceHandler.createFolderCompat)
	r.POST("/checkRepeat", datasourceHandler.CheckRepeat)
	r.POST("/checkApiDatasource", datasourceHandler.CheckAPIDatasource)
}

func registerDatasourceCompatFileRoutes(r gin.IRouter, datasourceHandler *DatasourceHandler) {
	r.POST("/loadRemoteFile", datasourceHandler.LoadRemoteFile)
	r.POST("/syncApiTable", datasourceHandler.SyncApiTable)
	r.POST("/syncApiDs", datasourceHandler.SyncApiDs)
	r.POST("/uploadFile", datasourceHandler.UploadFile)
	r.POST("/listSyncRecord/:dsId/:page/:limit", datasourceHandler.ListSyncRecord)
}

func registerDatasourceCompatDeleteRoutes(r gin.IRouter, datasourceHandler *DatasourceHandler) {
	r.GET("/delete/:id", datasourceHandler.Delete)
	r.POST("/delete/:id", datasourceHandler.Delete)
	r.POST("/perDelete/:id", datasourceHandler.PerDelete)
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

func (h *DatasourceHandler) moveCompat(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	id, ok := parseInt64Value(body["id"])
	if !ok || id <= 0 {
		response.Error(c, response.CodeInternalError, errInvalidDatasourceID)
		return
	}
	pid, ok := parseInt64Value(body["pid"])
	if !ok {
		pid = 0
	}
	result, err := h.service.Move(id, pid)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) renameCompat(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	id, ok := parseInt64Value(body["id"])
	if !ok || id <= 0 {
		response.Error(c, response.CodeInternalError, errInvalidDatasourceID)
		return
	}
	name, _ := body["name"].(string)
	result, err := h.service.Rename(id, name)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasourceHandler) createFolderCompat(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	name, _ := body["name"].(string)
	pid, ok := parseInt64Value(body["pid"])
	if !ok {
		pid = 0
	}
	result, err := h.service.CreateFolder(name, pid)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
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
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return nil, false
	}
	return result, true
}

func parseDatasourceIDParam(c *gin.Context) (int64, bool) {
	return parseDatasourceIDFromParam(c, "id")
}

func parseDatasourceIDFromParam(c *gin.Context, param string) (int64, bool) {
	return parseIDParamMsg(c, param, errInvalidDatasourceID)
}
