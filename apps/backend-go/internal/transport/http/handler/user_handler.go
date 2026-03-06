package handler

import (
	"net/url"
	"strconv"

	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService       *service.UserService
	userImportService *service.UserImportService
}

func NewUserHandler(userService *service.UserService, userImportService *service.UserImportService) *UserHandler {
	return &UserHandler{
		userService:       userService,
		userImportService: userImportService,
	}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	var req user.UserQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
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

	if err := h.userService.DeleteUser(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) GetUserOptions(c *gin.Context) {
	req := &user.UserQueryRequest{Current: 1, Size: 1000}

	result, err := h.userService.SearchUsers(req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	response.Success(c, result.List)
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	response.Success(c, map[string]interface{}{
		"id":       1,
		"name":     "admin",
		"oid":      1,
		"language": "zh-CN",
	})
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

	operator := "system"
	if username, exists := c.Get("username"); exists {
		if value, ok := username.(string); ok && value != "" {
			operator = value
		}
	}

	result, err := h.userImportService.ImportUsers(file, header, operator)
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

	if err = h.userService.ResetPassword(id, h.userService.ResolveDefaultPassword()); err != nil {
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
		userGroup.GET("/options", h.GetUserOptions)
	}

	r.GET("/user/info", h.GetUserInfo)
}
