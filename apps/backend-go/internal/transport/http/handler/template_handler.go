package handler

import (
	"dataease/backend/internal/domain/template"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type TemplateHandler struct {
	service *service.TemplateService
}

func NewTemplateHandler(service *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{service: service}
}

func (h *TemplateHandler) Create(c *gin.Context) {
	var req template.TemplateCreateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	createBy := templateCreateBy(c)

	result, err := h.service.CreateTemplate(&req, createBy)
	if err != nil {
		response.InternalError(c, "Failed to create template: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *TemplateHandler) Save(c *gin.Context) {
	var req template.TemplateSaveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	createBy := templateCreateBy(c)

	result, err := h.service.SaveTemplate(&req, createBy)
	if err != nil {
		response.InternalError(c, "Failed to save template: "+err.Error())
		return
	}

	response.Success(c, result)
}

func templateCreateBy(c *gin.Context) string {
	userID := int64(0)
	if uid, exists := c.Get("userId"); exists {
		if id, ok := uid.(int64); ok {
			userID = id
		}
	}

	if uid, exists := c.Get("userName"); exists {
		if name, ok := uid.(string); ok && name != "" {
			return name
		}
	}

	return strconv.FormatInt(userID, 10)
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
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
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
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
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

func (h *TemplateHandler) ListCategories(c *gin.Context) {
	if h.service == nil {
		response.Success(c, []interface{}{})
		return
	}

	var req struct {
		Level        string `json:"level" form:"level"`
		TemplateType string `json:"templateType" form:"templateType"`
	}
	var err error
	if c.Request.Method == http.MethodGet {
		err = c.ShouldBindQuery(&req)
	} else {
		err = c.ShouldBindBodyWith(&req, binding.JSON)
	}
	if err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.ListCategories(req.Level, req.TemplateType)
	if err != nil {
		response.InternalError(c, "Failed to list categories: "+err.Error())
		return
	}

	response.Success(c, result)
}

// DeleteWithCategory handles delete scoped to a category when categoryId is present (Java compatibility)
func (h *TemplateHandler) DeleteWithCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid template ID")
		return
	}
	categoryID := c.Param("categoryId")
	if err := h.service.DeleteWithCategory(id, categoryID); err != nil {
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

func (h *TemplateHandler) SearchTemplateMarket(c *gin.Context) {
	if h.service == nil {
		response.Success(c, map[string]interface{}{
			"baseUrl":    "",
			"contents":   []interface{}{},
			"categories": []interface{}{},
		})
		return
	}

	response.Success(c, h.buildTemplateMarketResponse(false))
}

func (h *TemplateHandler) SearchTemplateMarketRecommend(c *gin.Context) {
	if h.service == nil {
		response.Success(c, map[string]interface{}{
			"baseUrl":    "",
			"contents":   []interface{}{},
			"categories": []interface{}{},
		})
		return
	}

	response.Success(c, h.buildTemplateMarketResponse(false))
}

func (h *TemplateHandler) SearchTemplateMarketPreview(c *gin.Context) {
	if h.service == nil {
		response.Success(c, map[string]interface{}{
			"baseUrl":    "",
			"contents":   []interface{}{},
			"categories": []interface{}{},
		})
		return
	}

	response.Success(c, h.buildTemplateMarketResponse(true))
}

// DeleteCategory deletes a template category by ID
func (h *TemplateHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	deleted, err := h.service.DeleteCategory(idStr)
	if err != nil {
		response.InternalError(c, "Failed to delete category: "+err.Error())
		return
	}
	if deleted {
		response.Success(c, "success")
		return
	}
	response.Success(c, "failed")
}

// NameCheck checks if a template name already exists
func (h *TemplateHandler) NameCheck(c *gin.Context) {
	var req struct {
		Name    string `json:"name"`
		ID      string `json:"id"`
		OptType string `json:"optType"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.NameCheck(req.OptType, req.Name, req.ID)
	if err != nil {
		response.InternalError(c, "Failed to check template name: "+err.Error())
		return
	}
	response.Success(c, result)
}

// CategoryTemplateNameCheck checks category-template name uniqueness
func (h *TemplateHandler) CategoryTemplateNameCheck(c *gin.Context) {
	var req struct {
		Name          string   `json:"name"`
		CategoryID    string   `json:"categoryId"`
		Categories    []string `json:"categories"`
		TemplateNames []string `json:"templateNames"`
		TemplateArray []string `json:"templateArray"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	categories := req.Categories
	if len(categories) == 0 && req.CategoryID != "" {
		categories = []string{req.CategoryID}
	}
	result, err := h.service.CategoryTemplateNameCheck(req.Name, categories, req.TemplateNames, req.TemplateArray)
	if err != nil {
		response.InternalError(c, "Failed to check category template name: "+err.Error())
		return
	}
	response.Success(c, result)
}

// BatchDelete deletes multiple templates
func (h *TemplateHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs         []int64  `json:"ids"`
		TemplateIDs []string `json:"templateIds"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	deleteIDs := append([]int64{}, req.IDs...)
	for _, rawID := range req.TemplateIDs {
		if parsedID, err := strconv.ParseInt(rawID, 10, 64); err == nil && parsedID > 0 {
			deleteIDs = append(deleteIDs, parsedID)
		}
	}
	for _, id := range deleteIDs {
		if err := h.service.DeleteTemplate(id); err != nil {
			response.InternalError(c, "Failed to batch delete templates: "+err.Error())
			return
		}
	}
	response.Success(c, nil)
}

// BatchUpdate updates multiple templates (e.g., move category)
func (h *TemplateHandler) BatchUpdate(c *gin.Context) {
	var req struct {
		TemplateIDs []string `json:"templateIds"`
		Categories  []string `json:"categories"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.BatchUpdateCategories(req.TemplateIDs, req.Categories); err != nil {
		response.InternalError(c, "Failed to batch update templates: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TemplateHandler) FindCategoriesByTemplateIds(c *gin.Context) {
	var req struct {
		TemplateArray []string `json:"templateArray"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.FindCategoriesByTemplateIDs(req.TemplateArray)
	if err != nil {
		response.InternalError(c, "Failed to find template categories: "+err.Error())
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
		templateManage.POST("/save", h.Save)                                 // Java-compatible save (create/update)
		templateManage.GET("/findOne/:id", h.Get)                            // alias for /template/get/:id
		templateManage.POST("/delete/:id/:categoryId", h.DeleteWithCategory) // Java-compatible delete
		templateManage.POST("/find", h.List)                                 // alias for /template/list
		templateManage.POST("/findCategories", h.ListCategories)             // stub - returns empty array
		templateManage.POST("/deleteCategory/:id", h.DeleteCategory)
		templateManage.POST("/nameCheck", h.NameCheck)
		templateManage.POST("/categoryTemplateNameCheck", h.CategoryTemplateNameCheck)
		templateManage.POST("/batchDelete", h.BatchDelete)
		templateManage.POST("/batchUpdate", h.BatchUpdate)
		templateManage.POST("/findCategoriesByTemplateIds", h.FindCategoriesByTemplateIds)
	}

	// Java-compatible aliases: /templateMarket/*
	templateMarket := r.Group("/templateMarket")
	{
		templateMarket.GET("/search", h.SearchTemplateMarket)
		templateMarket.GET("/searchTemplate", h.SearchTemplateMarket)
		templateMarket.GET("/searchRecommend", h.SearchTemplateMarketRecommend)
		templateMarket.GET("/searchPreview", h.SearchTemplateMarketPreview)
		templateMarket.GET("/categories", h.ListCategories) // stub - returns empty array
		templateMarket.GET("/categoriesObject", h.ListCategories)
	}
}

func (h *TemplateHandler) buildTemplateMarketResponse(groupByCategory bool) map[string]interface{} {
	categories, err := h.service.ListCategories("0", "self")
	if err != nil || len(categories) == 0 {
		return map[string]interface{}{
			"baseUrl":    "",
			"contents":   []interface{}{},
			"categories": []interface{}{},
		}
	}

	categoryItems := make([]map[string]interface{}, 0, len(categories))
	flatContents := make([]map[string]interface{}, 0)
	groupedContents := make([]map[string]interface{}, 0, len(categories))
	for _, category := range categories {
		categoryItems = append(categoryItems, map[string]interface{}{
			"id":     category.ID,
			"label":  category.Name,
			"source": "manage",
		})

		templates, listErr := h.service.ListTemplates(&template.TemplateListRequest{CategoryID: strconv.FormatInt(category.ID, 10)})
		if listErr != nil {
			continue
		}

		categoryGroup := make([]map[string]interface{}, 0, len(templates.List))
		for _, item := range templates.List {
			marketItem := map[string]interface{}{
				"id":            item.ID,
				"title":         item.Name,
				"name":          item.Name,
				"thumbnail":     item.Snapshot,
				"templateType":  marketTemplateType(item.DvType),
				"source":        "manage",
				"classify":      "template",
				"categoryNames": []string{category.Name},
				"metas":         map[string]interface{}{},
			}
			flatContents = append(flatContents, marketItem)
			categoryGroup = append(categoryGroup, marketItem)
		}

		if groupByCategory {
			groupedContents = append(groupedContents, map[string]interface{}{
				"category": map[string]interface{}{
					"id":     category.ID,
					"label":  category.Name,
					"source": "manage",
				},
				"contents": categoryGroup,
			})
		}
	}

	contents := make([]interface{}, 0)
	if groupByCategory {
		for _, item := range groupedContents {
			contents = append(contents, item)
		}
	} else {
		for _, item := range flatContents {
			contents = append(contents, item)
		}
	}

	categoryResult := make([]interface{}, 0, len(categoryItems))
	for _, item := range categoryItems {
		categoryResult = append(categoryResult, item)
	}

	return map[string]interface{}{
		"baseUrl":    "",
		"contents":   contents,
		"categories": categoryResult,
	}
}

func marketTemplateType(dvType string) string {
	if dvType == "dataV" {
		return "SCREEN"
	}
	return "PANEL"
}
