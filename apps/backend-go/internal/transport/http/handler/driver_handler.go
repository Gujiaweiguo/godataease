package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type DriverHandler struct {
	service *service.DriverService
}

func NewDriverHandler(service *service.DriverService) *DriverHandler {
	return &DriverHandler{service: service}
}

func (h *DriverHandler) List(c *gin.Context) {
	defer recoverServicePanic(c)
	result, err := h.service.List()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DriverHandler) ListByType(c *gin.Context) {
	defer recoverServicePanic(c)
	dsType := c.Param("dsType")
	result, err := h.service.ListByType(dsType)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DriverHandler) GetByID(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid id")
	if !ok {
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
	defer recoverServicePanic(c)
	driverID, ok := parseIDParamMsg(c, "driverId", "Invalid driver id")
	if !ok {
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
