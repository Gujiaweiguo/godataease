//go:build integration
// +build integration

package repository

import (
	"fmt"
	"testing"
	"time"

	"dataease/backend/internal/domain/user"
)

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewUserRepository(testDB)
	cleanupTables("sys_user")

	u := &user.SysUser{
		Username:   "testuser",
		NickName:   "Test User",
		Email:      "test@example.com",
		Phone:      "13800138000",
		Status:     1,
		DelFlag:    user.DelFlagNormal,
		CreateTime: time.Now().Unix(),
	}

	err := repo.Create(u)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if u.UserID == 0 {
		t.Error("Expected UserID to be set after creation")
	}

	found, err := repo.GetByID(u.UserID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", found.Username)
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewUserRepository(testDB)
	cleanupTables("sys_user")

	u := &user.SysUser{
		Username:   "uniqueuser",
		NickName:   "Unique User",
		Status:     1,
		DelFlag:    user.DelFlagNormal,
		CreateTime: time.Now().Unix(),
	}
	_ = repo.Create(u)

	found, err := repo.GetByUsername("uniqueuser")
	if err != nil {
		t.Fatalf("GetByUsername failed: %v", err)
	}

	if found.UserID != u.UserID {
		t.Errorf("Expected UserID %d, got %d", u.UserID, found.UserID)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewUserRepository(testDB)
	cleanupTables("sys_user")

	u := &user.SysUser{
		Username:   "deleteuser",
		NickName:   "Delete User",
		Status:     1,
		DelFlag:    user.DelFlagNormal,
		CreateTime: time.Now().Unix(),
	}
	_ = repo.Create(u)

	err := repo.Delete(u.UserID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(u.UserID)
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}
}

func TestUserRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewUserRepository(testDB)
	cleanupTables("sys_user")

	u := &user.SysUser{
		Username:   "updateuser",
		NickName:   "Update User",
		Status:     1,
		DelFlag:    user.DelFlagNormal,
		CreateTime: time.Now().Unix(),
	}
	_ = repo.Create(u)

	u.NickName = "Updated NickName"
	err := repo.Update(u)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(u.UserID)
	if found.NickName != "Updated NickName" {
		t.Errorf("Expected NickName 'Updated NickName', got '%s'", found.NickName)
	}
}

func TestUserRepository_Query(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewUserRepository(testDB)
	cleanupTables("sys_user")

	for i := 1; i <= 5; i++ {
		u := &user.SysUser{
			Username:   fmt.Sprintf("queryuser%d", i),
			NickName:   fmt.Sprintf("Query User %d", i),
			Status:     1,
			DelFlag:    user.DelFlagNormal,
			CreateTime: time.Now().Unix(),
		}
		_ = repo.Create(u)
	}

	keyword := "query"
	req := &user.UserQueryRequest{
		Keyword: &keyword,
		Current: 1,
		Size:    10,
	}

	users, total, err := repo.Query(req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
}

func TestUserRepository_CountByUsername(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewUserRepository(testDB)
	cleanupTables("sys_user")

	u := &user.SysUser{
		Username:   "countuser",
		NickName:   "Count User",
		Status:     1,
		DelFlag:    user.DelFlagNormal,
		CreateTime: time.Now().Unix(),
	}
	_ = repo.Create(u)

	count, err := repo.CountByUsername("countuser")
	if err != nil {
		t.Fatalf("CountByUsername failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestUserRepository_CheckEmailExists(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewUserRepository(testDB)
	cleanupTables("sys_user")

	u := &user.SysUser{
		Username:   "emailuser",
		NickName:   "Email User",
		Email:      "exists@example.com",
		Status:     1,
		DelFlag:    user.DelFlagNormal,
		CreateTime: time.Now().Unix(),
	}
	_ = repo.Create(u)

	exists, err := repo.CheckEmailExists("exists@example.com", 0)
	if err != nil {
		t.Fatalf("CheckEmailExists failed: %v", err)
	}

	if !exists {
		t.Error("Expected email to exist")
	}

	exists, _ = repo.CheckEmailExists("notexists@example.com", 0)
	if exists {
		t.Error("Expected email to not exist")
	}
}

func TestUserRepository_ListUsersByIds(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewUserRepository(testDB)
	cleanupTables("sys_user")

	var ids []int64
	for i := 1; i <= 3; i++ {
		u := &user.SysUser{
			Username:   fmt.Sprintf("listuser%d", i),
			NickName:   fmt.Sprintf("List User %d", i),
			Status:     1,
			DelFlag:    user.DelFlagNormal,
			CreateTime: time.Now().Unix(),
		}
		_ = repo.Create(u)
		ids = append(ids, u.UserID)
	}

	users, err := repo.ListUsersByIds(ids)
	if err != nil {
		t.Fatalf("ListUsersByIds failed: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}
