package repository

import (
	"dataease/backend/internal/domain/embedded"
	"gorm.io/gorm"
)

type EmbeddedRepository struct {
	db *gorm.DB
}

func NewEmbeddedRepository(db *gorm.DB) *EmbeddedRepository {
	return &EmbeddedRepository{db: db}
}

func (r *EmbeddedRepository) Create(e *embedded.CoreEmbedded) error {
	return r.db.Create(e).Error
}

func (r *EmbeddedRepository) Update(e *embedded.CoreEmbedded) error {
	return r.db.Save(e).Error
}

func (r *EmbeddedRepository) Delete(id int64) error {
	return r.db.Delete(&embedded.CoreEmbedded{}, id).Error
}

func (r *EmbeddedRepository) DeleteBatch(ids []int64) error {
	return r.db.Delete(&embedded.CoreEmbedded{}, ids).Error
}

func (r *EmbeddedRepository) GetByID(id int64) (*embedded.CoreEmbedded, error) {
	var e embedded.CoreEmbedded
	err := r.db.First(&e, id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmbeddedRepository) GetByAppId(appId string) (*embedded.CoreEmbedded, error) {
	var e embedded.CoreEmbedded
	err := r.db.Where("app_id = ?", appId).First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmbeddedRepository) Query(keyword string, page, pageSize int) ([]*embedded.CoreEmbedded, int64, error) {
	var list []*embedded.CoreEmbedded
	var total int64

	db := r.db.Model(&embedded.CoreEmbedded{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	if err := db.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *EmbeddedRepository) ListDistinctDomains() ([]string, error) {
	var domains []string
	err := r.db.Model(&embedded.CoreEmbedded{}).
		Distinct("domain").
		Where("domain IS NOT NULL AND domain != ''").
		Pluck("domain", &domains).Error
	return domains, err
}
