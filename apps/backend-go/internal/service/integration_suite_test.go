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
		&org.SysOrg{},
		&menu.CoreMenu{},
		&permission.SysPerm{}, &permission.DataPermRow{}, &permission.DataPermColumn{},
		&share.Share{}, &share.ShareTicket{},
		&template.Template{},
		&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{},
		&visualization.DataVisualizationInfo{},
		&datasource.CoreDatasource{}, &auto.CoreDatasourceTaskLog{},
		&dataset.CoreDatasetGroup{},
		&auto.CoreExportTask{},
		&driver.Driver{}, &driver.DriverJar{},
		&engine.Engine{},
		&geo.GeometryArea{},
		&areamap.Area{}, &areamap.CoreAreaCustom{},
		&embedded.CoreEmbedded{},
		&static.StaticResource{}, &static.Store{}, &static.Typeface{},
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
	uuid VARCHAR(64),
	pwd VARCHAR(255),
	auto_pwd TINYINT(1) DEFAULT 1,
	ticket_require TINYINT(1) DEFAULT 0,
	INDEX idx_creator (creator),
	INDEX idx_resource_id (resource_id),
	UNIQUE INDEX idx_uuid (uuid)
)`).Error; err != nil {
		log.Fatalf("Failed to create core_share table: %v\n", err)
	}

	// Create core_share_ticket table manually
	if err = testDB.Exec(`CREATE TABLE IF NOT EXISTS core_share_ticket (
	id BIGINT AUTO_INCREMENT PRIMARY KEY,
	uuid VARCHAR(64),
	ticket VARCHAR(64),
	exp BIGINT,
	args TEXT,
	access_time DATETIME,
	INDEX idx_uuid (uuid),
	UNIQUE INDEX idx_ticket (ticket)
)`).Error; err != nil {
		log.Fatalf("Failed to create core_share_ticket table: %v\n", err)
	}

	// Create core_visualization_template table manually
	if err = testDB.Exec(`CREATE TABLE IF NOT EXISTS core_visualization_template (
	id BIGINT AUTO_INCREMENT PRIMARY KEY,
	name VARCHAR(255),
	pid BIGINT,
	level INT,
	dv_type VARCHAR(50),
	node_type VARCHAR(50),
	create_by VARCHAR(255),
	create_time DATETIME,
	snapshot LONGTEXT,
	template_type VARCHAR(50),
	template_style LONGTEXT,
	template_data LONGTEXT,
	dynamic_data LONGTEXT,
	app_data LONGTEXT,
	use_count INT DEFAULT 0,
	version INT DEFAULT 3,
	INDEX idx_pid (pid)
)`).Error; err != nil {
		log.Fatalf("Failed to create core_visualization_template table: %v\n", err)
	}

	// Create additional tables that don't have GORM models in the test suite
	if err = testDB.Exec(`CREATE TABLE IF NOT EXISTS core_ds_finish_page (
	id BIGINT PRIMARY KEY
)`).Error; err != nil {
		log.Fatalf("Failed to create core_ds_finish_page table: %v\n", err)
	}

	if err = testDB.Exec(`CREATE TABLE IF NOT EXISTS core_msg_setting (
	id BIGINT AUTO_INCREMENT PRIMARY KEY,
	msg_id VARCHAR(100),
	user_id BIGINT,
	status VARCHAR(20),
	read_at DATETIME,
	UNIQUE INDEX idx_msg_user (msg_id, user_id),
	INDEX idx_user_id (user_id)
)`).Error; err != nil {
		log.Fatalf("Failed to create core_msg_setting table: %v\n", err)
	}

	if err = testDB.Exec(`CREATE TABLE IF NOT EXISTS core_ticket (
	id BIGINT AUTO_INCREMENT PRIMARY KEY,
	uuid VARCHAR(255),
	ticket VARCHAR(255) UNIQUE,
	exp BIGINT,
	args TEXT,
	access_time BIGINT,
	create_time DATETIME,
	INDEX idx_uuid (uuid)
)`).Error; err != nil {
		log.Fatalf("Failed to create core_ticket table: %v\n", err)
	}

	code := m.Run()

	sqlDB, _ = testDB.DB()
	sqlDB.Close()
	os.Exit(code)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func cleanupTables(tables ...interface{}) {
	for _, table := range tables {
		switch table.(type) {
		case *share.Share, share.Share:
			testDB.Exec("DELETE FROM core_share")
		case *share.ShareTicket, share.ShareTicket:
			testDB.Exec("DELETE FROM core_share_ticket")
		case *template.Template, template.Template:
			testDB.Exec("DELETE FROM core_visualization_template")
		case *visualization.DataVisualizationInfo, visualization.DataVisualizationInfo:
			testDB.Exec("DELETE FROM data_visualization_info")
		case *datasource.CoreDatasource, datasource.CoreDatasource:
			testDB.Exec("DELETE FROM core_datasource")
			testDB.Exec("DELETE FROM core_datasource_task_log")
		case *auto.CoreDatasourceTaskLog, auto.CoreDatasourceTaskLog:
			testDB.Exec("DELETE FROM core_datasource_task_log")
		case *dataset.CoreDatasetGroup, dataset.CoreDatasetGroup:
			testDB.Exec("DELETE FROM core_dataset_group")
		default:
			// Use GORM's Unscoped delete for other types with soft delete support
			testDB.Unscoped().Where("1 = 1").Delete(table)
		}
	}
}
