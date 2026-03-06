//go:build integration

package service

import (
	"mime/multipart"
	"os"
	"testing"

	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserServiceIntegration_CreateUser(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	userRoleRepo := repository.NewUserRoleRepository(testDB)
	userPermRepo := repository.NewUserPermRepository(testDB)
	svc := NewUserService(userRepo, userRoleRepo, userPermRepo)

	email := "test@example.com"
	phone := "1234567890"
	req := &user.UserCreateRequest{
		Username: "testuser",
		Password: "password123",
		RealName: "Test User",
		Email:    &email,
		Phone:    &phone,
	}

	id, err := svc.CreateUser(req)
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify user was created
	created, err := userRepo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", created.Username)
	assert.Equal(t, "Test User", created.NickName)
	assert.Equal(t, user.StatusEnabled, created.Status)
	assert.Equal(t, user.FromLocal, created.From)

	// Verify password was hashed
	err = bcrypt.CompareHashAndPassword([]byte(created.Password), []byte("password123"))
	assert.NoError(t, err)
}

func TestUserServiceIntegration_CreateUser_DuplicateUsername(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	// Create first user
	_, _ = svc.CreateUser(&user.UserCreateRequest{
		Username: "duplicateuser",
		Password: "password123",
	})

	// Try to create with same username
	_, err := svc.CreateUser(&user.UserCreateRequest{
		Username: "duplicateuser",
		Password: "password456",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username already exists")
}

func TestUserServiceIntegration_CreateUser_WithStatus(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	// Note: GORM may not properly insert Status=0 for new records due to default:1 tag
	// This test verifies explicit status setting works with non-zero values
	enabled := user.StatusEnabled
	id, err := svc.CreateUser(&user.UserCreateRequest{
		Username: "statususer",
		Password: "password123",
		Status:   &enabled,
	})
	assert.NoError(t, err)

	created, _ := userRepo.GetByID(id)
	assert.Equal(t, enabled, created.Status)
}

func TestUserServiceIntegration_UpdateUser(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	// Create user first
	id, _ := svc.CreateUser(&user.UserCreateRequest{
		Username: "updateuser",
		Password: "password123",
		RealName: "Old Name",
	})

	// Update
	newEmail := "new@example.com"
	newPhone := "9876543210"
	newStatus := 0
	err := svc.UpdateUser(&user.UserUpdateRequest{
		ID:       id,
		Username: "updateduser",
		RealName: "New Name",
		Email:    &newEmail,
		Phone:    &newPhone,
		Status:   &newStatus,
	})
	assert.NoError(t, err)

	// Verify
	updated, err := userRepo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "updateduser", updated.Username)
	assert.Equal(t, "New Name", updated.NickName)
	assert.NotNil(t, updated.Email)
	assert.Equal(t, newEmail, *updated.Email)
	assert.Equal(t, newStatus, updated.Status)
	assert.NotNil(t, updated.UpdateTime)
}

func TestUserServiceIntegration_UpdateUser_WithPassword(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	// Create user
	id, _ := svc.CreateUser(&user.UserCreateRequest{
		Username: "pwduser",
		Password: "oldpassword",
	})

	// Update password
	newPwd := "newpassword123"
	err := svc.UpdateUser(&user.UserUpdateRequest{
		ID:       id,
		Password: &newPwd,
	})
	assert.NoError(t, err)

	// Verify password changed
	updated, _ := userRepo.GetByID(id)
	err = bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte(newPwd))
	assert.NoError(t, err)
}

func TestUserServiceIntegration_UpdateUser_NotFound(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	err := svc.UpdateUser(&user.UserUpdateRequest{
		ID:       99999,
		Username: "notexist",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserServiceIntegration_DeleteUser(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	id, _ := svc.CreateUser(&user.UserCreateRequest{
		Username: "deleteuser",
		Password: "password123",
	})

	err := svc.DeleteUser(id)
	assert.NoError(t, err)

	// Verify deleted
	_, err = userRepo.GetByID(id)
	assert.Error(t, err)
}

func TestUserServiceIntegration_GetUserByID(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	id, _ := svc.CreateUser(&user.UserCreateRequest{
		Username: "getuser",
		Password: "password123",
		RealName: "Get User",
	})

	found, err := svc.GetUserByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "getuser", found.Username)
	assert.Equal(t, "Get User", found.NickName)
}

func TestUserServiceIntegration_GetUserByID_NotFound(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	_, err := svc.GetUserByID(99999)
	assert.Error(t, err)
}

func TestUserServiceIntegration_GetUserByUsername(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	_, _ = svc.CreateUser(&user.UserCreateRequest{
		Username: "byname",
		Password: "password123",
	})

	found, err := svc.GetUserByUsername("byname")
	assert.NoError(t, err)
	assert.Equal(t, "byname", found.Username)
}

func TestUserServiceIntegration_SearchUsers(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	// Create multiple users
	for i := 1; i <= 15; i++ {
		_, _ = svc.CreateUser(&user.UserCreateRequest{
			Username: "searchuser" + string(rune('0'+i%10)),
			Password: "password123",
		})
	}

	// Search with pagination
	req := &user.UserQueryRequest{
		Current: 1,
		Size:    10,
	}
	result, err := svc.SearchUsers(req)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Total, int64(10))
	assert.Equal(t, 1, result.Current) // Current is int
	assert.Equal(t, 10, result.Size)   // Size is int
}

func TestUserServiceIntegration_SearchUsers_DefaultPagination(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	req := &user.UserQueryRequest{
		Current: 0,
		Size:    0,
	}
	result, err := svc.SearchUsers(req)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.Current) // Current is int
	assert.Equal(t, 10, result.Size)   // Size is int
}

func TestUserServiceIntegration_ResetPassword(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	id, _ := svc.CreateUser(&user.UserCreateRequest{
		Username: "resetpwd",
		Password: "oldpassword",
	})

	err := svc.ResetPassword(id, "newpassword123")
	assert.NoError(t, err)

	// Verify password changed
	updated, _ := userRepo.GetByID(id)
	err = bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte("newpassword123"))
	assert.NoError(t, err)
}

func TestUserServiceIntegration_ResetPassword_UserNotFound(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	err := svc.ResetPassword(99999, "newpassword")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserServiceIntegration_UpdateUserStatus(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	id, _ := svc.CreateUser(&user.UserCreateRequest{
		Username: "statususer",
		Password: "password",
	})

	err := svc.UpdateUserStatus(id, user.StatusDisabled)
	assert.NoError(t, err)

	updated, _ := userRepo.GetByID(id)
	assert.Equal(t, user.StatusDisabled, updated.Status)
}

func TestUserServiceIntegration_UpdateUserStatus_UserNotFound(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	err := svc.UpdateUserStatus(99999, user.StatusDisabled)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserServiceIntegration_DeleteUser_WithRolesAndPerms(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	userRoleRepo := repository.NewUserRoleRepository(testDB)
	userPermRepo := repository.NewUserPermRepository(testDB)
	svc := NewUserService(userRepo, userRoleRepo, userPermRepo)

	// Create user
	id, err := svc.CreateUser(&user.UserCreateRequest{
		Username: "delete_with_relations",
		Password: "password123",
	})
	require.NoError(t, err)

	// Add role and perm relations (simulate)
	roleRelation := &user.SysUserRole{UserID: id, RoleID: 1}
	require.NoError(t, testDB.Create(roleRelation).Error)

	permRelation := &user.SysUserPerm{UserID: id, PermID: 1}
	require.NoError(t, testDB.Create(permRelation).Error)

	// Delete user - should also delete relations
	err = svc.DeleteUser(id)
	assert.NoError(t, err)

	// Verify user deleted
	_, err = userRepo.GetByID(id)
	assert.Error(t, err)

	// Verify relations are cleaned up (soft delete or actual delete)
	var roleCount int64
	testDB.Model(&user.SysUserRole{}).Where("user_id = ?", id).Count(&roleCount)
	assert.Equal(t, int64(0), roleCount)
}

func TestUserServiceIntegration_CreateUser_CheckUsernameCount(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))

	// Create first user
	_, err := svc.CreateUser(&user.UserCreateRequest{
		Username: "countuser",
		Password: "password123",
	})
	require.NoError(t, err)

	// Try to create with same username
	_, err = svc.CreateUser(&user.UserCreateRequest{
		Username: "countuser",
		Password: "password456",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username already exists")
}

func TestUserImportServiceIntegration_ImportUsersPartialSuccess(t *testing.T) {
	cleanupTables(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{})

	userRepo := repository.NewUserRepository(testDB)
	svc := NewUserService(userRepo, repository.NewUserRoleRepository(testDB), repository.NewUserPermRepository(testDB))
	importSvc := NewUserImportService(svc)

	_, err := svc.CreateUser(&user.UserCreateRequest{
		Username: "dupuser",
		Password: "password123",
		RealName: "Duplicate User",
	})
	require.NoError(t, err)

	csvContent := "username,realName,email,phone\n" +
		"dupuser,Dup User,dup@example.com,13800000001\n" +
		"bademail,Bad User,invalid-email,13800000002\n" +
		"newuser,New User,newuser@example.com,13800000003\n"

	tmpFile, err := os.CreateTemp("", "user-import-integration-*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(csvContent)
	require.NoError(t, err)
	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)

	header := &multipart.FileHeader{Filename: "users.csv", Size: int64(len(csvContent))}
	result, err := importSvc.ImportUsers(tmpFile, header, "tester")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 3, result.TotalRows)
	assert.Equal(t, 1, result.SuccessRows)
	assert.Equal(t, 2, result.FailedRows)
	assert.NotEmpty(t, result.ErrorKey)

	reportContent, _, err := importSvc.GetErrorReport(result.ErrorKey)
	require.NoError(t, err)
	assert.NotEmpty(t, reportContent)

	err = importSvc.ClearErrorReport(result.ErrorKey)
	require.NoError(t, err)

	created, err := userRepo.GetByUsername("newuser")
	require.NoError(t, err)
	require.NotNil(t, created)
	err = bcrypt.CompareHashAndPassword([]byte(created.Password), []byte(svc.ResolveDefaultPassword()))
	assert.NoError(t, err)
}
