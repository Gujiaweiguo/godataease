package handler

import (
	"strings"

	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	transportmiddleware "dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type resourceGovernanceService interface {
	BackfillResources(req *service.ResourceGovernanceBackfillRequest) (*service.GovernanceBackfillReport, error)
}

type resourceGovernanceAdminChecker interface {
	IsAdmin(userID int64) bool
}

type ResourceGovernanceHandler struct {
	service      resourceGovernanceService
	adminChecker resourceGovernanceAdminChecker
}

func NewResourceGovernanceHandler(service resourceGovernanceService, adminChecker resourceGovernanceAdminChecker) *ResourceGovernanceHandler {
	return &ResourceGovernanceHandler{service: service, adminChecker: adminChecker}
}

type resourceGovernanceBackfillRequest struct {
	ResourceType string `json:"resourceType" binding:"required"`
	AfterID      int64  `json:"afterId"`
	Limit        int    `json:"limit"`
	OrgID        *int64 `json:"orgId,omitempty"`
}

func (h *ResourceGovernanceHandler) BackfillResources(c *gin.Context) {
	if h.service == nil {
		response.Error(c, "500000", "Failed: resource governance service is unavailable")
		return
	}
	userID := int64(transportmiddleware.GetUserID(c))
	if h.adminChecker != nil && !h.adminChecker.IsAdmin(userID) {
		response.Forbidden(c, "insufficient permissions")
		return
	}

	var req resourceGovernanceBackfillRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	report, err := h.service.BackfillResources(&service.ResourceGovernanceBackfillRequest{
		ResourceType: strings.TrimSpace(req.ResourceType),
		AfterID:      req.AfterID,
		Limit:        req.Limit,
		OrgID:        req.OrgID,
	})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, report)
}

func RegisterResourceGovernanceRoutes(r gin.IRouter, h *ResourceGovernanceHandler) {
	group := r.Group("/system/resource-governance")
	group.POST("/backfill", h.BackfillResources)
}
