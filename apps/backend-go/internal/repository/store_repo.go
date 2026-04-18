package repository

import (
	"dataease/backend/internal/domain/auto"

	"gorm.io/gorm"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

type FavoriteRow struct {
	StoreID    int64  `json:"storeId"`
	ResourceID int64  `json:"resourceId"`
	Name       string `json:"name"`
	Type       int32  `json:"type"`
	Creator    string `json:"creator"`
	Editor     string `json:"lastEditor"`
	EditTime   int64  `json:"editTime"`
	ExtFlag    int32  `json:"extFlag"`
	ExtFlag1   int32  `json:"extFlag1"`
}

func (r *FavoriteRepository) IsFavorited(resourceID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&auto.CoreStore{}).
		Where("resource_id = ? AND uid = ?", resourceID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *FavoriteRepository) DeleteFavorite(resourceID, userID int64) error {
	return r.db.Where("resource_id = ? AND uid = ?", resourceID, userID).
		Delete(&auto.CoreStore{}).Error
}

func (r *FavoriteRepository) CreateFavorite(store *auto.CoreStore) error {
	return r.db.Create(store).Error
}

func (r *FavoriteRepository) QueryFavorites(userID int64, resourceType int32, keyword string) ([]FavoriteRow, error) {
	query := `
		SELECT
			s.id AS store_id,
			s.resource_id,
			v.name,
			v.type,
			v.create_by AS creator,
			v.update_by AS editor,
			v.update_time AS edit_time,
			0 AS ext_flag,
			0 AS ext_flag1
		FROM core_store s
		INNER JOIN data_visualization_info v ON s.resource_id = v.id
		WHERE s.uid = ?
		  AND s.resource_id IS NOT NULL`

	args := []interface{}{userID}
	if resourceType > 0 {
		query += " AND s.resource_type = ?"
		args = append(args, resourceType)
	}
	if keyword != "" {
		query += " AND LOWER(v.name) LIKE LOWER(CONCAT('%', ?, '%'))"
		args = append(args, keyword)
	}
	query += " ORDER BY v.update_time DESC"

	var rows []FavoriteRow
	err := r.db.Raw(query, args...).Scan(&rows).Error
	return rows, err
}
