package repository

import (
	"time"

	"dataease/backend/internal/domain/share"

	"gorm.io/gorm"
)

type coreShare struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Creator       int64      `gorm:"column:creator;index" json:"creator"`
	ResourceID    int64      `gorm:"column:resource_id;index" json:"resourceId"`
	ResourceType  string     `gorm:"column:resource_type;size:50" json:"resourceType"`
	Time          *time.Time `gorm:"column:time" json:"time"`
	Exp           int64      `gorm:"column:exp" json:"exp"`
	UUID          string     `gorm:"column:uuid;size:64;uniqueIndex" json:"uuid"`
	Pwd           string     `gorm:"column:pwd;size:255" json:"pwd"`
	AutoPwd       bool       `gorm:"column:auto_pwd;default:true" json:"autoPwd"`
	TicketRequire bool       `gorm:"column:ticket_require;default:false" json:"ticketRequire"`
}

func (coreShare) TableName() string {
	return "core_share"
}

type coreShareTicket struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID       string     `gorm:"column:uuid;size:64;index" json:"uuid"`
	Ticket     string     `gorm:"column:ticket;size:64;uniqueIndex" json:"ticket"`
	Exp        int64      `gorm:"column:exp" json:"exp"`
	Args       string     `gorm:"column:args;type:text" json:"args"`
	AccessTime *time.Time `gorm:"column:access_time" json:"accessTime"`
}

func (coreShareTicket) TableName() string {
	return "core_share_ticket"
}

type ShareRepository struct {
	db *gorm.DB
}

func NewShareRepository(db *gorm.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

func (r *ShareRepository) Create(s *share.Share) error {
	now := time.Now()
	record := coreShare{
		Creator:       s.Creator,
		ResourceID:    s.ResourceID,
		ResourceType:  s.ResourceType,
		Time:          &now,
		Exp:           s.Exp,
		UUID:          s.UUID,
		Pwd:           s.Pwd,
		AutoPwd:       s.AutoPwd,
		TicketRequire: s.TicketRequire,
	}
	if err := r.db.Create(&record).Error; err != nil {
		return err
	}
	s.ID = record.ID
	if record.Time != nil {
		s.Time = record.Time
	}
	return nil
}

func (r *ShareRepository) GetByID(id int64) (*share.Share, error) {
	var record coreShare
	if err := r.db.Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return r.toShare(record), nil
}

func (r *ShareRepository) GetByUUID(uuid string) (*share.Share, error) {
	var record coreShare
	if err := r.db.Where("uuid = ?", uuid).First(&record).Error; err != nil {
		return nil, err
	}
	return r.toShare(record), nil
}

func (r *ShareRepository) GetByResourceID(resourceID int64) (*share.Share, error) {
	var record coreShare
	if err := r.db.Where("resource_id = ?", resourceID).First(&record).Error; err != nil {
		return nil, err
	}
	return r.toShare(record), nil
}

func (r *ShareRepository) Update(s *share.Share) error {
	return r.db.Model(&coreShare{}).Where("id = ?", s.ID).Updates(map[string]interface{}{
		"exp":            s.Exp,
		"pwd":            s.Pwd,
		"auto_pwd":       s.AutoPwd,
		"ticket_require": s.TicketRequire,
	}).Error
}

func (r *ShareRepository) Delete(id int64) error {
	return r.db.Delete(&coreShare{}, id).Error
}

func (r *ShareRepository) toShare(record coreShare) *share.Share {
	return &share.Share{
		ID:            record.ID,
		Creator:       record.Creator,
		ResourceID:    record.ResourceID,
		ResourceType:  record.ResourceType,
		Time:          record.Time,
		Exp:           record.Exp,
		UUID:          record.UUID,
		Pwd:           record.Pwd,
		AutoPwd:       record.AutoPwd,
		TicketRequire: record.TicketRequire,
	}
}

func (r *ShareRepository) CreateTicket(t *share.ShareTicket) error {
	record := coreShareTicket{
		UUID:       t.UUID,
		Ticket:     t.Ticket,
		Exp:        t.Exp,
		Args:       t.Args,
		AccessTime: t.AccessTime,
	}
	if err := r.db.Create(&record).Error; err != nil {
		return err
	}
	t.ID = record.ID
	return nil
}

func (r *ShareRepository) GetTicketByUUID(uuid string) (*share.ShareTicket, error) {
	var record coreShareTicket
	if err := r.db.Where("uuid = ?", uuid).First(&record).Error; err != nil {
		return nil, err
	}
	return r.toTicket(record), nil
}

func (r *ShareRepository) GetTicketByTicket(ticket string) (*share.ShareTicket, error) {
	var record coreShareTicket
	if err := r.db.Where("ticket = ?", ticket).First(&record).Error; err != nil {
		return nil, err
	}
	return r.toTicket(record), nil
}

func (r *ShareRepository) UpdateTicketAccessTime(ticket string) error {
	now := time.Now()
	return r.db.Model(&coreShareTicket{}).Where("ticket = ?", ticket).Update("access_time", now).Error
}

func (r *ShareRepository) DeleteTicket(id int64) error {
	return r.db.Delete(&coreShareTicket{}, id).Error
}

func (r *ShareRepository) toTicket(record coreShareTicket) *share.ShareTicket {
	return &share.ShareTicket{
		ID:         record.ID,
		UUID:       record.UUID,
		Ticket:     record.Ticket,
		Exp:        record.Exp,
		Args:       record.Args,
		AccessTime: record.AccessTime,
	}
}
