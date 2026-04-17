package handler

import (
	"encoding/json"
	"strconv"
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
