package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type RelationHandler struct {
	relationService *service.RelationService
}

func NewRelationHandler(svc *service.RelationService) *RelationHandler {
	return &RelationHandler{relationService: svc}
}

func (h *RelationHandler) GetDatasourceRelationship(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid id")
	if !ok {
		return
	}
	result, err := h.relationService.GetDatasourceRelationship(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "500000", "Failed to get datasource relationship: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *RelationHandler) GetDatasetRelationship(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid id")
	if !ok {
		return
	}
	result, err := h.relationService.GetDatasetRelationship(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "500000", "Failed to get dataset relationship: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *RelationHandler) GetPanelRelationship(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid id")
	if !ok {
		return
	}
	result, err := h.relationService.GetPanelRelationship(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "500000", "Failed to get panel relationship: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *RelationHandler) CheckPermission(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.relationService == nil {
		response.Error(c, response.CodeInternalError, "Service unavailable")
		return
	}
	id, ok := parseIDParamMsg(c, "id", "Invalid id")
	if !ok {
		return
	}
	result, err := h.relationService.CheckPermission(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "500000", "Failed to check permission: "+err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterRelationRoutes(r *gin.RouterGroup, h *RelationHandler) {
	relation := r.Group("/relation")
	{
		relation.POST("/datasource/:id", h.GetDatasourceRelationship)
		relation.POST("/dataset/:id", h.GetDatasetRelationship)
		relation.POST("/dv/:id", h.GetPanelRelationship)
	}

	resource := r.Group("/resource")
	{
		resource.POST("/checkPermission/:id", h.CheckPermission)
	}
}
