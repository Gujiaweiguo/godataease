package handler

import (
	"bytes"
	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	transportmiddleware "dataease/backend/internal/transport/http/middleware"
	"net/url"

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
	group.POST("/form/:formId/tableData", h.TableData)
	group.POST("/form/:formId/rowData/save", h.SaveRowData)
	group.GET("/form/:formId/delete/:id", h.DeleteRowData)
	group.POST("/form/:formId/batch-delete", h.BatchDeleteRowData)
	group.GET("/form/:formId/truncate", h.TruncateTableData)
	group.POST("/form/:formId/listColumnData", h.ListColumnData)
	group.GET("/form/:formId/excelTemplate", h.ExcelTemplate)
	group.POST("/form/:formId/uploadFile", h.ExcelUpload)
	group.POST("/form/:formId/confirmUpload", h.ConfirmUpload)
	group.POST("/form/extraDetails", h.ExtraDetails)
	group.POST("/form/:formId/options", h.ListDatasourceOptions)
	group.GET("/template/:itemId", h.GetTemplateByUserTaskItem)
	group.POST("/innerExport/:isDataEaseBi/:formId", h.ExportFormData)
	group.POST("/log/page/:goPage/:pageSize", h.LogPage)
	group.POST("/log/clear", h.LogClear)
	group.GET("/task/info/:taskId", h.GetTaskInfo)
	group.POST("/task/save", h.SaveTask)
	group.POST("/task/executeNow", h.ExecuteNowTask)
	group.POST("/form/:formId/task/page/:goPage/:pageSize", h.TaskPageList)
	group.GET("/form/:formId/task/:id/start", h.StartTask)
	group.GET("/form/:formId/task/:id/stop", h.StopTask)
	group.POST("/form/:formId/task/delete", h.DeleteTasks)
	group.POST("/sub-task/page/:goPage/:pageSize", h.SubTaskPageList)
	group.POST("/form/:formId/sub-task/delete", h.DeleteSubTasks)
	group.GET("/sub-task/:id/users/list/:type", h.SubTaskUsersList)
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
		response.Error(c, response.CodeInternalError, err.Error())
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
		response.Error(c, response.CodeInternalError, err.Error())
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
		response.Error(c, response.CodeInternalError, err.Error())
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
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) TableData(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	var req datafillingdomain.TableDataRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.SearchTableData(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) SaveRowData(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.SaveRowData(c.Request.Context(), formID, req, int64(transportmiddleware.GetUserID(c)), transportmiddleware.GetUsername(c))
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) DeleteRowData(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	rowID := c.Param("id")
	if err := h.service.DeleteRowData(c.Request.Context(), formID, rowID, int64(transportmiddleware.GetUserID(c)), transportmiddleware.GetUsername(c)); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) BatchDeleteRowData(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	var req datafillingdomain.BatchDeleteRowDataRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.BatchDeleteRowData(c.Request.Context(), formID, req.IDs, int64(transportmiddleware.GetUserID(c)), transportmiddleware.GetUsername(c)); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) TruncateTableData(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	if err := h.service.TruncateTableData(c.Request.Context(), formID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) ListColumnData(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	var req datafillingdomain.ListColumnDataRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.ListColumnData(c.Request.Context(), formID, req.ColumnName)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) ExcelTemplate(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	buf := bytes.NewBuffer(nil)
	if err := h.service.ExcelTemplateDownload(c.Request.Context(), formID, buf); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	filename := "template.xlsx"
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", mimeExcelOpenXML)
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(200, mimeExcelOpenXML, buf.Bytes())
}

func (h *DataFillingHandler) ExcelUpload(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to get uploaded file: "+err.Error())
		return
	}
	result, err := h.service.ExcelUpload(c.Request.Context(), formID, fileHeader)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) ConfirmUpload(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.ConfirmUpload(c.Request.Context(), formID, req.ID, int64(transportmiddleware.GetUserID(c)), transportmiddleware.GetUsername(c)); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) ExtraDetails(c *gin.Context) {
	defer recoverServicePanic(c)
	var req datafillingdomain.ExtraDetailsRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.ExtraDetails(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) ListDatasourceOptions(c *gin.Context) {
	defer recoverServicePanic(c)
	datasourceID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	var req datafillingdomain.DatasourceOptionsRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.ListDatasourceOptions(c.Request.Context(), datasourceID, &req)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) GetTemplateByUserTaskItem(c *gin.Context) {
	defer recoverServicePanic(c)
	itemID, ok := parseIDParamBadRequest(c, "itemId")
	if !ok {
		return
	}
	result, err := h.service.GetTemplateByUserTaskItem(c.Request.Context(), itemID)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) ExportFormData(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	buf := bytes.NewBuffer(nil)
	if err := h.service.ExportFormData(c.Request.Context(), formID, buf); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	filename := "form-data.xlsx"
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", mimeExcelOpenXML)
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(200, mimeExcelOpenXML, buf.Bytes())
}

func (h *DataFillingHandler) LogPage(c *gin.Context) {
	defer recoverServicePanic(c)
	goPage, pageSize, ok := parseThresholdPageParams(c)
	if !ok {
		return
	}
	var req datafillingdomain.CommitLogPageRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rows, total, err := h.service.ListCommitLogs(c.Request.Context(), req.FormID, goPage, pageSize)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"records": rows, "total": total, "current": goPage, "size": pageSize})
}

func (h *DataFillingHandler) LogClear(c *gin.Context) {
	defer recoverServicePanic(c)
	var req datafillingdomain.ClearCommitLogRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.ClearCommitLogs(c.Request.Context(), req.FormID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) GetTaskInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	taskID, ok := parseIDParamBadRequest(c, "taskId")
	if !ok {
		return
	}
	result, err := h.service.GetTaskInfo(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) SaveTask(c *gin.Context) {
	defer recoverServicePanic(c)
	var req datafillingdomain.TaskSaveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.SaveTask(c.Request.Context(), &req, int64(transportmiddleware.GetUserID(c)))
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) ExecuteNowTask(c *gin.Context) {
	defer recoverServicePanic(c)
	var req datafillingdomain.ExecuteNowRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.ExecuteNowTask(c.Request.Context(), req.TaskID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) TaskPageList(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	goPage, pageSize, ok := parseThresholdPageParams(c)
	if !ok {
		return
	}
	result, err := h.service.TaskPageList(c.Request.Context(), formID, goPage, pageSize)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) StartTask(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	taskID, ok := parseIDParamBadRequest(c, "id")
	if !ok {
		return
	}
	if err := h.service.StartTask(c.Request.Context(), formID, taskID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) StopTask(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	taskID, ok := parseIDParamBadRequest(c, "id")
	if !ok {
		return
	}
	if err := h.service.StopTask(c.Request.Context(), formID, taskID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) DeleteTasks(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	var req datafillingdomain.BatchDeleteTaskRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.DeleteTasks(c.Request.Context(), formID, req.IDs); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) SubTaskPageList(c *gin.Context) {
	defer recoverServicePanic(c)
	goPage, pageSize, ok := parseThresholdPageParams(c)
	if !ok {
		return
	}
	var req datafillingdomain.SubTaskPageRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.SubTaskPageList(c.Request.Context(), req.TaskID, goPage, pageSize)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) DeleteSubTasks(c *gin.Context) {
	defer recoverServicePanic(c)
	formID, ok := parseIDParamBadRequest(c, "formId")
	if !ok {
		return
	}
	var req datafillingdomain.BatchDeleteSubTaskRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.DeleteSubTasks(c.Request.Context(), formID, req.IDs); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *DataFillingHandler) SubTaskUsersList(c *gin.Context) {
	defer recoverServicePanic(c)
	subTaskID, ok := parseIDParamBadRequest(c, "id")
	if !ok {
		return
	}
	result, err := h.service.SubTaskUsersList(c.Request.Context(), subTaskID, c.Param("type"))
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
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
		response.Error(c, response.CodeInternalError, err.Error())
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
		response.Error(c, response.CodeInternalError, err.Error())
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
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) ListDatasourceList(c *gin.Context) {
	defer recoverServicePanic(c)
	result, err := h.service.ListDatasourceList(c.Request.Context())
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) ListDatasourceListAll(c *gin.Context) {
	defer recoverServicePanic(c)
	result, err := h.service.ListDatasourceListAll(c.Request.Context())
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DataFillingHandler) GetBuiltInTables(c *gin.Context) {
	defer recoverServicePanic(c)
	result, err := h.service.GetBuiltInTables(c.Request.Context())
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}
