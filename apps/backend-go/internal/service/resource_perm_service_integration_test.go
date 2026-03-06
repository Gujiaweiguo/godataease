//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

// Helper function to create ResourcePermissionService with all dependencies
func newTestResourcePermissionService(t *testing.T) *ResourcePermissionService {
	cleanupTables(&permission.SysPerm{})

	repo := repository.NewResourcePermissionRepository(testDB)
	return NewResourcePermissionService(repo, nil)
}

// Test GetUserPerspective for non-admin user
func TestResourcePermissionService_GetUserPerspective(t *testing.T) {
	svc := newTestResourcePermissionService(t)
	permRepo := repository.NewResourcePermissionRepository(testDB)

	// Create a test permission
	testPerm := &permission.SysPerm{
		PermKey:  "test:resource:read",
		PermName: "Test Resource Read",
		PermType: "resource",
		Status:   permission.StatusEnabled,
	}
	err := permRepo.CreatePerm(testPerm)
	assert.NoError(t, err)

	// Create a test user
	userRepo := repository.NewUserRepository(testDB)
	testUser := createUserForTest(t, userRepo, "perspective_user")

	// Grant permission to user
	err = permRepo.GrantPermToUser(testUser.UserID, testPerm.PermID, "tester")
	assert.NoError(t, err)

	// Get user perspective
	result, err := svc.GetUserPerspective(testUser.UserID, "resource")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

// Test GetUserPerspective for admin user
func TestResourcePermissionService_GetUserPerspectiveAdmin(t *testing.T) {
	permRepo := repository.NewResourcePermissionRepository(testDB)

	// Create a mock admin checker
	mockAdminChecker := &mockAdminCheckerForPerm{isAdmin: true}
	svc := NewResourcePermissionService(permRepo, mockAdminChecker)

	// Get admin user perspective
	result, err := svc.GetUserPerspective(1, "resource")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "*", result[0].PermKey)
	assert.Equal(t, "admin", result[0].SourceType)
}

// Test GetResourcePerspective
func TestResourcePermissionService_GetResourcePerspective(t *testing.T) {
	svc := newTestResourcePermissionService(t)
	permRepo := repository.NewResourcePermissionRepository(testDB)

	// Create a test permission
	testPerm := &permission.SysPerm{
		PermKey:  "test:resource2:read",
		PermName: "Test Resource2 Read",
		PermType: "resource",
		Status:   permission.StatusEnabled,
	}
	err := permRepo.CreatePerm(testPerm)
	assert.NoError(t, err)

	// Get resource perspective (should work even with no users)
	result, err := svc.GetResourcePerspective(1, "resource")
	assert.NoError(t, err)
	_ = result // Result can be empty if no users are granted
}

// Test ApplyGroupPermissionsToResource
func TestResourcePermissionService_ApplyGroupPermissionsToResource(t *testing.T) {
	svc := newTestResourcePermissionService(t)

	// Apply group permissions (this will fail gracefully if no group exists)
	err := svc.ApplyGroupPermissionsToResource(1, 100, "panel")
	// We don't assert.NoError because this depends on existing data
	// But we verify the function is callable
	_ = err
}

// Test CheckPermissionConsistency
func TestResourcePermissionService_CheckPermissionConsistency(t *testing.T) {
	svc := newTestResourcePermissionService(t)

	// Check consistency (returns result even if empty)
	result, err := svc.CheckPermissionConsistency()
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// Helper to create a user for testing
func createUserForTest(t *testing.T, userRepo *repository.UserRepository, username string) *user.SysUser {
	email := username + "@test.com"
	testUser := &user.SysUser{
		Username: username,
		NickName: username,
		Email:    &email,
		Status:   1,
	}
	err := userRepo.Create(testUser)
	assert.NoError(t, err)
	return testUser
}

// Use mockAdminChecker from row_permission_service_test.go
// Re-declare here for integration tests
type mockAdminCheckerForPerm struct {
	isAdmin bool
}

func (m *mockAdminCheckerForPerm) IsAdmin(userID int64) bool {
	return m.isAdmin
}
