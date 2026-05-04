package repository

import (
	"context"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"gorm.io/gorm"
)

type CommitLogRepo interface {
	Create(ctx context.Context, log *datafillingdomain.DfCommitLog) error
	ListByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DfCommitLog, int64, error)
	DeleteByFormID(ctx context.Context, formID int64) error
}

var _ CommitLogRepo = (*CommitLogRepository)(nil)

type CommitLogRepository struct {
	db *gorm.DB
}

func NewCommitLogRepository(db *gorm.DB) *CommitLogRepository {
	return &CommitLogRepository{db: db}
}

func (r *CommitLogRepository) Create(ctx context.Context, log *datafillingdomain.DfCommitLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *CommitLogRepository) ListByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DfCommitLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&datafillingdomain.DfCommitLog{}).Where("form_id = ?", formID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]*datafillingdomain.DfCommitLog, 0)
	offset := 0
	if page > 1 {
		offset = (page - 1) * pageSize
	}
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *CommitLogRepository) DeleteByFormID(ctx context.Context, formID int64) error {
	return r.db.WithContext(ctx).Where("form_id = ?", formID).Delete(&datafillingdomain.DfCommitLog{}).Error
}
