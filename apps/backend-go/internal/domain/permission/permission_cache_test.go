package permission

import (
	"context"
	"errors"
	"path"
	"testing"
	"time"
)

type mockCacheBackend struct {
	data       map[string]string
	expiration map[string]time.Duration
}

func newMockCacheBackend() *mockCacheBackend {
	return &mockCacheBackend{
		data:       make(map[string]string),
		expiration: make(map[string]time.Duration),
	}
}

func (m *mockCacheBackend) Get(ctx context.Context, key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", ErrCacheMiss
}

func (m *mockCacheBackend) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	switch v := value.(type) {
	case string:
		m.data[key] = v
	case []byte:
		m.data[key] = string(v)
	default:
		m.data[key] = ""
	}
	m.expiration[key] = expiration
	return nil
}

func (m *mockCacheBackend) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
		delete(m.expiration, key)
	}
	return nil
}

func (m *mockCacheBackend) DelByPattern(ctx context.Context, pattern string) error {
	for key := range m.data {
		matched, err := path.Match(pattern, key)
		if err != nil {
			return err
		}
		if matched {
			delete(m.data, key)
			delete(m.expiration, key)
		}
	}
	return nil
}

func (m *mockCacheBackend) Exists(ctx context.Context, keys ...string) (int64, error) {
	var count int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			count++
		}
	}
	return count, nil
}

var ErrCacheMiss = errors.New("cache miss")

func TestPermissionCacheService_UserPermissions(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)
	ctx := context.Background()

	userID := int64(123)
	permIDs := []int64{1, 2, 3}

	_, found := svc.GetUserPermissions(ctx, userID)
	if found {
		t.Error("Should not find permissions before setting")
	}

	err := svc.SetUserPermissions(ctx, userID, permIDs)
	if err != nil {
		t.Errorf("Failed to set user permissions: %v", err)
	}

	got, found := svc.GetUserPermissions(ctx, userID)
	if !found {
		t.Error("Should find permissions after setting")
	}
	if len(got) != len(permIDs) {
		t.Errorf("Expected %d permissions, got %d", len(permIDs), len(got))
	}

	err = svc.InvalidateUserPermissions(ctx, userID)
	if err != nil {
		t.Errorf("Failed to invalidate user permissions: %v", err)
	}

	_, found = svc.GetUserPermissions(ctx, userID)
	if found {
		t.Error("Should not find permissions after invalidation")
	}
}

func TestPermissionCacheService_RolePermissions(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)
	ctx := context.Background()

	roleID := int64(456)
	permIDs := []int64{4, 5, 6}

	err := svc.SetRolePermissions(ctx, roleID, permIDs)
	if err != nil {
		t.Errorf("Failed to set role permissions: %v", err)
	}

	got, found := svc.GetRolePermissions(ctx, roleID)
	if !found {
		t.Error("Should find permissions after setting")
	}
	if len(got) != len(permIDs) {
		t.Errorf("Expected %d permissions, got %d", len(permIDs), len(got))
	}

	err = svc.InvalidateRolePermissions(ctx, roleID)
	if err != nil {
		t.Errorf("Failed to invalidate role permissions: %v", err)
	}

	_, found = svc.GetRolePermissions(ctx, roleID)
	if found {
		t.Error("Should not find permissions after invalidation")
	}
}

func TestPermissionCacheService_ResourcePermission(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)
	ctx := context.Background()

	resourceType := "dashboard"
	resourceID := int64(789)
	permKey := "export"

	if _, found := svc.GetResourcePermission(ctx, resourceType, resourceID, permKey); found {
		t.Error("Should not find permission before setting")
	}

	err := svc.SetResourcePermission(ctx, resourceType, resourceID, permKey, true)
	if err != nil {
		t.Errorf("Failed to set resource permission: %v", err)
	}

	hasPerm, found := svc.GetResourcePermission(ctx, resourceType, resourceID, permKey)
	if !found {
		t.Error("Should find permission after setting")
	}
	if !hasPerm {
		t.Error("Permission should be true")
	}

	err = svc.SetResourcePermission(ctx, resourceType, resourceID, permKey, false)
	if err != nil {
		t.Errorf("Failed to update resource permission: %v", err)
	}

	if perm, _ := svc.GetResourcePermission(ctx, resourceType, resourceID, permKey); perm {
		t.Error("Permission should be false after update")
	}
}

func TestPermissionCacheService_BuildKey(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)

	tests := []struct {
		parts    []string
		expected string
	}{
		{[]string{"user_perms", "123"}, "de:perm:user_perms:123"},
		{[]string{"role_perms", "456"}, "de:perm:role_perms:456"},
		{[]string{"resource_perms", "dashboard", "789", "export"}, "de:perm:resource_perms:dashboard:789:export"},
	}

	for _, tt := range tests {
		key := svc.buildKey(tt.parts...)
		if key != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, key)
		}
	}
}

func TestPermissionCacheService_DefaultExpiration(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 0)

	if svc.expiresIn != CacheDefaultExpiration {
		t.Errorf("Expected default expiration %v, got %v", CacheDefaultExpiration, svc.expiresIn)
	}

	svc2 := NewPermissionCacheService(cache, -1)
	if svc2.expiresIn != CacheDefaultExpiration {
		t.Errorf("Expected default expiration for negative value, got %v", svc2.expiresIn)
	}
}

func TestPermissionCacheService_InvalidateByUserID(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)
	ctx := context.Background()

	userID := int64(123)

	if err := svc.SetUserPermissions(ctx, userID, []int64{1, 2, 3}); err != nil {
		t.Errorf("Failed to set user permissions: %v", err)
	}

	err := svc.InvalidateByUserID(ctx, userID)
	if err != nil {
		t.Errorf("Failed to invalidate by user ID: %v", err)
	}

	_, found := svc.GetUserPermissions(ctx, userID)
	if found {
		t.Error("User permissions should be invalidated")
	}
}

func TestPermissionCacheService_InvalidateByRoleID(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)
	ctx := context.Background()

	roleID := int64(456)

	if err := svc.SetRolePermissions(ctx, roleID, []int64{4, 5, 6}); err != nil {
		t.Errorf("Failed to set role permissions: %v", err)
	}

	err := svc.InvalidateByRoleID(ctx, roleID)
	if err != nil {
		t.Errorf("Failed to invalidate by role ID: %v", err)
	}

	_, found := svc.GetRolePermissions(ctx, roleID)
	if found {
		t.Error("Role permissions should be invalidated")
	}
}

func TestPermissionCacheService_InvalidateResourcePermissions(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)
	ctx := context.Background()

	requireNoError := func(err error) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	requireNoError(svc.SetResourcePermission(ctx, "dashboard", 789, "view", true))
	requireNoError(svc.SetResourcePermission(ctx, "dashboard", 789, "export", true))
	requireNoError(svc.SetResourcePermission(ctx, "dashboard", 790, "view", true))

	requireNoError(svc.InvalidateResourcePermissions(ctx, "dashboard", 789))

	if _, found := svc.GetResourcePermission(ctx, "dashboard", 789, "view"); found {
		t.Fatal("dashboard 789 view permission should be invalidated")
	}
	if _, found := svc.GetResourcePermission(ctx, "dashboard", 789, "export"); found {
		t.Fatal("dashboard 789 export permission should be invalidated")
	}
	if _, found := svc.GetResourcePermission(ctx, "dashboard", 790, "view"); !found {
		t.Fatal("dashboard 790 permission should remain")
	}
}

func TestPermissionCacheService_InvalidateRowPermissions(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)
	ctx := context.Background()

	obj := &DatasetRowPermissionsTreeObj{Logic: "OR"}
	if err := svc.SetRowPermissions(ctx, 100, 1, obj); err != nil {
		t.Fatalf("set row permissions failed: %v", err)
	}
	if err := svc.SetRowPermissions(ctx, 100, 2, obj); err != nil {
		t.Fatalf("set row permissions failed: %v", err)
	}
	if err := svc.SetRowPermissions(ctx, 101, 1, obj); err != nil {
		t.Fatalf("set row permissions failed: %v", err)
	}

	if err := svc.InvalidateRowPermissions(ctx, 100); err != nil {
		t.Fatalf("invalidate row permissions failed: %v", err)
	}

	if _, found := svc.GetRowPermissions(ctx, 100, 1); found {
		t.Fatal("dataset 100 user 1 row permissions should be invalidated")
	}
	if _, found := svc.GetRowPermissions(ctx, 100, 2); found {
		t.Fatal("dataset 100 user 2 row permissions should be invalidated")
	}
	if _, found := svc.GetRowPermissions(ctx, 101, 1); !found {
		t.Fatal("dataset 101 user 1 row permissions should remain")
	}
}

func TestPermissionCacheService_InvalidateAll(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)
	ctx := context.Background()

	if err := svc.SetUserPermissions(ctx, 123, []int64{1}); err != nil {
		t.Fatalf("set user permissions failed: %v", err)
	}
	if err := svc.SetRolePermissions(ctx, 456, []int64{2}); err != nil {
		t.Fatalf("set role permissions failed: %v", err)
	}
	if err := svc.SetColumnPermissions(ctx, 789, []*DataPermColumn{{DatasetID: 789, FieldName: "mobile"}}); err != nil {
		t.Fatalf("set column permissions failed: %v", err)
	}

	if err := svc.InvalidateAll(ctx); err != nil {
		t.Fatalf("invalidate all failed: %v", err)
	}
	if len(cache.data) != 0 {
		t.Fatalf("expected all cache entries removed, got %d", len(cache.data))
	}
}

func TestPermissionCacheService_InvalidateByUserID_RemovesWildcardRowPermissions(t *testing.T) {
	cache := newMockCacheBackend()
	svc := NewPermissionCacheService(cache, 30*time.Minute)
	ctx := context.Background()

	obj := &DatasetRowPermissionsTreeObj{Logic: "OR"}
	if err := svc.SetUserPermissions(ctx, 123, []int64{1, 2, 3}); err != nil {
		t.Fatalf("set user permissions failed: %v", err)
	}
	if err := svc.SetRowPermissions(ctx, 100, 123, obj); err != nil {
		t.Fatalf("set row permissions failed: %v", err)
	}
	if err := svc.SetRowPermissions(ctx, 101, 123, obj); err != nil {
		t.Fatalf("set row permissions failed: %v", err)
	}
	if err := svc.SetRowPermissions(ctx, 101, 456, obj); err != nil {
		t.Fatalf("set row permissions failed: %v", err)
	}

	if err := svc.InvalidateByUserID(ctx, 123); err != nil {
		t.Fatalf("invalidate by user id failed: %v", err)
	}

	if _, found := svc.GetUserPermissions(ctx, 123); found {
		t.Fatal("user permissions should be invalidated")
	}
	if _, found := svc.GetRowPermissions(ctx, 100, 123); found {
		t.Fatal("dataset 100 row permissions for user 123 should be invalidated")
	}
	if _, found := svc.GetRowPermissions(ctx, 101, 123); found {
		t.Fatal("dataset 101 row permissions for user 123 should be invalidated")
	}
	if _, found := svc.GetRowPermissions(ctx, 101, 456); !found {
		t.Fatal("row permissions for other users should remain")
	}
}
