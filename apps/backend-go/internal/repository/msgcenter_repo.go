package repository

import (
	"time"

	"gorm.io/gorm"
)

type coreMsgSetting struct {
	ID     int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MsgID  string     `gorm:"column:msg_id;size:100;uniqueIndex:idx_msg_user" json:"msgId"`
	UserID int64      `gorm:"column:user_id;index;uniqueIndex:idx_msg_user" json:"userId"`
	Status string     `gorm:"column:status;size:20" json:"status"`
	ReadAt *time.Time `gorm:"column:read_at" json:"readAt"`
}

func (coreMsgSetting) TableName() string {
	return "core_msg_setting"
}

type MsgCenterRepository struct {
	db *gorm.DB
}

func NewMsgCenterRepository(db *gorm.DB) *MsgCenterRepository {
	return &MsgCenterRepository{db: db}
}

// MarkAsRead marks a message as read for a specific user
func (r *MsgCenterRepository) MarkAsRead(msgID string, userID int64) error {
	now := time.Now()
	var existing coreMsgSetting
	err := r.db.Where("msg_id = ? AND user_id = ?", msgID, userID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		record := coreMsgSetting{
			MsgID:  msgID,
			UserID: userID,
			Status: "read",
			ReadAt: &now,
		}
		return r.db.Create(&record).Error
	}
	if err != nil {
		return err
	}
	existing.Status = "read"
	existing.ReadAt = &now
	return r.db.Save(&existing).Error
}

// MarkBatchAsRead marks multiple messages as read for a specific user
func (r *MsgCenterRepository) MarkBatchAsRead(msgIDs []string, userID int64) (int, error) {
	if len(msgIDs) == 0 {
		return 0, nil
	}

	now := time.Now()
	updated := 0

	for _, msgID := range msgIDs {
		if msgID == "" {
			continue
		}

		var existing coreMsgSetting
		err := r.db.Where("msg_id = ? AND user_id = ?", msgID, userID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			record := coreMsgSetting{
				MsgID:  msgID,
				UserID: userID,
				Status: "read",
				ReadAt: &now,
			}
			if createErr := r.db.Create(&record).Error; createErr == nil {
				updated++
			}
			continue
		}
		if err != nil {
			continue
		}
		// Skip if already read
		if existing.Status == "read" {
			continue
		}
		// Update existing record
		existing.Status = "read"
		existing.ReadAt = &now
		if saveErr := r.db.Save(&existing).Error; saveErr == nil {
			updated++
		}
	}

	return updated, nil
}

// IsRead checks if a message has been read by a specific user
func (r *MsgCenterRepository) IsRead(msgID string, userID int64) (bool, error) {
	var record coreMsgSetting
	err := r.db.Where("msg_id = ? AND user_id = ? AND status = ?", msgID, userID, "read").First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetReadStatusMap returns a map of msgID to read status for a specific user
func (r *MsgCenterRepository) GetReadStatusMap(msgIDs []string, userID int64) (map[string]bool, error) {
	result := make(map[string]bool)
	if len(msgIDs) == 0 {
		return result, nil
	}

	var records []coreMsgSetting
	err := r.db.Where("msg_id IN ? AND user_id = ? AND status = ?", msgIDs, userID, "read").Find(&records).Error
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		result[record.MsgID] = true
	}

	return result, nil
}
