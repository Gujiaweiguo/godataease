package handler

import (
	"strconv"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type MenuHandler struct {
	service *service.MenuService
}

func NewMenuHandler(service *service.MenuService) *MenuHandler {
	return &MenuHandler{service: service}
}

func (h *MenuHandler) Query(c *gin.Context) {
	result, err := h.service.Query()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	applyMenuTitles(result, requestLocale(c, nil))
	response.Success(c, result)
}

func applyMenuTitles(menus []*menu.MenuVO, locale string) {
	for _, current := range menus {
		if current == nil {
			continue
		}
		if current.Meta != nil {
			titleKey := current.Name
			if current.Meta.Title != "" {
				titleKey = current.Meta.Title
			}
			current.Meta.Title = ResolveMenuTitle(titleKey, locale)
		}
		if len(current.Children) > 0 {
			applyMenuTitles(current.Children, locale)
		}
	}
}

type CreateMenuRequest struct {
	Pid          int64                  `json:"pid"`
	Type         int                    `json:"type"`
	Name         string                 `json:"name" binding:"required"`
	Component    string                 `json:"component"`
	MenuSort     int                    `json:"menuSort"`
	Icon         string                 `json:"icon"`
	Path         string                 `json:"path" binding:"required"`
	Hidden       bool                   `json:"hidden"`
	InLayout     bool                   `json:"inLayout"`
	Auth         bool                   `json:"auth"`
	MenuLocation string                 `json:"menuLocation"`
	MenuType     string                 `json:"menuType"`
	ActionConfig map[string]interface{} `json:"actionConfig"`
}

func (h *MenuHandler) Create(c *gin.Context) {
	var req CreateMenuRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	var actionConfig menu.JSON
	if req.ActionConfig != nil {
		actionConfig = menu.JSON(req.ActionConfig)
	}

	m := &menu.CoreMenu{
		Pid:          req.Pid,
		Type:         req.Type,
		Name:         req.Name,
		Component:    req.Component,
		MenuSort:     req.MenuSort,
		Icon:         req.Icon,
		Path:         req.Path,
		Hidden:       req.Hidden,
		InLayout:     req.InLayout,
		Auth:         req.Auth,
		MenuLocation: req.MenuLocation,
		MenuType:     req.MenuType,
		ActionConfig: actionConfig,
	}

	if err := h.service.Create(m); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, m.ID)
}

type UpdateMenuRequest struct {
	ID           int64                  `json:"id" binding:"required"`
	Pid          int64                  `json:"pid"`
	Type         int                    `json:"type"`
	Name         string                 `json:"name" binding:"required"`
	Component    string                 `json:"component"`
	MenuSort     int                    `json:"menuSort"`
	Icon         string                 `json:"icon"`
	Path         string                 `json:"path" binding:"required"`
	Hidden       bool                   `json:"hidden"`
	InLayout     bool                   `json:"inLayout"`
	Auth         bool                   `json:"auth"`
	MenuLocation string                 `json:"menuLocation"`
	MenuType     string                 `json:"menuType"`
	ActionConfig map[string]interface{} `json:"actionConfig"`
}

func (h *MenuHandler) Update(c *gin.Context) {
	var req UpdateMenuRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	var actionConfig menu.JSON
	if req.ActionConfig != nil {
		actionConfig = menu.JSON(req.ActionConfig)
	}

	m := &menu.CoreMenu{
		ID:           req.ID,
		Pid:          req.Pid,
		Type:         req.Type,
		Name:         req.Name,
		Component:    req.Component,
		MenuSort:     req.MenuSort,
		Icon:         req.Icon,
		Path:         req.Path,
		Hidden:       req.Hidden,
		InLayout:     req.InLayout,
		Auth:         req.Auth,
		MenuLocation: req.MenuLocation,
		MenuType:     req.MenuType,
		ActionConfig: actionConfig,
	}

	if err := h.service.Update(m); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid menu ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

type UpdateSortRequest struct {
	ID   int64 `json:"id" binding:"required"`
	Sort int   `json:"sort" binding:"required"`
}

func (h *MenuHandler) UpdateSort(c *gin.Context) {
	var req UpdateSortRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.UpdateSort(req.ID, req.Sort); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

type UpdateHiddenRequest struct {
	ID     int64 `json:"id" binding:"required"`
	Hidden bool  `json:"hidden"`
}

func (h *MenuHandler) UpdateHidden(c *gin.Context) {
	var req UpdateHiddenRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if err := h.service.UpdateHidden(req.ID, req.Hidden); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *MenuHandler) Detail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid menu ID")
		return
	}

	result, err := h.service.GetByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func RegisterMenuReadRoutes(r *gin.RouterGroup, h *MenuHandler) {
	menuGroup := r.Group("/menu")
	{
		menuGroup.GET("/query", h.Query)
		menuGroup.GET("/detail/:id", h.Detail)
	}
}

func RegisterMenuWriteRoutes(r *gin.RouterGroup, h *MenuHandler) {
	menuGroup := r.Group("/menu")
	{
		menuGroup.POST("/create", h.Create)
		menuGroup.POST("/update", h.Update)
		menuGroup.POST("/delete/:id", h.Delete)
		menuGroup.POST("/updateSort", h.UpdateSort)
		menuGroup.POST("/updateHidden", h.UpdateHidden)
	}
}
