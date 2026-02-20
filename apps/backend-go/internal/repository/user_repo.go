package repository

import (
	"dataease/backend/internal/domain/user"
	"gorm.io/gorm"
)

// UserRepository provides CRUD operations for SysUser with soft delete.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(u *user.SysUser) error {
	return r.db.Create(u).Error
}

// Update 更新用户
func (r *UserRepository) Update(u *user.SysUser) error {
	return r.db.Save(u).Error
}

// Delete 软删除用户（设置 del_flag = 1）
func (r *UserRepository) Delete(userID int64) error {
	return r.db.Model(&user.SysUser{}).
		Where("user_id = ?", userID).
		Update("del_flag", user.DelFlagDeleted).Error
}

// GetByID 根据ID查询用户
func (r *UserRepository) GetByID(userID int64) (*user.SysUser, error) {
	var u user.SysUser
	err := r.db.Where("user_id = ? AND del_flag = ?", userID, user.DelFlagNormal).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByUsername 根据用户名查询用户
func (r *UserRepository) GetByUsername(username string) (*user.SysUser, error) {
	var u user.SysUser
	err := r.db.Where("username = ? AND del_flag = ?", username, user.DelFlagNormal).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Query 查询用户列表（支持 keyword、orgId、status 过滤）
func (r *UserRepository) Query(req *user.UserQueryRequest) ([]*user.SysUser, int64, error) {
	var users []*user.SysUser
	var total int64

	// 基础条件：未删除
	db := r.db.Model(&user.SysUser{}).Where("del_flag = ?", user.DelFlagNormal)

	// 关键字筛选（用户名、昵称、邮箱）
	if req.Keyword != nil && *req.Keyword != "" {
		keyword := "%" + *req.Keyword + "%"
		db = db.Where("username LIKE ? OR nick_name LIKE ? OR email LIKE ?", keyword, keyword, keyword)
	}

	// 状态筛选
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}

	// 机构/组织筛选：通过 sys_user_role 表进行联结筛选
	if req.OrgID != nil {
		db = db.Joins("JOIN sys_user_role ur ON ur.user_id = sys_user.user_id").Where("ur.org_id = ?", *req.OrgID)
	}

	// 总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页参数
	page := req.Current
	pageSize := req.Size
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	if err := db.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// CountByUsername 统计用户名是否存在
func (r *UserRepository) CountByUsername(username string) (int64, error) {
	var count int64
	err := r.db.Model(&user.SysUser{}).
		Where("username = ? AND del_flag = ?", username, user.DelFlagNormal).
		Count(&count).Error
	return count, err
}

// CheckEmailExists 检查邮箱是否存在（排除指定用户ID）
func (r *UserRepository) CheckEmailExists(email string, excludeUserID int64) (bool, error) {
	var count int64
	query := r.db.Model(&user.SysUser{}).
		Where("email = ? AND del_flag = ?", email, user.DelFlagNormal)
	if excludeUserID > 0 {
		query = query.Where("user_id != ?", excludeUserID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// ListUsersByIds 根据ID列表查询用户
func (r *UserRepository) ListUsersByIds(ids []int64) ([]*user.SysUser, error) {
	var users []*user.SysUser
	err := r.db.Where("user_id IN ? AND del_flag = ?", ids, user.DelFlagNormal).
		Find(&users).Error
	return users, err
}

// UserRoleRepository 关联表操作（用户-角色）
type UserRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

func (r *UserRoleRepository) Create(role *user.SysUserRole) error {
	return r.db.Create(role).Error
}

func (r *UserRoleRepository) DeleteByUserID(userID int64) error {
	return r.db.Where("user_id = ?", userID).Delete(&user.SysUserRole{}).Error
}

func (r *UserRoleRepository) GetByUserID(userID int64) ([]*user.SysUserRole, error) {
	var roles []*user.SysUserRole
	err := r.db.Where("user_id = ?", userID).Find(&roles).Error
	return roles, err
}

type UserPermRepository struct {
	db *gorm.DB
}

func NewUserPermRepository(db *gorm.DB) *UserPermRepository {
	return &UserPermRepository{db: db}
}

func (r *UserPermRepository) Create(perm *user.SysUserPerm) error {
	return r.db.Create(perm).Error
}

func (r *UserPermRepository) DeleteByUserID(userID int64) error {
	return r.db.Where("user_id = ?", userID).Delete(&user.SysUserPerm{}).Error
}

func (r *UserPermRepository) GetByUserID(userID int64) ([]*user.SysUserPerm, error) {
	var perms []*user.SysUserPerm
	err := r.db.Where("user_id = ?", userID).Find(&perms).Error
	return perms, err
}
