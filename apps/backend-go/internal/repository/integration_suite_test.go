//go:build integration
// +build integration

package repository

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		host := getEnv("TEST_DB_HOST", "localhost")
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

	if err = testDB.AutoMigrate(
		&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{},
		&role.SysRole{},
	); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
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

func cleanupTables(tables ...string) {
	for _, table := range tables {
		testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table))
	}
}
