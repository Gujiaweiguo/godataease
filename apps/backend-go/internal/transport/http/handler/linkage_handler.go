package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type LinkageHandler struct {
	service *service.LinkageService
}

func NewLinkageHandler(service *service.LinkageService) *LinkageHandler {
	return &LinkageHandler{service: service}
}

func (h *LinkageHandler) GetViewLinkageGather(c *gin.Context) {
	defer recoverServicePanic(c)
	var req service.LinkageRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.GetViewLinkageGather(&req)
	if err != nil {
		response.Error(c, "500000", "Failed to get linkage gather: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *LinkageHandler) GetViewLinkageGatherArray(c *gin.Context) {
	defer recoverServicePanic(c)
	var req service.LinkageRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.GetViewLinkageGatherArray(&req)
	if err != nil {
		response.Error(c, "500000", "Failed to get linkage gather: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *LinkageHandler) SaveLinkage(c *gin.Context) {
	defer recoverServicePanic(c)
	var req service.LinkageRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.SaveLinkage(&req); err != nil {
		response.Error(c, "500000", "Failed to save linkage: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *LinkageHandler) GetVisualizationAllLinkageInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	dvID, ok := parseIDParamMsg(c, "dvId", "Invalid dvId")
	if !ok {
		return
	}
	resourceTable := c.Param("resourceTable")
	result, err := h.service.GetVisualizationAllLinkageInfo(dvID, resourceTable)
	if err != nil {
		response.Error(c, "500000", "Failed to get linkage info: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *LinkageHandler) UpdateLinkageActive(c *gin.Context) {
	defer recoverServicePanic(c)
	var req service.LinkageRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.UpdateLinkageActive(&req)
	if err != nil {
		response.Error(c, "500000", "Failed to update linkage active: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *LinkageHandler) RemoveLinkage(c *gin.Context) {
	defer recoverServicePanic(c)
	var req service.LinkageRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.RemoveLinkage(&req); err != nil {
		response.Error(c, "500000", "Failed to remove linkage: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func RegisterLinkageRoutes(r gin.IRouter, h *LinkageHandler) {
	g := r.Group("/linkage")
	{
		g.POST("/getViewLinkageGather", h.GetViewLinkageGather)
		g.POST("/getViewLinkageGatherArray", h.GetViewLinkageGatherArray)
		g.POST("/saveLinkage", h.SaveLinkage)
		g.GET("/getVisualizationAllLinkageInfo/:dvId/:resourceTable", h.GetVisualizationAllLinkageInfo)
		g.POST("/updateLinkageActive", h.UpdateLinkageActive)
		g.POST("/removeLinkage", h.RemoveLinkage)
	}
}
