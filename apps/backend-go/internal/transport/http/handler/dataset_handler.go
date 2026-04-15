package handler

import (
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DatasetHandler struct {
	service *service.DatasetService
}

func NewDatasetHandler(service *service.DatasetService) *DatasetHandler {
	return &DatasetHandler{service: service}
}

func (h *DatasetHandler) Tree(c *gin.Context) {
	var req dataset.TreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Tree(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) Fields(c *gin.Context) {
	var req dataset.FieldsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Fields(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) Preview(c *gin.Context) {
	var req dataset.PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Preview(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) PreviewWithPermission(c *gin.Context) {
	var req dataset.PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	userID := int64(middleware.GetUserID(c))
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	result, err := h.service.PreviewWithPermission(&req, userID)
	if err != nil {
		if errors.Is(err, service.ErrDatasetDatasourcePermissionDenied) {
			response.Forbidden(c, err.Error())
			return
		}
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *DatasetHandler) Save(c *gin.Context) {
	defer recoverDatasetServicePanic(c)
	req, ok := parseDatasetWriteRequest(c, true)
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

func (h *DatasetHandler) Create(c *gin.Context) {
	defer recoverDatasetServicePanic(c)
	req, ok := parseDatasetWriteRequest(c, true)
	if !ok {
		return
	}
	result, err := h.service.Create(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) Rename(c *gin.Context) {
	defer recoverDatasetServicePanic(c)
	req, ok := parseDatasetWriteRequest(c, true)
	if !ok {
		return
	}
	if req.ID <= 0 {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	result, err := h.service.Rename(req.ID, req.Name)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) Move(c *gin.Context) {
	defer recoverDatasetServicePanic(c)
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
	result, err := h.service.Move(req.ID, pid)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DatasetHandler) Delete(c *gin.Context) {
	defer recoverDatasetServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	if err = h.service.Delete(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DatasetHandler) PerDelete(c *gin.Context) {
	defer recoverDatasetServicePanic(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid dataset ID")
		return
	}
	result, err := h.service.PerDelete(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func recoverDatasetServicePanic(c *gin.Context) {
	if r := recover(); r != nil {
		response.Error(c, "500000", "Service unavailable")
	}
}

func RegisterDatasetRoutes(r *gin.RouterGroup, h *DatasetHandler) {
	datasetGroup := r.Group("/dataset")
	{
		datasetGroup.POST("/tree", h.Tree)
		datasetGroup.POST("/fields", h.Fields)
		datasetGroup.POST("/preview", h.Preview)
		datasetGroup.POST("/previewWithPerm", h.PreviewWithPermission)
		datasetGroup.POST("/save", h.Save)
		datasetGroup.POST("/create", h.Create)
		datasetGroup.POST("/rename", h.Rename)
		datasetGroup.POST("/move", h.Move)
		datasetGroup.POST("/delete/:id", h.Delete)
		datasetGroup.POST("/perDelete/:id", h.PerDelete)
	}
}
