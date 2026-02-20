package handler

import (
	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type SystemParamHandler struct {
	service *service.SystemParamService
}

func NewSystemParamHandler(service *service.SystemParamService) *SystemParamHandler {
	return &SystemParamHandler{service: service}
}

func (h *SystemParamHandler) QueryBasic(c *gin.Context) {
	result, err := h.service.QueryBasic()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemParamHandler) SaveBasic(c *gin.Context) {
	var req []system.SettingItem
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.SaveBasic(req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SystemParamHandler) QueryOnlineMap(c *gin.Context) {
	result, err := h.service.QueryOnlineMap()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemParamHandler) QueryOnlineMapByType(c *gin.Context) {
	mapType := c.Param("type")
	result, err := h.service.QueryOnlineMapByType(mapType)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemParamHandler) SaveOnlineMap(c *gin.Context) {
	var req system.OnlineMapEditor
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.SaveOnlineMap(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SystemParamHandler) QuerySQLBot(c *gin.Context) {
	result, err := h.service.QuerySQLBot()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemParamHandler) SaveSQLBot(c *gin.Context) {
	var req system.SQLBotConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.SaveSQLBot(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SystemParamHandler) ShareBase(c *gin.Context) {
	result, err := h.service.ShareBase()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemParamHandler) RequestTimeOut(c *gin.Context) {
	result, err := h.service.RequestTimeOut()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemParamHandler) DefaultSettings(c *gin.Context) {
	result, err := h.service.DefaultSettings()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemParamHandler) UI(c *gin.Context) {
	result, err := h.service.UI()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemParamHandler) DefaultLogin(c *gin.Context) {
	result, err := h.service.DefaultLogin()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SystemParamHandler) I18nOptions(c *gin.Context) {
	result, err := h.service.I18nOptions()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterSystemParamRoutes(r gin.IRouter, h *SystemParamHandler) {
	sys := r.Group("/sysParameter")
	{
		sys.GET("/basic/query", h.QueryBasic)
		sys.POST("/basic/save", h.SaveBasic)

		sys.GET("/queryOnlineMap", h.QueryOnlineMap)
		sys.GET("/queryOnlineMap/:type", h.QueryOnlineMapByType)
		sys.POST("/saveOnlineMap", h.SaveOnlineMap)

		sys.GET("/sqlbot", h.QuerySQLBot)
		sys.POST("/sqlbot", h.SaveSQLBot)

		sys.GET("/shareBase", h.ShareBase)

		sys.GET("/requestTimeOut", h.RequestTimeOut)
		sys.GET("/defaultSettings", h.DefaultSettings)
		sys.GET("/ui", h.UI)
		sys.GET("/defaultLogin", h.DefaultLogin)
		sys.GET("/i18nOptions", h.I18nOptions)
	}
}
