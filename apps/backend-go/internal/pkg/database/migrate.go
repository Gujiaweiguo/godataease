package database

import (
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/static"
	"dataease/backend/internal/domain/user"
	applogger "dataease/backend/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	models := []interface{}{
		&user.SysUser{},
		&user.SysUserRole{},
		&user.SysUserPerm{},
		&role.SysRole{},
		&org.SysOrg{},
		&permission.SysPerm{},
		&static.StaticResource{},
		&static.Store{},
		&static.Typeface{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		return err
	}

	applogger.Info("Database migration completed",
		zap.Int("tables", len(models)),
	)

	return nil
}
