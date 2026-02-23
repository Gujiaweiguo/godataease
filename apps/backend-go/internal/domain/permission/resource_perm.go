package permission

import "time"

// Permission keys for resource operations
const (
	PermKeyView   = "view"
	PermKeyEdit   = "edit"
	PermKeyExport = "export"
	PermKeyManage = "manage"
)

// Resource types
const (
	ResourceTypeMenu       = "menu"
	ResourceTypeDatasource = "datasource"
	ResourceTypeDataset    = "dataset"
	ResourceTypeDashboard  = "dashboard"
	ResourceTypeScreen     = "screen"
)

// SysResource 系统资源表 - 映射 sys_resource 表
type SysResource struct {
	ResourceID   int64      `gorm:"column:resource_id;primaryKey;autoIncrement" json:"resourceId"`
	ResourceName string     `gorm:"column:resource_name;size:255" json:"resourceName"`
	ResourceType string     `gorm:"column:resource_type;size:50" json:"resourceType"`
	ParentID     *int64     `gorm:"column:parent_id" json:"parentId"`
	CreateBy     *string    `gorm:"column:create_by;size:255" json:"createBy"`
	CreateTime   *time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateBy     *string    `gorm:"column:update_by;size:255" json:"updateBy"`
	UpdateTime   *time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (SysResource) TableName() string {
	return "sys_resource"
}

// SysResourcePerm 资源权限关联表 - 映射 sys_resource_perm 表
type SysResourcePerm struct {
	ID         int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ResourceID int64 `gorm:"column:resource_id" json:"resourceId"`
	PermID     int64 `gorm:"column:perm_id" json:"permId"`
}

func (SysResourcePerm) TableName() string {
	return "sys_resource_perm"
}

// SysUserPerm 用户权限关联表 - 映射 sys_user_perm 表
type SysUserPerm struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     int64      `gorm:"column:user_id" json:"userId"`
	OrgID      *int64     `gorm:"column:org_id" json:"orgId"`
	PermID     int64      `gorm:"column:perm_id" json:"permId"`
	CreateBy   *string    `gorm:"column:create_by;size:255" json:"createBy"`
	CreateTime *time.Time `gorm:"column:create_time" json:"createTime"`
	Status     int        `gorm:"column:status;default:1" json:"status"`
	DelFlag    int        `gorm:"column:del_flag;default:0" json:"delFlag"`
}

func (SysUserPerm) TableName() string {
	return "sys_user_perm"
}

// SysRolePerm 角色权限关联表 - 映射 sys_role_perm 表
type SysRolePerm struct {
	ID     int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoleID int64 `gorm:"column:role_id" json:"roleId"`
	PermID int64 `gorm:"column:perm_id" json:"permId"`
}

func (SysRolePerm) TableName() string {
	return "sys_role_perm"
}

// PermissionCheckResult 权限检查结果
type PermissionCheckResult struct {
	HasPermission bool
	Reason        string
}

// ResourcePermissionRequest 资源权限检查请求
type ResourcePermissionRequest struct {
	ResourceType string `json:"resourceType"`
	ResourceID   int64  `json:"resourceId"`
	Permission   string `json:"permission"`
	UserID       int64  `json:"userId"`
}
