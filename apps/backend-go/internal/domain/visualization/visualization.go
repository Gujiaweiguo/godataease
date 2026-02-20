package visualization

type DataVisualizationInfo struct {
	ID              int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name            string  `gorm:"column:name" json:"name"`
	PID             *int64  `gorm:"column:pid" json:"pid"`
	OrgID           *int64  `gorm:"column:org_id" json:"orgId"`
	Level           *int    `gorm:"column:level" json:"level"`
	NodeType        *string `gorm:"column:node_type" json:"nodeType"`
	Type            *string `gorm:"column:type" json:"type"`
	CanvasStyleData *string `gorm:"column:canvas_style_data" json:"canvasStyleData"`
	ComponentData   *string `gorm:"column:component_data" json:"componentData"`
	MobileLayout    *bool   `gorm:"column:mobile_layout" json:"mobileLayout"`
	Status          *int    `gorm:"column:status" json:"status"`
	Sort            *int    `gorm:"column:sort" json:"sort"`
	CreateTime      *int64  `gorm:"column:create_time" json:"createTime"`
	CreateBy        *string `gorm:"column:create_by" json:"createBy"`
	UpdateTime      *int64  `gorm:"column:update_time" json:"updateTime"`
	UpdateBy        *string `gorm:"column:update_by" json:"updateBy"`
	DeleteFlag      *bool   `gorm:"column:delete_flag" json:"deleteFlag"`
	DeleteTime      *int64  `gorm:"column:delete_time" json:"deleteTime"`
	DeleteBy        *string `gorm:"column:delete_by" json:"deleteBy"`
	Version         *int    `gorm:"column:version" json:"version"`
	ContentID       *string `gorm:"column:content_id" json:"contentId"`
	CheckVersion    *string `gorm:"column:check_version" json:"checkVersion"`
}

func (DataVisualizationInfo) TableName() string {
	return "data_visualization_info"
}

type SaveRequest struct {
	Name            string  `json:"name" binding:"required"`
	PID             *int64  `json:"pid"`
	Type            *string `json:"type"`
	NodeType        *string `json:"nodeType"`
	CanvasStyleData *string `json:"canvasStyleData"`
	ComponentData   *string `json:"componentData"`
	MobileLayout    *bool   `json:"mobileLayout"`
}

type UpdateRequest struct {
	ID              int64   `json:"id" binding:"required"`
	Name            *string `json:"name"`
	PID             *int64  `json:"pid"`
	Type            *string `json:"type"`
	CanvasStyleData *string `json:"canvasStyleData"`
	ComponentData   *string `json:"componentData"`
	MobileLayout    *bool   `json:"mobileLayout"`
	Status          *int    `json:"status"`
}

type DetailRequest struct {
	ID int64 `json:"id" binding:"required"`
}

type ListRequest struct {
	Keyword *string `json:"keyword"`
	Type    *string `json:"type"`
	Current int     `json:"current"`
	Size    int     `json:"size"`
}

type ListResponse struct {
	List    []*DataVisualizationInfo `json:"list"`
	Total   int64                    `json:"total"`
	Current int                      `json:"current"`
	Size    int                      `json:"size"`
}
