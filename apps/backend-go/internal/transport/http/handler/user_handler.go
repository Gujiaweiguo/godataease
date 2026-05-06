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
	"github.com/gin-gonic/gin/binding"
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
		response.Error(c, response.CodeInternalError, errInvalidOrgContext)
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
	defer recoverServicePanic(c)
	var req user.UserQueryRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
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
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	defer recoverServicePanic(c)
	var req user.UserCreateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
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
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, id)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	defer recoverServicePanic(c)
	var req user.UserUpdateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if err := h.userService.EnsureUserInOrg(req.ID, orgID); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	if err := h.userService.UpdateUser(&req); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", errInvalidUserID)
	if !ok {
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if err := h.userService.EnsureUserInOrg(id, orgID); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	if err := h.userService.DeleteUser(id); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) GetUserOptions(c *gin.Context) {
	defer recoverServicePanic(c)
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	req := &user.UserQueryRequest{Current: 1, Size: 1000, OrgID: int64Ptr(orgID)}

	result, err := h.userService.SearchUsers(req)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, result.List)
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	userID := int64(middleware.GetUserID(c))
	if userID <= 0 {
		response.Unauthorized(c, "invalid user context")
		return
	}
	if h.buildBootstrap == nil {
		response.Error(c, response.CodeInternalError, "identity bootstrap resolver is not configured")
		return
	}

	bootstrap, err := h.buildBootstrap(userID, middleware.GetOrgID(c), c.GetHeader("Accept-Language"))
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, bootstrap)
}

func (h *UserHandler) PersonInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	userID := int64(middleware.GetUserID(c))
	if userID <= 0 {
		response.Unauthorized(c, "invalid user context")
		return
	}

	account, name := h.resolveWatermarkIdentity(userID, middleware.GetUsername(c))
	response.Success(c, gin.H{
		"id":      userID,
		"account": account,
		"name":    name,
		"ip":      c.ClientIP(),
		"model":   "de",
	})
}

func (h *UserHandler) IPInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	userID := int64(middleware.GetUserID(c))
	if userID <= 0 {
		response.Unauthorized(c, "invalid user context")
		return
	}

	account, name := h.resolveWatermarkIdentity(userID, middleware.GetUsername(c))
	response.Success(c, gin.H{
		"account": account,
		"name":    name,
		"ip":      c.ClientIP(),
	})
}

func (h *UserHandler) resolveWatermarkIdentity(userID int64, fallbackUsername string) (string, string) {
	account := fallbackUsername
	name := fallbackUsername

	if h.loadUserByID != nil {
		if u, err := h.loadUserByID(userID); err == nil && u != nil {
			if u.Username != "" {
				account = u.Username
			}
			if u.NickName != "" {
				name = u.NickName
			} else if u.Username != "" {
				name = u.Username
			}
		}
	}

	if account == "" {
		account = "admin"
	}
	if name == "" {
		name = account
	}

	return account, name
}

func (h *UserHandler) SwitchOrg(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.switchOrg == nil {
		response.Error(c, response.CodeInternalError, "org switcher is not configured")
		return
	}

	targetOrgID, ok := parseIDParamMsg(c, "id", errInvalidOrganizationID)
	if !ok {
		return
	}

	userID := int64(middleware.GetUserID(c))
	if userID <= 0 {
		response.Unauthorized(c, "invalid user context")
		return
	}

	tokenVO, err := h.switchOrg(userID, targetOrgID, c.GetHeader("Accept-Language"))
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, tokenVO)
}

func (h *UserHandler) SwitchLanguage(c *gin.Context) {
	defer recoverServicePanic(c)
	var req user.LangSwitchRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}

	userID := int64(middleware.GetUserID(c))
	if userID <= 0 {
		response.Error(c, response.CodeInternalError, errInvalidUserID)
		return
	}

	if err := h.userService.SwitchLanguage(userID, req.Lang); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) DownloadExcelTemplate(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.userImportService == nil {
		response.Error(c, response.CodeInternalError, "user import service is not configured")
		return
	}

	content, filename, err := h.userImportService.GenerateTemplate()
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", mimeExcelOpenXML)
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(200, mimeExcelOpenXML, content)
}

func (h *UserHandler) BatchImportUsers(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.userImportService == nil {
		response.Error(c, response.CodeInternalError, "user import service is not configured")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to get uploaded file: "+err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	if header.Size > service.MaxUserImportFileSize {
		response.Error(c, response.CodeInternalError, "file size exceeds 10MB limit")
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
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

func (h *UserHandler) DownloadErrorRecord(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.userImportService == nil {
		response.Error(c, response.CodeInternalError, "user import service is not configured")
		return
	}

	key := c.Param("key")
	if key == "" {
		response.Error(c, response.CodeInternalError, "error report key is required")
		return
	}

	content, filename, err := h.userImportService.GetErrorReport(key)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", mimeExcelOpenXML)
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(200, mimeExcelOpenXML, content)
}

func (h *UserHandler) ClearErrorRecord(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.userImportService == nil {
		response.Error(c, response.CodeInternalError, "user import service is not configured")
		return
	}

	key := c.Param("key")
	if key == "" {
		response.Error(c, response.CodeInternalError, "error report key is required")
		return
	}

	if err := h.userImportService.ClearErrorReport(key); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) GetDefaultPassword(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.userService == nil {
		response.Error(c, response.CodeInternalError, "user service is not configured")
		return
	}

	response.Success(c, gin.H{"defaultPwd": h.userService.ResolveDefaultPassword()})
}

func (h *UserHandler) ResetPasswordCompat(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.userService == nil {
		response.Error(c, response.CodeInternalError, "user service is not configured")
		return
	}

	idParam := c.Param("uid")
	if idParam == "" {
		idParam = c.Param("id")
	}
	if idParam == "" {
		response.Error(c, response.CodeInternalError, errInvalidUserID)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, response.CodeInternalError, errInvalidUserID)
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if err = h.userService.EnsureUserInOrg(id, orgID); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	if err = h.userService.ResetPassword(id, h.userService.ResolveDefaultPassword()); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) SwitchEnable(c *gin.Context) {
	defer recoverServicePanic(c)
	if h.userService == nil {
		response.Error(c, response.CodeInternalError, "user service is not configured")
		return
	}

	var req user.UserUpdateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}
	if req.ID <= 0 || req.Status == nil {
		response.Error(c, response.CodeInternalError, "Invalid request: id and status are required")
		return
	}
	orgID, ok := requireCurrentOrg(c)
	if !ok {
		return
	}
	if err := h.userService.EnsureUserInOrg(req.ID, orgID); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
		return
	}

	if err := h.userService.UpdateUserStatus(req.ID, *req.Status); err != nil {
		response.Error(c, response.CodeInternalError, "Failed: "+err.Error())
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
	r.POST("/user/excelTemplate", h.DownloadExcelTemplate)
	r.POST("/user/batchImport", h.BatchImportUsers)
	r.GET("/user/errorRecord/:key", h.DownloadErrorRecord)
	r.GET("/user/clearErrorRecord/:key", h.ClearErrorRecord)
	r.GET("/user/personInfo", h.PersonInfo)
	r.GET("/user/ipInfo", h.IPInfo)
	r.POST("/user/byCurOrg", h.ListUsers)
}
