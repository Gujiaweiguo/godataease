package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// LinkJumpHandler handles HTTP requests for chart-to-dashboard jump navigation.
type LinkJumpHandler struct {
	service *service.LinkJumpService
}

func NewLinkJumpHandler(service *service.LinkJumpService) *LinkJumpHandler {
	return &LinkJumpHandler{service: service}
}

// GetTableFieldWithViewID returns dataset fields for a chart view.
func (h *LinkJumpHandler) GetTableFieldWithViewID(c *gin.Context) {
	defer recoverServicePanic(c)
	viewID, ok := parseIDParamMsg(c, "viewId", "Invalid viewId")
	if !ok {
		return
	}
	result, err := h.service.GetTableFieldWithViewID(viewID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to get table fields: "+err.Error())
		return
	}
	response.Success(c, result)
}

// QueryWithViewId returns the jump config for a specific view in a dashboard.
func (h *LinkJumpHandler) QueryWithViewId(c *gin.Context) {
	defer recoverServicePanic(c)
	dvID, ok := parseIDParamMsg(c, "dvId", errInvalidDvID)
	if !ok {
		return
	}
	viewID, ok := parseIDParamMsg(c, "viewId", "Invalid viewId")
	if !ok {
		return
	}
	result, err := h.service.QueryWithViewId(dvID, viewID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to query jump info: "+err.Error())
		return
	}
	response.Success(c, result)
}

// UpdateJumpSet saves a complete jump configuration.
func (h *LinkJumpHandler) UpdateJumpSet(c *gin.Context) {
	defer recoverServicePanic(c)
	var dto service.LinkJumpDTO
	if err := c.ShouldBindBodyWith(&dto, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.UpdateJumpSet(&dto); err != nil {
		response.Error(c, response.CodeInternalError, "Failed to update jump set: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// QueryTargetVisualizationJumpInfo returns sourceInfo→targetInfo mappings for target navigation.
func (h *LinkJumpHandler) QueryTargetVisualizationJumpInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	var req service.LinkJumpRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.QueryTargetVisualizationJumpInfo(&req)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to query target jump info: "+err.Error())
		return
	}
	response.Success(c, result)
}

// QueryVisualizationJumpInfo returns all active jump info for a dashboard.
func (h *LinkJumpHandler) QueryVisualizationJumpInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	dvID, ok := parseIDParamMsg(c, "dvId", errInvalidDvID)
	if !ok {
		return
	}
	resourceTable := c.Param("resourceTable")
	result, err := h.service.QueryVisualizationJumpInfo(dvID, resourceTable)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to query visualization jump info: "+err.Error())
		return
	}
	response.Success(c, result)
}

// ViewTableDetailList returns chart views with field details for a dashboard.
func (h *LinkJumpHandler) ViewTableDetailList(c *gin.Context) {
	defer recoverServicePanic(c)
	dvID, ok := parseIDParamMsg(c, "dvId", errInvalidDvID)
	if !ok {
		return
	}
	result, err := h.service.ViewTableDetailList(dvID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to get view table details: "+err.Error())
		return
	}
	response.Success(c, result)
}

// UpdateJumpSetActive toggles jump_active on a chart.
func (h *LinkJumpHandler) UpdateJumpSetActive(c *gin.Context) {
	defer recoverServicePanic(c)
	var req service.LinkJumpRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.UpdateJumpActive(&req)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to update jump active: "+err.Error())
		return
	}
	response.Success(c, result)
}

// RemoveJumpSet removes all jump data for a view.
func (h *LinkJumpHandler) RemoveJumpSet(c *gin.Context) {
	defer recoverServicePanic(c)
	var dto service.LinkJumpDTO
	if err := c.ShouldBindBodyWith(&dto, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.RemoveJumpSet(&dto); err != nil {
		response.Error(c, response.CodeInternalError, "Failed to remove jump set: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// RegisterLinkJumpRoutes registers all link jump routes on the given router.
func RegisterLinkJumpRoutes(r gin.IRouter, h *LinkJumpHandler) {
	g := r.Group("/linkJump")
	{
		g.GET("/getTableFieldWithViewId/:viewId", h.GetTableFieldWithViewID)
		g.GET("/queryWithViewId/:dvId/:viewId", h.QueryWithViewId)
		g.POST("/updateJumpSet", h.UpdateJumpSet)
		g.POST("/queryTargetVisualizationJumpInfo", h.QueryTargetVisualizationJumpInfo)
		g.GET("/queryVisualizationJumpInfo/:dvId/:resourceTable", h.QueryVisualizationJumpInfo)
		g.GET("/viewTableDetailList/:dvId", h.ViewTableDetailList)
		g.POST("/updateJumpSetActive", h.UpdateJumpSetActive)
		g.POST("/removeJumpSet", h.RemoveJumpSet)
	}
}
