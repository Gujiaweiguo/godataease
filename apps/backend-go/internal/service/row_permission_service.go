package service

import (
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/logger"

	"go.uber.org/zap"
)

type RowPermissionService struct{}

func NewRowPermissionService() *RowPermissionService {
	return &RowPermissionService{}
}

func (s *RowPermissionService) GetRowPermissionsTree(datasetID, userID int64) (*permission.RowPermissionFilter, error) {
	logger.Info("GetRowPermissionsTree called",
		zap.Int64("datasetId", datasetID),
		zap.Int64("userId", userID),
	)

	return nil, nil
}

func (s *RowPermissionService) BuildWhereClause(filter *permission.RowPermissionFilter) (string, []interface{}, error) {
	if filter == nil || len(filter.Rules) == 0 {
		return "", nil, nil
	}

	logger.Warn("BuildWhereClause called but not fully implemented",
		zap.Int64("datasetId", filter.DatasetID),
		zap.Int("ruleCount", len(filter.Rules)),
	)

	return "", nil, nil
}

func (s *RowPermissionService) MergeRowPermissionFilters(filters []*permission.RowPermissionFilter) *permission.RowPermissionFilter {
	if len(filters) == 0 {
		return nil
	}

	if len(filters) == 1 {
		return filters[0]
	}

	merged := &permission.RowPermissionFilter{
		DatasetID: filters[0].DatasetID,
		UserID:    filters[0].UserID,
		Rules:     make([]permission.RowPermissionTree, 0),
	}

	for _, f := range filters {
		merged.Rules = append(merged.Rules, f.Rules...)
	}

	return merged
}

func (s *RowPermissionService) IsAdmin(userID int64) bool {
	logger.Debug("IsAdmin check", zap.Int64("userId", userID))
	return false
}

func (s *RowPermissionService) GetUserRoleIDs(userID int64) []int64 {
	logger.Debug("GetUserRoleIDs called", zap.Int64("userId", userID))
	return nil
}
