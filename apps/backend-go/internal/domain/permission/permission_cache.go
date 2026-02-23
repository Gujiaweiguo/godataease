package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	CacheKeyPrefix         = "de:perm:"
	CacheKeyUserPerms      = "user_perms"
	CacheKeyRolePerms      = "role_perms"
	CacheKeyResourcePerms  = "resource_perms"
	CacheKeyRowPerms       = "row_perms"
	CacheKeyColumnPerms    = "column_perms"
	CacheDefaultExpiration = 30 * time.Minute
)

type CacheBackend interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
}

type PermissionCacheService struct {
	cache     CacheBackend
	expiresIn time.Duration
}

func NewPermissionCacheService(cache CacheBackend, expiresIn time.Duration) *PermissionCacheService {
	if expiresIn <= 0 {
		expiresIn = CacheDefaultExpiration
	}
	return &PermissionCacheService{
		cache:     cache,
		expiresIn: expiresIn,
	}
}

func (s *PermissionCacheService) buildKey(parts ...string) string {
	return CacheKeyPrefix + fmt.Sprintf("%s", joinKey(parts...))
}

func joinKey(parts ...string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ":"
		}
		result += p
	}
	return result
}

func (s *PermissionCacheService) GetUserPermissions(ctx context.Context, userID int64) ([]int64, bool) {
	key := s.buildKey(CacheKeyUserPerms, fmt.Sprintf("%d", userID))
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, false
	}

	var permIDs []int64
	if err := json.Unmarshal([]byte(data), &permIDs); err != nil {
		return nil, false
	}
	return permIDs, true
}

func (s *PermissionCacheService) SetUserPermissions(ctx context.Context, userID int64, permIDs []int64) error {
	key := s.buildKey(CacheKeyUserPerms, fmt.Sprintf("%d", userID))
	data, err := json.Marshal(permIDs)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, key, data, s.expiresIn)
}

func (s *PermissionCacheService) InvalidateUserPermissions(ctx context.Context, userID int64) error {
	key := s.buildKey(CacheKeyUserPerms, fmt.Sprintf("%d", userID))
	return s.cache.Del(ctx, key)
}

func (s *PermissionCacheService) GetRolePermissions(ctx context.Context, roleID int64) ([]int64, bool) {
	key := s.buildKey(CacheKeyRolePerms, fmt.Sprintf("%d", roleID))
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, false
	}

	var permIDs []int64
	if err := json.Unmarshal([]byte(data), &permIDs); err != nil {
		return nil, false
	}
	return permIDs, true
}

func (s *PermissionCacheService) SetRolePermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	key := s.buildKey(CacheKeyRolePerms, fmt.Sprintf("%d", roleID))
	data, err := json.Marshal(permIDs)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, key, data, s.expiresIn)
}

func (s *PermissionCacheService) InvalidateRolePermissions(ctx context.Context, roleID int64) error {
	key := s.buildKey(CacheKeyRolePerms, fmt.Sprintf("%d", roleID))
	return s.cache.Del(ctx, key)
}

func (s *PermissionCacheService) GetResourcePermission(ctx context.Context, resourceType string, resourceID int64, permKey string) (bool, bool) {
	key := s.buildKey(CacheKeyResourcePerms, resourceType, fmt.Sprintf("%d", resourceID), permKey)
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return false, false
	}
	return data == "1", true
}

func (s *PermissionCacheService) SetResourcePermission(ctx context.Context, resourceType string, resourceID int64, permKey string, hasPermission bool) error {
	key := s.buildKey(CacheKeyResourcePerms, resourceType, fmt.Sprintf("%d", resourceID), permKey)
	value := "0"
	if hasPermission {
		value = "1"
	}
	return s.cache.Set(ctx, key, value, s.expiresIn)
}

func (s *PermissionCacheService) InvalidateResourcePermissions(ctx context.Context, resourceType string, resourceID int64) error {
	pattern := s.buildKey(CacheKeyResourcePerms, resourceType, fmt.Sprintf("%d", resourceID), "*")
	return s.cache.Del(ctx, pattern)
}

func (s *PermissionCacheService) GetRowPermissions(ctx context.Context, datasetID, userID int64) (*DatasetRowPermissionsTreeObj, bool) {
	key := s.buildKey(CacheKeyRowPerms, fmt.Sprintf("ds:%d:user:%d", datasetID, userID))
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, false
	}

	var obj DatasetRowPermissionsTreeObj
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return nil, false
	}
	return &obj, true
}

func (s *PermissionCacheService) SetRowPermissions(ctx context.Context, datasetID, userID int64, obj *DatasetRowPermissionsTreeObj) error {
	key := s.buildKey(CacheKeyRowPerms, fmt.Sprintf("ds:%d:user:%d", datasetID, userID))
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, key, data, s.expiresIn)
}

func (s *PermissionCacheService) InvalidateRowPermissions(ctx context.Context, datasetID int64) error {
	key := s.buildKey(CacheKeyRowPerms, fmt.Sprintf("ds:%d:*", datasetID))
	return s.cache.Del(ctx, key)
}

func (s *PermissionCacheService) GetColumnPermissions(ctx context.Context, datasetID int64) ([]*DataPermColumn, bool) {
	key := s.buildKey(CacheKeyColumnPerms, fmt.Sprintf("ds:%d", datasetID))
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, false
	}

	var perms []*DataPermColumn
	if err := json.Unmarshal([]byte(data), &perms); err != nil {
		return nil, false
	}
	return perms, true
}

func (s *PermissionCacheService) SetColumnPermissions(ctx context.Context, datasetID int64, perms []*DataPermColumn) error {
	key := s.buildKey(CacheKeyColumnPerms, fmt.Sprintf("ds:%d", datasetID))
	data, err := json.Marshal(perms)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, key, data, s.expiresIn)
}

func (s *PermissionCacheService) InvalidateColumnPermissions(ctx context.Context, datasetID int64) error {
	key := s.buildKey(CacheKeyColumnPerms, fmt.Sprintf("ds:%d", datasetID))
	return s.cache.Del(ctx, key)
}

func (s *PermissionCacheService) InvalidateAll(ctx context.Context) error {
	return s.cache.Del(ctx, CacheKeyPrefix+"*")
}

func (s *PermissionCacheService) InvalidateByUserID(ctx context.Context, userID int64) error {
	keys := []string{
		s.buildKey(CacheKeyUserPerms, fmt.Sprintf("%d", userID)),
		s.buildKey(CacheKeyRowPerms, fmt.Sprintf("*:user:%d", userID)),
	}
	return s.cache.Del(ctx, keys...)
}

func (s *PermissionCacheService) InvalidateByRoleID(ctx context.Context, roleID int64) error {
	return s.cache.Del(ctx, s.buildKey(CacheKeyRolePerms, fmt.Sprintf("%d", roleID)))
}
