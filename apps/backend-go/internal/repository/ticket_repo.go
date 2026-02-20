package repository

import (
	"time"

	"dataease/backend/internal/domain/ticket"

	"gorm.io/gorm"
)

type coreTicket struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID       string    `gorm:"column:uuid;size:255;index" json:"uuid"`
	Ticket     string    `gorm:"column:ticket;size:255;uniqueIndex" json:"ticket"`
	Exp        int64     `gorm:"column:exp" json:"exp"`
	Args       string    `gorm:"column:args;type:text" json:"args"`
	AccessTime int64     `gorm:"column:access_time" json:"accessTime"`
	CreatedAt  time.Time `gorm:"column:create_time" json:"createTime"`
}

func (coreTicket) TableName() string {
	return "core_ticket"
}

type TicketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(t *ticket.Ticket) error {
	record := coreTicket{
		UUID:   t.UUID,
		Ticket: t.Ticket,
		Exp:    t.Exp,
		Args:   t.Args,
	}
	if err := r.db.Create(&record).Error; err != nil {
		return err
	}
	t.ID = record.ID
	t.AccessTime = record.AccessTime
	return nil
}

func (r *TicketRepository) FindByTicket(ticketStr string) (*ticket.Ticket, error) {
	var record coreTicket
	err := r.db.Where("ticket = ?", ticketStr).First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ticket.Ticket{
		ID:         record.ID,
		UUID:       record.UUID,
		Ticket:     record.Ticket,
		Exp:        record.Exp,
		Args:       record.Args,
		AccessTime: record.AccessTime,
	}, nil
}

func (r *TicketRepository) FindByID(id int64) (*ticket.Ticket, error) {
	var record coreTicket
	err := r.db.Where("id = ?", id).First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ticket.Ticket{
		ID:         record.ID,
		UUID:       record.UUID,
		Ticket:     record.Ticket,
		Exp:        record.Exp,
		Args:       record.Args,
		AccessTime: record.AccessTime,
	}, nil
}

func (r *TicketRepository) Delete(ticketStr string) error {
	return r.db.Where("ticket = ?", ticketStr).Delete(&coreTicket{}).Error
}

func (r *TicketRepository) UpdateAccessTime(ticketStr string, accessTime int64) error {
	return r.db.Model(&coreTicket{}).
		Where("ticket = ?", ticketStr).
		Update("access_time", accessTime).Error
}

func (r *TicketRepository) ListByUUID(uuid string, page, pageSize int) ([]ticket.Ticket, int64, error) {
	var records []coreTicket
	var total int64

	query := r.db.Model(&coreTicket{}).Where("uuid = ?", uuid)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	result := make([]ticket.Ticket, len(records))
	for i, record := range records {
		result[i] = ticket.Ticket{
			ID:         record.ID,
			UUID:       record.UUID,
			Ticket:     record.Ticket,
			Exp:        record.Exp,
			Args:       record.Args,
			AccessTime: record.AccessTime,
		}
	}

	return result, total, nil
}
