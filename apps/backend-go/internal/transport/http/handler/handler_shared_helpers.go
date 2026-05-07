package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func parseInt64Value(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case json.Number:
		parsed, err := n.Int64()
		if err == nil {
			return parsed, true
		}
		fallback, fallbackErr := strconv.ParseFloat(n.String(), 64)
		if fallbackErr != nil {
			return 0, false
		}
		return int64(fallback), true
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case string:
		parsed, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func parsePageParams(c *gin.Context) (int, int, bool) {
	page, err := strconv.Atoi(strings.TrimSpace(firstNonEmptyParam(c.Param("page"), c.Param("current"))))
	if err != nil || page < 1 {
		response.Error(c, response.CodeInternalError, errInvalidPage)
		return 0, 0, false
	}
	size, err := strconv.Atoi(strings.TrimSpace(firstNonEmptyParam(c.Param("limit"), c.Param("size"))))
	if err != nil || size < 1 {
		response.Error(c, response.CodeInternalError, "Invalid size")
		return 0, 0, false
	}
	return page, size, true
}

func parseIDParam(c *gin.Context, key string) (int64, bool) {
	return parseIDParamMsg(c, key, errInvalidID)
}

func parseIDParamBadRequest(c *gin.Context, key string) (int64, bool) {
	return parseIDParamMsgBadRequest(c, key, errInvalidID)
}

func parseIDParamMsg(c *gin.Context, key, errMsg string) (int64, bool) {
	value := strings.TrimSpace(c.Param(key))
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, response.CodeInternalError, errMsg)
		return 0, false
	}
	return id, true
}

func parseIDParamMsgBadRequest(c *gin.Context, key, errMsg string) (int64, bool) {
	value := strings.TrimSpace(c.Param(key))
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, errMsg)
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
