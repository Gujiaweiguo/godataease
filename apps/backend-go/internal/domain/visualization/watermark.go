package visualization

type Watermark struct {
	ID             string `gorm:"column:id;primaryKey" json:"id"`
	Version        string `gorm:"column:version" json:"version"`
	SettingContent string `gorm:"column:setting_content" json:"settingContent"`
	CreateBy       string `gorm:"column:create_by" json:"createBy"`
	CreateTime     int64  `gorm:"column:create_time" json:"createTime"`
}

func (Watermark) TableName() string {
	return "visualization_watermark"
}

type WatermarkSaveRequest struct {
	SettingContent string `json:"settingContent"`
}
