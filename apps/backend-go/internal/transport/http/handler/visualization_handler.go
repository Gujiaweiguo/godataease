package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type VisualizationHandler struct {
	service *service.VisualizationService
}

const maxVisualizationTemplateUploadBytes = 35 << 20

func NewVisualizationHandler(service *service.VisualizationService) *VisualizationHandler {
	return &VisualizationHandler{service: service}
}

func (h *VisualizationHandler) FindByID(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.DetailRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Detail(&req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Visualization not found")
			return
		}
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}

	var canvasViewInfo map[string]interface{}
	chartViews, err := h.service.ViewDetailList(req.ID.Int64())
	if err != nil {
		// Non-fatal: log and continue with empty canvasViewInfo
		canvasViewInfo = map[string]interface{}{}
	} else {
		canvasViewInfo = buildCanvasViewInfo(chartViews)
	}

	// Enrich response with fields the frontend canvas expects.
	// The raw DataVisualizationInfo lacks watermarkInfo, canvasViewInfo,
	// weight, creatorName, etc. that initCanvasDataPrepare() requires.
	enriched := buildEnrichedVisualizationResponse(result, canvasViewInfo)
	response.Success(c, enriched)
}

// buildEnrichedVisualizationResponse wraps the raw domain model into a
// map containing all fields the frontend's initCanvasDataPrepare expects.
func buildEnrichedVisualizationResponse(v *visualization.DataVisualizationInfo, canvasViewInfo map[string]interface{}) map[string]interface{} {
	var mobileLayout bool
	if v.MobileLayout != nil {
		mobileLayout = *v.MobileLayout
	}
	var createTime int64
	if v.CreateTime != nil {
		createTime = *v.CreateTime
	}
	var updateTime int64
	if v.UpdateTime != nil {
		updateTime = *v.UpdateTime
	}
	resp := map[string]interface{}{
		"id":                  v.ID,
		"name":                v.Name,
		"pid":                 v.PID,
		"status":              v.Status,
		"type":                v.Type,
		"nodeType":            v.NodeType,
		"componentData":       v.ComponentData,
		"canvasStyleData":     v.CanvasStyleData,
		"mobileLayout":        mobileLayout,
		"version":             v.Version,
		"contentId":           v.ContentID,
		"selfWatermarkStatus": true,
		"watermarkInfo":       map[string]interface{}{"id": "1", "settingContent": "{}"},
		"weight":              9,
		"ext":                 map[string]interface{}{},
		"canvasViewInfo":      canvasViewInfo,
		"creatorName":         "admin",
		"updateName":          "admin",
		"createTime":          createTime,
		"updateTime":          updateTime,
	}
	return resp
}

func buildCanvasViewInfo(chartViews []chart.CoreChartView) map[string]interface{} {
	result := make(map[string]interface{})
	for _, view := range chartViews {
		// Serialize via JSON to get camelCase keys from struct tags
		jsonBytes, err := json.Marshal(view)
		if err != nil {
			continue
		}
		var viewMap map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &viewMap); err != nil {
			continue
		}
		parseJSONStrings(viewMap)

		viewID := ""
		if id, ok := viewMap["id"]; ok {
			switch v := id.(type) {
			case float64:
				viewID = strconv.FormatInt(int64(v), 10)
			case string:
				viewID = v
			}
		}
		if viewID != "" {
			result[viewID] = viewMap
		}
	}
	return result
}

// parseJSONStrings converts JSON-encoded string values into their actual
// object/array representations so the frontend receives parsed objects
// rather than JSON strings.
func parseJSONStrings(m map[string]interface{}) {
	for key, val := range m {
		if str, ok := val.(string); ok {
			str = strings.TrimSpace(str)
			if strings.HasPrefix(str, "{") || strings.HasPrefix(str, "[") {
				var parsed interface{}
				if err := json.Unmarshal([]byte(str), &parsed); err == nil {
					m[key] = parsed
				}
			}
		}
	}
}

func (h *VisualizationHandler) List(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.ListRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.List(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) FindRecent(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.WorkbranchQueryRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	req.QueryFrom = "recent"
	uid := int64(middleware.GetUserID(c))
	result, err := h.service.FindRecent(&req, uid)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

type treeRequest struct {
	BusiFlag string `json:"busiFlag"`
	Leaf     *bool  `json:"leaf"`
}

type treeNode struct {
	ID         string     `json:"id"`
	PID        string     `json:"pid"`
	Name       string     `json:"name"`
	Leaf       bool       `json:"leaf"`
	Weight     int        `json:"weight"`
	ExtraFlag  int        `json:"extraFlag"`
	ExtraFlag1 int        `json:"extraFlag1"`
	Children   []treeNode `json:"children,omitempty"`
}

const (
	visualizationNodeTypeFolder = "folder"
	visualizationNodeTypePanel  = "panel"
	visualizationNodeTypeLeaf   = "leaf"
)

func (h *VisualizationHandler) Tree(c *gin.Context) {
	defer recoverServicePanic(c)
	var req treeRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	types, err := resolveBusiTypes(req.BusiFlag)
	if err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	all := make([]*visualization.DataVisualizationInfo, 0)
	for _, typ := range types {
		t := typ
		list, err := h.service.List(&visualization.ListRequest{Type: &t, Current: 1, Size: 2000})
		if err != nil {
			response.Error(c, "500000", "Failed: "+err.Error())
			return
		}
		all = append(all, list.List...)
	}

	nodes, err := buildVisualizationTree(all, req.Leaf)
	if err != nil {
		response.Error(c, "500000", "Invalid tree payload: "+err.Error())
		return
	}
	root := treeNode{
		ID:         "0",
		PID:        "-1",
		Name:       "root",
		Leaf:       false,
		Weight:     9,
		ExtraFlag:  0,
		ExtraFlag1: 0,
		Children:   nodes,
	}

	if err := validateTreeNodes([]treeNode{root}); err != nil {
		response.Error(c, "500000", "Invalid tree payload: "+err.Error())
		return
	}

	response.Success(c, []treeNode{root})
}

func resolveBusiTypes(busiFlag string) ([]string, error) {
	flag := strings.TrimSpace(busiFlag)
	if flag == "" || flag == "dashboard-dataV" {
		return []string{"dashboard", "dataV"}, nil
	}
	if flag == visualizationNodeTypePanel {
		return []string{"dashboard"}, nil
	}
	if flag == "screen" {
		return []string{"dataV"}, nil
	}
	if flag == "dashboard" || flag == "dataV" {
		return []string{flag}, nil
	}

	return nil, fmt.Errorf("unsupported busiFlag: %s", flag)
}

func buildVisualizationTree(items []*visualization.DataVisualizationInfo, leafFilter *bool) ([]treeNode, error) {
	childrenMap := make(map[string][]treeNode)
	var roots []treeNode

	for _, item := range items {
		if item == nil {
			continue
		}
		if item.ID <= 0 {
			return nil, fmt.Errorf("node id is required")
		}
		if strings.TrimSpace(item.Name) == "" {
			return nil, fmt.Errorf("node name is required")
		}
		nodeType := ""
		if item.NodeType != nil {
			nodeType = *item.NodeType
		}
		if nodeType == visualizationNodeTypeLeaf {
			nodeType = visualizationNodeTypePanel
		}
		if nodeType != visualizationNodeTypeFolder && nodeType != visualizationNodeTypePanel {
			return nil, fmt.Errorf("invalid nodeType: %s", nodeType)
		}
		leaf := nodeType != visualizationNodeTypeFolder
		if leafFilter != nil && *leafFilter != leaf {
			continue
		}

		pid := "0"
		if item.PID != nil {
			pid = strconv.FormatInt(*item.PID, 10)
		}

		node := treeNode{
			ID:         strconv.FormatInt(item.ID, 10),
			PID:        pid,
			Name:       item.Name,
			Leaf:       leaf,
			Weight:     9,
			ExtraFlag:  visualizationExtraFlag(item.MobileLayout),
			ExtraFlag1: visualizationPublishFlag(item.Status),
			Children:   []treeNode{},
		}

		childrenMap[pid] = append(childrenMap[pid], node)
	}

	var attach func(parentID string) []treeNode
	attach = func(parentID string) []treeNode {
		nodes := childrenMap[parentID]
		for idx := range nodes {
			kids := attach(nodes[idx].ID)
			if len(kids) > 0 {
				nodes[idx].Children = kids
				nodes[idx].Leaf = false
			}
		}
		return nodes
	}

	roots = attach("0")
	return roots, nil
}

func visualizationExtraFlag(mobileLayout *bool) int {
	if mobileLayout != nil && *mobileLayout {
		return 1
	}
	return 0
}

func visualizationPublishFlag(status *int) int {
	if status != nil && *status > 0 {
		return 1
	}
	return 0
}

func validateTreeNodes(nodes []treeNode) error {
	for _, node := range nodes {
		if strings.TrimSpace(node.ID) == "" {
			return fmt.Errorf("node id is empty")
		}
		if strings.TrimSpace(node.Name) == "" {
			return fmt.Errorf("node name is empty")
		}
		if node.Leaf && len(node.Children) > 0 {
			return fmt.Errorf("leaf node cannot have children")
		}
		if len(node.Children) > 0 {
			if err := validateTreeNodes(node.Children); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *VisualizationHandler) SaveCanvas(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.SaveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	updateBy := h.getUpdateBy(c)
	id, err := h.service.Save(&req, updateBy)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, strconv.FormatInt(id, 10))
}

func (h *VisualizationHandler) FindDvType(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid ID")
	if !ok {
		return
	}
	var err error
	result, err := h.service.FindDvType(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) UpdateCheckVersion(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid ID")
	if !ok {
		return
	}
	var err error
	result, err := h.service.Detail(&visualization.DetailRequest{ID: visualization.FlexInt(id)})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	if result == nil || result.CheckVersion == nil {
		response.Success(c, "")
		return
	}
	response.Success(c, *result.CheckVersion)
}

func (h *VisualizationHandler) Copy(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.CopyRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	updateBy := h.getUpdateBy(c)
	id, err := h.service.Copy(&req, updateBy)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, id)
}

func (h *VisualizationHandler) CheckCanvasChange(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.CanvasChangeRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.CheckCanvasChange(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) UpdateCanvas(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.UpdateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	updateBy := h.getUpdateBy(c)
	if err := h.service.Update(&req, updateBy); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *VisualizationHandler) UpdateBase(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.UpdateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	updateBy := h.getUpdateBy(c)
	if err := h.service.UpdateBase(&req, updateBy); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *VisualizationHandler) Move(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.MoveRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	updateBy := h.getUpdateBy(c)
	if err := h.service.Move(&req, updateBy); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *VisualizationHandler) UpdatePublishStatus(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.UpdateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	updateBy := h.getUpdateBy(c)
	result, err := h.service.UpdatePublishStatus(&req, updateBy)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) RecoverToPublished(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.DetailRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	updateBy := h.getUpdateBy(c)
	result, err := h.service.RecoverToPublished(req.ID.Int64(), updateBy)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) DeleteLogic(c *gin.Context) {
	defer recoverServicePanic(c)
	id, ok := parseIDParamMsg(c, "id", "Invalid ID")
	if !ok {
		return
	}

	updateBy := h.getUpdateBy(c)
	if err := h.service.DeleteLogic(id, updateBy); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *VisualizationHandler) NameCheck(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.NameCheckRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.NameCheck(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) getUpdateBy(c *gin.Context) string {
	if userID, exists := c.Get("userId"); exists {
		switch v := userID.(type) {
		case int64:
			return strconv.FormatInt(v, 10)
		case int:
			return strconv.Itoa(v)
		case string:
			return v
		}
	}
	return "system"
}

func (h *VisualizationHandler) FindCopyResource(c *gin.Context) {
	defer recoverServicePanic(c)
	dvID, ok := parseIDParamMsg(c, "dvId", "Invalid dvId")
	if !ok {
		return
	}
	result, err := h.service.Detail(&visualization.DetailRequest{ID: visualization.FlexInt(dvID)})
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	if result != nil && result.PID != nil && *result.PID == -1 {
		response.Success(c, result)
		return
	}
	response.Success(c, nil)
}

func (h *VisualizationHandler) ViewDetailList(c *gin.Context) {
	defer recoverServicePanic(c)
	dvID, ok := parseIDParamMsg(c, "dvId", "Invalid dvId")
	if !ok {
		return
	}
	result, err := h.service.ViewDetailList(dvID)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) AppCanvasNameCheck(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.AppCanvasNameCheckRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.AppCanvasNameCheck(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) GetComponentInfo(c *gin.Context) {
	defer recoverServicePanic(c)
	response.Success(c, nil)
}

func (h *VisualizationHandler) Export2AppCheck(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.Export2AppCheckRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Export2AppCheck(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) ExportLogApp(c *gin.Context) {
	defer recoverServicePanic(c)
	h.recordExportLog(c, "app")
}

func (h *VisualizationHandler) ExportLogTemplate(c *gin.Context) {
	defer recoverServicePanic(c)
	h.recordExportLog(c, "template")
}

func (h *VisualizationHandler) ExportLogPDF(c *gin.Context) {
	defer recoverServicePanic(c)
	h.recordExportLog(c, "pdf")
}

func (h *VisualizationHandler) ExportLogImg(c *gin.Context) {
	defer recoverServicePanic(c)
	h.recordExportLog(c, "img")
}

func (h *VisualizationHandler) recordExportLog(c *gin.Context, logType string) {
	var req visualization.ExportLogRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	var userID *int64
	if uid, exists := c.Get("userId"); exists {
		switch v := uid.(type) {
		case int64:
			userID = &v
		case int:
			converted := int64(v)
			userID = &converted
		}
	}
	var username *string
	if raw, exists := c.Get("userName"); exists {
		if name, ok := raw.(string); ok && strings.TrimSpace(name) != "" {
			username = &name
		}
	}
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	if err := h.service.RecordExportLog(&req, userID, username, &ipAddress, &userAgent, logType); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *VisualizationHandler) Decompression(c *gin.Context) {
	defer recoverServicePanic(c)
	var req visualization.DecompressionRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Decompression(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) DecompressionLocalFile(c *gin.Context) {
	defer recoverServicePanic(c)

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, "500000", "Failed to read upload file: "+err.Error())
		return
	}

	opened, err := file.Open()
	if err != nil {
		response.Error(c, "500000", "Failed to open upload file: "+err.Error())
		return
	}
	defer func() { _ = opened.Close() }()

	content, err := io.ReadAll(io.LimitReader(opened, maxVisualizationTemplateUploadBytes+1))
	if err != nil {
		response.Error(c, "500000", "Failed to read file content: "+err.Error())
		return
	}
	if len(content) > maxVisualizationTemplateUploadBytes {
		response.Error(c, "500000", fmt.Sprintf("Template file exceeds %d bytes", maxVisualizationTemplateUploadBytes))
		return
	}

	result, err := h.service.DecompressionLocalFile(content)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func RegisterVisualizationRoutes(r *gin.RouterGroup, h *VisualizationHandler, permMiddleware *middleware.PermissionMiddleware) {
	vg := r.Group("/dataVisualization")
	{
		vg.GET("/findDvType/:id", h.FindDvType)
		vg.GET("/updateCheckVersion/:id", h.UpdateCheckVersion)
		vg.POST("/tree", h.Tree)
		vg.POST("/nameCheck", h.NameCheck)
		vg.POST("/checkCanvasChange", h.CheckCanvasChange)
		vg.POST("/list", h.List)
		if permMiddleware != nil {
			vg.POST("/save", permMiddleware.CheckVisualizationParentEdit(), h.SaveCanvas)
		} else {
			vg.POST("/save", h.SaveCanvas)
		}
		if permMiddleware != nil {
			vg.POST("/copy", permMiddleware.CheckVisualizationCopy(), h.Copy)
		} else {
			vg.POST("/copy", h.Copy)
		}
		if permMiddleware != nil {
			vg.POST("/updateBase", permMiddleware.CheckVisualizationEdit(), h.UpdateBase)
			vg.POST("/move", permMiddleware.CheckVisualizationEdit(), h.Move)
			vg.POST("/updatePublishStatus", permMiddleware.CheckVisualizationEdit(), h.UpdatePublishStatus)
			vg.POST("/recoverToPublished", permMiddleware.CheckVisualizationEdit(), h.RecoverToPublished)
		} else {
			vg.POST("/updateBase", h.UpdateBase)
			vg.POST("/move", h.Move)
			vg.POST("/updatePublishStatus", h.UpdatePublishStatus)
			vg.POST("/recoverToPublished", h.RecoverToPublished)
		}
		if permMiddleware != nil {
			vg.POST("/saveCanvas", permMiddleware.CheckVisualizationParentEdit(), h.SaveCanvas)
			vg.POST("/findById", permMiddleware.CheckVisualizationView(), h.FindByID)
			vg.POST("/updateCanvas", permMiddleware.CheckVisualizationEdit(), h.UpdateCanvas)
			vg.POST("/deleteLogic/:id", permMiddleware.CheckVisualizationEdit(), h.DeleteLogic)
			vg.POST("/deleteLogic/:id/:busiFlag", permMiddleware.CheckVisualizationEdit(), h.DeleteLogic)
		} else {
			vg.POST("/saveCanvas", h.SaveCanvas)
			vg.POST("/findById", h.FindByID)
			vg.POST("/updateCanvas", h.UpdateCanvas)
			vg.POST("/deleteLogic/:id", h.DeleteLogic)
			vg.POST("/deleteLogic/:id/:busiFlag", h.DeleteLogic)
		}
		vg.GET("/findCopyResource/:dvId/:busiFlag", h.FindCopyResource)
		vg.GET("/viewDetailList/:dvId", h.ViewDetailList)
		vg.POST("/appCanvasNameCheck", h.AppCanvasNameCheck)
		vg.POST("/decompression", h.Decompression)
		vg.POST("/decompressionLocalFile", h.DecompressionLocalFile)
		vg.POST("/export2AppCheck", h.Export2AppCheck)
		vg.POST("/exportLogApp", h.ExportLogApp)
		vg.POST("/exportLogTemplate", h.ExportLogTemplate)
		vg.POST("/exportLogPDF", h.ExportLogPDF)
		vg.POST("/exportLogImg", h.ExportLogImg)
	}
}
