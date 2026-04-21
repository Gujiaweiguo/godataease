package database

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
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

// SeedDemoData ensures a demo datasource and dataset exist on first launch.
// It is idempotent — safe to call on every startup.
func SeedDemoData(db *gorm.DB) error {
	now := time.Now().UnixMilli()

	demoDatasource, err := ensureDemoDatasource(db, now)
	if err != nil {
		return err
	}

	demoFolder, err := ensureDemoDatasetFolder(db, now)
	if err != nil {
		return err
	}

	teaSalesDataset, err := ensureTeaSalesDataset(db, demoFolder.ID, now)
	if err != nil {
		return err
	}

	demoTable, err := ensureTeaSalesTable(db, demoDatasource.ID, teaSalesDataset.ID)
	if err != nil {
		return err
	}

	if err := ensureTeaSalesFields(db, demoDatasource.ID, teaSalesDataset.ID, demoTable.ID); err != nil {
		return err
	}

	return nil
}

func ensureDemoDatasource(db *gorm.DB, now int64) (*datasource.CoreDatasource, error) {
	var out datasource.CoreDatasource
	rootPID := int64(0)
	result := db.Where("name = ? AND pid = ? AND del_flag = 0", "Demo MySQL", rootPID).First(&out)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		configJSON, err := json.Marshal(datasource.ConnectionConfig{
			Database: "dataease_demo",
			Host:     "127.0.0.1",
			Port:     3306,
			Username: "root",
			Password: "Admin168",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal demo datasource config: %w", err)
		}

		encodedConfig := base64.StdEncoding.EncodeToString(configJSON)
		out = datasource.CoreDatasource{
			PID:            ptrInt64(rootPID),
			Name:           "Demo MySQL",
			Type:           "MySQL",
			EditType:       ptrString("0"),
			Configuration:  ptrString(encodedConfig),
			Status:         ptrString(datasource.StatusSuccess),
			EnableDataFill: ptrBool(false),
			CreateTime:     ptrInt64(now),
			UpdateTime:     ptrInt64(now),
			CreateBy:       ptrString("system"),
			DelFlag:        ptrInt(0),
		}
		if err := db.Create(&out).Error; err != nil {
			return nil, fmt.Errorf("failed to create demo datasource: %w", err)
		}
	} else if result.Error != nil {
		return nil, fmt.Errorf("failed to query demo datasource: %w", result.Error)
	}

	return &out, nil
}

func ensureDemoDatasetFolder(db *gorm.DB, now int64) (*dataset.CoreDatasetGroup, error) {
	var out dataset.CoreDatasetGroup
	rootPID := int64(0)
	folderType := dataset.NodeTypeFolder
	result := db.Where("name = ? AND pid = ? AND node_type = ? AND del_flag = 0", "Demo Datasets", rootPID, folderType).First(&out)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		out = dataset.CoreDatasetGroup{
			Name:           "Demo Datasets",
			PID:            ptrInt64(rootPID),
			Level:          ptrInt(0),
			NodeType:       ptrString(folderType),
			DelFlag:        ptrInt(0),
			CreateBy:       "system",
			CreateTime:     now,
			UpdateBy:       "system",
			LastUpdateTime: now,
		}
		if err := db.Create(&out).Error; err != nil {
			return nil, fmt.Errorf("failed to create demo dataset folder: %w", err)
		}
	} else if result.Error != nil {
		return nil, fmt.Errorf("failed to query demo dataset folder: %w", result.Error)
	}

	return &out, nil
}

func ensureTeaSalesDataset(db *gorm.DB, folderID int64, now int64) (*dataset.CoreDatasetGroup, error) {
	var out dataset.CoreDatasetGroup
	datasetNodeType := dataset.NodeTypeDataset
	datasetType := "db"
	result := db.Where("name = ? AND pid = ? AND node_type = ? AND type = ? AND del_flag = 0", "Tea Sales", folderID, datasetNodeType, datasetType).First(&out)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		out = dataset.CoreDatasetGroup{
			Name:           "Tea Sales",
			PID:            ptrInt64(folderID),
			Level:          ptrInt(1),
			NodeType:       ptrString(datasetNodeType),
			Type:           ptrString(datasetType),
			DelFlag:        ptrInt(0),
			CreateBy:       "system",
			CreateTime:     now,
			UpdateBy:       "system",
			LastUpdateTime: now,
		}
		if err := db.Create(&out).Error; err != nil {
			return nil, fmt.Errorf("failed to create tea sales dataset: %w", err)
		}
	} else if result.Error != nil {
		return nil, fmt.Errorf("failed to query tea sales dataset: %w", result.Error)
	}

	return &out, nil
}

func ensureTeaSalesTable(db *gorm.DB, datasourceID, datasetGroupID int64) (*dataset.CoreDatasetTable, error) {
	var out dataset.CoreDatasetTable
	result := db.Where("datasource_id = ? AND dataset_group_id = ? AND table_name = ?", datasourceID, datasetGroupID, "tea_sales").First(&out)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		tableType := "db"
		out = dataset.CoreDatasetTable{
			Name:           ptrString("tea_sales"),
			DatasourceID:   ptrInt64(datasourceID),
			DatasetGroupID: datasetGroupID,
			PhysicalTable:  ptrString("tea_sales"),
			Type:           ptrString(tableType),
		}
		if err := db.Create(&out).Error; err != nil {
			return nil, fmt.Errorf("failed to create tea sales dataset table: %w", err)
		}
	} else if result.Error != nil {
		return nil, fmt.Errorf("failed to query tea sales dataset table: %w", result.Error)
	}

	return &out, nil
}

func ensureTeaSalesFields(db *gorm.DB, datasourceID, datasetGroupID, datasetTableID int64) error {
	fields := []dataset.CoreDatasetTableField{
		newDemoDatasetField(datasourceID, datasetGroupID, datasetTableID, "id", "q", "BIGINT", 2),
		newDemoDatasetField(datasourceID, datasetGroupID, datasetTableID, "product_name", "d", "LONGTEXT", 0),
		newDemoDatasetField(datasourceID, datasetGroupID, datasetTableID, "category", "d", "LONGTEXT", 0),
		newDemoDatasetField(datasourceID, datasetGroupID, datasetTableID, "region", "d", "LONGTEXT", 0),
		newDemoDatasetField(datasourceID, datasetGroupID, datasetTableID, "sales_amount", "q", "LONGTEXT", 2),
		newDemoDatasetField(datasourceID, datasetGroupID, datasetTableID, "quantity", "q", "BIGINT", 2),
		newDemoDatasetField(datasourceID, datasetGroupID, datasetTableID, "sale_date", "d", "DATETIME", 1),
		newDemoDatasetField(datasourceID, datasetGroupID, datasetTableID, "salesperson", "d", "LONGTEXT", 0),
	}

	for _, field := range fields {
		result := db.Where("dataset_table_id = ? AND origin_name = ?", datasetTableID, *field.OriginName).First(&dataset.CoreDatasetTableField{})
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fieldCopy := field
			if err := db.Create(&fieldCopy).Error; err != nil {
				return fmt.Errorf("failed to create demo dataset field %s: %w", *field.OriginName, err)
			}
		} else if result.Error != nil {
			return fmt.Errorf("failed to query demo dataset field %s: %w", *field.OriginName, result.Error)
		}
	}

	return nil
}

func newDemoDatasetField(datasourceID, datasetGroupID, datasetTableID int64, originName, groupType, fieldType string, deType int) dataset.CoreDatasetTableField {
	dataeaseName := "f_" + originName
	return dataset.CoreDatasetTableField{
		DatasourceID:   ptrInt64(datasourceID),
		DatasetTableID: ptrInt64(datasetTableID),
		DatasetGroupID: datasetGroupID,
		OriginName:     ptrString(originName),
		Name:           ptrString(originName),
		DataeaseName:   ptrString(dataeaseName),
		FieldShortName: ptrString(dataeaseName),
		GroupType:      ptrString(groupType),
		Type:           ptrString(fieldType),
		DeType:         ptrInt(deType),
		DeExtractType:  ptrInt(deType),
		ExtField:       ptrInt(0),
		Checked:        ptrBool(true),
	}
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

// CleanupStaleMenuData removes obsolete/duplicate menu records and their
// role_menu bindings that were left behind by earlier migration iterations.
// It also fixes top-level navigation menus whose hidden flag was incorrectly
// set to true by older migration scripts.
// It is idempotent — safe to call on every startup.
func CleanupStaleMenuData(db *gorm.DB) {
	staleMenuIDs := []int64{11, 12, 15, 16, 19, 31, 64, 70, 71, 200}

	db.Exec("DELETE FROM sys_role_menu WHERE menu_id IN ?", staleMenuIDs)
	db.Exec("DELETE FROM core_menu WHERE id IN ?", staleMenuIDs)

	db.Exec("DELETE FROM core_menu WHERE pid IN ?", staleMenuIDs)

	// Fix top-level directory menus that should be visible in the header bar.
	// Older migration iterations left hidden=1 on these records.
	db.Exec("UPDATE core_menu SET hidden = 0 WHERE id IN (1, 4, 100, 101) AND hidden = 1")
}

func ptrString(s string) *string { return &s }

func ptrInt(i int) *int { return &i }

func ptrInt64(i int64) *int64 { return &i }

func ptrBool(b bool) *bool { return &b }
