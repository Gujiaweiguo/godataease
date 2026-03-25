//go:build integration

package service

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/driver"
	"dataease/backend/internal/domain/embedded"
	"dataease/backend/internal/domain/engine"
	"dataease/backend/internal/domain/geo"
	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/share"
	"dataease/backend/internal/domain/static"
	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/domain/template"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/domain/visualization"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		host := getEnv("TEST_DB_HOST", "mysql8")
		port := getEnv("TEST_DB_PORT", "3306")
		dbUser := getEnv("TEST_DB_USER", "root")
		password := getEnv("TEST_DB_PASSWORD", "Admin168")
		dbname := getEnv("TEST_DB_NAME", "dataease_test")
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, password, host, port, dbname)
	}

	var err error
	testDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	sqlDB, err := testDB.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate
	if err = testDB.AutoMigrate(
		&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{},
		&role.SysRole{}, &role.RoleMenu{},
		&permission.SysRolePerm{},
		&org.SysOrg{},
		&menu.CoreMenu{},
		&permission.SysPerm{}, &permission.SysResource{}, &permission.SysResourcePerm{}, &permission.DataPermRow{}, &permission.DataPermColumn{},
		&share.Share{}, &share.ShareTicket{},
		&template.Template{},
		&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{},
		&visualization.DataVisualizationInfo{},
		&visualization.Watermark{},
		&datasource.CoreDatasource{}, &auto.CoreDatasourceTaskLog{},
		&dataset.CoreDatasetGroup{},
		&auto.CoreExportTask{},
		&driver.Driver{}, &driver.DriverJar{},
		&engine.Engine{},
		&geo.GeometryArea{},
		&areamap.Area{}, &areamap.CoreAreaCustom{},
		&embedded.CoreEmbedded{},
		&static.StaticResource{}, &static.Store{}, &static.Typeface{},
		&system.SysVariable{}, &system.SysVariableValue{},
	); err != nil {
		log.Fatalf("Failed to migrate: %v\n", err)
	}

	// Create core_share table manually (repository uses internal coreShare type with gorm tags)
	if err = testDB.Exec(`CREATE TABLE IF NOT EXISTS core_share (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    creator BIGINT,
    resource_id BIGINT,
    resource_type VARCHAR(50),
    time DATETIME,
    exp BIGINT,
    pwd VARCHAR(255),
    uuid VARCHAR(36)
    )`).Error; err != nil {
		log.Fatalf("Failed to create core_share table: %v", err)
	}

	// Create visualization_watermark table manually
	if err = testDB.Exec(`CREATE TABLE IF NOT EXISTS visualization_watermark (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    setting_content LONGTEXT
    )`).Error; err != nil {
		log.Fatalf("Failed to create visualization_watermark table: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func cleanupTables(tables ...interface{}) {
	if err := testDB.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		panic(fmt.Sprintf("disable foreign key checks failed: %v", err))
	}
	defer func() {
		if err := testDB.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
			panic(fmt.Sprintf("enable foreign key checks failed: %v", err))
		}
	}()

	for _, table := range tables {
		switch table.(type) {
		case *share.Share, share.Share:
			mustExecCleanup("DELETE FROM core_share")
		case *share.ShareTicket, share.ShareTicket:
			thenExecCleanup("DELETE FROM core_share_ticket")
		case *template.Template, template.Template:
			mustExecCleanup("DELETE FROM core_visualization_template")
		case *visualization.DataVisualizationInfo, visualization.DataVisualizationInfo:
			mustExecCleanup("DELETE FROM data_visualization_info")
		case *visualization.Watermark, visualization.Watermark:
			mustExecCleanup("DELETE FROM visualization_watermark")
		case *datasource.CoreDatasource, datasource.CoreDatasource:
			mustExecCleanup("DELETE FROM core_datasource")
			thenExecCleanup("DELETE FROM core_datasource_task_log")
		case *auto.CoreDatasourceTaskLog, auto.CoreDatasourceTaskLog:
			mustExecCleanup("DELETE FROM core_datasource_task_log")
		case *dataset.CoreDatasetGroup, dataset.CoreDatasetGroup:
			mustExecCleanup("DELETE FROM core_dataset_table_field")
			thenExecCleanup("DELETE FROM core_dataset_table")
			mustExecCleanup("DELETE FROM core_dataset_group")
		case *dataset.CoreDatasetTable, dataset.CoreDatasetTable:
			mustExecCleanup("DELETE FROM core_dataset_table_field")
			thenExecCleanup("DELETE FROM core_dataset_table")
		case *dataset.CoreDatasetTableField, dataset.CoreDatasetTableField:
			mustExecCleanup("DELETE FROM core_dataset_table_field")
		case *permission.DataPermRow, permission.DataPermRow:
			mustExecCleanup("DELETE FROM data_perm_row")
		case *permission.DataPermColumn, permission.DataPermColumn:
			mustExecCleanup("DELETE FROM data_perm_column")
		case *system.SysVariable, system.SysVariable:
			mustExecCleanup("DELETE FROM sys_variable_value")
			mustExecCleanup("DELETE FROM sys_variable")
		case *system.SysVariableValue, system.SysVariableValue:
			mustExecCleanup("DELETE FROM sys_variable_value")
		case *permission.SysRolePerm, permission.SysRolePerm:
			mustExecCleanup("DELETE FROM sys_role_perm")
		}
	}
}

func mustExecCleanup(sql string) {
	if err := testDB.Exec(sql).Error; err != nil {
		panic(fmt.Sprintf("cleanup exec failed for %q: %v", sql, err))
	}
}

func thenExecCleanup(sql string) {
	if err := testDB.Exec(sql).Error; err != nil {
		log.Printf("cleanup exec warning for %q: %v", sql, err)
	}
}
