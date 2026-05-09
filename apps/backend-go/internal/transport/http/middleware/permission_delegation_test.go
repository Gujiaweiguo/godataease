package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// helper builds a PermissionMiddleware with a full mock stack for delegation tests.
func newDelegationTestMiddleware(hasPerm bool) *PermissionMiddleware {
	repo := &mockResourcePermRepo{hasPermission: hasPerm}
	adminChecker := NewDefaultAdminChecker([]int64{})
	resourcePermSvc := service.NewResourcePermissionService(repo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	return NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)
}

func TestPermissionMiddleware_CheckDelegates_ReturnNonNilHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	handlers := map[string]gin.HandlerFunc{
		"CheckDatasetBatchView": pm.CheckDatasetBatchView(),
		"CheckDatasetEdit":      pm.CheckDatasetEdit(),
		"CheckDatasetExport":    pm.CheckDatasetExport(),
		"CheckDashboardView":    pm.CheckDashboardView(),
		"CheckDashboardEdit":    pm.CheckDashboardEdit(),
		"CheckDashboardExport":  pm.CheckDashboardExport(),
		"CheckScreenView":       pm.CheckScreenView(),
		"CheckScreenEdit":       pm.CheckScreenEdit(),
		"CheckScreenExport":     pm.CheckScreenExport(),
		"CheckDatasourceView":   pm.CheckDatasourceView(),
		"CheckDatasourceEdit":   pm.CheckDatasourceEdit(),
	}

	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			assert.NotNil(t, handler, name+" should return a non-nil handler")
		})
	}
}

func TestPermissionMiddleware_CheckDelegates_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	resourcePermDelegates := map[string]gin.HandlerFunc{
		"CheckDatasetEdit":     pm.CheckDatasetEdit(),
		"CheckDatasetExport":   pm.CheckDatasetExport(),
		"CheckDashboardView":   pm.CheckDashboardView(),
		"CheckDashboardEdit":   pm.CheckDashboardEdit(),
		"CheckDashboardExport": pm.CheckDashboardExport(),
		"CheckScreenView":      pm.CheckScreenView(),
		"CheckScreenEdit":      pm.CheckScreenEdit(),
		"CheckScreenExport":    pm.CheckScreenExport(),
		"CheckDatasourceView":  pm.CheckDatasourceView(),
		"CheckDatasourceEdit":  pm.CheckDatasourceEdit(),
	}

	for name, handler := range resourcePermDelegates {
		t.Run(name, func(t *testing.T) {
			r := gin.New()
			r.GET("/test/:id", handler, func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)
			assert.Equal(t, http.StatusUnauthorized, resp.Code, name+" should reject unauthenticated requests with 401")
		})
	}
}

func TestPermissionMiddleware_CheckDatasetBatchView_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	r := gin.New()
	r.POST("/dataset/batchView", pm.CheckDatasetBatchView(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/dataset/batchView", strings.NewReader(`{"ids":[1,2,3]}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestPermissionMiddleware_CheckDatasetBatchView_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(false)

	r := gin.New()
	r.POST("/dataset/batchView", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, pm.CheckDatasetBatchView(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/dataset/batchView", strings.NewReader(`{"ids":[10,20]}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestPermissionMiddleware_CheckDatasetBatchView_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	r := gin.New()
	r.POST("/dataset/batchView", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, pm.CheckDatasetBatchView(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/dataset/batchView", strings.NewReader(`{"ids":[10,20]}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestPermissionMiddleware_CheckDatasetEdit_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(false)

	r := gin.New()
	r.POST("/dataset/edit/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, pm.CheckDatasetEdit(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/dataset/edit/999", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestPermissionMiddleware_CheckDashboardExport_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	r := gin.New()
	r.GET("/dashboard/export/:id", pm.CheckDashboardExport(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/dashboard/export/123", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestPermissionMiddleware_CheckDashboardExport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	r := gin.New()
	r.GET("/dashboard/export/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, pm.CheckDashboardExport(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/dashboard/export/456", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestPermissionMiddleware_CheckScreenEdit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	r := gin.New()
	r.POST("/screen/edit/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, pm.CheckScreenEdit(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/screen/edit/789", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestPermissionMiddleware_CheckScreenEdit_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(false)

	r := gin.New()
	r.POST("/screen/edit/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, pm.CheckScreenEdit(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/screen/edit/789", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestPermissionMiddleware_CheckScreenExport_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	r := gin.New()
	r.GET("/screen/export/:id", pm.CheckScreenExport(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/screen/export/100", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestPermissionMiddleware_CheckScreenExport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	r := gin.New()
	r.GET("/screen/export/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, pm.CheckScreenExport(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/screen/export/100", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestPermissionMiddleware_CheckDatasourceEdit_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	r := gin.New()
	r.POST("/datasource/edit/:id", pm.CheckDatasourceEdit(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/datasource/edit/50", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestPermissionMiddleware_CheckDatasourceEdit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(true)

	r := gin.New()
	r.POST("/datasource/edit/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, pm.CheckDatasourceEdit(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/datasource/edit/50", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestPermissionMiddleware_CheckDatasourceEdit_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pm := newDelegationTestMiddleware(false)

	r := gin.New()
	r.POST("/datasource/edit/:id", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, pm.CheckDatasourceEdit(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/datasource/edit/50", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestPermissionMiddleware_CheckDelegates_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{42})
	resourcePermSvc := service.NewResourcePermissionService(repo, adminChecker)
	exportPermSvc := service.NewExportPermissionService(resourcePermSvc, nil)
	pm := NewPermissionMiddleware(resourcePermSvc, exportPermSvc, adminChecker)

	delegates := map[string]gin.HandlerFunc{
		"CheckDatasetEdit":     pm.CheckDatasetEdit(),
		"CheckDashboardView":   pm.CheckDashboardView(),
		"CheckDashboardEdit":   pm.CheckDashboardEdit(),
		"CheckDashboardExport": pm.CheckDashboardExport(),
		"CheckScreenView":      pm.CheckScreenView(),
		"CheckScreenEdit":      pm.CheckScreenEdit(),
		"CheckScreenExport":    pm.CheckScreenExport(),
		"CheckDatasourceView":  pm.CheckDatasourceView(),
		"CheckDatasourceEdit":  pm.CheckDatasourceEdit(),
	}

	for name, handler := range delegates {
		t.Run(name, func(t *testing.T) {
			r := gin.New()
			r.GET("/test/:id", func(c *gin.Context) {
				c.Set("user_id", uint64(42))
				c.Next()
			}, handler, func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/test/1", nil)
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)
			assert.Equal(t, http.StatusOK, resp.Code, name+" should allow admin user")
		})
	}
}

func TestGetExportableResourceIDs_Set(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("exportable_resource_ids", []int64{1, 2, 3})
	ids := GetExportableResourceIDs(c)
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

func TestGetExportableResourceIDs_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	ids := GetExportableResourceIDs(c)
	assert.Nil(t, ids)
}

func TestGetExportableResourceIDs_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("exportable_resource_ids", []int64{})
	ids := GetExportableResourceIDs(c)
	assert.Equal(t, []int64{}, ids)
}
