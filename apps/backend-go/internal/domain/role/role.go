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

type SysRole struct {
	RoleID     int64      `gorm:"column:role_id;primaryKey;autoIncrement" json:"roleId"`
	RoleName   string     `gorm:"column:role_name;size:100;not null" json:"roleName"`
	RoleCode   string     `gorm:"column:role_code;size:100" json:"roleCode"`
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
	Name     string  `json:"name" binding:"required"`
	TypeCode int     `json:"typeCode" binding:"required"`
	Desc     *string `json:"desc"`
}

type RoleEditor struct {
	ID   int64   `json:"id" binding:"required"`
	Name string  `json:"name" binding:"required"`
	Desc *string `json:"desc"`
}

type RoleVO struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ReadOnly bool   `json:"readonly"`
	Root     bool   `json:"root"`
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
	OrgId int64   `json:"orgId" binding:"required"`
	Over  bool    `json:"over"`
}
