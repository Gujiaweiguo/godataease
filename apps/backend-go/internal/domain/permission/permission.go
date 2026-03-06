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

// ========== 双视角接口 DTOs ==========

// UserResourcePermVO 用户资源权限视图（按用户视角）
type UserResourcePermVO struct {
	ResourceID   int64  `json:"resourceId"`
	ResourceName string `json:"resourceName"`
	ResourceType string `json:"resourceType"`
	PermKey      string `json:"permKey"`
	PermName     string `json:"permName"`
	SourceType   string `json:"sourceType"` // direct=直接授权, role=角色继承, group=分组继承
	SourceID     int64  `json:"sourceId"`   // 授权来源ID（角色ID或分组ID）
	SourceName   string `json:"sourceName"` // 授权来源名称
}

// ResourceUserPermVO 资源用户权限视图（按资源视角）
type ResourceUserPermVO struct {
	UserID       int64  `json:"userId"`
	Username     string `json:"username"`
	NickName     string `json:"nickName"`
	PermKey      string `json:"permKey"`
	PermName     string `json:"permName"`
	SourceType   string `json:"sourceType"` // direct=直接授权, role=角色继承
	SourceID     int64  `json:"sourceId"`   // 授权来源ID
	SourceName   string `json:"sourceName"` // 授权来源名称
}

// UserPerspectiveRequest 按用户视角查询请求
type UserPerspectiveRequest struct {
	UserID       int64   `json:"userId" binding:"required"`
	ResourceType *string `json:"resourceType"` // 可选过滤：dashboard, dataset, datasource
}

// ResourcePerspectiveRequest 按资源视角查询请求
type ResourcePerspectiveRequest struct {
	ResourceID   int64   `json:"resourceId" binding:"required"`
	ResourceType string  `json:"resourceType" binding:"required"`
}

// ResourceGroupPermVO 资源分组权限
type ResourceGroupPermVO struct {
	GroupID      int64                        `json:"groupId"`
	GroupName    string                       `json:"groupName"`
	ResourceType string                       `json:"resourceType"`
	Permissions  []*ResourceGroupPermItemVO   `json:"permissions"`
}

// ResourceGroupPermItemVO 资源分组权限项
type ResourceGroupPermItemVO struct {
	TargetType string `json:"targetType"` // user=用户, role=角色
	TargetID   int64  `json:"targetId"`
	TargetName string `json:"targetName"`
	PermKey    string `json:"permKey"`
	PermName   string `json:"permName"`
}

// ApplyGroupPermRequest 应用分组权限到资源请求
type ApplyGroupPermRequest struct {
	ResourceID   int64  `json:"resourceId" binding:"required"`
	ResourceType string `json:"resourceType" binding:"required"`
	GroupID      int64  `json:"groupId" binding:"required"`
}

// PermissionConsistencyResult 双视角一致性校验结果
type PermissionConsistencyResult struct {
	Consistent     bool                         `json:"consistent"`
	UserCount      int                          `json:"userCount"`
	ResourceCount  int                          `json:"resourceCount"`
	Inconsistencies []*PermissionInconsistencyVO `json:"inconsistencies"`
}

// PermissionInconsistencyVO 权限不一致项
type PermissionInconsistencyVO struct {
	UserID       int64  `json:"userId"`
	ResourceID   int64  `json:"resourceId"`
	ResourceType string `json:"resourceType"`
	UserView     string `json:"userView"`     // 用户视角的权限状态
	ResourceView string `json:"resourceView"` // 资源视角的权限状态
	Description  string `json:"description"`
}

