package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dsn = flag.String("dsn", "", "Database DSN (required)")

func main() {
	flag.Parse()

	if *dsn == "" {
		*dsn = "root:Admin168@tcp(localhost:3306)/dataease_dev?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(*dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	fmt.Println("Running AutoMigrate...")

	models := []interface{}{
		&user.SysUser{},
		&user.SysUserRole{},
		&user.SysUserPerm{},
		&role.SysRole{},
		&org.SysOrg{},
		&permission.SysPerm{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	fmt.Printf("Migrated %d tables\n", len(models))

	fmt.Println("Creating default admin user...")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	adminUser := user.SysUser{
		Username:   "admin",
		Password:   string(hashedPassword),
		NickName:   "Administrator",
		Status:     1,
		DelFlag:    0,
		From:       0,
		CreateTime: time.Now(),
	}

	result := db.Where("username = ?", "admin").FirstOrCreate(&adminUser)
	if result.Error != nil {
		log.Printf("Warning: Failed to create admin user: %v", result.Error)
	} else {
		fmt.Println("Admin user created/verified (password: admin123)")
	}

	fmt.Println("Creating default organization...")

	defaultOrg := org.SysOrg{
		OrgName:    "Default Organization",
		ParentID:   0,
		Level:      1,
		Status:     1,
		DelFlag:    0,
		CreateTime: time.Now(),
	}

	result = db.Where("org_name = ?", "Default Organization").FirstOrCreate(&defaultOrg)
	if result.Error != nil {
		log.Printf("Warning: Failed to create org: %v", result.Error)
	} else {
		fmt.Println("Default organization created/verified")
	}

	fmt.Println("Creating default role...")

	adminRole := role.SysRole{
		RoleName: "Administrator",
		RoleCode: "admin",
		RoleDesc: strPtr("System Administrator"),
		Status:   1,
	}

	result = db.Where("role_code = ?", "admin").FirstOrCreate(&adminRole)
	if result.Error != nil {
		log.Printf("Warning: Failed to create role: %v", result.Error)
	} else {
		fmt.Println("Default role created/verified")
	}

	fmt.Println("Migration completed successfully!")
}

func strPtr(s string) *string {
	return &s
}
