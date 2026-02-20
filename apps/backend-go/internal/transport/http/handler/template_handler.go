package handler

import (
	"dataease/backend/internal/domain/template"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	service *service.TemplateService
}

func NewTemplateHandler(service *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{service: service}
}

func (h *TemplateHandler) Create(c *gin.Context) {
	var req template.TemplateCreateRequest
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

	createBy := ""
	if uid, exists := c.Get("userName"); exists {
		if name, ok := uid.(string); ok {
			createBy = name
		}
	}
	if createBy == "" {
		createBy = strconv.FormatInt(userID, 10)
	}

	result, err := h.service.CreateTemplate(&req, createBy)
	if err != nil {
		response.InternalError(c, "Failed to create template: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *TemplateHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid template ID")
		return
	}

	result, err := h.service.GetTemplate(id)
	if err != nil {
		response.InternalError(c, "Failed to get template: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *TemplateHandler) List(c *gin.Context) {
	var req template.TemplateListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.ListTemplates(&req)
	if err != nil {
		response.InternalError(c, "Failed to list templates: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *TemplateHandler) Update(c *gin.Context) {
	var req template.TemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.UpdateTemplate(&req)
	if err != nil {
		response.InternalError(c, "Failed to update template: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *TemplateHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid template ID")
		return
	}

	if err := h.service.DeleteTemplate(id); err != nil {
		response.InternalError(c, "Failed to delete template: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// Java-compatible stub handlers

// ListCategories returns empty array (stub for Java compatibility)
func (h *TemplateHandler) ListCategories(c *gin.Context) {
	response.Success(c, []interface{}{})
}

// DeleteWithCategory handles delete with optional categoryId param (Java compatibility)
func (h *TemplateHandler) DeleteWithCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid template ID")
		return
	}
	// categoryId is ignored for now, just delete by id
	if err := h.service.DeleteTemplate(id); err != nil {
		response.InternalError(c, "Failed to delete template: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// SearchTemplates handles GET search (alias for List)
func (h *TemplateHandler) SearchTemplates(c *gin.Context) {
	// For GET requests, use query params instead of JSON body
	var req template.TemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.ListTemplates(&req)
	if err != nil {
		response.InternalError(c, "Failed to search templates: "+err.Error())
		return
	}

	response.Success(c, result)
}

func RegisterTemplateRoutes(r gin.IRouter, h *TemplateHandler) {
	// Original Go routes
	group := r.Group("/template")
	{
		group.POST("/create", h.Create)
		group.GET("/get/:id", h.Get)
		group.POST("/list", h.List)
		group.POST("/update", h.Update)
		group.DELETE("/delete/:id", h.Delete)
	}

	// Java-compatible aliases: /templateManage/*
	templateManage := r.Group("/templateManage")
	{
		templateManage.POST("/templateList", h.List)                         // alias for /template/list
		templateManage.POST("/save", h.Create)                               // alias for /template/create
		templateManage.GET("/findOne/:id", h.Get)                            // alias for /template/get/:id
		templateManage.POST("/delete/:id/:categoryId", h.DeleteWithCategory) // Java-compatible delete
		templateManage.POST("/find", h.List)                                 // alias for /template/list
		templateManage.POST("/findCategories", h.ListCategories)             // stub - returns empty array
	}

	// Java-compatible aliases: /templateMarket/*
	templateMarket := r.Group("/templateMarket")
	{
		templateMarket.GET("/search", h.SearchTemplates) // GET search alias
		templateMarket.GET("/searchTemplate", h.SearchTemplates)
		templateMarket.GET("/categories", h.ListCategories) // stub - returns empty array
	}
}
