package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DriverHandler struct {
	service *service.DriverService
}

func NewDriverHandler(service *service.DriverService) *DriverHandler {
	return &DriverHandler{service: service}
}

func (h *DriverHandler) List(c *gin.Context) {
	result, err := h.service.List()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DriverHandler) ListByType(c *gin.Context) {
	dsType := c.Param("dsType")
	result, err := h.service.ListByType(dsType)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DriverHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid id")
		return
	}
	result, err := h.service.GetByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DriverHandler) ListDriverJars(c *gin.Context) {
	driverIDStr := c.Param("driverId")
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid driver id")
		return
	}
	result, err := h.service.ListDriverJars(driverID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterDriverRoutes(r *gin.RouterGroup, h *DriverHandler) {
	dg := r.Group("/driver")
	{
		dg.GET("/list", h.List)
		dg.GET("/list/:dsType", h.ListByType)
		dg.GET("/get/:id", h.GetByID)
		dg.GET("/listDriverJar/:driverId", h.ListDriverJars)
	}
}
