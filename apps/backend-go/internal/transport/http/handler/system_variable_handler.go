package handler

import (
	"strconv"

	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type SystemVariableHandler struct {
	service *service.SystemVariableService
}

func NewSystemVariableHandler(service *service.SystemVariableService) *SystemVariableHandler {
	return &SystemVariableHandler{service: service}
}

func (h *SystemVariableHandler) Create(c *gin.Context) {
	defer recoverServicePanic(c)
	var req system.SysVariable
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Create(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemVariableHandler) Edit(c *gin.Context) {
	defer recoverServicePanic(c)
	var req system.SysVariable
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Edit(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemVariableHandler) Detail(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	result, err := h.service.Detail(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemVariableHandler) Delete(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var err error
	if err = h.service.Delete(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SystemVariableHandler) Query(c *gin.Context) {
	defer recoverServicePanic(c)
	var req system.SysVariableQueryRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Query(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemVariableHandler) CreateValue(c *gin.Context) {
	defer recoverServicePanic(c)
	var req system.SysVariableValue
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.CreateValue(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemVariableHandler) EditValue(c *gin.Context) {
	defer recoverServicePanic(c)
	var req system.SysVariableValue
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.EditValue(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemVariableHandler) DeleteValue(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var err error
	if err = h.service.DeleteValue(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SystemVariableHandler) SelectedValues(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	result, err := h.service.SelectedValues(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemVariableHandler) SelectedValuePage(c *gin.Context) {
	defer recoverServicePanic(c)
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		response.Error(c, "500000", "Invalid page")
		return
	}
	size, err := strconv.Atoi(c.Param("limit"))
	if err != nil || size <= 0 {
		response.Error(c, "500000", "Invalid size")
		return
	}
	var req system.SysVariableValueQueryRequest
	if err = c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.SelectedValuePage(page, size, &req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemVariableHandler) BatchDeleteValues(c *gin.Context) {
	defer recoverServicePanic(c)
	var ids []int64
	if err := c.ShouldBindBodyWith(&ids, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.BatchDeleteValues(ids); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func RegisterSystemVariableRoutes(r gin.IRouter, h *SystemVariableHandler) {
	group := r.Group("/sysVariable")
	{
		group.POST("/create", h.Create)
		group.POST("/edit", h.Edit)
		group.GET("/detail/:id", h.Detail)
		group.GET("/delete/:id", h.Delete)
		group.POST("/query", h.Query)
		group.POST("/value/create", h.CreateValue)
		group.POST("/value/edit", h.EditValue)
		group.GET("/value/delete/:id", h.DeleteValue)
		group.GET("/value/selected/:id", h.SelectedValues)
		group.POST("/value/selected/:page/:limit", h.SelectedValuePage)
		group.POST("/value/batchDel", h.BatchDeleteValues)
	}
}
