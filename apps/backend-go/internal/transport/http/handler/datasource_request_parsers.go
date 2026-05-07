package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type datasourceTreeNode struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	PID       string               `json:"pid,omitempty"`
	Type      string               `json:"type,omitempty"`
	Leaf      bool                 `json:"leaf"`
	Weight    int                  `json:"weight"`
	ExtraFlag int                  `json:"extraFlag"`
	Children  []datasourceTreeNode `json:"children,omitempty"`
}

func buildDatasourceTreeResponse(list []*datasource.CoreDatasource) []datasourceTreeNode {
	childrenByPID := make(map[int64][]datasourceTreeNode)
	for _, item := range list {
		if item == nil {
			continue
		}
		pid := int64(0)
		if item.PID != nil {
			pid = *item.PID
		}
		leaf := !strings.EqualFold(strings.TrimSpace(item.Type), datasource.TypeFolder)
		node := datasourceTreeNode{
			ID:        strconv.FormatInt(item.ID, 10),
			Name:      item.Name,
			PID:       strconv.FormatInt(pid, 10),
			Type:      item.Type,
			Leaf:      leaf,
			Weight:    9,
			ExtraFlag: 1,
			Children:  []datasourceTreeNode{},
		}
		childrenByPID[pid] = append(childrenByPID[pid], node)
	}

	var attach func(pid int64) []datasourceTreeNode
	attach = func(pid int64) []datasourceTreeNode {
		nodes := childrenByPID[pid]
		for i := range nodes {
			nodes[i].Children = attach(toInt64ID(nodes[i].ID))
			nodes[i].Leaf = len(nodes[i].Children) == 0 && !strings.EqualFold(strings.TrimSpace(nodes[i].Type), datasource.TypeFolder)
		}
		return nodes
	}

	root := datasourceTreeNode{
		ID:        "0",
		PID:       "0",
		Name:      "root",
		Leaf:      false,
		Weight:    9,
		ExtraFlag: 1,
		Children:  attach(0),
	}
	return []datasourceTreeNode{root}
}

func toInt64ID(id interface{}) int64 {
	switch v := id.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case uint64:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func sanitizeDatasourceResponse(ds *datasource.CoreDatasource, dsService *service.DatasourceService) gin.H {
	if ds == nil {
		return gin.H{}
	}
	pid := "0"
	if ds.PID != nil {
		pid = strconv.FormatInt(*ds.PID, 10)
	}

	creator := ""
	if ds.CreateBy != nil {
		creator = dsService.ResolveUserName(*ds.CreateBy)
	}
	updater := ""
	if ds.UpdateBy != nil {
		updater = dsService.ResolveUserName(strconv.FormatInt(*ds.UpdateBy, 10))
	}

	return gin.H{
		"id":             strconv.FormatInt(ds.ID, 10),
		"name":           ds.Name,
		"description":    ds.Description,
		"type":           ds.Type,
		"pid":            pid,
		"editType":       ds.EditType,
		"configuration":  ds.Configuration,
		"createTime":     ds.CreateTime,
		"updateTime":     ds.UpdateTime,
		"updateBy":       ds.UpdateBy,
		"createBy":       ds.CreateBy,
		"creator":        creator,
		"updater":        updater,
		"status":         ds.Status,
		"qrtzInstance":   ds.QrtzInstance,
		"taskStatus":     ds.TaskStatus,
		"enableDataFill": ds.EnableDataFill,
		"delFlag":        ds.DelFlag,
	}
}

func parseTableRequest(c *gin.Context) (*datasource.TableRequest, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return nil, false
	}

	request := &datasource.TableRequest{}
	if id, ok := parseInt64Value(body["datasourceId"]); ok {
		request.DatasourceID = id
	}
	if tableName, ok := body["tableName"].(string); ok {
		request.TableName = tableName
	}
	if limit, ok := parseInt64Value(body["limit"]); ok {
		request.Limit = int(limit)
	}

	return request, true
}

func parseRequestBody(c *gin.Context) (map[string]interface{}, bool) {
	body := make(map[string]interface{})
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return nil, false
	}
	if len(bytes.TrimSpace(raw)) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
			return nil, false
		}
	}

	return body, true
}

func parseEditType(body map[string]interface{}) *string {
	if editType, ok := body["editType"].(string); ok {
		return &editType
	}
	if editTypeNum, ok := parseInt64Value(body["editType"]); ok {
		tmp := strconv.FormatInt(editTypeNum, 10)
		return &tmp
	}

	return nil
}

func parseDatasourceConfiguration(c *gin.Context, body map[string]interface{}) (*string, bool) {
	if cfg, ok := body["configuration"].(string); ok {
		return &cfg, true
	}

	var rawConfig interface{}
	switch cfg := body["configuration"].(type) {
	case map[string]interface{}:
		rawConfig = cfg
	case []interface{}:
		rawConfig = cfg
	default:
		return nil, true
	}

	b, err := json.Marshal(rawConfig)
	if err != nil {
		response.Error(c, response.CodeInternalError, "Invalid configuration")
		return nil, false
	}
	tmp := string(b)
	return &tmp, true
}

func parseDatasourceWriteRequest(c *gin.Context, requireName bool) (*datasource.WriteRequest, bool) {
	body, ok := parseRequestBody(c)
	if !ok {
		return nil, false
	}

	req := &datasource.WriteRequest{}
	if id, ok := parseInt64Value(body["id"]); ok {
		req.ID = id
	}

	if pidVal, exists := body["pid"]; exists {
		if pid, ok := parseInt64Value(pidVal); ok {
			req.PID = &pid
		} else {
			pid := int64(0)
			req.PID = &pid
		}
	}

	if name, ok := body["name"].(string); ok {
		req.Name = name
	}
	if desc, ok := body["description"].(string); ok {
		req.Description = &desc
	} else if desc, ok := body["desc"].(string); ok {
		req.Description = &desc
	}
	if dsType, ok := body["type"].(string); ok {
		req.Type = dsType
	}
	if nodeType, ok := body["nodeType"].(string); ok {
		req.NodeType = nodeType
	}
	req.EditType = parseEditType(body)

	configuration, ok := parseDatasourceConfiguration(c, body)
	if !ok {
		return nil, false
	}
	req.Configuration = configuration

	if enable, ok := body["enableDataFill"].(bool); ok {
		req.EnableDataFill = &enable
	}

	if requireName && strings.TrimSpace(req.Name) == "" {
		response.Error(c, response.CodeInternalError, "datasource name is required")
		return nil, false
	}

	return req, true
}
