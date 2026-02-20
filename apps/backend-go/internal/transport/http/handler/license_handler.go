package handler

import (
	"errors"
	"io"

	"dataease/backend/internal/domain/license"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type LicenseHandler struct {
	service *service.LicenseService
}

func NewLicenseHandler(service *service.LicenseService) *LicenseHandler {
	return &LicenseHandler{service: service}
}

func (h *LicenseHandler) Validate(c *gin.Context) {
	var req license.LicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Validate(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *LicenseHandler) Update(c *gin.Context) {
	var req license.LicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Update(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *LicenseHandler) Version(c *gin.Context) {
	response.Success(c, h.service.Version())
}

func (h *LicenseHandler) Revert(c *gin.Context) {
	if err := h.service.Revert(); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *LicenseHandler) GetExpiryWarning(c *gin.Context) {
	warning := h.service.GetExpiryWarning()
	response.Success(c, warning)
}

func RegisterLicenseRoutes(r gin.IRouter, h *LicenseHandler) {
	group := r.Group("/license")
	{
		group.POST("/validate", h.Validate)
		group.POST("/update", h.Update)
		group.GET("/version", h.Version)
		group.POST("/revert", h.Revert)
		group.GET("/expiryWarning", h.GetExpiryWarning)
	}
}
