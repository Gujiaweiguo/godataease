package handler

import (
	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	transportmiddleware "dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type UserTaskHandler struct {
	service *service.DataFillingService
}

func NewUserTaskHandler(service *service.DataFillingService) *UserTaskHandler {
	return &UserTaskHandler{service: service}
}

func RegisterDataFillingUserTaskRoutes(r *gin.RouterGroup, h *UserTaskHandler, authMiddleware, menuAuthMiddleware gin.HandlerFunc) {
	userTask := r.Group("/data-filling/user-task")
	if authMiddleware != nil {
		userTask.Use(authMiddleware)
	}
	if menuAuthMiddleware != nil {
		userTask.Use(menuAuthMiddleware)
	}
	userTask.POST("/page/:goPage/:pageSize", h.UserTaskList)
	userTask.POST("/todo/count", h.UserTaskTodoCount)
	userTask.GET("/list/:id", h.UserTaskData)
	userTask.POST("/saveData/:id", h.UserTaskSave)
	userTask.POST("/appendData/:id", h.UserTaskAppend)
	userTask.GET("/:taskInstanceId/deleteData/:id", h.UserTaskDelete)
}

func (h *UserTaskHandler) UserTaskList(c *gin.Context) {
	defer recoverServicePanic(c)
	goPage, pageSize, ok := parseThresholdPageParams(c)
	if !ok {
		return
	}
	var req datafillingdomain.UserTaskPageRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil && !isEOFBindError(err) {
		response.BadRequest(c, err.Error())
		return
	}
	rows, total, err := h.service.UserTaskPageList(c.Request.Context(), int64(transportmiddleware.GetUserID(c)), goPage, pageSize, &req)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, gin.H{"records": rows, "total": total, "current": goPage, "size": pageSize})
}

func (h *UserTaskHandler) UserTaskTodoCount(c *gin.Context) {
	defer recoverServicePanic(c)
	count, err := h.service.UserTaskTodoCount(c.Request.Context(), int64(transportmiddleware.GetUserID(c)))
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, count)
}

func (h *UserTaskHandler) UserTaskData(c *gin.Context) {
	defer recoverServicePanic(c)
	subTaskID, ok := parseIDParamBadRequest(c, "id")
	if !ok {
		return
	}
	data, err := h.service.GetUserTaskData(c.Request.Context(), int64(transportmiddleware.GetUserID(c)), subTaskID)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, data)
}

func (h *UserTaskHandler) UserTaskSave(c *gin.Context) {
	defer recoverServicePanic(c)
	subTaskID, ok := parseIDParamBadRequest(c, "id")
	if !ok {
		return
	}
	var req datafillingdomain.UserTaskSaveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.SaveUserTaskData(c.Request.Context(), int64(transportmiddleware.GetUserID(c)), subTaskID, req.Data); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *UserTaskHandler) UserTaskAppend(c *gin.Context) {
	defer recoverServicePanic(c)
	subTaskID, ok := parseIDParamBadRequest(c, "id")
	if !ok {
		return
	}
	var req datafillingdomain.UserTaskSaveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.AppendUserTaskData(c.Request.Context(), int64(transportmiddleware.GetUserID(c)), subTaskID, req.Data); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *UserTaskHandler) UserTaskDelete(c *gin.Context) {
	defer recoverServicePanic(c)
	subTaskID, ok := parseIDParamBadRequest(c, "taskInstanceId")
	if !ok {
		return
	}
	dataID := c.Param("id")
	if err := h.service.DeleteUserTaskData(c.Request.Context(), int64(transportmiddleware.GetUserID(c)), subTaskID, []string{dataID}); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, nil)
}
