package handler

import (
	"errors"
	"strconv"

	"dataease/backend/internal/domain/governance"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type RoleHandler struct {
	service             *service.RoleService
	governancePolicySvc *service.GovernancePolicyService
	adminChecker        roleGovernanceAdminChecker
}

type roleGovernanceAdminChecker interface {
	IsAdmin(userID int64) bool
}

type updateLastRolePolicyRequest struct {
	OrgID  int64  `json:"orgId"`
	Policy string `json:"policy" binding:"required"`
}

func NewRoleHandler(service *service.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func (h *RoleHandler) SetGovernancePolicyService(svc *service.GovernancePolicyService) {
	h.governancePolicySvc = svc
}

func (h *RoleHandler) SetAdminChecker(checker roleGovernanceAdminChecker) {
	h.adminChecker = checker
}

func (h *RoleHandler) Query(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.RoleQueryRequest
	if err := shouldBindOptionalJSON(c, &req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.QueryRoles(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"list": result})
}

func (h *RoleHandler) QueryByCurrentOrg(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.RoleQueryRequest
	if err := shouldBindOptionalJSON(c, &req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	orgID := middleware.GetOrgID(c)
	if orgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}

	keyword := ""
	if req.Keyword != nil {
		keyword = *req.Keyword
	}

	result, err := h.service.QueryRolesByOrgID(orgID, keyword)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"list": result})
}

func (h *RoleHandler) Page(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.RolePageRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.QueryRolesPage(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *RoleHandler) Create(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.RoleCreator
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	createBy := h.getCreateBy(c)
	callerOrgID := middleware.GetOrgID(c)
	if callerOrgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}
	id, err := h.service.CreateRole(&req, createBy, callerOrgID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, id)
}

func (h *RoleHandler) Edit(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.RoleEditor
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	updateBy := h.getCreateBy(c)
	callerOrgID := middleware.GetOrgID(c)
	if callerOrgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}
	if err := h.service.EditRole(&req, updateBy, callerOrgID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid role ID")
	if !ok {
		return
	}
	callerOrgID := middleware.GetOrgID(c)
	if callerOrgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}

	if err := h.service.DeleteRole(id, callerOrgID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *RoleHandler) Detail(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid role ID")
	if !ok {
		return
	}
	var err error

	result, err := h.service.GetRoleByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *RoleHandler) QueryWithOrgID(c *gin.Context) {
	defer recoverServicePanic(c)
	oid, ok := parseIDParamMsg(c, "oid", "Invalid org ID")
	if !ok {
		return
	}
	var err error

	keyword := c.Query("keyword")
	result, err := h.service.QueryRolesByOrgID(oid, keyword)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *RoleHandler) getCreateBy(c *gin.Context) string {
	if userId, exists := c.Get("userId"); exists {
		switch v := userId.(type) {
		case string:
			return v
		case int64:
			return strconv.FormatInt(v, 10)
		case int:
			return strconv.Itoa(v)
		}
	}
	return embeddedDefaultUpdateBy
}

// MountUser 绑定用户到角色
func (h *RoleHandler) MountUser(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.MountUserRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if req.OrgId <= 0 {
		req.OrgId = middleware.GetOrgID(c)
	}
	if req.OrgId <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}

	if err := h.service.MountUsers(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// MountExternalUser 绑定组织外用户
func (h *RoleHandler) MountExternalUser(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.MountExternalUserRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	orgID := middleware.GetOrgID(c)
	if orgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}
	if err := h.service.MountExternalUser(&req, orgID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// UnmountUser 解绑用户
func (h *RoleHandler) UnmountUser(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.UnmountUserRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if req.OrgId <= 0 {
		req.OrgId = middleware.GetOrgID(c)
	}
	if req.OrgId <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}

	if err := h.service.UnmountUser(&req); err != nil {
		if errors.Is(err, service.ErrLastRoleRemovalBlocked) {
			response.Error(c, "500000", service.ErrLastRoleRemovalBlocked.Error())
			return
		}
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// BeforeUnmountInfo 解绑前检查
func (h *RoleHandler) BeforeUnmountInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.UnmountUserRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	if req.OrgId <= 0 {
		req.OrgId = middleware.GetOrgID(c)
	}
	if req.OrgId <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}

	count, err := h.service.BeforeUnmountInfo(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, count)
}

func (h *RoleHandler) GetLastRolePolicy(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.governancePolicySvc == nil {
		response.Error(c, "500000", "Failed: governance policy service is unavailable")
		return
	}

	orgID, err := strconv.ParseInt(c.Query("orgId"), 10, 64)
	if err != nil || orgID <= 0 {
		orgID = middleware.GetOrgID(c)
	}
	if orgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}

	policy, err := h.governancePolicySvc.GetLastRolePolicy(orgID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"orgId": orgID, "policy": policy})
}

func (h *RoleHandler) UpdateLastRolePolicy(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.governancePolicySvc == nil {
		response.Error(c, "500000", "Failed: governance policy service is unavailable")
		return
	}
	userID := int64(middleware.GetUserID(c))
	if h.adminChecker != nil && !h.adminChecker.IsAdmin(userID) {
		response.Forbidden(c, "insufficient permissions")
		return
	}

	var req updateLastRolePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if req.OrgID <= 0 {
		req.OrgID = middleware.GetOrgID(c)
	}
	if req.OrgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}

	policy := governance.LastRolePolicy(req.Policy)
	if err := h.governancePolicySvc.SetLastRolePolicy(req.OrgID, policy, userID); err != nil {
		if errors.Is(err, service.ErrInvalidLastRolePolicy) {
			response.Error(c, "500000", "Invalid request: "+err.Error())
			return
		}
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"orgId": req.OrgID, "policy": policy})
}

// SearchExternalUser 搜索组织外用户
func (h *RoleHandler) SearchExternalUser(c *gin.Context) {
	defer recoverServicePanic(c)
	keyword := c.Param("keyword")
	excludeOrgID := middleware.GetOrgID(c)
	if excludeOrgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}

	result, err := h.service.SearchExternalUser(keyword, excludeOrgID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

// OptionForUser 用户可选角色
func (h *RoleHandler) OptionForUser(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.RoleRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	orgID := middleware.GetOrgID(c)
	if orgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return
	}
	result, err := h.service.OptionForUser(&req, orgID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

// SelectedForUser 用户已选角色
func (h *RoleHandler) SelectedForUser(c *gin.Context) {
	defer recoverServicePanic(c)
	var req role.RoleRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.SelectedForUser(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func RegisterRoleRoutes(r *gin.RouterGroup, h *RoleHandler) {
	roleGroup := r.Group("/role")
	{
		roleGroup.POST("/query", h.Query)
		roleGroup.POST("/page", h.Page)
		roleGroup.POST("/byCurOrg", h.QueryByCurrentOrg)
		roleGroup.GET("/queryWithOid/:oid", h.QueryWithOrgID)
		roleGroup.POST("/create", h.Create)
		roleGroup.POST("/edit", h.Edit)
		roleGroup.POST("/delete/:id", h.Delete)
		roleGroup.POST("/mountUser", h.MountUser)
		roleGroup.POST("/mountExternalUser", h.MountExternalUser)
		roleGroup.POST("/unMountUser", h.UnmountUser)
		roleGroup.POST("/beforeUnmountInfo", h.BeforeUnmountInfo)
		roleGroup.GET("/searchExternalUser/:keyword", h.SearchExternalUser)
		roleGroup.POST("/user/option", h.OptionForUser)
		roleGroup.POST("/user/selected", h.SelectedForUser)
	}

	governanceGroup := r.Group("/governance")
	{
		governanceGroup.GET("/last-role-policy", h.GetLastRolePolicy)
		governanceGroup.PUT("/last-role-policy", h.UpdateLastRolePolicy)
	}

	// System role routes - frontend compatibility aliases
	systemRoleGroup := r.Group("/system/role")
	{
		systemRoleGroup.POST("/create", h.Create)
		systemRoleGroup.POST("/update", h.Edit) // Map to edit
		systemRoleGroup.POST("/delete/:id", h.Delete)
	}
}
