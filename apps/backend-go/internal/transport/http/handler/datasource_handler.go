package handler

import (
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type DatasourceHandler struct {
	service *service.DatasourceService
}

const datasourceMenuPath = "/datasource"

const datasourceValidateRateLimitWindow = time.Minute

const datasourceValidateRateLimitRequests = 30

func NewDatasourceHandler(service *service.DatasourceService) *DatasourceHandler {
	return &DatasourceHandler{service: service}
}

func (h *DatasourceHandler) List(c *gin.Context) {
	var req datasource.ListRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.List(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasourceHandler) Validate(c *gin.Context) {
	var req datasource.ValidateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Validate(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasourceHandler) ValidateByID(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid datasource ID")
		return
	}

	result, err := h.service.ValidateByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasourceHandler) Tree(c *gin.Context) {
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

func (h *DatasourceHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid datasource ID")
		return
	}

	result, err := h.service.GetByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, sanitizeDatasourceResponse(result, h.service))
}

func (h *DatasourceHandler) HidePw(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid datasource ID")
		return
	}

	result, err := h.service.GetByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, sanitizeDatasourceResponse(result, h.service))
}

func (h *DatasourceHandler) GetSimpleDs(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid datasource ID")
		return
	}

	result, err := h.service.GetByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"id": strconv.FormatInt(result.ID, 10), "name": result.Name, "type": result.Type})
}

func (h *DatasourceHandler) Save(c *gin.Context) {
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

func (h *DatasourceHandler) Update(c *gin.Context) {
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

func (h *DatasourceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid datasource ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *DatasourceHandler) PerDelete(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid datasource ID")
		return
	}

	result, err := h.service.PerDelete(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasourceHandler) Tables(c *gin.Context) {
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

func (h *DatasourceHandler) TableStatus(c *gin.Context) {
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

func (h *DatasourceHandler) Schema(c *gin.Context) {
	result, err := h.service.GetSchema()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasourceHandler) TableField(c *gin.Context) {
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

func (h *DatasourceHandler) PreviewData(c *gin.Context) {
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

func (h *DatasourceHandler) SyncApiTable(c *gin.Context) {
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

func (h *DatasourceHandler) SyncApiDs(c *gin.Context) {
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

func (h *DatasourceHandler) LoadRemoteFile(c *gin.Context) {
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

func (h *DatasourceHandler) UploadFile(c *gin.Context) {
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

func (h *DatasourceHandler) CheckRepeat(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

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

func (h *DatasourceHandler) Move(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	req, ok := parseDatasourceWriteRequest(c, false)
	if !ok {
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

func (h *DatasourceHandler) Rename(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	req, ok := parseDatasourceWriteRequest(c, false)
	if !ok {
		return
	}

	result, err := h.service.Rename(req.ID, req.Name)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasourceHandler) CreateFolder(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	req, ok := parseDatasourceWriteRequest(c, false)
	if !ok {
		return
	}

	pid := int64(0)
	if req.PID != nil {
		pid = *req.PID
	}

	result, err := h.service.CreateFolder(req.Name, pid)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasourceHandler) CheckAPIDatasource(c *gin.Context) {
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

func (h *DatasourceHandler) ShowFinishPage(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	userID := getCurrentUserID(c)
	result, err := h.service.ShowFinishPage(userID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasourceHandler) SetShowFinishPage(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	userID := getCurrentUserID(c)
	if err := h.service.SetShowFinishPage(userID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *DatasourceHandler) LatestUse(c *gin.Context) {
	defer recoverDatasourceServicePanic(c)

	username := getCurrentUsername(c)
	result, err := h.service.LatestTypes(username)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func RegisterDatasourceRoutes(r *gin.RouterGroup, h *DatasourceHandler, permMiddleware *middleware.PermissionMiddleware, menuAuthMiddlewares ...*middleware.MenuAuthMiddleware) {
	var menuAuthMiddleware *middleware.MenuAuthMiddleware
	if len(menuAuthMiddlewares) > 0 {
		menuAuthMiddleware = menuAuthMiddlewares[0]
	}
	dsGroup := r.Group("/ds")
	validateGroup := dsGroup.Group("")
	validateGroup.Use(middleware.RateLimit(
		"datasource-validate",
		datasourceValidateRateLimitRequests,
		datasourceValidateRateLimitWindow,
		middleware.AuthenticatedUserKey,
	))
	{
		dsGroup.POST("/list", h.List)
		dsGroup.POST("/tree", h.Tree)
		if menuAuthMiddleware != nil {
			validateGroup.POST("/validate", menuAuthMiddleware.RequireMenuAuth(datasourceMenuPath), h.Validate)
		} else {
			validateGroup.POST("/validate", h.Validate)
		}
		if permMiddleware != nil {
			validateGroup.GET("/validate/:id", permMiddleware.CheckDatasourceView(), h.ValidateByID)
		} else {
			validateGroup.GET("/validate/:id", h.ValidateByID)
		}
		dsGroup.GET("/:id", h.Get)
		dsGroup.GET("/hidePw/:id", h.HidePw)
		dsGroup.GET("/simple/:id", h.GetSimpleDs)
		dsGroup.POST("/save", h.Save)
		dsGroup.POST("/update", h.Update)
		dsGroup.POST("/delete/:id", h.Delete)
		dsGroup.POST("/perDelete/:id", h.PerDelete)
		dsGroup.POST("/move", h.Move)
		dsGroup.POST("/reName", h.Rename)
		dsGroup.POST("/createFolder", h.CreateFolder)
		dsGroup.POST("/checkRepeat", h.CheckRepeat)
		dsGroup.POST("/checkApiDatasource", h.CheckAPIDatasource)
		dsGroup.GET("/types", func(c *gin.Context) {
			defer recoverDatasourceServicePanic(c)
			response.Success(c, []map[string]string{
				{"type": "MySQL", "name": "MySQL"},
				{"type": "PostgreSQL", "name": "PostgreSQL"},
				{"type": "SQLServer", "name": "SQL Server"},
				{"type": "Oracle", "name": "Oracle"},
				{"type": "Excel", "name": "Excel"},
			})
		})
		dsGroup.GET("/showFinishPage", h.ShowFinishPage)
		dsGroup.POST("/showFinishPage", h.SetShowFinishPage)
		dsGroup.POST("/latestUse", h.LatestUse)
		dsGroup.POST("/tables", h.Tables)
		dsGroup.POST("/tableStatus", h.TableStatus)
		dsGroup.POST("/tableField", h.TableField)
		dsGroup.POST("/schema", h.Schema)
		dsGroup.POST("/previewData", h.PreviewData)
		dsGroup.POST("/syncApiTable", h.SyncApiTable)
		dsGroup.POST("/syncApiDs", h.SyncApiDs)
		dsGroup.POST("/loadRemoteFile", h.LoadRemoteFile)
		dsGroup.POST("/syncRecord/:dsId/:page/:limit", func(c *gin.Context) {
			defer recoverDatasourceServicePanic(c)
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
			result, err := h.service.ListSyncRecord(dsID, page, limit)
			if err != nil {
				response.Error(c, "500000", "Failed: "+err.Error())
				return
			}
			response.Success(c, result)
		})
		dsGroup.POST("/uploadFile", h.UploadFile)
	}
}

func getCurrentUserID(c *gin.Context) int64 {
	if uid, exists := c.Get("userId"); exists {
		if id, ok := uid.(int64); ok {
			return id
		}
	}
	return 0
}

func getCurrentUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if s, ok := username.(string); ok {
			return s
		}
	}
	return ""
}
