package handler

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func parseDatasetWriteRequest(c *gin.Context, requireName bool) (*dataset.WriteRequest, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return nil, false
	}

	req := &dataset.WriteRequest{}
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
	if nodeType, ok := body["nodeType"].(string); ok {
		req.NodeType = nodeType
	}
	if dsType, ok := body["type"].(string); ok {
		tmp := dsType
		req.Type = &tmp
	}
	if isCross, ok := body["isCross"].(bool); ok {
		req.IsCross = &isCross
	}

	if requireName && strings.TrimSpace(req.Name) == "" {
		response.Error(c, response.CodeInternalError, "dataset name is required")
		return nil, false
	}

	return req, true
}

func parseDatasetIDs(c *gin.Context) ([]int64, bool) {
	var body map[string]interface{}
	err := c.ShouldBindBodyWith(&body, binding.JSON)
	if err != nil && !errors.Is(err, io.EOF) {
		var bodyList []interface{}
		if listErr := c.ShouldBindBodyWith(&bodyList, binding.JSON); listErr == nil {
			ids := make([]int64, 0, len(bodyList))
			for _, item := range bodyList {
				if id, ok := parseInt64Value(item); ok {
					ids = append(ids, id)
				}
			}
			return dedupeDatasetIDs(ids), true
		}
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return nil, false
	}

	ids := make([]int64, 0)
	if id, ok := parseInt64Value(body["id"]); ok {
		ids = append(ids, id)
	}
	if id, ok := parseInt64Value(body["datasetGroupId"]); ok {
		ids = append(ids, id)
	}
	if arr, ok := body["ids"].([]interface{}); ok {
		for _, item := range arr {
			if id, ok := parseInt64Value(item); ok {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return []int64{}, true
	}

	return dedupeDatasetIDs(ids), true
}

func dedupeDatasetIDs(ids []int64) []int64 {
	uniq := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		result = append(result, id)
	}

	return result
}

func parseEnumValueRequest(c *gin.Context) (*dataset.EnumValueRequest, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return nil, false
	}

	req := &dataset.EnumValueRequest{}
	if v, ok := parseInt64Value(body["queryId"]); ok {
		req.QueryID = v
	}
	if v, ok := parseInt64Value(body["displayId"]); ok {
		req.DisplayID = v
	}
	if v, ok := parseInt64Value(body["sortId"]); ok {
		req.SortID = v
	}
	if sort, ok := body["sort"].(string); ok {
		req.Sort = sort
	}
	if searchText, ok := body["searchText"].(string); ok {
		req.SearchText = searchText
	}
	if resultMode, ok := parseInt64Value(body["resultMode"]); ok {
		req.ResultMode = int(resultMode)
	}
	req.Filter = parseEnumFilters(body["filter"])

	if req.QueryID <= 0 {
		response.Success(c, []gin.H{})
		return nil, false
	}

	return req, true
}

func parseMultFieldValuesRequest(c *gin.Context) (*dataset.MultFieldValuesRequest, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return nil, false
	}

	req := &dataset.MultFieldValuesRequest{FieldIDs: make([]int64, 0), ResultMode: 0}
	if list, ok := body["fieldIds"].([]interface{}); ok {
		for _, item := range list {
			if id, idOK := parseInt64Value(item); idOK {
				req.FieldIDs = append(req.FieldIDs, id)
			}
		}
	}
	if resultMode, ok := parseInt64Value(body["resultMode"]); ok {
		req.ResultMode = int(resultMode)
	}
	req.Filter = parseEnumFilters(body["filter"])

	uniq := make(map[int64]struct{}, len(req.FieldIDs))
	uniqueIDs := make([]int64, 0, len(req.FieldIDs))
	for _, id := range req.FieldIDs {
		if id <= 0 {
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	req.FieldIDs = uniqueIDs

	return req, true
}

func parseEnumFieldID(c *gin.Context) (int64, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return 0, false
	}

	field, ok := body["field"].(map[string]interface{})
	if !ok {
		return 0, true
	}
	id, _ := parseInt64Value(field["id"])
	return id, true
}

func parseEnumFilters(v interface{}) []dataset.EnumFilter {
	items, ok := v.([]interface{})
	if !ok || len(items) == 0 {
		return []dataset.EnumFilter{}
	}

	filters := make([]dataset.EnumFilter, 0, len(items))
	for _, item := range items {
		obj, objOK := item.(map[string]interface{})
		if !objOK {
			continue
		}
		filter := dataset.EnumFilter{Value: make([]interface{}, 0)}
		if fieldID, exists := obj["fieldId"]; exists {
			switch val := fieldID.(type) {
			case string:
				filter.FieldID = strings.TrimSpace(val)
			default:
				if parsed, parsedOK := parseInt64Value(val); parsedOK {
					filter.FieldID = strconv.FormatInt(parsed, 10)
				}
			}
		}
		if op, ok := obj["operator"].(string); ok {
			filter.Operator = strings.TrimSpace(op)
		}
		if values, ok := obj["value"].([]interface{}); ok {
			filter.Value = values
		}
		filters = append(filters, filter)
	}
	return filters
}
