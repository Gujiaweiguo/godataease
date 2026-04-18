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
	id := c.Param("id")
	response.Success(c, map[string]interface{}{
		"id":           id,
		"busiFlag":     "datasource",
		"relationList": []interface{}{},
	})
}

func (h *RelationHandler) GetDatasetRelationship(c *gin.Context) {
	id := c.Param("id")
	response.Success(c, map[string]interface{}{
		"id":           id,
		"busiFlag":     "dataset",
		"relationList": []interface{}{},
	})
}

func (h *RelationHandler) GetPanelRelationship(c *gin.Context) {
	id := c.Param("id")
	response.Success(c, map[string]interface{}{
		"id":           id,
		"busiFlag":     "dashboard",
		"relationList": []interface{}{},
	})
}

func (h *RelationHandler) CheckPermission(c *gin.Context) {
	id := c.Param("id")
	response.Success(c, map[string]interface{}{
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
