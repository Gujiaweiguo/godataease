package repository

import (
	"time"

	"dataease/backend/internal/domain/audit"

	"gorm.io/gorm"
)

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(log *audit.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditLogRepository) CreateBatch(logs []*audit.AuditLog) error {
	return r.db.CreateInBatches(logs, 100).Error
}

func (r *AuditLogRepository) GetByID(id int64) (*audit.AuditLog, error) {
	var log audit.AuditLog
	err := r.db.First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *AuditLogRepository) GetByUserID(userID int64, page, pageSize int) ([]*audit.AuditLog, int64, error) {
	var logs []*audit.AuditLog
	var total int64

	query := r.db.Model(&audit.AuditLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AuditLogRepository) Query(query *audit.AuditLogQuery) ([]*audit.AuditLog, int64, error) {
	var logs []*audit.AuditLog
	var total int64

	db := r.db.Model(&audit.AuditLog{})

	if query.UserID != nil {
		db = db.Where("user_id = ?", *query.UserID)
	}
	if query.Username != nil {
		db = db.Where("username LIKE ?", "%"+*query.Username+"%")
	}
	if query.ActionType != nil {
		db = db.Where("action_type = ?", *query.ActionType)
	}
	if query.ResourceType != nil {
		db = db.Where("resource_type = ?", *query.ResourceType)
	}
	if query.OrganizationID != nil {
		db = db.Where("organization_id = ?", *query.OrganizationID)
	}
	if query.StartTime != nil {
		db = db.Where("create_time >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("create_time <= ?", *query.EndTime)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := query.Page
	pageSize := query.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	if err := db.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AuditLogRepository) DeleteBeforeDate(beforeTime time.Time) (int64, error) {
	result := r.db.Where("create_time < ?", beforeTime).Delete(&audit.AuditLog{})
	return result.RowsAffected, result.Error
}

func (r *AuditLogRepository) GetByIDs(ids []int64) ([]*audit.AuditLog, error) {
	var logs []*audit.AuditLog
	if err := r.db.Where("id IN ?", ids).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

type LoginFailureRepository struct {
	db *gorm.DB
}

func NewLoginFailureRepository(db *gorm.DB) *LoginFailureRepository {
	return &LoginFailureRepository{db: db}
}

func (r *LoginFailureRepository) Create(failure *audit.LoginFailure) error {
	return r.db.Create(failure).Error
}

func (r *LoginFailureRepository) GetByUsername(username string, limit int) ([]*audit.LoginFailure, error) {
	var failures []*audit.LoginFailure
	err := r.db.Where("username = ?", username).
		Order("create_time DESC").
		Limit(limit).
		Find(&failures).Error
	return failures, err
}

func (r *LoginFailureRepository) CountSinceTime(username string, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&audit.LoginFailure{}).
		Where("username = ? AND create_time >= ?", username, since).
		Count(&count).Error
	return count, err
}

type AuditLogDetailRepository struct {
	db *gorm.DB
}

func NewAuditLogDetailRepository(db *gorm.DB) *AuditLogDetailRepository {
	return &AuditLogDetailRepository{db: db}
}

func (r *AuditLogDetailRepository) Create(detail *audit.AuditLogDetail) error {
	return r.db.Create(detail).Error
}

func (r *AuditLogDetailRepository) CreateBatch(details []*audit.AuditLogDetail) error {
	return r.db.CreateInBatches(details, 100).Error
}

func (r *AuditLogDetailRepository) GetByAuditLogID(auditLogID int64) ([]*audit.AuditLogDetail, error) {
	var details []*audit.AuditLogDetail
	err := r.db.Where("audit_log_id = ?", auditLogID).Find(&details).Error
	return details, err
}

func (r *AuditLogDetailRepository) DeleteByAuditLogID(auditLogID int64) error {
	return r.db.Where("audit_log_id = ?", auditLogID).Delete(&audit.AuditLogDetail{}).Error
}
