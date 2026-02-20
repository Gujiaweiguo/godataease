package permission

import "time"

// 权限类型常量
const (
	PermTypeMenu   = "menu"
	PermTypeButton = "button"
	PermTypeData   = "data"
)

// 状态常量
const (
	StatusEnabled  = 1
	StatusDisabled = 0
)

// 删除标记常量
const (
	DelFlagNormal  = 0
	DelFlagDeleted = 1
)

// SysPerm 权限实体 - 映射 sys_perm 表
type SysPerm struct {
	PermID     int64      `gorm:"column:perm_id;primaryKey;autoIncrement" json:"permId"`
	PermName   string     `gorm:"column:perm_name;size:100;not null" json:"permName"`
	PermKey    string     `gorm:"column:perm_key;size:100;not null" json:"permKey"`
	PermType   string     `gorm:"column:perm_type;size:20" json:"permType"`
	PermDesc   *string    `gorm:"column:perm_desc;size:500" json:"permDesc"`
	Status     int        `gorm:"column:status;default:1" json:"status"`
	CreateBy   *string    `gorm:"column:create_by;size:100" json:"createBy"`
	CreateTime time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateBy   *string    `gorm:"column:update_by;size:100" json:"updateBy"`
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime"`
	DelFlag    int        `gorm:"column:del_flag;default:0" json:"delFlag"`
}

func (SysPerm) TableName() string {
	return "sys_perm"
}

// PermCreateRequest 创建权限请求
type PermCreateRequest struct {
	PermName string  `json:"permName" binding:"required"`
	PermKey  string  `json:"permKey" binding:"required"`
	PermType string  `json:"permType"`
	PermDesc *string `json:"permDesc"`
	Status   *int    `json:"status"`
}

// PermUpdateRequest 更新权限请求
type PermUpdateRequest struct {
	PermID   int64   `json:"permId" binding:"required"`
	PermName string  `json:"permName"`
	PermKey  string  `json:"permKey"`
	PermType string  `json:"permType"`
	PermDesc *string `json:"permDesc"`
	Status   *int    `json:"status"`
}

// PermQueryRequest 查询权限请求
type PermQueryRequest struct {
	Current int `json:"current"`
	Size    int `json:"size"`
}

// PermListResponse 权限列表响应
type PermListResponse struct {
	List    interface{} `json:"list"`
	Total   int64       `json:"total"`
	Current int         `json:"current"`
	Size    int         `json:"size"`
}
