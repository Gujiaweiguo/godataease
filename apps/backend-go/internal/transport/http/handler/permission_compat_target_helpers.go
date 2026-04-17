package handler

// Target permission resolution and save helpers for PermissionCompatHandler.
// Extracted from permission_compat_handler.go for readability.

import (
	"fmt"
	"strings"

	"dataease/backend/internal/domain/permission"
)

type targetPermissionTarget struct {
	TargetType  string                  `json:"targetType"`
	SourceType  string                  `json:"sourceType"`
	TargetID    int64                   `json:"targetId"`
	SourceID    int64                   `json:"sourceId"`
	PermIDs     []int64                 `json:"permIds"`
	Permissions []targetPermissionEntry `json:"permissions"`
}

type targetPermissionEntry struct {
	ID int64 `json:"id"`
}

func (h *PermissionCompatHandler) collectMatchedTargetPermIDs(target targetPermissionTarget, resourceType string) ([]int64, error) {
	targetType := normalizeTargetType(target.TargetType, target.SourceType)
	targetID := normalizeTargetID(target.TargetID, target.SourceID)
	if targetType != permission.AuthTargetTypeRole || targetID <= 0 {
		return nil, fmt.Errorf("only role targets are supported in the current resource-perspective save slice")
	}

	permIDs := extractTargetPermIDs(target)
	matchedPermIDs := make([]int64, 0, len(permIDs))
	for _, permID := range permIDs {
		matches, err := h.permissionMatchesResourceType(permID, resourceType)
		if err != nil {
			return nil, err
		}
		if matches {
			matchedPermIDs = append(matchedPermIDs, permID)
		}
	}
	return matchedPermIDs, nil
}

func extractTargetPermIDs(target targetPermissionTarget) []int64 {
	if len(target.PermIDs) > 0 || len(target.Permissions) == 0 {
		return target.PermIDs
	}

	permIDs := make([]int64, 0, len(target.Permissions))
	for _, item := range target.Permissions {
		if item.ID > 0 {
			permIDs = append(permIDs, item.ID)
		}
	}
	return permIDs
}

func (h *PermissionCompatHandler) collectDirectResourcePermIDs(resourceID int64, resourceType string) ([]int64, error) {
	items, err := h.resourcePermService.GetResourcePerspective(resourceID, resourceType)
	if err != nil {
		return nil, err
	}
	permIDs := make([]int64, 0)
	seen := make(map[int64]struct{})
	for _, item := range items {
		if item == nil || item.SourceType != "direct" || strings.TrimSpace(item.PermKey) == "" {
			continue
		}
		perm, resolveErr := h.resourcePermService.ResolvePermission(resourceType, item.PermKey)
		if resolveErr != nil || perm == nil || perm.PermID <= 0 {
			continue
		}
		if _, ok := seen[perm.PermID]; ok {
			continue
		}
		seen[perm.PermID] = struct{}{}
		permIDs = append(permIDs, perm.PermID)
	}
	return permIDs, nil
}

func uniqueInt64(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (h *PermissionCompatHandler) saveRolePerms(roleID int64, targetPermIDs []int64) error {
	target := make(map[int64]struct{}, len(targetPermIDs))
	for _, id := range targetPermIDs {
		if id > 0 {
			target[id] = struct{}{}
		}
	}

	currentPermIDs, err := h.resourcePermService.GetRolePermissionIDs(roleID)
	if err != nil {
		return err
	}

	current := make(map[int64]struct{}, len(currentPermIDs))
	for _, id := range currentPermIDs {
		current[id] = struct{}{}
	}

	for id := range target {
		if _, exists := current[id]; !exists {
			if grantErr := h.resourcePermService.GrantPermissionToRole(roleID, id); grantErr != nil {
				return grantErr
			}
		}
	}

	for id := range current {
		if _, keep := target[id]; !keep {
			if revokeErr := h.resourcePermService.RevokePermissionFromRole(roleID, id); revokeErr != nil {
				return revokeErr
			}
		}
	}

	return nil
}

func (h *PermissionCompatHandler) saveRolePermsForResourceType(roleID int64, targetPermIDs []int64, resourceType string) error {
	target := make(map[int64]struct{}, len(targetPermIDs))
	for _, id := range targetPermIDs {
		if id <= 0 {
			continue
		}
		matches, err := h.permissionMatchesResourceType(id, resourceType)
		if err != nil {
			return err
		}
		if !matches {
			return fmt.Errorf("permission %d does not belong to resource type %s", id, resourceType)
		}
		target[id] = struct{}{}
	}

	currentPermIDs, err := h.resourcePermService.GetRolePermissionIDs(roleID)
	if err != nil {
		return err
	}

	current := make(map[int64]struct{}, len(currentPermIDs))
	for _, id := range currentPermIDs {
		matches, matchErr := h.permissionMatchesResourceType(id, resourceType)
		if matchErr != nil {
			return matchErr
		}
		if matches {
			current[id] = struct{}{}
		}
	}

	for id := range target {
		if _, exists := current[id]; !exists {
			if grantErr := h.resourcePermService.GrantPermissionToRole(roleID, id); grantErr != nil {
				return grantErr
			}
		}
	}

	for id := range current {
		if _, keep := target[id]; !keep {
			if revokeErr := h.resourcePermService.RevokePermissionFromRole(roleID, id); revokeErr != nil {
				return revokeErr
			}
		}
	}

	return nil
}

func (h *PermissionCompatHandler) permissionMatchesResourceType(permID int64, resourceType string) (bool, error) {
	if h.resourcePermService == nil {
		return false, fmt.Errorf("resource permission service is unavailable")
	}

	perm, err := h.resourcePermService.GetPermissionByID(permID)
	if err != nil {
		return false, err
	}
	if perm == nil {
		return false, fmt.Errorf("permission %d not found", permID)
	}

	return strings.HasPrefix(perm.PermKey, resourceType+":"), nil
}

func normalizeTargetType(targetType, sourceType string) string {
	if targetType != "" {
		return targetType
	}
	return sourceType
}

func normalizeTargetID(targetID, sourceID int64) int64 {
	if targetID > 0 {
		return targetID
	}
	return sourceID
}
