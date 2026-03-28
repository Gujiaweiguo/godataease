package service

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/permission"
)

func TestRunGovernanceBackfillWithOptions(t *testing.T) {
	t.Run("returns fetch error", func(t *testing.T) {
		wantErr := errors.New("fetch failed")
		_, err := runGovernanceBackfillWithOptions(nil, permission.ResourceTypeDataset, func(options GovernanceBackfillOptions) ([]*int, error) {
			if options.Limit != DefaultGovernanceBackfillLimit {
				t.Fatalf("expected default limit %d, got %d", DefaultGovernanceBackfillLimit, options.Limit)
			}
			return nil, wantErr
		}, func(item *int) governanceBackfillItem {
			return governanceBackfillItem{}
		}, func(parentID, resourceID int64, resourceName string) (bool, error) {
			return false, nil
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected fetch error %v, got %v", wantErr, err)
		}
	})

	t.Run("builds report from normalized options and empty items", func(t *testing.T) {
		orgID := int64(9)
		report, err := runGovernanceBackfillWithOptions(&GovernanceBackfillOptions{AfterID: -5, Limit: 0, OrgID: &orgID}, "", func(options GovernanceBackfillOptions) ([]*int, error) {
			if options.AfterID != 0 {
				t.Fatalf("expected normalized afterID 0, got %d", options.AfterID)
			}
			if options.Limit != DefaultGovernanceBackfillLimit {
				t.Fatalf("expected normalized limit %d, got %d", DefaultGovernanceBackfillLimit, options.Limit)
			}
			return []*int{nil}, nil
		}, func(item *int) governanceBackfillItem {
			return governanceBackfillItem{resourceID: int64(*item)}
		}, func(parentID, resourceID int64, resourceName string) (bool, error) {
			return false, nil
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if report.ResourceType != permission.ResourceTypeDatasource {
			t.Fatalf("expected default resource type %s, got %s", permission.ResourceTypeDatasource, report.ResourceType)
		}
		if report.OrgID == nil || *report.OrgID != orgID {
			t.Fatalf("expected orgID %d to be preserved", orgID)
		}
		if report.Scanned != 0 || report.Governed != 0 || report.Skipped != 0 {
			t.Fatalf("expected empty counters, got scanned=%d governed=%d skipped=%d", report.Scanned, report.Governed, report.Skipped)
		}
		if len(report.ResourceIDs) != 0 || len(report.SkippedItems) != 0 {
			t.Fatalf("expected empty report slices, got resourceIDs=%v skippedItems=%v", report.ResourceIDs, report.SkippedItems)
		}
	})
}

func TestNormalizeGovernanceBackfillOptions(t *testing.T) {
	t.Run("nil uses default limit", func(t *testing.T) {
		got := normalizeGovernanceBackfillOptions(nil)
		if got.Limit != DefaultGovernanceBackfillLimit {
			t.Fatalf("expected default limit %d, got %d", DefaultGovernanceBackfillLimit, got.Limit)
		}
		if got.AfterID != 0 {
			t.Fatalf("expected afterID 0, got %d", got.AfterID)
		}
	})

	t.Run("normalizes negative afterID and non-positive limit", func(t *testing.T) {
		orgID := int64(7)
		got := normalizeGovernanceBackfillOptions(&GovernanceBackfillOptions{AfterID: -9, Limit: 0, OrgID: &orgID})
		if got.AfterID != 0 {
			t.Fatalf("expected normalized afterID 0, got %d", got.AfterID)
		}
		if got.Limit != DefaultGovernanceBackfillLimit {
			t.Fatalf("expected normalized limit %d, got %d", DefaultGovernanceBackfillLimit, got.Limit)
		}
		if got.OrgID == nil || *got.OrgID != orgID {
			t.Fatalf("expected orgID %d to be preserved", orgID)
		}
	})
}

func TestExecuteGovernanceBackfill(t *testing.T) {
	parentID := int64(10)
	blockedParentID := int64(20)
	items := []governanceBackfillItem{
		{resourceID: 0, parentID: &parentID, resourceName: "ignored-invalid"},
		{resourceID: 101, parentID: nil, resourceName: "missing-parent"},
		{resourceID: 102, parentID: &blockedParentID, resourceName: "parent-not-governed"},
		{resourceID: 103, parentID: &parentID, resourceName: "governed"},
	}

	inheritCalls := 0
	report, err := executeGovernanceBackfill(permission.ResourceTypeDataset, GovernanceBackfillOptions{AfterID: 5, Limit: 3}, items, func(parentID, resourceID int64, resourceName string) (bool, error) {
		inheritCalls++
		if resourceID == 102 {
			return false, nil
		}
		if resourceID == 103 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inheritCalls != 2 {
		t.Fatalf("expected 2 inherit calls, got %d", inheritCalls)
	}
	if report.Scanned != 3 {
		t.Fatalf("expected scanned 3, got %d", report.Scanned)
	}
	if report.Governed != 1 {
		t.Fatalf("expected governed 1, got %d", report.Governed)
	}
	if report.Skipped != 2 {
		t.Fatalf("expected skipped 2, got %d", report.Skipped)
	}
	if report.NextAfterID != 103 {
		t.Fatalf("expected nextAfterID 103, got %d", report.NextAfterID)
	}
	if len(report.ResourceIDs) != 1 || report.ResourceIDs[0] != 103 {
		t.Fatalf("expected governed resourceIDs [103], got %#v", report.ResourceIDs)
	}
	if len(report.SkippedItems) != 2 {
		t.Fatalf("expected 2 skipped items, got %d", len(report.SkippedItems))
	}
	if report.SkippedItems[0].Reason != GovernanceBackfillSkipReasonMissingParent {
		t.Fatalf("expected first skip reason missing_parent, got %s", report.SkippedItems[0].Reason)
	}
	if report.SkippedItems[0].Remediation != GovernanceBackfillRemediationDataCleanup {
		t.Fatalf("expected first remediation data_cleanup, got %s", report.SkippedItems[0].Remediation)
	}
	if report.SkippedItems[1].Reason != GovernanceBackfillSkipReasonParentNotGoverned {
		t.Fatalf("expected second skip reason parent_not_governed, got %s", report.SkippedItems[1].Reason)
	}
	if report.SkippedItems[1].Remediation != GovernanceBackfillRemediationGovernParent {
		t.Fatalf("expected second remediation govern_parent, got %s", report.SkippedItems[1].Remediation)
	}
}

func TestExecuteGovernanceBackfill_ReturnsInheritError(t *testing.T) {
	parentID := int64(10)
	wantErr := errors.New("inherit failed")
	_, err := executeGovernanceBackfill(permission.ResourceTypeDatasource, GovernanceBackfillOptions{}, []governanceBackfillItem{{resourceID: 100, parentID: &parentID, resourceName: "broken"}}, func(parentID, resourceID int64, resourceName string) (bool, error) {
		return false, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected inherit error %v, got %v", wantErr, err)
	}
}

func TestBuildGovernanceBackfillItems(t *testing.T) {
	type sample struct {
		id   int64
		name string
	}
	parentID := int64(5)
	items := []*sample{{id: 1, name: "one"}, nil, {id: 2, name: "two"}}
	got := buildGovernanceBackfillItems(items, func(item *sample) governanceBackfillItem {
		return governanceBackfillItem{resourceID: item.id, parentID: &parentID, resourceName: item.name}
	})
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].resourceID != 1 || got[1].resourceID != 2 {
		t.Fatalf("unexpected resource ids %#v", got)
	}
}

func TestNewGovernanceBackfillReport_Defaults(t *testing.T) {
	orgID := int64(3)
	report := newGovernanceBackfillReport("", GovernanceBackfillOptions{AfterID: 11, Limit: 7, OrgID: &orgID})
	if report.ResourceType != permission.ResourceTypeDatasource {
		t.Fatalf("expected default resource type %s, got %s", permission.ResourceTypeDatasource, report.ResourceType)
	}
	if report.AfterID != 11 || report.Limit != 7 {
		t.Fatalf("unexpected report options afterID=%d limit=%d", report.AfterID, report.Limit)
	}
	if report.OrgID == nil || *report.OrgID != orgID {
		t.Fatalf("expected orgID %d to be preserved", orgID)
	}
	if report.RollbackBoundary != "current_request_batch" {
		t.Fatalf("unexpected rollback boundary %s", report.RollbackBoundary)
	}
	if report.RerunStrategy != "idempotent_recompute" {
		t.Fatalf("unexpected rerun strategy %s", report.RerunStrategy)
	}
	if len(report.ResourceIDs) != 0 || len(report.SkippedItems) != 0 {
		t.Fatalf("expected initialized empty slices, got resourceIDs=%v skippedItems=%v", report.ResourceIDs, report.SkippedItems)
	}
}

func TestGovernanceBackfillReportHelpers(t *testing.T) {
	t.Run("nil receiver and invalid IDs are ignored", func(t *testing.T) {
		var nilReport *GovernanceBackfillReport
		nilReport.observe(1)
		nilReport.addGoverned(1)
		nilReport.addSkipped(1, permission.ResourceTypeDatasource, 2, GovernanceBackfillSkipReasonInvalidResource)

		report := newGovernanceBackfillReport(permission.ResourceTypeDatasource, GovernanceBackfillOptions{})
		report.observe(0)
		report.addGoverned(0)
		report.addSkipped(0, permission.ResourceTypeDatasource, 2, GovernanceBackfillSkipReasonInvalidResource)
		if report.Scanned != 0 || report.Governed != 0 || report.Skipped != 0 {
			t.Fatalf("expected counters to stay zero, got scanned=%d governed=%d skipped=%d", report.Scanned, report.Governed, report.Skipped)
		}
	})

	t.Run("records observation and invalid-resource skip", func(t *testing.T) {
		report := newGovernanceBackfillReport(permission.ResourceTypeDashboard, GovernanceBackfillOptions{})
		report.observe(12)
		report.addGoverned(12)
		report.addSkipped(13, permission.ResourceTypeDashboard, 4, GovernanceBackfillSkipReasonInvalidResource)
		if report.Scanned != 1 || report.Governed != 1 || report.Skipped != 1 {
			t.Fatalf("unexpected counters scanned=%d governed=%d skipped=%d", report.Scanned, report.Governed, report.Skipped)
		}
		if report.NextAfterID != 12 {
			t.Fatalf("expected nextAfterID 12, got %d", report.NextAfterID)
		}
		if len(report.ResourceIDs) != 1 || report.ResourceIDs[0] != 12 {
			t.Fatalf("unexpected governed resourceIDs %v", report.ResourceIDs)
		}
		if len(report.SkippedItems) != 1 {
			t.Fatalf("expected one skipped item, got %d", len(report.SkippedItems))
		}
		if report.SkippedItems[0].Reason != GovernanceBackfillSkipReasonInvalidResource {
			t.Fatalf("expected invalid_resource reason, got %s", report.SkippedItems[0].Reason)
		}
		if report.SkippedItems[0].Remediation != GovernanceBackfillRemediationNeedsChange {
			t.Fatalf("expected needs_change remediation, got %s", report.SkippedItems[0].Remediation)
		}
	})
}
