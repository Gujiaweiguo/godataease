package handler

import (
	"strconv"

	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type OrgHandler struct {
	orgService *service.OrgService
}

func NewOrgHandler(orgService *service.OrgService) *OrgHandler {
	return &OrgHandler{
		orgService: orgService,
	}
}

func (h *OrgHandler) CreateOrg(c *gin.Context) {
	var req org.OrgCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	err := h.orgService.CreateOrg(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrgHandler) UpdateOrg(c *gin.Context) {
	var req org.OrgUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	err := h.orgService.UpdateOrg(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrgHandler) DeleteOrg(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid organization ID")
		return
	}

	err = h.orgService.DeleteOrg(orgID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrgHandler) ListOrgs(c *gin.Context) {
	orgs, err := h.orgService.ListOrgs()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, orgs)
}

func (h *OrgHandler) GetOrgByID(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid organization ID")
		return
	}

	org, err := h.orgService.GetOrgByID(orgID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, org)
}

func (h *OrgHandler) GetOrgTree(c *gin.Context) {
	tree, err := h.orgService.GetOrgTree()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, tree)
}

func (h *OrgHandler) CheckOrgName(c *gin.Context) {
	orgName := c.Query("orgName")
	if orgName == "" {
		response.Error(c, "500000", "orgName is required")
		return
	}

	exists, err := h.orgService.CheckOrgNameExists(orgName)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"exists": exists})
}

func (h *OrgHandler) UpdateOrgStatus(c *gin.Context) {
	var req org.OrgStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	err := h.orgService.UpdateOrgStatus(req.OrgID, req.Status)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrgHandler) GetChildOrgs(c *gin.Context) {
	parentIDStr := c.Param("parentId")
	parentID, err := strconv.ParseInt(parentIDStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid parent ID")
		return
	}

	children, err := h.orgService.ListByParentID(parentID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, children)
}

func RegisterOrgRoutes(r *gin.RouterGroup, h *OrgHandler) {
	orgGroup := r.Group("/system/organization")
	{
		orgGroup.POST("/create", h.CreateOrg)
		orgGroup.POST("/update", h.UpdateOrg)
		orgGroup.POST("/delete/:orgId", h.DeleteOrg)
		orgGroup.GET("/list", h.ListOrgs)
		orgGroup.GET("/info/:orgId", h.GetOrgByID)
		orgGroup.GET("/tree", h.GetOrgTree)
		orgGroup.GET("/checkName", h.CheckOrgName)
		orgGroup.POST("/updateStatus", h.UpdateOrgStatus)
		orgGroup.GET("/children/:parentId", h.GetChildOrgs)
	}
}
