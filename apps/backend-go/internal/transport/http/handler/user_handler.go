package handler

import (
	"net/url"
	"strconv"

	"dataease/backend/internal/domain/audit"
	domainauth "dataease/backend/internal/domain/auth"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService       *service.UserService
	userImportService *service.UserImportService
	loadUserByID      userByIDLoader
	buildBootstrap    identityBootstrapBuilder
	switchOrg         orgSwitcher
}

type identityBootstrapBuilder func(userID int64, selectedOrgID int64, requestLanguage string) (*domainauth.IdentityBootstrap, error)
type orgSwitcher func(userID int64, targetOrgID int64, requestLanguage string) (*domainauth.TokenVO, error)

func requireCurrentOrg(c *gin.Context) (int64, bool) {
	orgID := middleware.GetOrgID(c)
	if orgID <= 0 {
		response.Error(c, "500000", "Invalid org context")
		return 0, false
	}
	return orgID, true
}

func int64Ptr(v int64) *int64 {
	return &v
}

func NewUserHandler(userService *service.UserService, userImportService *service.UserImportService) *UserHandler {
	var loadUserByID userByIDLoader
	if userService != nil {
		loadUserByID = userService.GetUserByID
	}

	return &UserHandler{
		userService:       userService,
		userImportService: userImportService,
		loadUserByID:      loadUserByID,
	}
}

func (h *UserHandler) SetAuthService(authService *service.AuthService) {
	if authService == nil {
		h.buildBootstrap = nil
		h.switchOrg = nil
		return
	}
	h.buildBootstrap = authService.BuildIdentityBootstrapForOrg
	h.switchOrg = authService.SwitchOrg
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	var req user.UserQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if req.OrgID == nil {
		req.OrgID = int64Ptr(orgID)
	}

	result, err := h.userService.SearchUsers(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req user.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if req.OrgID == nil && req.OrganizationID == nil {
		req.OrgID = int64Ptr(orgID)
	}

	id, err := h.userService.CreateUser(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, id)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req user.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if err := h.userService.EnsureUserInOrg(req.ID, orgID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	if err := h.userService.UpdateUser(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid user ID")
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if err := h.userService.EnsureUserInOrg(id, orgID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	if err := h.userService.DeleteUser(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) GetUserOptions(c *gin.Context) {
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	req := &user.UserQueryRequest{Current: 1, Size: 1000, OrgID: int64Ptr(orgID)}

	result, err := h.userService.SearchUsers(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result.List)
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID := int64(middleware.GetUserID(c))
	if userID <= 0 {
		response.Unauthorized(c, "invalid user context")
		return
	}
	if h.buildBootstrap == nil {
		response.Error(c, "500000", "identity bootstrap resolver is not configured")
		return
	}

	bootstrap, err := h.buildBootstrap(userID, middleware.GetOrgID(c), c.GetHeader("Accept-Language"))
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, bootstrap)
}

func (h *UserHandler) SwitchOrg(c *gin.Context) {
	if h.switchOrg == nil {
		response.Error(c, "500000", "org switcher is not configured")
		return
	}

	targetOrgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetOrgID <= 0 {
		response.Error(c, "500000", "Invalid organization ID")
		return
	}

	userID := int64(middleware.GetUserID(c))
	if userID <= 0 {
		response.Unauthorized(c, "invalid user context")
		return
	}

	tokenVO, err := h.switchOrg(userID, targetOrgID, c.GetHeader("Accept-Language"))
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, tokenVO)
}

func (h *UserHandler) SwitchLanguage(c *gin.Context) {
	var req user.LangSwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	userID := int64(middleware.GetUserID(c))
	if userID == 0 {
		response.Error(c, "500000", "Invalid user ID")
		return
	}

	if err := h.userService.SwitchLanguage(userID, req.Lang); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) DownloadExcelTemplate(c *gin.Context) {
	if h.userImportService == nil {
		response.Error(c, "500000", "user import service is not configured")
		return
	}

	content, filename, err := h.userImportService.GenerateTemplate()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", content)
}

func (h *UserHandler) BatchImportUsers(c *gin.Context) {
	if h.userImportService == nil {
		response.Error(c, "500000", "user import service is not configured")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, "500000", "Failed to get uploaded file: "+err.Error())
		return
	}
	defer file.Close()

	if header.Size > service.MaxUserImportFileSize {
		response.Error(c, "500000", "file size exceeds 10MB limit")
		return
	}

	operator := embeddedDefaultUpdateBy
	if username, exists := c.Get("username"); exists {
		if value, ok := username.(string); ok && value != "" {
			operator = value
		}
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}

	result, err := h.userImportService.ImportUsers(file, header, operator, orgID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *UserHandler) DownloadErrorRecord(c *gin.Context) {
	if h.userImportService == nil {
		response.Error(c, "500000", "user import service is not configured")
		return
	}

	key := c.Param("key")
	if key == "" {
		response.Error(c, "500000", "error report key is required")
		return
	}

	content, filename, err := h.userImportService.GetErrorReport(key)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", content)
}

func (h *UserHandler) ClearErrorRecord(c *gin.Context) {
	if h.userImportService == nil {
		response.Error(c, "500000", "user import service is not configured")
		return
	}

	key := c.Param("key")
	if key == "" {
		response.Error(c, "500000", "error report key is required")
		return
	}

	if err := h.userImportService.ClearErrorReport(key); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) GetDefaultPassword(c *gin.Context) {
	if h.userService == nil {
		response.Error(c, "500000", "user service is not configured")
		return
	}

	response.Success(c, gin.H{"defaultPwd": h.userService.ResolveDefaultPassword()})
}

func (h *UserHandler) ResetPasswordCompat(c *gin.Context) {
	if h.userService == nil {
		response.Error(c, "500000", "user service is not configured")
		return
	}

	idParam := c.Param("uid")
	if idParam == "" {
		idParam = c.Param("id")
	}
	if idParam == "" {
		response.Error(c, "500000", "Invalid user ID")
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid user ID")
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if err = h.userService.EnsureUserInOrg(id, orgID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	if err = h.userService.ResetPassword(id, h.userService.ResolveDefaultPassword()); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) SwitchEnable(c *gin.Context) {
	if h.userService == nil {
		response.Error(c, "500000", "user service is not configured")
		return
	}

	var req user.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if req.ID <= 0 || req.Status == nil {
		response.Error(c, "500000", "Invalid request: id and status are required")
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if err := h.userService.EnsureUserInOrg(req.ID, orgID); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	if err := h.userService.UpdateUserStatus(req.ID, *req.Status); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func RegisterUserRoutes(r *gin.RouterGroup, h *UserHandler) {
	userGroup := r.Group("/system/user")
	{
		userGroup.POST("/list", h.ListUsers)
		userGroup.POST("/create", h.CreateUser)
		userGroup.POST("/update", h.UpdateUser)
		userGroup.POST("/delete/:id", h.DeleteUser)
		userGroup.POST("/enable", h.SwitchEnable)
		userGroup.GET("/options", h.GetUserOptions)
		userGroup.GET("/defaultPwd", h.GetDefaultPassword)
		userGroup.POST("/resetPwd/:id", middleware.AuditLog(middleware.AuditConfig{
			ActionType:   audit.ActionTypeUserAction,
			ActionName:   "RESET_USER_PASSWORD",
			ResourceType: audit.ResourceTypeUser,
		}), h.ResetPasswordCompat)
	}

	r.GET("/user/info", h.GetUserInfo)
	r.POST("/user/switch/:id", h.SwitchOrg)
}
