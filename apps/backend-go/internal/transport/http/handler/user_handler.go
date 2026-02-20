package handler

import (
	"strconv"

	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
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
