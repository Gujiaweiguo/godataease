package handler

import (
	"dataease/backend/internal/domain/ticket"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type TicketHandler struct {
	service *service.TicketService
}

func NewTicketHandler(service *service.TicketService) *TicketHandler {
	return &TicketHandler{service: service}
}

func (h *TicketHandler) Create(c *gin.Context) {
	defer recoverServicePanic(c)
	var req ticket.TicketCreateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.CreateTicket(&req)
	if err != nil {
		response.Error(c, "500000", "Failed to create ticket: "+err.Error())
		return
	}

	response.Success(c, result.Ticket)
}

func (h *TicketHandler) Validate(c *gin.Context) {
	defer recoverServicePanic(c)
	ticketStr := c.Param("id")
	if ticketStr == "" {
		response.Error(c, "500000", "Ticket ID is required")
		return
	}

	result := h.service.ValidateTicket(ticketStr)
	response.Success(c, result)
}

func (h *TicketHandler) Delete(c *gin.Context) {
	defer recoverServicePanic(c)
	ticketStr := c.Param("id")
	if ticketStr == "" {
		response.Error(c, "500000", "Ticket ID is required")
		return
	}

	req := &ticket.TicketDeleteRequest{Ticket: ticketStr}
	if err := h.service.DeleteTicket(req); err != nil {
		response.Error(c, "500000", "Failed to delete ticket: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) List(c *gin.Context) {
	defer recoverServicePanic(c)
	uuid := c.Param("uuid")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	result := h.service.ListTickets(uuid, page, pageSize)
	response.Success(c, result)
}

func (h *TicketHandler) TempTicket(c *gin.Context) {
	defer recoverServicePanic(c)
	ticketStr := h.service.TempTicket()
	response.Success(c, ticketStr)
}

func RegisterTicketRoutes(r gin.IRouter, h *TicketHandler) {
	group := r.Group("/ticket")
	{
		group.POST("/create", h.Create)
		group.GET("/validate/:id", h.Validate)
		group.DELETE("/delete/:id", h.Delete)
		group.GET("/list/:uuid", h.List)
		group.GET("/temp", h.TempTicket)
	}
}
