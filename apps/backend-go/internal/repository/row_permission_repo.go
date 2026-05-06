package repository

import (
	"dataease/backend/internal/domain/permission"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type RowPermissionRepository struct {
	db *gorm.DB
}

func NewRowPermissionRepository(db *gorm.DB) *RowPermissionRepository {
	return &RowPermissionRepository{db: db}
}

func (r *RowPermissionRepository) DB() *gorm.DB {
	if r == nil {
		return nil
	}
	return r.db
}

func (r *RowPermissionRepository) ListByDatasetID(datasetID int64) ([]*permission.DataPermRow, error) {
	var perms []*permission.DataPermRow
	err := r.db.Where("dataset_id = ? AND status = 1", datasetID).
		Find(&perms).Error
	return perms, err
}

func (r *RowPermissionRepository) PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermRow, int64, error) {
	query := r.db.Model(&permission.DataPermRow{}).Where("dataset_id = ? AND status = 1", datasetID)
	return r.pageRows(query, page, size)
}

func (r *RowPermissionRepository) PagerByDatasetIDAndTarget(datasetID int64, targetType string, targetID int64, page, size int) ([]*permission.DataPermRow, int64, error) {
	query := r.db.Model(&permission.DataPermRow{}).
		Where("dataset_id = ? AND auth_target_type = ? AND auth_target_id = ? AND status = 1", datasetID, targetType, targetID)
	return r.pageRows(query, page, size)
}

func (r *RowPermissionRepository) pageRows(query *gorm.DB, page, size int) ([]*permission.DataPermRow, int64, error) {
	var (
		perms []*permission.DataPermRow
		total int64
	)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&perms).Error
	return perms, total, err
}

func (r *RowPermissionRepository) GetByID(id int64) (*permission.DataPermRow, error) {
	var perm permission.DataPermRow
	err := r.db.Where("id = ?", id).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *RowPermissionRepository) Create(perm *permission.DataPermRow) error {
	now := time.Now()
	if perm.DatasetGroupID == 0 {
		perm.DatasetGroupID = perm.DatasetID
	}
	if perm.CreateTime == nil {
		perm.CreateTime = &now
	}
	if perm.Status == 0 {
		perm.Status = 1
	}
	return r.db.Create(perm).Error
}

func (r *RowPermissionRepository) Update(perm *permission.DataPermRow) error {
	now := time.Now()
	if perm.DatasetGroupID == 0 {
		perm.DatasetGroupID = perm.DatasetID
	}
	perm.UpdateTime = &now
	return r.db.Save(perm).Error
}

func (r *RowPermissionRepository) Delete(id int64) error {
	now := time.Now()
	return r.db.Model(&permission.DataPermRow{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": 0, "update_time": &now}).Error
}

func (r *RowPermissionRepository) ListByDatasetIDAndUserID(datasetID, userID int64) ([]*permission.DataPermRow, error) {
	var perms []*permission.DataPermRow
	err := r.db.Where("dataset_id = ? AND auth_target_type = ? AND auth_target_id = ? AND status = 1",
		datasetID, permission.AuthTargetTypeUser, userID).
		Find(&perms).Error
	return perms, err
}

func (r *RowPermissionRepository) ListByDatasetIDAndRoleIDs(datasetID int64, roleIDs []int64) ([]*permission.DataPermRow, error) {
	if len(roleIDs) == 0 {
		return []*permission.DataPermRow{}, nil
	}
	var perms []*permission.DataPermRow
	err := r.db.Where("dataset_id = ? AND auth_target_type = ? AND auth_target_id IN ? AND status = 1",
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

func (r *ColumnPermissionRepository) DB() *gorm.DB {
	if r == nil {
		return nil
	}
	return r.db
}

func (r *ColumnPermissionRepository) ListByDatasetID(datasetID int64) ([]*permission.DataPermColumn, error) {
	var perms []*permission.DataPermColumn
	err := r.db.Where("dataset_id = ? AND status = 1", datasetID).
		Find(&perms).Error
	return perms, err
}

func (r *ColumnPermissionRepository) PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermColumn, int64, error) {
	var (
		perms []*permission.DataPermColumn
		total int64
	)

	query := r.db.Model(&permission.DataPermColumn{}).Where("dataset_id = ? AND status = 1", datasetID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&perms).Error
	return perms, total, err
}

func (r *ColumnPermissionRepository) GetByID(id int64) (*permission.DataPermColumn, error) {
	var perm permission.DataPermColumn
	err := r.db.Where("id = ?", id).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *ColumnPermissionRepository) Create(perm *permission.DataPermColumn) error {
	now := time.Now()
	if perm.DatasetGroupID == 0 {
		perm.DatasetGroupID = perm.DatasetID
	}
	if perm.CreateTime == nil {
		perm.CreateTime = &now
	}
	if perm.Status == 0 {
		perm.Status = 1
	}
	return r.db.Create(perm).Error
}

func (r *ColumnPermissionRepository) Update(perm *permission.DataPermColumn) error {
	now := time.Now()
	if perm.DatasetGroupID == 0 {
		perm.DatasetGroupID = perm.DatasetID
	}
	perm.UpdateTime = &now
	return r.db.Save(perm).Error
}

func (r *ColumnPermissionRepository) Delete(id int64) error {
	now := time.Now()
	return r.db.Model(&permission.DataPermColumn{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": 0, "update_time": &now}).Error
}

func (r *ColumnPermissionRepository) ListAllColumnNamesByDatasetID(datasetID int64) ([]string, error) {
	var names []string
	err := r.db.Model(&permission.DataPermColumn{}).
		Where("dataset_id = ?", datasetID).
		Distinct("field_name").
		Pluck("field_name", &names).Error
	return names, err
}
