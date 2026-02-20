package user

import "time"

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

// 用户来源常量
const (
	FromLocal      = 0
	FromThirdParty = 1
)

// SysUser 用户实体 - 映射 sys_user 表
type SysUser struct {
	UserID     int64      `gorm:"column:user_id;primaryKey;autoIncrement" json:"userId"`
	Username   string     `gorm:"column:username;size:100;not null" json:"username"`
	Password   string     `gorm:"column:password;size:255" json:"-"`
	NickName   string     `gorm:"column:nick_name;size:100" json:"nickName"`
	Email      *string    `gorm:"column:email;size:100" json:"email"`
	Phone      *string    `gorm:"column:phone;size:50" json:"phone"`
	From       int        `gorm:"column:from;default:0" json:"from"`
	Sub        *string    `gorm:"column:sub;size:255" json:"sub"`
	Avatar     *string    `gorm:"column:avatar;size:500" json:"avatar"`
	DeptID     *int64     `gorm:"column:dept_id" json:"deptId"`
	Status     int        `gorm:"column:status;default:1" json:"status"`
	DelFlag    int        `gorm:"column:del_flag;default:0" json:"delFlag"`
	CreateBy   *string    `gorm:"column:create_by;size:100" json:"createBy"`
	CreateTime time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateBy   *string    `gorm:"column:update_by;size:100" json:"updateBy"`
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime"`
	Language   *string    `gorm:"column:language;size:20" json:"language"`
}

func (SysUser) TableName() string {
	return "sys_user"
}

// SysUserRole 用户角色关联 - 映射 sys_user_role 表
type SysUserRole struct {
	ID     int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID int64 `gorm:"column:user_id;not null" json:"userId"`
	RoleID int64 `gorm:"column:role_id;not null" json:"roleId"`
	OrgID  int64 `gorm:"column:org_id;not null" json:"orgId"`
}

func (SysUserRole) TableName() string {
	return "sys_user_role"
}

// SysUserPerm 用户权限关联 - 映射 sys_user_perm 表
type SysUserPerm struct {
	ID      int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID  int64 `gorm:"column:user_id;not null" json:"userId"`
	OrgID   int64 `gorm:"column:org_id;not null" json:"orgId"`
	PermID  int64 `gorm:"column:perm_id;not null" json:"permId"`
	Status  int   `gorm:"column:status;default:1" json:"status"`
	DelFlag int   `gorm:"column:del_flag;default:0" json:"delFlag"`
}

func (SysUserPerm) TableName() string {
	return "sys_user_perm"
}

// UserCreateRequest 创建用户请求
type UserCreateRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	RealName string  `json:"realName"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Status   *int    `json:"status"`
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	ID       int64   `json:"id" binding:"required"`
	Username string  `json:"username"`
	RealName string  `json:"realName"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Password *string `json:"password"`
	Status   *int    `json:"status"`
}

// UserQueryRequest 查询用户请求
type UserQueryRequest struct {
	Keyword *string `json:"keyword"`
	OrgID   *int64  `json:"orgId"`
	Status  *int    `json:"status"`
	Current int     `json:"current"`
	Size    int     `json:"size"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	List    interface{} `json:"list"`
	Total   int64       `json:"total"`
	Current int         `json:"current"`
	Size    int         `json:"size"`
}
