package repository

import (
	"context"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"gorm.io/gorm"
)

type DataFillingRepositoryInterface interface {
	Create(ctx context.Context, form *datafillingdomain.DataFillingForm) error
	GetByID(ctx context.Context, id int64) (*datafillingdomain.DataFillingForm, error)
	Update(ctx context.Context, form *datafillingdomain.DataFillingForm) error
	DeleteByID(ctx context.Context, id int64) error
	Rename(ctx context.Context, id int64, name string) error
	Move(ctx context.Context, id int64, pid int64) error
	GetTree(ctx context.Context) ([]*datafillingdomain.DataFillingForm, error)
	GetByPID(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error)
	GetChildren(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error)
}

var _ DataFillingRepositoryInterface = (*DataFillingRepository)(nil)

type DataFillingRepository struct {
	db *gorm.DB
}

func NewDataFillingRepository(db *gorm.DB) *DataFillingRepository {
	return &DataFillingRepository{db: db}
}

func (r *DataFillingRepository) Create(ctx context.Context, form *datafillingdomain.DataFillingForm) error {
	return r.db.WithContext(ctx).Create(form).Error
}

func (r *DataFillingRepository) GetByID(ctx context.Context, id int64) (*datafillingdomain.DataFillingForm, error) {
	var form datafillingdomain.DataFillingForm
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&form).Error; err != nil {
		return nil, err
	}
	return &form, nil
}

func (r *DataFillingRepository) Update(ctx context.Context, form *datafillingdomain.DataFillingForm) error {
	return r.db.WithContext(ctx).Save(form).Error
}

func (r *DataFillingRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&datafillingdomain.DataFillingForm{}).Error
}

func (r *DataFillingRepository) Rename(ctx context.Context, id int64, name string) error {
	return r.db.WithContext(ctx).Model(&datafillingdomain.DataFillingForm{}).Where("id = ?", id).Update("name", name).Error
}

func (r *DataFillingRepository) Move(ctx context.Context, id int64, pid int64) error {
	level, err := r.resolveLevel(ctx, pid)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&datafillingdomain.DataFillingForm{}).Where("id = ?", id).Updates(map[string]any{
		"pid":   pid,
		"level": level,
	}).Error
}

func (r *DataFillingRepository) GetTree(ctx context.Context) ([]*datafillingdomain.DataFillingForm, error) {
	var rows []*datafillingdomain.DataFillingForm
	err := r.db.WithContext(ctx).
		Order("level ASC").
		Order("CASE WHEN node_type = 'folder' THEN 0 ELSE 1 END ASC").
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *DataFillingRepository) GetByPID(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error) {
	return r.GetChildren(ctx, pid)
}

func (r *DataFillingRepository) GetChildren(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error) {
	var rows []*datafillingdomain.DataFillingForm
	err := r.db.WithContext(ctx).
		Where("pid = ?", pid).
		Order("CASE WHEN node_type = 'folder' THEN 0 ELSE 1 END ASC").
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *DataFillingRepository) resolveLevel(ctx context.Context, pid int64) (int, error) {
	if pid <= 0 {
		return 0, nil
	}
	parent, err := r.GetByID(ctx, pid)
	if err != nil {
		return 0, err
	}
	return parent.Level + 1, nil
}
