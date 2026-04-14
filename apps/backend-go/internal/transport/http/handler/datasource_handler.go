package handler

import (
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"errors"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DatasourceHandler struct {
	service *service.DatasourceService
}

func NewDatasourceHandler(service *service.DatasourceService) *DatasourceHandler {
	return &DatasourceHandler{service: service}
}

func (h *DatasourceHandler) List(c *gin.Context) {
	var req datasource.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	if err := c.ShouldBindJSON(&req); err != nil {
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

func (h *DatasourceHandler) Tree(c *gin.Context) {
	var req datasource.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
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
	if err := c.ShouldBindJSON(&req); err != nil {
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
	if err := c.ShouldBindJSON(&req); err != nil {
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

func RegisterDatasourceRoutes(r *gin.RouterGroup, h *DatasourceHandler) {
	dsGroup := r.Group("/ds")
	{
		dsGroup.POST("/list", h.List)
		dsGroup.POST("/tree", h.Tree)
		dsGroup.POST("/validate", h.Validate)
		dsGroup.GET("/:id", h.Get)
		dsGroup.POST("/save", h.Save)
		dsGroup.POST("/update", h.Update)
		dsGroup.POST("/delete/:id", h.Delete)
		dsGroup.POST("/tables", h.Tables)
		dsGroup.POST("/tableStatus", h.TableStatus)
		dsGroup.POST("/tableField", h.TableField)
		dsGroup.POST("/schema", h.Schema)
		dsGroup.POST("/previewData", h.PreviewData)
		dsGroup.POST("/syncApiTable", h.SyncApiTable)
		dsGroup.POST("/syncApiDs", h.SyncApiDs)
	}
}
