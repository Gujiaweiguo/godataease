package handler

import (
	"fmt"
	"strconv"
	"strings"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type VisualizationHandler struct {
	service *service.VisualizationService
}

func NewVisualizationHandler(service *service.VisualizationService) *VisualizationHandler {
	return &VisualizationHandler{service: service}
}

func (h *VisualizationHandler) FindByID(c *gin.Context) {
	var req visualization.DetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.Detail(&req)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) List(c *gin.Context) {
	var req visualization.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
)

func (h *VisualizationHandler) Tree(c *gin.Context) {
	var req treeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
			ExtraFlag:  0,
			ExtraFlag1: 0,
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
	var req visualization.SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid ID")
		return
	}
	result, err := h.service.FindDvType(id)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) CheckCanvasChange(c *gin.Context) {
	var req visualization.CanvasChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	var req visualization.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	var req visualization.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	var req visualization.MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	var req visualization.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	var req visualization.DetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	updateBy := h.getUpdateBy(c)
	result, err := h.service.RecoverToPublished(req.ID, updateBy)
	if err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, result)
}

func (h *VisualizationHandler) DeleteLogic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "500000", "Invalid ID")
		return
	}

	updateBy := h.getUpdateBy(c)
	if err = h.service.DeleteLogic(id, updateBy); err != nil {
		response.Error(c, "500000", "Failed: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *VisualizationHandler) NameCheck(c *gin.Context) {
	var req visualization.NameCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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

func RegisterVisualizationRoutes(r *gin.RouterGroup, h *VisualizationHandler) {
	vg := r.Group("/dataVisualization")
	{
		vg.GET("/findDvType/:id", h.FindDvType)
		vg.POST("/tree", h.Tree)
		vg.POST("/nameCheck", h.NameCheck)
		vg.POST("/checkCanvasChange", h.CheckCanvasChange)
		vg.POST("/findById", h.FindByID)
		vg.POST("/list", h.List)
		vg.POST("/updateBase", h.UpdateBase)
		vg.POST("/move", h.Move)
		vg.POST("/updatePublishStatus", h.UpdatePublishStatus)
		vg.POST("/recoverToPublished", h.RecoverToPublished)
		vg.POST("/saveCanvas", h.SaveCanvas)
		vg.POST("/updateCanvas", h.UpdateCanvas)
		vg.POST("/deleteLogic/:id", h.DeleteLogic)
		vg.POST("/deleteLogic/:id/:busiFlag", h.DeleteLogic)
	}
}
