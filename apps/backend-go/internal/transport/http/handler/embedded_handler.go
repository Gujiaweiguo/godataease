package handler

import (
	"strconv"

	"dataease/backend/internal/domain/embedded"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type EmbeddedHandler struct {
	service *service.EmbeddedService
}

func NewEmbeddedHandler(service *service.EmbeddedService) *EmbeddedHandler {
	return &EmbeddedHandler{service: service}
}

func (h *EmbeddedHandler) QueryGrid(c *gin.Context) {
	goPage, _ := strconv.Atoi(c.Param("goPage"))
	pageSize, _ := strconv.Atoi(c.Param("pageSize"))
	if goPage < 1 {
		goPage = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	var req embedded.KeywordRequest
	_ = c.ShouldBindJSON(&req)

	keyword := ""
	if req.Keyword != nil {
		keyword = *req.Keyword
	}

	result, err := h.service.QueryGrid(keyword, goPage, pageSize)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *EmbeddedHandler) Create(c *gin.Context) {
	var req embedded.EmbeddedCreator
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	updateBy := h.getUpdateBy(c)
	id, err := h.service.Create(&req, updateBy)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, id)
}

func (h *EmbeddedHandler) Edit(c *gin.Context) {
	var req embedded.EmbeddedEditor
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	updateBy := h.getUpdateBy(c)
	if err := h.service.Edit(&req, updateBy); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *EmbeddedHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *EmbeddedHandler) BatchDelete(c *gin.Context) {
	var req struct {
		Ids []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if len(req.Ids) == 0 {
		response.Error(c, "500000", "IDs list cannot be empty")
		return
	}

	if err := h.service.BatchDelete(req.Ids); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *EmbeddedHandler) Reset(c *gin.Context) {
	var req embedded.EmbeddedResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	updateBy := h.getUpdateBy(c)
	if err := h.service.ResetSecret(&req, updateBy); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *EmbeddedHandler) DomainList(c *gin.Context) {
	domains, err := h.service.GetDomainList()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, domains)
}

func (h *EmbeddedHandler) InitIframe(c *gin.Context) {
	var req embedded.EmbeddedOrigin
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	domains, err := h.service.InitIframe(req.Token, req.Origin)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, domains)
}

func (h *EmbeddedHandler) GetTokenArgs(c *gin.Context) {
	userId, orgId := h.getCurrentUser(c)
	result := h.service.GetTokenArgs(userId, orgId)
	response.Success(c, result)
}

func (h *EmbeddedHandler) GetLimitCount(c *gin.Context) {
	count := h.service.GetLimitCount()
	response.Success(c, count)
}

func (h *EmbeddedHandler) getUpdateBy(c *gin.Context) string {
	if userId, exists := c.Get("userId"); exists {
		return toString(userId)
	}
	return "system"
}

func (h *EmbeddedHandler) getCurrentUser(c *gin.Context) (int64, int64) {
	userId := int64(1)
	orgId := int64(1)
	if uid, exists := c.Get("userId"); exists {
		userId = toInt64(uid)
	}
	if oid, exists := c.Get("orgId"); exists {
		orgId = toInt64(oid)
	}
	return userId, orgId
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int64:
		return strconv.FormatInt(val, 10)
	case int:
		return strconv.Itoa(val)
	default:
		return ""
	}
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	default:
		return 0
	}
}

func RegisterEmbeddedRoutes(r *gin.RouterGroup, h *EmbeddedHandler) {
	embeddedGroup := r.Group("/embedded")
	{
		embeddedGroup.POST("/pager/:goPage/:pageSize", h.QueryGrid)
		embeddedGroup.POST("/create", h.Create)
		embeddedGroup.POST("/edit", h.Edit)
		embeddedGroup.POST("/delete/:id", h.Delete)
		embeddedGroup.POST("/batchDelete", h.BatchDelete)
		embeddedGroup.POST("/reset", h.Reset)
		embeddedGroup.GET("/domainList", h.DomainList)
		embeddedGroup.POST("/initIframe", h.InitIframe)
		embeddedGroup.GET("/getTokenArgs", h.GetTokenArgs)
		embeddedGroup.GET("/limitCount", h.GetLimitCount)
	}
}
