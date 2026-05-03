package service

import (
	"context"
	"encoding/json"
	"strings"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
)

type ColumnPermissionService struct {
	columnPermRepo *repository.ColumnPermissionRepository
	cache          *permission.PermissionCacheService
}

const (
	defaultMaskValue     = "******"
	defaultRangeMaskText = "*** ***"
)

func NewColumnPermissionService(columnPermRepo *repository.ColumnPermissionRepository, cache *permission.PermissionCacheService) *ColumnPermissionService {
	return &ColumnPermissionService{columnPermRepo: columnPermRepo, cache: cache}
}

func (s *ColumnPermissionService) GetColumnPermissions(datasetID int64) ([]*permission.DataPermColumn, error) {
	if s.cache != nil {
		if perms, ok := s.cache.GetColumnPermissions(context.Background(), datasetID); ok {
			return perms, nil
		}
	}

	perms, err := s.columnPermRepo.ListByDatasetID(datasetID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		if err := s.cache.SetColumnPermissions(context.Background(), datasetID, perms); err != nil {
			logger.Warn("Failed to populate column permission cache", zap.Int64("datasetId", datasetID), zap.Error(err))
		}
	}

	return perms, nil
}

func (s *ColumnPermissionService) GetDisabledColumns(datasetID int64) (map[string]bool, error) {
	perms, err := s.GetColumnPermissions(datasetID)
	if err != nil {
		return nil, err
	}

	disabled := make(map[string]bool)
	for _, perm := range perms {
		if perm.PermType == permission.PermTypeDisable {
			disabled[perm.FieldName] = true
		}
	}
	return disabled, nil
}

func (s *ColumnPermissionService) GetMaskRules(datasetID int64) (map[string]*permission.DesensitizationRule, error) {
	perms, err := s.GetColumnPermissions(datasetID)
	if err != nil {
		return nil, err
	}

	rules := make(map[string]*permission.DesensitizationRule)
	for _, perm := range perms {
		if perm.PermType == permission.PermTypeMask {
			if perm.MaskRule == "" {
				rules[perm.FieldName] = &permission.DesensitizationRule{BuiltInRule: permission.BuiltInRuleCompleteDesensitization}
				continue
			}
			rule := s.parseMaskRule(perm.MaskRule)
			if rule != nil {
				rules[perm.FieldName] = rule
			}
		}
	}
	return rules, nil
}

func (s *ColumnPermissionService) parseMaskRule(maskRuleJSON string) *permission.DesensitizationRule {
	if maskRuleJSON == "" {
		return nil
	}

	var rule permission.DesensitizationRule
	if err := json.Unmarshal([]byte(maskRuleJSON), &rule); err != nil {
		return nil
	}
	return &rule
}

func (s *ColumnPermissionService) ApplyMask(value string, rule *permission.DesensitizationRule) string {
	if rule == nil {
		return defaultMaskValue
	}

	if rule.BuiltInRule != permission.BuiltInRuleCustom {
		return s.applyBuiltInRule(value, rule.BuiltInRule)
	}
	return s.applyCustomRule(value, rule)
}

func (s *ColumnPermissionService) applyBuiltInRule(value, builtInRule string) string {
	switch builtInRule {
	case permission.BuiltInRuleCompleteDesensitization:
		return defaultMaskValue
	case permission.BuiltInRuleKeepFirstAndLastThree:
		return s.keepFirstAndLastThree(value)
	case permission.BuiltInRuleKeepMiddleThree:
		return s.keepMiddleThree(value)
	default:
		return defaultMaskValue
	}
}

func (s *ColumnPermissionService) keepFirstAndLastThree(value string) string {
	if value == "" || len(value) < 7 {
		return "XXX***XXX"
	}
	return value[:3] + "***" + value[len(value)-3:]
}

func (s *ColumnPermissionService) keepMiddleThree(value string) string {
	if value == "" || len(value) < 4 {
		return "***XXX***"
	}
	mid := len(value) / 2
	return "***" + value[mid-1:mid+2] + "***"
}

func (s *ColumnPermissionService) applyCustomRule(value string, rule *permission.DesensitizationRule) string {
	if rule == nil {
		return defaultMaskValue
	}

	switch rule.CustomBuiltInRule {
	case permission.CustomRuleRetainBeforeMAndAfterN:
		return s.retainBeforeMAndAfterN(value, rule.M, rule.N)
	case permission.CustomRuleRetainMToN:
		return s.retainMToN(value, rule.M, rule.N)
	default:
		return defaultMaskValue
	}
}

func (s *ColumnPermissionService) retainBeforeMAndAfterN(value string, m, n int) string {
	if m <= 0 && n <= 0 {
		return defaultMaskValue
	}
	if m < 0 {
		m = 0
	}
	if n < 0 {
		n = 0
	}

	if value == "" || len(value) < m+n {
		prefix := strings.Repeat("X", m)
		suffix := strings.Repeat("X", n)
		return prefix + "***" + suffix
	}

	return value[:m] + "***" + value[len(value)-n:]
}

func (s *ColumnPermissionService) retainMToN(value string, m, n int) string {
	if m <= 0 && n <= 0 {
		return defaultMaskValue
	}
	if m < 1 {
		m = 1
	}
	if n < m {
		return defaultRangeMaskText
	}
	if value == "" || len(value) < m {
		return defaultRangeMaskText
	}

	endIdx := n
	if endIdx > len(value) {
		endIdx = len(value)
	}

	if m == 1 {
		return value[0:endIdx] + "***"
	}
	return "***" + value[m-1:endIdx] + "***"
}

func (s *ColumnPermissionService) MaskRowData(row map[string]interface{}, maskRules map[string]*permission.DesensitizationRule) map[string]interface{} {
	if len(maskRules) == 0 {
		return row
	}

	result := make(map[string]interface{})
	for k, v := range row {
		if rule, ok := maskRules[k]; ok {
			strVal := ""
			if v != nil {
				strVal = toString(v)
			}
			result[k] = s.ApplyMask(strVal, rule)
		} else {
			result[k] = v
		}
	}
	return result
}

func (s *ColumnPermissionService) FilterDisabledColumns(row map[string]interface{}, disabledColumns map[string]bool) map[string]interface{} {
	if len(disabledColumns) == 0 {
		return row
	}

	result := make(map[string]interface{})
	for k, v := range row {
		if !disabledColumns[k] {
			result[k] = v
		}
	}
	return result
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return ""
	}
}
