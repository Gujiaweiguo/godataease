package handler

import (
	"dataease/backend/internal/domain/msgcenter"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type MsgCenterHandler struct {
	service *service.MsgCenterService
}

func NewMsgCenterHandler(service *service.MsgCenterService) *MsgCenterHandler {
	return &MsgCenterHandler{service: service}
}

func (h *MsgCenterHandler) Count(c *gin.Context) {
	var req msgcenter.CountRequest
	_ = c.ShouldBindJSON(&req)
	response.Success(c, h.service.Count(&req))
}

func (h *MsgCenterHandler) List(c *gin.Context) {
	var req msgcenter.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	response.Success(c, h.service.List(&req))
}

func (h *MsgCenterHandler) Read(c *gin.Context) {
	var req msgcenter.ReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	userID := int64(0)
	if uid, exists := c.Get("userId"); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	response.Success(c, h.service.Read(&req, userID))
}

func (h *MsgCenterHandler) ReadBatch(c *gin.Context) {
	var req msgcenter.ReadBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	userID := int64(0)
	if uid, exists := c.Get("userId"); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	response.Success(c, h.service.ReadBatch(&req, userID))
}

func RegisterMsgCenterRoutes(r gin.IRouter, h *MsgCenterHandler) {
	group := r.Group("/msg-center")
	{
		group.POST("/count", h.Count)
		group.POST("/list", h.List)
		group.POST("/read", h.Read)
		group.POST("/read/batch", h.ReadBatch)
	}
}
