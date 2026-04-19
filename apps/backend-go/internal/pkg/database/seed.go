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
	// Step 1: Ensure admin role exists
	var adminRole role.SysRole
	result := db.Where("role_code = ?", "admin").First(&adminRole)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		adminRole = role.SysRole{
			RoleName:  "系统管理员",
			RoleCode:  "admin",
			RoleDesc:  ptrString("系统管理员，拥有所有权限"),
			Level:     ptrInt(1),
			DataScope: ptrString("all"),
			Status:    role.StatusEnabled,
			CreateBy:  ptrString("system"),
		}
		if err := db.Create(&adminRole).Error; err != nil {
			return fmt.Errorf("failed to create admin role: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query admin role: %w", result.Error)
	}

	// Step 2: Ensure user role exists
	var userRole role.SysRole
	result = db.Where("role_code = ?", "user").First(&userRole)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		userRole = role.SysRole{
			RoleName:  "普通用户",
			RoleCode:  "user",
			RoleDesc:  ptrString("普通用户，拥有基本权限"),
			Level:     ptrInt(1),
			DataScope: ptrString("self"),
			Status:    role.StatusEnabled,
			CreateBy:  ptrString("system"),
		}
		if err := db.Create(&userRole).Error; err != nil {
			return fmt.Errorf("failed to create user role: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query user role: %w", result.Error)
	}

	// Step 3: Ensure default organization exists
	var defaultOrg org.SysOrg
	result = db.Where("org_name = ? AND parent_id = 0 AND del_flag = 0", "默认组织").First(&defaultOrg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		defaultOrg = org.SysOrg{
			OrgName:  "默认组织",
			ParentID: org.RootParentID,
			Level:    1,
			Status:   org.StatusEnabled,
			DelFlag:  org.DelFlagNormal,
			CreateBy: ptrString("system"),
		}
		if err := db.Create(&defaultOrg).Error; err != nil {
			return fmt.Errorf("failed to create default org: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query default org: %w", result.Error)
	}

	// Step 4: Ensure admin user exists
	var adminUser user.SysUser
	result = db.Where("username = ? AND del_flag = 0", "admin").First(&adminUser)
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
		adminUser = user.SysUser{
			Username: "admin",
			Password: string(hashedPwd),
			NickName: "Admin",
			From:     user.FromLocal,
			Status:   user.StatusEnabled,
			DelFlag:  user.DelFlagNormal,
			Language: &lang,
			CreateBy: ptrString("system"),
		}
		if err := db.Create(&adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query admin user: %w", result.Error)
	}

	// Step 5: Ensure admin user-role-org binding exists
	var userRoleBinding user.SysUserRole
	result = db.Where("user_id = ? AND role_id = ? AND org_id = ?",
		adminUser.UserID, adminRole.RoleID, defaultOrg.OrgID).First(&userRoleBinding)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		userRoleBinding = user.SysUserRole{
			UserID: adminUser.UserID,
			RoleID: adminRole.RoleID,
			OrgID:  defaultOrg.OrgID,
		}
		if err := db.Create(&userRoleBinding).Error; err != nil {
			return fmt.Errorf("failed to bind admin to role and org: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to query admin user-role binding: %w", result.Error)
	}

	return nil
}

func ptrString(s string) *string { return &s }

func ptrInt(i int) *int { return &i }
