package role

import "time"

const (
	StatusEnabled  = 1
	StatusDisabled = 0
)

const (
	DataScopeAll          = "all"
	DataScopeCustom       = "custom"
	DataScopeDept         = "dept"
	DataScopeDeptAndChild = "dept_and_child"
	DataScopeSelf         = "self"
)

const (
	RoleTypeSystem       = "system"
	RoleTypeCustom       = "custom"
	RoleTypeOrganization = "organization"
)

const (
	BuiltInOrgUserRoleCode = "ROLE_ORG_DEFAULT_USER"
	BuiltInOrgUserRoleName = "普通用户"
)

type SysRole struct {
	RoleID     int64      `gorm:"column:role_id;primaryKey;autoIncrement" json:"roleId"`
	RoleName   string     `gorm:"column:role_name;size:100;not null" json:"roleName"`
	RoleCode   string     `gorm:"column:role_code;size:100" json:"roleCode"`
	RoleType   *string    `gorm:"column:role_type;size:50" json:"roleType"`
	RoleDesc   *string    `gorm:"column:role_desc;size:255" json:"roleDesc"`
	ParentID   *int64     `gorm:"column:parent_id" json:"parentId"`
	Level      *int       `gorm:"column:level" json:"level"`
	DataScope  *string    `gorm:"column:data_scope;size:50" json:"dataScope"`
	Status     int        `gorm:"column:status;default:1" json:"status"`
	CreateBy   *string    `gorm:"column:create_by;size:100" json:"createBy"`
	CreateTime *time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateBy   *string    `gorm:"column:update_by;size:100" json:"updateBy"`
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (SysRole) TableName() string {
	return "sys_role"
}

type RoleCreator struct {
	RoleName string  `json:"roleName"`
	Name     string  `json:"name"`
	RoleKey  string  `json:"roleKey"`
	TypeCode int     `json:"typeCode"`
	RoleDesc *string `json:"roleDesc"`
	Desc     *string `json:"desc"`
	Status   *int    `json:"status"`
	ParentID *int64  `json:"parentId"`
}

type RoleEditor struct {
	ID       int64   `json:"id"`
	RoleID   int64   `json:"roleId"`
	RoleName string  `json:"roleName"`
	Name     string  `json:"name"`
	RoleDesc *string `json:"roleDesc"`
	Desc     *string `json:"desc"`
	Status   *int    `json:"status"`
	ParentID *int64  `json:"parentId"`
}

type RoleVO struct {
	ID       int64   `json:"roleId"`
	Name     string  `json:"roleName"`
	Code     string  `json:"roleKey"`
	RoleType *string `json:"roleType"`
	Desc     *string `json:"roleDesc"`
	Status   int     `json:"status"`
	ReadOnly bool    `json:"readonly"`
	Root     bool    `json:"root"`
}

type RoleDetailVO struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Code      string  `json:"code"`
	Desc      *string `json:"desc"`
	ParentID  *int64  `json:"parentId"`
	Level     *int    `json:"level"`
	DataScope *string `json:"dataScope"`
	Status    int     `json:"status"`
}

type RoleQueryRequest struct {
	Keyword *string `json:"keyword"`
}

type MountUserRequest struct {
	Rid   int64   `json:"rid" binding:"required"`
	Uids  []int64 `json:"uids" binding:"required"`
	OrgId int64   `json:"orgId"`
	Over  bool    `json:"over"`
}

// UnmountUserRequest 用户解绑请求
type UnmountUserRequest struct {
	Rid int64 `json:"rid" binding:"required"`
	Uid int64 `json:"uid" binding:"required"`
}

// MountExternalUserRequest 绑定组织外用户请求
type MountExternalUserRequest struct {
	Rid int64 `json:"rid" binding:"required"`
	Uid int64 `json:"uid" binding:"required"`
}

// RoleRequest 角色过滤器
type RoleRequest struct {
	Keyword *string `json:"keyword"`
	Uid     *int64  `json:"uid"`
}

type ExternalUserVO struct {
	Uid     int64   `json:"uid"`
	Account string  `json:"account"`
	Name    string  `json:"name"`
	Email   *string `json:"email"`
	Phone   *string `json:"phone"`
}

type RolePageRequest struct {
	Keyword  *string `json:"keyword"`
	RoleType *string `json:"roleType"`
	Current  int     `json:"current"`
	Size     int     `json:"size"`
}

type RolePageResult struct {
	List    []*RoleVO `json:"list"`
	Total   int64     `json:"total"`
	Current int       `json:"current"`
	Size    int       `json:"size"`
}
