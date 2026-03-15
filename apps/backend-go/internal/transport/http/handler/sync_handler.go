package handler

import (
	"strconv"
	"strings"

	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/syncmodule"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type SyncHandler struct {
	service *service.SyncService
}

func NewSyncHandler(service *service.SyncService) *SyncHandler {
	return &SyncHandler{service: service}
}

func (h *SyncHandler) SourceDatasourcePager(c *gin.Context) {
	page, size, ok := parsePageParams(c)
	if !ok {
		return
	}
	var req datasource.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil && !isEOFBindError(err) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.SourcePager(page, size, &req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) TargetDatasourcePager(c *gin.Context) {
	page, size, ok := parsePageParams(c)
	if !ok {
		return
	}
	var req datasource.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil && !isEOFBindError(err) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.TargetPager(page, size, &req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) LatestUse(c *gin.Context) {
	creator := ""
	if username, exists := c.Get("username"); exists {
		if s, ok := username.(string); ok {
			creator = s
		}
	}
	result, err := h.service.LatestUse(c.Param("sourceType"), creator)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) ValidateDatasource(c *gin.Context) {
	var req datasource.ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.ValidateDatasource(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) ValidateDatasourceByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	result, err := h.service.ValidateDatasourceByID(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) GetSchemas(c *gin.Context) {
	result, err := h.service.GetSchemas()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) SaveDatasource(c *gin.Context) {
	var req datasource.WriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.SaveDatasource(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) GetDatasource(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	result, err := h.service.GetDatasource(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) UpdateDatasource(c *gin.Context) {
	var req datasource.WriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.UpdateDatasource(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) DeleteDatasource(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteDatasource(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) BatchDeleteDatasource(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	parsedIDs, err := parseIDList(ids)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	if err = h.service.BatchDeleteDatasource(parsedIDs); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) GetDatasourceFields(c *gin.Context) {
	var req syncmodule.SyncDatasourceFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.GetDatasourceFields(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) ListDatasourceByType(c *gin.Context) {
	result, err := h.service.ListDatasourceByType(c.Param("type"))
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) ListDatasourceTables(c *gin.Context) {
	id, ok := parseIDParam(c, "dsId")
	if !ok {
		return
	}
	result, err := h.service.ListDatasourceTables(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) TaskPager(c *gin.Context) {
	page, size, ok := parsePageParams(c)
	if !ok {
		return
	}
	var req syncmodule.TaskGridRequest
	if err := c.ShouldBindJSON(&req); err != nil && !isEOFBindError(err) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.TaskPager(page, size, &req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) GetTask(c *gin.Context) {
	id, ok := parseIDParam(c, "taskId")
	if !ok {
		return
	}
	result, err := h.service.GetTask(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) AddTask(c *gin.Context) {
	var req syncmodule.TaskInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.AddTask(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) UpdateTask(c *gin.Context) {
	var req syncmodule.TaskInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.UpdateTask(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) RemoveTask(c *gin.Context) {
	id, ok := parseIDParam(c, "taskId")
	if !ok {
		return
	}
	if err := h.service.RemoveTask(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) BatchDeleteTasks(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	parsedIDs, err := parseIDList(ids)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	if err = h.service.BatchDeleteTasks(parsedIDs); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) ExecuteTask(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	result, err := h.service.ExecuteTask(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) StartTask(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.StartTask(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) StopTask(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.StopTask(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) TaskLogPager(c *gin.Context) {
	page, size, ok := parsePageParams(c)
	if !ok {
		return
	}
	var req syncmodule.TaskLogGridRequest
	if err := c.ShouldBindJSON(&req); err != nil && !isEOFBindError(err) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.TaskLogPager(page, size, &req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) TaskLogDetail(c *gin.Context) {
	id, ok := parseIDParam(c, "logId")
	if !ok {
		return
	}
	fromLineNum, err := strconv.Atoi(c.Param("fromLineNum"))
	if err != nil {
		response.Error(c, "500000", "Invalid fromLineNum")
		return
	}
	result, err := h.service.TaskLogDetail(id, fromLineNum)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) DeleteTaskLog(c *gin.Context) {
	id, ok := parseIDParam(c, "logId")
	if !ok {
		return
	}
	if err := h.service.DeleteTaskLog(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) ClearTaskLog(c *gin.Context) {
	var req syncmodule.TaskLog
	if err := c.ShouldBindJSON(&req); err != nil && !isEOFBindError(err) {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if err := h.service.ClearTaskLog(&req); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) TerminateTask(c *gin.Context) {
	id, ok := parseIDParam(c, "logId")
	if !ok {
		return
	}
	if err := h.service.TerminateTaskByLogID(id); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SyncHandler) ResourceCount(c *gin.Context) {
	result, err := h.service.ResourceCount()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SyncHandler) LogChartData(c *gin.Context) {
	result, err := h.service.LogChartData()
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterSyncRoutes(r *gin.RouterGroup, h *SyncHandler) {
	syncGroup := r.Group("/sync")
	{
		datasourceGroup := syncGroup.Group("/datasource")
		{
			datasourceGroup.POST("/source/pager/:page/:limit", h.SourceDatasourcePager)
			datasourceGroup.POST("/target/pager/:page/:limit", h.TargetDatasourcePager)
			datasourceGroup.POST("/latestUse/:sourceType", h.LatestUse)
			datasourceGroup.POST("/validate", h.ValidateDatasource)
			datasourceGroup.GET("/validate/:id", h.ValidateDatasourceByID)
			datasourceGroup.POST("/getSchema", h.GetSchemas)
			datasourceGroup.POST("/save", h.SaveDatasource)
			datasourceGroup.GET("/get/:id", h.GetDatasource)
			datasourceGroup.POST("/update", h.UpdateDatasource)
			datasourceGroup.POST("/delete/:id", h.DeleteDatasource)
			datasourceGroup.POST("/batchDel", h.BatchDeleteDatasource)
			datasourceGroup.POST("/fields", h.GetDatasourceFields)
			datasourceGroup.GET("/list/:type", h.ListDatasourceByType)
			datasourceGroup.GET("/table/list/:dsId", h.ListDatasourceTables)
		}

		taskGroup := syncGroup.Group("/task")
		{
			taskGroup.POST("/pager/:current/:size", h.TaskPager)
			taskGroup.GET("/get/:taskId", h.GetTask)
			taskGroup.POST("/add", h.AddTask)
			taskGroup.POST("/update", h.UpdateTask)
			taskGroup.POST("/remove/:taskId", h.RemoveTask)
			taskGroup.POST("/batch/del", h.BatchDeleteTasks)
			taskGroup.GET("/execute/:id", h.ExecuteTask)
			taskGroup.GET("/start/:id", h.StartTask)
			taskGroup.GET("/stop/:id", h.StopTask)

			logGroup := taskGroup.Group("/log")
			{
				logGroup.POST("/pager/:current/:size", h.TaskLogPager)
				logGroup.GET("/detail/:logId/:fromLineNum", h.TaskLogDetail)
				logGroup.POST("/delete/:logId", h.DeleteTaskLog)
				logGroup.POST("/clear", h.ClearTaskLog)
				logGroup.POST("/terminationTask/:logId", h.TerminateTask)
			}
		}

		summaryGroup := syncGroup.Group("/summary")
		{
			summaryGroup.GET("/resourceCount", h.ResourceCount)
			summaryGroup.POST("/logChartData", h.LogChartData)
		}
	}
}

func parsePageParams(c *gin.Context) (int, int, bool) {
	page, err := strconv.Atoi(strings.TrimSpace(firstNonEmptyParam(c.Param("page"), c.Param("current"))))
	if err != nil || page < 1 {
		response.Error(c, "500000", "Invalid page")
		return 0, 0, false
	}
	size, err := strconv.Atoi(strings.TrimSpace(firstNonEmptyParam(c.Param("limit"), c.Param("size"))))
	if err != nil || size < 1 {
		response.Error(c, "500000", "Invalid size")
		return 0, 0, false
	}
	return page, size, true
}

func parseIDParam(c *gin.Context, key string) (int64, bool) {
	value := strings.TrimSpace(c.Param(key))
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, "500000", "Invalid id")
		return 0, false
	}
	return id, true
}

func parseIDList(values []string) ([]int64, error) {
	result := make([]int64, 0, len(values))
	for _, value := range values {
		id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || id <= 0 {
			return nil, strconv.ErrSyntax
		}
		result = append(result, id)
	}
	return result, nil
}

func isEOFBindError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "EOF")
}

func firstNonEmptyParam(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
