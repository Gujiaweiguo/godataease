package handler

import (
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type PdfTemplateHandler struct{}

func NewPdfTemplateHandler() *PdfTemplateHandler {
	return &PdfTemplateHandler{}
}

func (h *PdfTemplateHandler) QueryAll(c *gin.Context) {
	defer recoverServicePanic(c)
	response.Success(c, []any{})
}

func RegisterPdfTemplateRoutes(r *gin.RouterGroup, h *PdfTemplateHandler) {
	pdfGroup := r.Group("/pdf-template")
	{
		pdfGroup.GET("/queryAll", h.QueryAll)
	}
}
