package handler

import (
	"strconv"
	"strings"

	thresholddomain "dataease/backend/internal/domain/threshold"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	transportmiddleware "dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type ThresholdHandler struct {
	thresholdService *service.ThresholdService
}

func NewThresholdHandler(svc *service.ThresholdService) *ThresholdHandler {
	return &ThresholdHandler{thresholdService: svc}
}

func RegisterThresholdRoutes(r gin.IRouter, h *ThresholdHandler, authMiddleware, menuAuthMiddleware gin.HandlerFunc) {
	group := r.Group("/threshold")
	if authMiddleware != nil {
		group.Use(authMiddleware)
	}

	authGroup := group.Group("")
	if menuAuthMiddleware != nil {
		authGroup.Use(menuAuthMiddleware)
	}

	authGroup.POST("/save", h.Save)
	authGroup.POST("/edit", h.Edit)
	authGroup.POST("/pager/:goPage/:pageSize", h.Pager)
	authGroup.GET("/formInfo/:id/:resourceTable", h.FormInfo)
	authGroup.POST("/switch", h.SwitchEnable)
	authGroup.POST("/delete/:resourceTable", h.Delete)
	authGroup.POST("/batchReci", h.BatchReci)
	authGroup.POST("/instancePager/:goPage/:pageSize", h.InstancePager)
	authGroup.POST("/preview", h.Preview)
	authGroup.GET("/deleteWithChart/:chartId/:resourceTable", h.DeleteWithChart)

	group.GET("/anyThreshold/:chartId/:resourceTable", h.AnyThreshold)
}

func (h *ThresholdHandler) Save(c *gin.Context) {
	defer recoverServicePanic(c)
	var req thresholddomain.CreateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, userName, oid := getThresholdUserInfo(c)
	result, err := h.thresholdService.Create(c.Request.Context(), &req, userID, userName, oid)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ThresholdHandler) Edit(c *gin.Context) {
	defer recoverServicePanic(c)
	var req thresholddomain.CreateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.thresholdService.Edit(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ThresholdHandler) Pager(c *gin.Context) {
	defer recoverServicePanic(c)
	goPage, pageSize, ok := parseThresholdPageParams(c)
	if !ok {
		return
	}

	var req thresholddomain.GridRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.thresholdService.Pager(c.Request.Context(), &req, goPage, pageSize)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ThresholdHandler) FormInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamBadRequest(c, "id")
	if !ok {
		return
	}

	result, err := h.thresholdService.FormInfo(c.Request.Context(), id, strings.TrimSpace(c.Param("resourceTable")))
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ThresholdHandler) SwitchEnable(c *gin.Context) {
	defer recoverServicePanic(c)
	var req thresholddomain.SwitchRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.thresholdService.SwitchEnable(c.Request.Context(), &req); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ThresholdHandler) Delete(c *gin.Context) {
	defer recoverServicePanic(c)
	resourceTable := strings.TrimSpace(c.Param("resourceTable"))
	var ids []int64
	if err := c.ShouldBindBodyWith(&ids, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.thresholdService.Delete(c.Request.Context(), ids, resourceTable); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ThresholdHandler) BatchReci(c *gin.Context) {
	defer recoverServicePanic(c)
	var req thresholddomain.BatchReciRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.thresholdService.BatchReci(c.Request.Context(), &req); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ThresholdHandler) InstancePager(c *gin.Context) {
	defer recoverServicePanic(c)
	goPage, pageSize, ok := parseThresholdPageParams(c)
	if !ok {
		return
	}

	var req thresholddomain.InstanceRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.thresholdService.InstancePager(c.Request.Context(), &req, goPage, pageSize)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ThresholdHandler) Preview(c *gin.Context) {
	defer recoverServicePanic(c)
	var req thresholddomain.PreviewRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.thresholdService.Preview(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ThresholdHandler) AnyThreshold(c *gin.Context) {
	defer recoverServicePanic(c)
	chartID, ok := parseIDParamBadRequest(c, "chartId")
	if !ok {
		return
	}

	result, err := h.thresholdService.AnyThreshold(c.Request.Context(), chartID, strings.TrimSpace(c.Param("resourceTable")))
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ThresholdHandler) DeleteWithChart(c *gin.Context) {
	defer recoverServicePanic(c)
	chartID, ok := parseIDParamBadRequest(c, "chartId")
	if !ok {
		return
	}

	if err := h.thresholdService.DeleteWithChart(c.Request.Context(), chartID, strings.TrimSpace(c.Param("resourceTable"))); err != nil {
		response.Error(c, "500000", err.Error())
		return
	}

	response.Success(c, nil)
}

func getThresholdUserInfo(c *gin.Context) (int64, string, int64) {
	return int64(transportmiddleware.GetUserID(c)), transportmiddleware.GetUsername(c), transportmiddleware.GetOrgID(c)
}

func parseThresholdPageParams(c *gin.Context) (int, int, bool) {
	goPage, err := strconv.Atoi(strings.TrimSpace(c.Param("goPage")))
	if err != nil || goPage < 1 {
		response.BadRequest(c, "Invalid goPage")
		return 0, 0, false
	}

	pageSize, err := strconv.Atoi(strings.TrimSpace(c.Param("pageSize")))
	if err != nil || pageSize < 1 {
		response.BadRequest(c, "Invalid pageSize")
		return 0, 0, false
	}

	return goPage, pageSize, true
}
