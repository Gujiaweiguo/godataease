package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// createTestMenuAuthMiddleware creates a MenuAuthMiddleware with nil services
// for testing cases that don't require service calls (admin bypass, validation errors)
func createTestMenuAuthMiddleware() *MenuAuthMiddleware {
	return &MenuAuthMiddleware{
		roleMenuService: nil,
		menuService:     nil,
	}
}

func TestCheckMenuAccess_NoRole(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/menu/:menuId", middleware.CheckMenuAccess(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/menu/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected status 403 for no role, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if resp["msg"] != "No role assigned" {
		t.Errorf("expected 'No role assigned' message, got %v", resp["msg"])
	}
}

func TestCheckMenuAccess_EmptyRole(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/menu/:menuId", func(c *gin.Context) {
		c.Set("role_ids", []int64{})
		c.Next()
	}, middleware.CheckMenuAccess(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/menu/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected status 403 for empty role, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if resp["msg"] != "No valid role" {
		t.Errorf("expected 'No valid role' message, got %v", resp["msg"])
	}
}

func TestCheckMenuAccess_InvalidRoleType(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/menu/:menuId", func(c *gin.Context) {
		c.Set("role_ids", "not-a-slice")
		c.Next()
	}, middleware.CheckMenuAccess(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/menu/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected status 403 for invalid role type, got %d", w.Code)
	}
}

func TestCheckMenuAccess_AdminBypass(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/menu/:menuId", func(c *gin.Context) {
		c.Set("role_ids", []int64{1}) // Admin role
		c.Next()
	}, middleware.CheckMenuAccess(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/menu/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 for admin bypass, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("expected success true, got %v", resp["success"])
	}
}

func TestCheckMenuAccess_AdminBypassWithMultipleRoles(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/menu/:menuId", func(c *gin.Context) {
		c.Set("role_ids", []int64{2, 3, 1, 4}) // Contains admin role
		c.Next()
	}, middleware.CheckMenuAccess(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/menu/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 for admin bypass with multiple roles, got %d", w.Code)
	}
}

func TestCheckMenuAccess_NoMenuId_Allowed(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	// When no menuId is specified, the middleware allows the request through
	r.GET("/menu", func(c *gin.Context) {
		c.Set("role_ids", []int64{2, 3}) // Non-admin roles
		c.Next()
	}, middleware.CheckMenuAccess(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/menu", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 when no menuId specified, got %d", w.Code)
	}
}

func TestCheckMenuAccess_MenuIdFromQuery(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/menu", func(c *gin.Context) {
		c.Set("role_ids", []int64{1}) // Admin role to bypass auth check
		c.Next()
	}, middleware.CheckMenuAccess(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/menu?menuId=456", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCheckMenuAccess_InvalidMenuId(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/menu/:menuId", func(c *gin.Context) {
		c.Set("role_ids", []int64{2, 3}) // Non-admin roles
		c.Next()
	}, middleware.CheckMenuAccess(), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/menu/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		// Invalid menu ID returns an error response
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
		if resp["code"] != "500000" {
			t.Errorf("expected error code 500000 for invalid menu ID, got %v", resp["code"])
		}
	}
}

func TestRequireMenuAuth_NoRole(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/protected", middleware.RequireMenuAuth("/dashboard"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected status 403 for no role, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if resp["msg"] != "No role assigned" {
		t.Errorf("expected 'No role assigned' message, got %v", resp["msg"])
	}
}

func TestRequireMenuAuth_EmptyRole(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/protected", func(c *gin.Context) {
		c.Set("role_ids", []int64{})
		c.Next()
	}, middleware.RequireMenuAuth("/dashboard"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected status 403 for empty role, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if resp["msg"] != "No valid role" {
		t.Errorf("expected 'No valid role' message, got %v", resp["msg"])
	}
}

func TestRequireMenuAuth_AdminBypass(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/protected", func(c *gin.Context) {
		c.Set("role_ids", []int64{1}) // Admin role
		c.Next()
	}, middleware.RequireMenuAuth("/dashboard"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 for admin bypass, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("expected success true, got %v", resp["success"])
	}
}

func TestRequireMenuAuth_NonAdminDenied(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/protected", func(c *gin.Context) {
		c.Set("role_ids", []int64{2, 3}) // Non-admin roles
		c.Next()
	}, middleware.RequireMenuAuth("/dashboard"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected status 403 for non-admin, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	expectedMsg := "Menu access denied: /dashboard"
	if resp["msg"] != expectedMsg {
		t.Errorf("expected '%s' message, got %v", expectedMsg, resp["msg"])
	}
}

func TestRequireMenuAuth_AdminBypassWithMultipleRoles(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/protected", func(c *gin.Context) {
		c.Set("role_ids", []int64{2, 1, 3}) // Contains admin role
		c.Next()
	}, middleware.RequireMenuAuth("/dashboard"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 for admin bypass with multiple roles, got %d", w.Code)
	}
}

func TestRequireMenuAuth_InvalidRoleType(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	r.GET("/protected", func(c *gin.Context) {
		c.Set("role_ids", "not-a-slice")
		c.Next()
	}, middleware.RequireMenuAuth("/dashboard"), func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected status 403 for invalid role type, got %d", w.Code)
	}
}

func TestNewMenuAuthMiddleware(t *testing.T) {
	middleware := NewMenuAuthMiddleware(nil, nil)
	if middleware == nil {
		t.Error("expected non-nil middleware")
	}
}

// Integration-style tests that verify the middleware chain behavior
func TestCheckMenuAccess_MiddlewareChain(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	// Test that middleware properly chains to next handler
	r.GET("/menu/:menuId",
		func(c *gin.Context) {
			c.Set("role_ids", []int64{1})
			c.Next()
		},
		middleware.CheckMenuAccess(),
		func(c *gin.Context) {
			c.Set("middleware_executed", true)
			c.Next()
		},
		func(c *gin.Context) {
			executed, exists := c.Get("middleware_executed")
			if !exists || !executed.(bool) {
				c.JSON(500, gin.H{"error": "middleware chain broken"})
				return
			}
			c.JSON(200, gin.H{"success": true})
		},
	)

	req := httptest.NewRequest("GET", "/menu/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequireMenuAuth_MiddlewareChain(t *testing.T) {
	r := gin.New()
	middleware := createTestMenuAuthMiddleware()

	// Test that middleware properly chains to next handler for admin
	r.GET("/protected",
		func(c *gin.Context) {
			c.Set("role_ids", []int64{1})
			c.Next()
		},
		middleware.RequireMenuAuth("/admin"),
		func(c *gin.Context) {
			c.Set("middleware_executed", true)
			c.Next()
		},
		func(c *gin.Context) {
			executed, exists := c.Get("middleware_executed")
			if !exists || !executed.(bool) {
				c.JSON(500, gin.H{"error": "middleware chain broken"})
				return
			}
			c.JSON(200, gin.H{"success": true})
		},
	)

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
