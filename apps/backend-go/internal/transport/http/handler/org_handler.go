package handler

import (
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type transferUserRequest struct {
	SourceOrgID int64 `json:"sourceOrgId" binding:"required"`
	TargetOrgID int64 `json:"targetOrgId" binding:"required"`
	UserID      int64 `json:"userId" binding:"required"`
}

type OrgHandler struct {
	orgService *service.OrgService
}

func NewOrgHandler(orgService *service.OrgService) *OrgHandler {
	return &OrgHandler{
		orgService: orgService,
	}
}

func (h *OrgHandler) CreateOrg(c *gin.Context) {
	defer recoverServicePanic(c)
	var req org.OrgCreateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	callerOrgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}

	err := h.orgService.CreateOrg(&req, callerOrgID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrgHandler) UpdateOrg(c *gin.Context) {
	defer recoverServicePanic(c)
	var req org.OrgUpdateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	callerOrgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}

	err := h.orgService.UpdateOrg(&req, callerOrgID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrgHandler) DeleteOrg(c *gin.Context) {
	defer recoverServicePanic(c)
	orgID, ok := parseIDParamMsg(c, "orgId", errInvalidOrganizationID)
	if !ok {
		return
	}
	callerOrgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}

	// 获取操作者信息
	operatorID := int64(0)
	operatorName := embeddedDefaultUpdateBy
	if userIDValue, exists := c.Get("user_id"); exists {
		switch v := userIDValue.(type) {
		case uint64:
			operatorID = int64(v)
		case int64:
			operatorID = v
		case int:
			operatorID = int64(v)
		}
	} else if userId, exists := c.Get(middleware.ContextUserID); exists {
		switch v := userId.(type) {
		case int64:
			operatorID = v
		case int:
			operatorID = int64(v)
		}
	}
	if username, exists := c.Get("username"); exists {
		if u, ok := username.(string); ok {
			operatorName = u
		}
	}

	// 获取 IP 地址
	ipAddress := c.ClientIP()

	err := h.orgService.DeleteOrg(orgID, callerOrgID, operatorID, operatorName, ipAddress)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrgHandler) ListOrgs(c *gin.Context) {
	defer recoverServicePanic(c)
	orgs, err := h.orgService.ListOrgs()
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, orgs)
}

func (h *OrgHandler) GetOrgByID(c *gin.Context) {
	defer recoverServicePanic(c)
	orgID, ok := parseIDParamMsg(c, "orgId", errInvalidOrganizationID)
	if !ok {
		return
	}

	org, err := h.orgService.GetOrgByID(orgID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, org)
}

func (h *OrgHandler) GetOrgTree(c *gin.Context) {
	defer recoverServicePanic(c)
	tree, err := h.orgService.GetOrgTree()
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, tree)
}

func (h *OrgHandler) CheckOrgName(c *gin.Context) {
	defer recoverServicePanic(c)
	orgName := c.Query("orgName")
	if orgName == "" {
		response.Error(c, response.CodeInternalError, "orgName is required")
		return
	}

	exists, err := h.orgService.CheckOrgNameExists(orgName)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"exists": exists})
}

func (h *OrgHandler) UpdateOrgStatus(c *gin.Context) {
	defer recoverServicePanic(c)
	var req org.OrgStatusRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	callerOrgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}

	err := h.orgService.UpdateOrgStatus(req.OrgID, req.Status, callerOrgID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrgHandler) TransferUser(c *gin.Context) {
	defer recoverServicePanic(c)
	var req transferUserRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}

	actorID := int64(middleware.GetUserID(c))
	if err := h.orgService.TransferUserOrg(req.SourceOrgID, req.TargetOrgID, req.UserID, actorID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *OrgHandler) GetChildOrgs(c *gin.Context) {
	defer recoverServicePanic(c)
	parentID, ok := parseIDParamMsg(c, "parentId", "Invalid parent ID")
	if !ok {
		return
	}

	children, err := h.orgService.ListByParentID(parentID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
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
		orgGroup.POST("/transfer-user", h.TransferUser)
		orgGroup.GET("/children/:parentId", h.GetChildOrgs)
	}

	r.POST("/org/mounted", h.ListOrgs)
	r.POST("/organization/transfer-user", h.TransferUser)
}
