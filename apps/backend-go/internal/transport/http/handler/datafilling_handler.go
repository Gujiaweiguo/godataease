package handler

import (
	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	transportmiddleware "dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type DataFillingHandler struct {
	service *service.DataFillingService
}

func NewDataFillingHandler(svc *service.DataFillingService) *DataFillingHandler {
	return &DataFillingHandler{service: svc}
}

func RegisterDataFillingRoutes(r gin.IRouter, h *DataFillingHandler, authMiddleware, menuAuthMiddleware gin.HandlerFunc) {
	group := r.Group("/data-filling")
	if authMiddleware != nil {
		group.Use(authMiddleware)
	}
	if menuAuthMiddleware != nil {
		group.Use(menuAuthMiddleware)
	}
	group.POST("/tree", h.Tree)
	group.GET("/get/:id", h.Get)
	group.POST("/save", h.Save)
	group.POST("/update", h.Update)
	group.POST("/rename", h.Rename)
	group.POST("/move", h.Move)
	group.GET("/delete/:id", h.Delete)
	group.GET("/datasource/list", h.ListDatasourceList)
	group.GET("/datasource/listAll", h.ListDatasourceListAll)
	group.POST("/getBuiltInTables", h.GetBuiltInTables)
}

func (h *DataFillingHandler) Save(c *gin.Context) {
	defer recoverServicePanic(c)
	var req datafillingdomain.CreateFormRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.Save(c.Request.Context(), &req, int64(transportmiddleware.GetUserID(c)))
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) Get(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamBadRequest(c, "id")
	if !ok {
		return
	}
	result, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) Update(c *gin.Context) {
	defer recoverServicePanic(c)
	var req datafillingdomain.UpdateFormRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.Update(c.Request.Context(), &req, int64(transportmiddleware.GetUserID(c)))
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) Delete(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamBadRequest(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) Rename(c *gin.Context) {
	defer recoverServicePanic(c)
	var req datafillingdomain.RenameRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.Rename(c.Request.Context(), req.ID, req.Name)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) Move(c *gin.Context) {
	defer recoverServicePanic(c)
	var req datafillingdomain.MoveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.Move(c.Request.Context(), req.ID, req.PID)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) Tree(c *gin.Context) {
	defer recoverServicePanic(c)
	var req datafillingdomain.TreeRequest
	if err := shouldBindOptionalJSON(c, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.Tree(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) ListDatasourceList(c *gin.Context) {
	defer recoverServicePanic(c)
	result, err := h.service.ListDatasourceList(c.Request.Context())
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) ListDatasourceListAll(c *gin.Context) {
	defer recoverServicePanic(c)
	result, err := h.service.ListDatasourceListAll(c.Request.Context())
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) GetBuiltInTables(c *gin.Context) {
	defer recoverServicePanic(c)
	result, err := h.service.GetBuiltInTables(c.Request.Context())
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}
