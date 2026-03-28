package service

import "dataease/backend/internal/domain/permission"

const (
	DefaultGovernanceBackfillLimit = 100
	defaultLanguageZhCN            = "zh-CN"
	systemActor                    = "system"
)

type GovernanceBackfillOptions struct {
	AfterID int64  `json:"afterId"`
	Limit   int    `json:"limit"`
	OrgID   *int64 `json:"orgId,omitempty"`
}

type GovernanceBackfillSkipReason string

type GovernanceBackfillRemediation string

const (
	GovernanceBackfillSkipReasonMissingParent     GovernanceBackfillSkipReason = "missing_parent"
	GovernanceBackfillSkipReasonParentNotGoverned GovernanceBackfillSkipReason = "parent_not_governed"
	GovernanceBackfillSkipReasonInvalidResource   GovernanceBackfillSkipReason = "invalid_resource"
)

const (
	GovernanceBackfillRemediationDataCleanup  GovernanceBackfillRemediation = "data_cleanup"
	GovernanceBackfillRemediationGovernParent GovernanceBackfillRemediation = "govern_parent"
	GovernanceBackfillRemediationNeedsChange  GovernanceBackfillRemediation = "needs_change"
)

type GovernanceBackfillSkippedItem struct {
	ResourceID   int64                         `json:"resourceId"`
	ResourceType string                        `json:"resourceType"`
	ParentID     int64                         `json:"parentId,omitempty"`
	Reason       GovernanceBackfillSkipReason  `json:"reason"`
	Remediation  GovernanceBackfillRemediation `json:"remediation"`
}

type GovernanceBackfillReport struct {
	ResourceType     string                           `json:"resourceType"`
	AfterID          int64                            `json:"afterId"`
	Limit            int                              `json:"limit"`
	OrgID            *int64                           `json:"orgId,omitempty"`
	Scanned          int                              `json:"scanned"`
	Governed         int                              `json:"governed"`
	Skipped          int                              `json:"skipped"`
	NextAfterID      int64                            `json:"nextAfterId"`
	RollbackBoundary string                           `json:"rollbackBoundary"`
	RerunStrategy    string                           `json:"rerunStrategy"`
	ResourceIDs      []int64                          `json:"resourceIds"`
	SkippedItems     []*GovernanceBackfillSkippedItem `json:"skippedItems"`
}

type DatasourceGovernanceBackfillReport = GovernanceBackfillReport
type DatasetGovernanceBackfillReport = GovernanceBackfillReport
type VisualizationGovernanceBackfillReport = GovernanceBackfillReport

type governanceBackfillItem struct {
	resourceID   int64
	parentID     *int64
	resourceName string
}

func buildGovernanceBackfillItems[T any](items []*T, toItem func(*T) governanceBackfillItem) []governanceBackfillItem {
	backfillItems := make([]governanceBackfillItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		backfillItems = append(backfillItems, toItem(item))
	}
	return backfillItems
}

func runGovernanceBackfillWithOptions[T any](options *GovernanceBackfillOptions, resourceType string, fetch func(GovernanceBackfillOptions) ([]*T, error), toItem func(*T) governanceBackfillItem, inherit func(parentID, resourceID int64, resourceName string) (bool, error)) (*GovernanceBackfillReport, error) {
	normalized := normalizeGovernanceBackfillOptions(options)
	items, err := fetch(normalized)
	if err != nil {
		return nil, err
	}
	return executeGovernanceBackfill(resourceType, normalized, buildGovernanceBackfillItems(items, toItem), inherit)
}

func normalizeGovernanceBackfillOptions(options *GovernanceBackfillOptions) GovernanceBackfillOptions {
	if options == nil {
		return GovernanceBackfillOptions{Limit: DefaultGovernanceBackfillLimit}
	}
	normalized := *options
	if normalized.Limit <= 0 {
		normalized.Limit = DefaultGovernanceBackfillLimit
	}
	if normalized.AfterID < 0 {
		normalized.AfterID = 0
	}
	return normalized
}

func newGovernanceBackfillReport(resourceType string, options GovernanceBackfillOptions) *GovernanceBackfillReport {
	if resourceType == "" {
		resourceType = permission.ResourceTypeDatasource
	}
	return &GovernanceBackfillReport{
		ResourceType:     resourceType,
		AfterID:          options.AfterID,
		Limit:            options.Limit,
		OrgID:            options.OrgID,
		RollbackBoundary: "current_request_batch",
		RerunStrategy:    "idempotent_recompute",
		ResourceIDs:      make([]int64, 0),
		SkippedItems:     make([]*GovernanceBackfillSkippedItem, 0),
	}
}

func (r *GovernanceBackfillReport) observe(resourceID int64) {
	if r == nil || resourceID <= 0 {
		return
	}
	r.Scanned++
	r.NextAfterID = resourceID
}

func (r *GovernanceBackfillReport) addGoverned(resourceID int64) {
	if r == nil || resourceID <= 0 {
		return
	}
	r.Governed++
	r.ResourceIDs = append(r.ResourceIDs, resourceID)
}

func (r *GovernanceBackfillReport) addSkipped(resourceID int64, resourceType string, parentID int64, reason GovernanceBackfillSkipReason) {
	if r == nil || resourceID <= 0 {
		return
	}
	r.Skipped++
	r.SkippedItems = append(r.SkippedItems, &GovernanceBackfillSkippedItem{
		ResourceID:   resourceID,
		ResourceType: resourceType,
		ParentID:     parentID,
		Reason:       reason,
		Remediation:  remediationForGovernanceBackfillReason(reason),
	})
}

func remediationForGovernanceBackfillReason(reason GovernanceBackfillSkipReason) GovernanceBackfillRemediation {
	switch reason {
	case GovernanceBackfillSkipReasonMissingParent:
		return GovernanceBackfillRemediationDataCleanup
	case GovernanceBackfillSkipReasonParentNotGoverned:
		return GovernanceBackfillRemediationGovernParent
	default:
		return GovernanceBackfillRemediationNeedsChange
	}
}

func executeGovernanceBackfill(resourceType string, options GovernanceBackfillOptions, items []governanceBackfillItem, inherit func(parentID, resourceID int64, resourceName string) (bool, error)) (*GovernanceBackfillReport, error) {
	report := newGovernanceBackfillReport(resourceType, options)
	for _, item := range items {
		if item.resourceID <= 0 {
			continue
		}
		report.observe(item.resourceID)
		if item.parentID == nil || *item.parentID <= 0 {
			report.addSkipped(item.resourceID, resourceType, 0, GovernanceBackfillSkipReasonMissingParent)
			continue
		}
		inherited, err := inherit(*item.parentID, item.resourceID, item.resourceName)
		if err != nil {
			return nil, err
		}
		if !inherited {
			report.addSkipped(item.resourceID, resourceType, *item.parentID, GovernanceBackfillSkipReasonParentNotGoverned)
			continue
		}
		report.addGoverned(item.resourceID)
	}
	return report, nil
}
