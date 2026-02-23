package repository

import (
	"dataease/backend/internal/domain/permission"
	"encoding/json"

	"gorm.io/gorm"
)

type RowPermissionRepository struct {
	db *gorm.DB
}

func NewRowPermissionRepository(db *gorm.DB) *RowPermissionRepository {
	return &RowPermissionRepository{db: db}
}

func (r *RowPermissionRepository) ListByDatasetID(datasetID int64) ([]*permission.DataPermRow, error) {
	var perms []*permission.DataPermRow
	err := r.db.Where("dataset_group_id = ? AND status = 1", datasetID).
		Find(&perms).Error
	return perms, err
}

func (r *RowPermissionRepository) ListByDatasetIDAndUserID(datasetID, userID int64) ([]*permission.DataPermRow, error) {
	var perms []*permission.DataPermRow
	err := r.db.Where("dataset_group_id = ? AND auth_target_type = ? AND auth_target_id = ? AND status = 1",
		datasetID, permission.AuthTargetTypeUser, userID).
		Find(&perms).Error
	return perms, err
}

func (r *RowPermissionRepository) ListByDatasetIDAndRoleIDs(datasetID int64, roleIDs []int64) ([]*permission.DataPermRow, error) {
	if len(roleIDs) == 0 {
		return []*permission.DataPermRow{}, nil
	}
	var perms []*permission.DataPermRow
	err := r.db.Where("dataset_group_id = ? AND auth_target_type = ? AND auth_target_id IN ? AND status = 1",
		datasetID, permission.AuthTargetTypeRole, roleIDs).
		Find(&perms).Error
	return perms, err
}

func (r *RowPermissionRepository) GetUserPermissions(datasetID, userID int64) ([]*permission.DataPermRow, error) {
	userPerms, err := r.ListByDatasetIDAndUserID(datasetID, userID)
	if err != nil {
		return nil, err
	}
	return userPerms, nil
}

func (r *RowPermissionRepository) ParseExpressionTree(exprTree string) (*permission.DatasetRowPermissionsTreeObj, error) {
	if exprTree == "" {
		return nil, nil
	}
	var obj permission.DatasetRowPermissionsTreeObj
	if err := json.Unmarshal([]byte(exprTree), &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

type ColumnPermissionRepository struct {
	db *gorm.DB
}

func NewColumnPermissionRepository(db *gorm.DB) *ColumnPermissionRepository {
	return &ColumnPermissionRepository{db: db}
}

func (r *ColumnPermissionRepository) ListByDatasetID(datasetID int64) ([]*permission.DataPermColumn, error) {
	var perms []*permission.DataPermColumn
	err := r.db.Where("dataset_group_id = ? AND status = 1", datasetID).
		Find(&perms).Error
	return perms, err
}

func (r *ColumnPermissionRepository) ListAllColumnNamesByDatasetID(datasetID int64) ([]string, error) {
	var names []string
	err := r.db.Model(&permission.DataPermColumn{}).
		Where("dataset_group_id = ?", datasetID).
		Distinct("field_name").
		Pluck("field_name", &names).Error
	return names, err
}
