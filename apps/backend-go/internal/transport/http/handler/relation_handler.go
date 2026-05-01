package handler

import (
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type RelationHandler struct{}

func NewRelationHandler() *RelationHandler {
	return &RelationHandler{}
}

func (h *RelationHandler) GetDatasourceRelationship(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	response.Success(c, map[string]any{
		"id":           id,
		"busiFlag":     "datasource",
		"relationList": []any{},
	})
}

func (h *RelationHandler) GetDatasetRelationship(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	response.Success(c, map[string]any{
		"id":           id,
		"busiFlag":     "dataset",
		"relationList": []any{},
	})
}

func (h *RelationHandler) GetPanelRelationship(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	response.Success(c, map[string]any{
		"id":           id,
		"busiFlag":     "dashboard",
		"relationList": []any{},
	})
}

func (h *RelationHandler) CheckPermission(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	response.Success(c, map[string]any{
		"id":        id,
		"editable":  true,
		"creatable": true,
	})
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
