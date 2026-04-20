package database

import (
	"errors"
	"fmt"
	"os"

	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedDefaults ensures the minimum data needed for admin login exists.
// It is idempotent — safe to call on every startup.
func SeedDefaults(db *gorm.DB) error {
	var adminRole role.SysRole
	if err := ensureAdminRole(db, &adminRole); err != nil {
		return err
	}

	var userRole role.SysRole
	if err := ensureUserRole(db, &userRole); err != nil {
		return err
	}

	var defaultOrg org.SysOrg
	if err := ensureDefaultOrg(db, &defaultOrg); err != nil {
		return err
	}

	var adminUser user.SysUser
	if err := ensureAdminUser(db, &adminUser); err != nil {
		return err
	}

	return ensureAdminBinding(db, adminUser.UserID, adminRole.RoleID, defaultOrg.OrgID)
}

func ensureAdminRole(db *gorm.DB, out *role.SysRole) error {
	result := db.Where("role_code = ?", "admin").First(out)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		*out = role.SysRole{
			RoleName:  "系统管理员",
			RoleCode:  "admin",
			RoleDesc:  ptrString("系统管理员，拥有所有权限"),
			Level:     ptrInt(1),
			DataScope: ptrString("all"),
			Status:    role.StatusEnabled,
			CreateBy:  ptrString("system"),
		}
		if err := db.Create(out).Error; err != nil {
			return fmt.Errorf("failed to create admin role: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query admin role: %w", result.Error)
	}
	return nil
}

func ensureUserRole(db *gorm.DB, out *role.SysRole) error {
	result := db.Where("role_code = ?", "user").First(out)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		*out = role.SysRole{
			RoleName:  "普通用户",
			RoleCode:  "user",
			RoleDesc:  ptrString("普通用户，拥有基本权限"),
			Level:     ptrInt(1),
			DataScope: ptrString("self"),
			Status:    role.StatusEnabled,
			CreateBy:  ptrString("system"),
		}
		if err := db.Create(out).Error; err != nil {
			return fmt.Errorf("failed to create user role: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query user role: %w", result.Error)
	}
	return nil
}

func ensureDefaultOrg(db *gorm.DB, out *org.SysOrg) error {
	result := db.Where("org_name = ? AND parent_id = 0 AND del_flag = 0", "默认组织").First(out)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		*out = org.SysOrg{
			OrgName:  "默认组织",
			ParentID: org.RootParentID,
			Level:    1,
			Status:   org.StatusEnabled,
			DelFlag:  org.DelFlagNormal,
			CreateBy: ptrString("system"),
		}
		if err := db.Create(out).Error; err != nil {
			return fmt.Errorf("failed to create default org: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query default org: %w", result.Error)
	}
	return nil
}

func ensureAdminUser(db *gorm.DB, out *user.SysUser) error {
	result := db.Where("username = ? AND del_flag = 0", "admin").First(out)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		pwd := os.Getenv("ADMIN_PASSWORD")
		if pwd == "" {
			pwd = "admin123"
		}
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), 10)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}
		lang := "zh-CN"
		*out = user.SysUser{
			Username: "admin",
			Password: string(hashedPwd),
			NickName: "Admin",
			From:     user.FromLocal,
			Status:   user.StatusEnabled,
			DelFlag:  user.DelFlagNormal,
			Language: &lang,
			CreateBy: ptrString("system"),
		}
		if err := db.Create(out).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query admin user: %w", result.Error)
	}
	return nil
}

func ensureAdminBinding(db *gorm.DB, userID, roleID, orgID int64) error {
	var binding user.SysUserRole
	result := db.Where("user_id = ? AND role_id = ? AND org_id = ?", userID, roleID, orgID).First(&binding)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		binding = user.SysUserRole{
			UserID: userID,
			RoleID: roleID,
			OrgID:  orgID,
		}
		if err := db.Create(&binding).Error; err != nil {
			return fmt.Errorf("failed to bind admin to role and org: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query admin user-role binding: %w", result.Error)
	}
	return nil
}

func ptrString(s string) *string { return &s }

func ptrInt(i int) *int { return &i }
