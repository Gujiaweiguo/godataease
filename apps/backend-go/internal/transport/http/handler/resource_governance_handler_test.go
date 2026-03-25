package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type fakeResourceGovernanceService struct {
	report *service.GovernanceBackfillReport
	err    error
	last   *service.ResourceGovernanceBackfillRequest
}

func (f *fakeResourceGovernanceService) BackfillResources(req *service.ResourceGovernanceBackfillRequest) (*service.GovernanceBackfillReport, error) {
	f.last = req
	return f.report, f.err
}

type fakeResourceGovernanceAdminChecker struct {
	admin bool
}

func (f *fakeResourceGovernanceAdminChecker) IsAdmin(userID int64) bool {
	return f.admin && userID > 0
}

func TestResourceGovernanceHandler_BackfillResources_AdminOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeResourceGovernanceService{report: &service.GovernanceBackfillReport{ResourceType: "visualization", Governed: 1}}
	h := NewResourceGovernanceHandler(svc, &fakeResourceGovernanceAdminChecker{admin: true})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
	})
	RegisterResourceGovernanceRoutes(r.Group("/api"), h)

	orgID := int64(6)
	body, _ := json.Marshal(map[string]any{"resourceType": "visualization", "afterId": 10, "limit": 5, "orgId": orgID})
	req := httptest.NewRequest(http.MethodPost, "/api/system/resource-governance/backfill", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if svc.last == nil || svc.last.ResourceType != "visualization" || svc.last.AfterID != 10 || svc.last.Limit != 5 || svc.last.OrgID == nil || *svc.last.OrgID != orgID {
		t.Fatalf("expected request to be forwarded, got %+v", svc.last)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"code":"000000"`)) {
		t.Fatalf("expected success envelope, got %s", w.Body.String())
	}
}

func TestResourceGovernanceHandler_BackfillResources_ReturnsSkippedItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeResourceGovernanceService{report: &service.GovernanceBackfillReport{
		ResourceType:     "datasource",
		RollbackBoundary: "current_request_batch",
		RerunStrategy:    "idempotent_recompute",
		Scanned:          2,
		Governed:         1,
		Skipped:          1,
		SkippedItems: []*service.GovernanceBackfillSkippedItem{{
			ResourceID:   101,
			ResourceType: "datasource",
			ParentID:     88,
			Reason:       service.GovernanceBackfillSkipReasonParentNotGoverned,
			Remediation:  service.GovernanceBackfillRemediationGovernParent,
		}},
	}}
	h := NewResourceGovernanceHandler(svc, &fakeResourceGovernanceAdminChecker{admin: true})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
	})
	RegisterResourceGovernanceRoutes(r.Group("/api"), h)

	body, _ := json.Marshal(map[string]any{"resourceType": "datasource"})
	req := httptest.NewRequest(http.MethodPost, "/api/system/resource-governance/backfill", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"skippedItems"`)) {
		t.Fatalf("expected skippedItems in response, got %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"resourceId":101`)) {
		t.Fatalf("expected skipped resource id in response, got %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"parentId":88`)) {
		t.Fatalf("expected skipped parent id in response, got %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"reason":"parent_not_governed"`)) {
		t.Fatalf("expected skip reason in response, got %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"remediation":"govern_parent"`)) {
		t.Fatalf("expected remediation in response, got %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"rollbackBoundary":"current_request_batch"`)) {
		t.Fatalf("expected rollback boundary in response, got %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"rerunStrategy":"idempotent_recompute"`)) {
		t.Fatalf("expected rerun strategy in response, got %s", w.Body.String())
	}
}

func TestResourceGovernanceHandler_BackfillResources_RejectsNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewResourceGovernanceHandler(&fakeResourceGovernanceService{}, &fakeResourceGovernanceAdminChecker{admin: false})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(2))
	})
	RegisterResourceGovernanceRoutes(r.Group("/api"), h)

	body, _ := json.Marshal(map[string]any{"resourceType": "datasource"})
	req := httptest.NewRequest(http.MethodPost, "/api/system/resource-governance/backfill", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
