package handler

import (
	"dataease/backend/internal/domain/share"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ShareHandler struct {
	service *service.ShareService
}

func NewShareHandler(service *service.ShareService) *ShareHandler {
	return &ShareHandler{service: service}
}

func (h *ShareHandler) Create(c *gin.Context) {
	var req share.ShareCreateRequest
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

	result, err := h.service.CreateShare(&req, userID)
	if err != nil {
		response.InternalError(c, "Failed to create share: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ShareHandler) Validate(c *gin.Context) {
	var req share.ShareValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.ValidateShare(&req)
	if err != nil {
		response.InternalError(c, "Failed to validate share: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ShareHandler) Revoke(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid share ID")
		return
	}

	userID := int64(0)
	if uid, exists := c.Get("userId"); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	result, err := h.service.RevokeShare(id, userID)
	if err != nil {
		response.InternalError(c, "Failed to revoke share: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ShareHandler) Status(c *gin.Context) {
	resourceIDStr := c.Param("resourceId")
	resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid resource ID")
		return
	}

	detail, err := h.service.GetDetail(resourceID)
	if err != nil {
		response.InternalError(c, "Failed to get share status: "+err.Error())
		return
	}

	response.Success(c, detail != nil)
}

func (h *ShareHandler) Detail(c *gin.Context) {
	resourceIDStr := c.Param("resourceId")
	resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid resource ID")
		return
	}

	detail, err := h.service.GetDetail(resourceID)
	if err != nil {
		response.InternalError(c, "Failed to get share detail: "+err.Error())
		return
	}

	response.Success(c, detail)
}

func (h *ShareHandler) Switcher(c *gin.Context) {
	resourceIDStr := c.Param("resourceId")
	resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid resource ID")
		return
	}

	userID := int64(0)
	if uid, exists := c.Get("userId"); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	if err := h.service.SwitchStatus(resourceID, userID); err != nil {
		response.InternalError(c, "Failed to switch share status: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ShareHandler) EditUUID(c *gin.Context) {
	var req share.ShareEditUUIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	msg, err := h.service.EditUUID(&req, shareUserID(c))
	if err != nil {
		response.InternalError(c, "Failed to edit share uuid: "+err.Error())
		return
	}

	response.Success(c, msg)
}

func (h *ShareHandler) EditExp(c *gin.Context) {
	var req share.ShareEditExpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.EditExp(&req, shareUserID(c)); err != nil {
		response.InternalError(c, "Failed to edit share expiration: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ShareHandler) EditPwd(c *gin.Context) {
	var req share.ShareEditPwdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.EditPwd(&req, shareUserID(c)); err != nil {
		response.InternalError(c, "Failed to edit share password: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ShareHandler) CreateTicket(c *gin.Context) {
	var req share.TicketCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.CreateTicket(&req)
	if err != nil {
		response.InternalError(c, "Failed to create ticket: "+err.Error())
		return
	}

	response.Success(c, result.Ticket)
}

func (h *ShareHandler) ValidateTicket(c *gin.Context) {
	var req share.TicketValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.ValidateTicket(&req)
	if err != nil {
		response.InternalError(c, "Failed to validate ticket: "+err.Error())
		return
	}

	response.Success(c, result)
}

func RegisterShareRoutes(r gin.IRouter, h *ShareHandler) {
	group := r.Group("/share")
	{
		group.POST("/create", h.Create)
		group.POST("/validate", h.Validate)
		group.DELETE("/revoke/:id", h.Revoke)
		group.GET("/status/:resourceId", h.Status)
		group.GET("/detail/:resourceId", h.Detail)
		group.POST("/switcher/:resourceId", h.Switcher)
		group.POST("/editUuid", h.EditUUID)
		group.POST("/editExp", h.EditExp)
		group.POST("/editPwd", h.EditPwd)
		group.POST("/ticket/create", h.CreateTicket)
		group.POST("/ticket/validate", h.ValidateTicket)
	}
}

func shareUserID(c *gin.Context) int64 {
	if uid, exists := c.Get("userId"); exists {
		if id, ok := uid.(int64); ok {
			return id
		}
		if id, ok := uid.(uint64); ok {
			return int64(id)
		}
	}
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(uint64); ok {
			return int64(id)
		}
		if id, ok := uid.(int64); ok {
			return id
		}
	}
	return 0
}
