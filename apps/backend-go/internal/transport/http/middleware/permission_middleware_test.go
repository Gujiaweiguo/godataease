package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestExtractResourceID_FromPathParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.DELETE("/resource/:id", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("DELETE", "/resource/12345", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if id, ok := resp["id"].(float64); !ok || int64(id) != 12345 {
		t.Errorf("expected id 12345, got %v", resp["id"])
	}
}

func TestExtractResourceID_FromQueryParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/resource", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("GET", "/resource?id=54321", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if id, ok := resp["id"].(float64); !ok || int64(id) != 54321 {
		t.Errorf("expected id 54321, got %v", resp["id"])
	}
}

func TestExtractResourceID_FromResourceIdQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/resource", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("GET", "/resource?resourceId=99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if id, ok := resp["id"].(float64); !ok || int64(id) != 99999 {
		t.Errorf("expected id 99999, got %v", resp["id"])
	}
}

func TestExtractResourceID_FromJSONBody_Id(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/resource/action", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("POST", "/resource/action", strings.NewReader(`{"id":11111}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if id, ok := resp["id"].(float64); !ok || int64(id) != 11111 {
		t.Errorf("expected id 11111, got %v", resp["id"])
	}
}

func TestExtractResourceID_FromJSONBody_DatasetGroupId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/dataset/preview", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("POST", "/dataset/preview", strings.NewReader(`{"datasetGroupId":67890,"limit":100}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if id, ok := resp["id"].(float64); !ok || int64(id) != 67890 {
		t.Errorf("expected id 67890, got %v", resp["id"])
	}
}

func TestExtractResourceID_FromJSONBody_DatasetId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/data/preview", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("POST", "/data/preview", strings.NewReader(`{"datasetId":33333}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if id, ok := resp["id"].(float64); !ok || int64(id) != 33333 {
		t.Errorf("expected id 33333, got %v", resp["id"])
	}
}

func TestExtractResourceID_FromJSONBody_Array(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/datasetTree/detailWithPerm", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("POST", "/datasetTree/detailWithPerm", strings.NewReader(`[67890,67891]`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if id, ok := resp["id"].(float64); !ok || int64(id) != 67890 {
		t.Errorf("expected id 67890, got %v", resp["id"])
	}
}

func TestExtractResourceID_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/resource/action", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("POST", "/resource/action", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected status 400 for missing resource id, got %d", w.Code)
	}
}

func TestExtractResourceID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/resource/:id", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("GET", "/resource/not-a-number", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected status 400 for invalid resource id, got %d", w.Code)
	}
}

func TestExtractResourceID_Priority(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/resource/:id", func(c *gin.Context) {
		id, err := extractResourceID(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id})
	})

	req := httptest.NewRequest("POST", "/resource/100?resourceId=200", strings.NewReader(`{"id":300}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if id, ok := resp["id"].(float64); !ok || int64(id) != 100 {
		t.Errorf("expected id 100 (path param should have priority), got %v", resp["id"])
	}
}

func TestParseResourceIDFromAny_Float64(t *testing.T) {
	id, err := parseResourceIDFromAny(float64(12345))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 12345 {
		t.Errorf("expected 12345, got %d", id)
	}
}

func TestParseResourceIDFromAny_Int64(t *testing.T) {
	id, err := parseResourceIDFromAny(int64(54321))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 54321 {
		t.Errorf("expected 54321, got %d", id)
	}
}

func TestParseResourceIDFromAny_Int(t *testing.T) {
	id, err := parseResourceIDFromAny(99999)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 99999 {
		t.Errorf("expected 99999, got %d", id)
	}
}

func TestParseResourceIDFromAny_String(t *testing.T) {
	id, err := parseResourceIDFromAny("77777")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 77777 {
		t.Errorf("expected 77777, got %d", id)
	}
}

func TestParseResourceIDFromAny_InvalidType(t *testing.T) {
	_, err := parseResourceIDFromAny([]int{1, 2, 3})
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestParseResourceID(t *testing.T) {
	id, err := parseResourceID("12345")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 12345 {
		t.Errorf("expected 12345, got %d", id)
	}
}

func TestParseResourceID_Invalid(t *testing.T) {
	_, err := parseResourceID("not-a-number")
	if err == nil {
		t.Error("expected error for invalid number")
	}
}

func TestCheckResourcePermission_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminChecker := NewDefaultAdminChecker([]int64{1})
	permMiddleware := NewPermissionMiddleware(nil, nil, adminChecker)

	r := gin.New()
	r.GET("/protected", permMiddleware.CheckResourcePermission("dataset", "view"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/protected?id=123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected status 401 for unauthenticated request, got %d", w.Code)
	}
}

func TestCheckResourcePermission_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminChecker := NewDefaultAdminChecker([]int64{1})
	permMiddleware := NewPermissionMiddleware(nil, nil, adminChecker)

	r := gin.New()
	r.GET("/protected", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckResourcePermission("dataset", "view"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/protected?id=123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 for admin bypass, got %d", w.Code)
	}
}
