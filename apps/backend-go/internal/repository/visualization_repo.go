package repository

import (
	"time"

	"dataease/backend/internal/domain/visualization"

	"gorm.io/gorm"
)

type VisualizationRepository struct {
	db *gorm.DB
}

func NewVisualizationRepository(db *gorm.DB) *VisualizationRepository {
	return &VisualizationRepository{db: db}
}

func (r *VisualizationRepository) Create(v *visualization.DataVisualizationInfo) error {
	return r.db.Create(v).Error
}

func (r *VisualizationRepository) Update(v *visualization.DataVisualizationInfo) error {
	return r.db.Save(v).Error
}

func (r *VisualizationRepository) GetByID(id int64) (*visualization.DataVisualizationInfo, error) {
	var item visualization.DataVisualizationInfo
	err := r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("id = ? AND COALESCE(delete_flag, 0) = 0", id).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *VisualizationRepository) DeleteLogic(id int64, deletedBy string) error {
	now := time.Now().UnixMilli()
	return r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"delete_flag": true,
			"delete_time": now,
			"delete_by":   deletedBy,
			"update_time": now,
			"update_by":   deletedBy,
		}).Error
}

func (r *VisualizationRepository) Query(req *visualization.ListRequest) ([]*visualization.DataVisualizationInfo, int64, error) {
	var list []*visualization.DataVisualizationInfo
	var total int64

	page := req.Current
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	q := r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("COALESCE(delete_flag, 0) = 0")
	if req.Keyword != nil && *req.Keyword != "" {
		kw := "%" + *req.Keyword + "%"
		q = q.Where("name LIKE ?", kw)
	}
	if req.Type != nil && *req.Type != "" {
		q = q.Where("type = ?", *req.Type)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("update_time DESC").Offset(offset).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *VisualizationRepository) ListAllByTypes(types []string) ([]*visualization.DataVisualizationInfo, error) {
	var list []*visualization.DataVisualizationInfo

	q := r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("COALESCE(delete_flag, 0) = 0")
	if len(types) > 0 {
		q = q.Where("type IN ?", types)
	}

	err := q.Order("COALESCE(pid, 0) ASC").Order("COALESCE(sort, 0) ASC").Order("update_time DESC").Find(&list).Error
	return list, err
}

func (r *VisualizationRepository) CountByNameAndPID(name string, pid *int64, excludeID *int64) (int64, error) {
	var count int64
	normalizedPID := int64(0)
	if pid != nil {
		normalizedPID = *pid
	}

	query := r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("name = ? AND COALESCE(pid, 0) = ? AND COALESCE(delete_flag, 0) = 0", name, normalizedPID)
	if excludeID != nil && *excludeID > 0 {
		query = query.Where("id != ?", *excludeID)
	}

	err := query.Count(&count).Error
	return count, err
}
