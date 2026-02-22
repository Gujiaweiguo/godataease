package repository

import (
	"dataease/backend/internal/domain/role"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RoleMenuRepository provides CRUD operations for role-menu associations.
type RoleMenuRepository struct {
	db *gorm.DB
}

func NewRoleMenuRepository(db *gorm.DB) *RoleMenuRepository {
	return &RoleMenuRepository{db: db}
}

// GetMenuIDsByRoleID 获取角色关联的所有菜单ID
func (r *RoleMenuRepository) GetMenuIDsByRoleID(roleID int64) ([]int64, error) {
	var menuIDs []int64
	err := r.db.Model(&role.RoleMenu{}).
		Where("role_id = ?", roleID).
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

// GetMenuIDsByRoleIDs 获取多个角色关联的所有菜单ID（去重）
func (r *RoleMenuRepository) GetMenuIDsByRoleIDs(roleIDs []int64) ([]int64, error) {
	if len(roleIDs) == 0 {
		return []int64{}, nil
	}
	var menuIDs []int64
	err := r.db.Model(&role.RoleMenu{}).
		Where("role_id IN ?", roleIDs).
		Distinct("menu_id").
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

// GetByRoleID 获取角色的所有菜单关联记录
func (r *RoleMenuRepository) GetByRoleID(roleID int64) ([]*role.RoleMenu, error) {
	var roleMenus []*role.RoleMenu
	err := r.db.Where("role_id = ?", roleID).Find(&roleMenus).Error
	return roleMenus, err
}

// SaveRoleMenus 保存角色菜单关联（幂等操作：先删后插）
func (r *RoleMenuRepository) SaveRoleMenus(roleID int64, menuIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除该角色的所有现有菜单关联
		if err := tx.Where("role_id = ?", roleID).Delete(&role.RoleMenu{}).Error; err != nil {
			return err
		}

		// 如果没有新的菜单ID，直接返回
		if len(menuIDs) == 0 {
			return nil
		}

		// 批量插入新的关联
		roleMenus := make([]*role.RoleMenu, 0, len(menuIDs))
		for _, menuID := range menuIDs {
			roleMenus = append(roleMenus, &role.RoleMenu{
				RoleID: roleID,
				MenuID: menuID,
			})
		}

		return tx.Create(&roleMenus).Error
	})
}

// UpsertRoleMenus 幂等保存（使用 GORM 的 OnConflict）
func (r *RoleMenuRepository) UpsertRoleMenus(roleID int64, menuIDs []int64) error {
	if len(menuIDs) == 0 {
		// 清空该角色的所有菜单关联
		return r.db.Where("role_id = ?", roleID).Delete(&role.RoleMenu{}).Error
	}

	roleMenus := make([]*role.RoleMenu, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		roleMenus = append(roleMenus, &role.RoleMenu{
			RoleID: roleID,
			MenuID: menuID,
		})
	}

	// 使用 clause.OnConflict 实现幂等插入
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}, {Name: "menu_id"}},
		DoNothing: true,
	}).CreateInBatches(roleMenus, 100).Error
}

// DeleteByRoleID 删除角色的所有菜单关联
func (r *RoleMenuRepository) DeleteByRoleID(roleID int64) error {
	return r.db.Where("role_id = ?", roleID).Delete(&role.RoleMenu{}).Error
}

// DeleteByMenuID 删除菜单的所有角色关联
func (r *RoleMenuRepository) DeleteByMenuID(menuID int64) error {
	return r.db.Where("menu_id = ?", menuID).Delete(&role.RoleMenu{}).Error
}

// IsMenuAuthorizedForRole 检查角色是否有指定菜单的授权
func (r *RoleMenuRepository) IsMenuAuthorizedForRole(roleID int64, menuID int64) (bool, error) {
	var count int64
	err := r.db.Model(&role.RoleMenu{}).
		Where("role_id = ? AND menu_id = ?", roleID, menuID).
		Count(&count).Error
	return count > 0, err
}

// IsMenuAuthorizedForRoles 检查多个角色中是否有任意一个拥有指定菜单的授权
func (r *RoleMenuRepository) IsMenuAuthorizedForRoles(roleIDs []int64, menuID int64) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}
	var count int64
	err := r.db.Model(&role.RoleMenu{}).
		Where("role_id IN ? AND menu_id = ?", roleIDs, menuID).
		Count(&count).Error
	return count > 0, err
}

// GetRoleMenuCount 获取角色的菜单数量
func (r *RoleMenuRepository) GetRoleMenuCount(roleID int64) (int64, error) {
	var count int64
	err := r.db.Model(&role.RoleMenu{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	return count, err
}
