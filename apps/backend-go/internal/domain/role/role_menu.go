package role

import "time"

// RoleMenu 角色-菜单关联实体
// 映射 sys_role_menu 表，用于管理角色与菜单的多对多关系
type RoleMenu struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoleID    int64     `gorm:"column:role_id;not null;index:idx_role_id" json:"roleId"`
	MenuID    int64     `gorm:"column:menu_id;not null;index:idx_menu_id" json:"menuId"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (RoleMenu) TableName() string {
	return "sys_role_menu"
}
