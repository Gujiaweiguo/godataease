package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type stubUserOrgResolver struct {
	inOrg bool
	err   error
}

func (s *stubUserOrgResolver) IsUserInOrg(userID, orgID int64) (bool, error) {
	return s.inOrg, s.err
}

type stubDatasetOrgValidator struct {
	allowed bool
	err     error
}

func (s *stubDatasetOrgValidator) DatasetBelongsToOrg(datasetID, orgID int64) (bool, error) {
	return s.allowed, s.err
}

func TestPermissionMiddleware_RejectsMissingOrgContextWhenResolverConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	permMiddleware := NewPermissionMiddleware(service.NewResourcePermissionService(repo, adminChecker), service.NewExportPermissionService(nil, nil), adminChecker)
	permMiddleware.SetUserOrgResolver(&stubUserOrgResolver{inOrg: true})

	r := gin.New()
	r.POST("/dataset/check", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Next()
	}, permMiddleware.CheckDatasetView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/dataset/check", strings.NewReader(`{"datasetGroupId":9}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected http 403, got status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "70001" {
		t.Fatalf("expected response code 403, got %v", resp["code"])
	}
}

func TestPermissionMiddleware_RejectsCrossOrgUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	permMiddleware := NewPermissionMiddleware(service.NewResourcePermissionService(repo, adminChecker), service.NewExportPermissionService(nil, nil), adminChecker)
	permMiddleware.SetUserOrgResolver(&stubUserOrgResolver{inOrg: false})

	r := gin.New()
	r.POST("/dataset/check", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("org_id", int64(7))
		c.Next()
	}, permMiddleware.CheckDatasetView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/dataset/check", strings.NewReader(`{"datasetGroupId":9}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected http 403, got status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "70001" {
		t.Fatalf("expected response code 403, got %v", resp["code"])
	}
}

func TestPermissionMiddleware_AdminBypassesOrgValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockResourcePermRepo{hasPermission: false}
	adminChecker := NewDefaultAdminChecker([]int64{1})
	permMiddleware := NewPermissionMiddleware(service.NewResourcePermissionService(repo, adminChecker), service.NewExportPermissionService(nil, nil), adminChecker)
	permMiddleware.SetUserOrgResolver(&stubUserOrgResolver{inOrg: false})

	r := gin.New()
	r.POST("/dataset/check", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	}, permMiddleware.CheckDatasetView(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/dataset/check", strings.NewReader(`{"datasetGroupId":9}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected admin bypass, got status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRowPermissionMiddleware_RejectsMismatchedDatasetOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetRowPermissionAdminChecker(NewDefaultAdminChecker([]int64{}))
	SetRowPermissionDatasetOrgValidator(&stubDatasetOrgValidator{allowed: false})
	t.Cleanup(func() {
		SetRowPermissionAdminChecker(nil)
		SetRowPermissionDatasetOrgValidator(nil)
	})

	r := gin.New()
	r.POST("/row/check", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("org_id", int64(7))
		c.Set(DatasetIDKey, int64(9))
		c.Next()
	}, RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/row/check", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected http 403, got status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "70001" {
		t.Fatalf("expected dataset org rejection, got %v", resp)
	}
}

func TestRowPermissionMiddleware_AllowsMatchingDatasetOrgAndSeedsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetRowPermissionAdminChecker(NewDefaultAdminChecker([]int64{}))
	SetRowPermissionDatasetOrgValidator(&stubDatasetOrgValidator{allowed: true})
	t.Cleanup(func() {
		SetRowPermissionAdminChecker(nil)
		SetRowPermissionDatasetOrgValidator(nil)
	})

	r := gin.New()
	r.POST("/row/check", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("org_id", int64(7))
		c.Next()
	}, RowPermissionMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"datasetId":  GetRowPermissionDatasetID(c),
			"datasetIds": GetRowPermissionDatasetIDs(c),
			"orgId":      GetOrgID(c),
		})
	})

	req := httptest.NewRequest("POST", "/row/check", strings.NewReader(`{"datasetId":9}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected success, got status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		DatasetID  int64   `json:"datasetId"`
		DatasetIDs []int64 `json:"datasetIds"`
		OrgID      int64   `json:"orgId"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.DatasetID != 9 || len(resp.DatasetIDs) != 1 || resp.OrgID != 7 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestPermissionMiddleware_PreservesOrgContextForDownstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockResourcePermRepo{hasPermission: true}
	adminChecker := NewDefaultAdminChecker([]int64{})
	permMiddleware := NewPermissionMiddleware(service.NewResourcePermissionService(repo, adminChecker), service.NewExportPermissionService(nil, nil), adminChecker)
	permMiddleware.SetUserOrgResolver(&stubUserOrgResolver{inOrg: true})

	r := gin.New()
	r.POST("/dataset/check", func(c *gin.Context) {
		c.Set("user_id", uint64(100))
		c.Set("org_id", int64(7))
		c.Next()
	}, permMiddleware.CheckResourcePermission(permission.ResourceTypeDataset, permission.PermKeyView), func(c *gin.Context) {
		c.JSON(200, gin.H{"orgId": GetOrgID(c)})
	})

	req := httptest.NewRequest("POST", "/dataset/check", strings.NewReader(`{"datasetId":9}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected success, got status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"orgId":7`) {
		t.Fatalf("expected downstream org context, got %s", w.Body.String())
	}
}
