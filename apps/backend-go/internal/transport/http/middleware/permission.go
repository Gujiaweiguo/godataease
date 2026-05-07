package middleware

import (
	"dataease/backend/internal/pkg/errno"
	"dataease/backend/internal/pkg/response"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

func Permission(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		if role == "" {
			response.Unauthorized(c, response.MsgAuthenticationRequired)
			return
		}

		if role != "admin" && role != requiredRole {
			response.Forbidden(c, response.MsgInsufficientPermission)
			return
		}

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return Permission("admin")
}

const (
	RowPermissionDatasetIDKey  = "row_permission_dataset_id"
	RowPermissionDatasetIDsKey = "row_permission_dataset_ids"
	RowPermissionFilterKey     = "row_permission_filter"
)

var (
	rowPermissionAdminChecker       AdminChecker
	rowPermissionDatasetOrgVerifier DatasetOrgScopeValidator
)

func SetRowPermissionAdminChecker(adminChecker AdminChecker) {
	rowPermissionAdminChecker = adminChecker
}

func SetRowPermissionDatasetOrgValidator(validator DatasetOrgScopeValidator) {
	rowPermissionDatasetOrgVerifier = validator
}

func RowPermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, response.MsgAuthenticationRequired)
			c.Abort()
			return
		}

		orgID := GetOrgID(c)
		isAdmin := rowPermissionAdminChecker != nil && rowPermissionAdminChecker.IsAdmin(int64(userID))
		if orgID > 0 {
			c.Set("org_id", orgID)
		}
		if !isAdmin && rowPermissionAdminChecker != nil && orgID <= 0 {
			response.Forbidden(c, "invalid org context")
			c.Abort()
			return
		}

		datasetIDs, err := extractRowPermissionDatasetIDs(c)
		if err != nil {
			response.BadRequest(c, err.Error())
			c.Abort()
			return
		}

		if !isAdmin && orgID > 0 {
			for _, datasetID := range datasetIDs {
				if rowPermissionDatasetOrgVerifier == nil {
					zap.L().Warn("Skipping dataset org validation: validator unavailable",
						zap.Int64(DatasetIDKey, datasetID),
						zap.Int64("org_id", orgID),
					)
					continue
				}
				allowed, validateErr := rowPermissionDatasetOrgVerifier.DatasetBelongsToOrg(datasetID, orgID)
				if validateErr != nil {
					zap.L().Warn("Dataset org validation failed; allowing request due to unsupported dataset org boundary",
						zap.Int64(DatasetIDKey, datasetID),
						zap.Int64("org_id", orgID),
						zap.Error(validateErr),
					)
					continue
				}
				if !allowed {
					response.Forbidden(c, "dataset does not belong to current organization")
					c.Abort()
					return
				}
			}
		}

		c.Set(RowPermissionDatasetIDKey, datasetIDs[0])
		c.Set(RowPermissionDatasetIDsKey, datasetIDs)
		c.Next()
	}
}

func extractRowPermissionDatasetIDs(c *gin.Context) ([]int64, error) {
	if datasetIDs := GetRowPermissionDatasetIDs(c); len(datasetIDs) > 0 {
		return uniquePositiveIDs(datasetIDs), nil
	}

	ids := make([]int64, 0, 4)
	if datasetID := GetDatasetID(c); datasetID > 0 {
		ids = append(ids, datasetID)
	}
	if resourceID := GetResourceID(c); resourceID > 0 {
		ids = append(ids, resourceID)
	}

	bodyIDs, err := parseDatasetIDsFromBody(c)
	if err != nil {
		return nil, err
	}
	ids = append(ids, bodyIDs...)

	ids = uniquePositiveIDs(ids)
	if len(ids) == 0 {
		return nil, fmt.Errorf(errno.ErrDatasetIDRequired)
	}

	return ids, nil
}

func parseDatasetIDsFromBody(c *gin.Context) ([]int64, error) {
	var payload map[string]interface{}
	err := c.ShouldBindBodyWith(&payload, binding.JSON)
	if err == nil {
		ids := make([]int64, 0, 4)
		for _, key := range []string{"id", "datasetGroupId", "datasetId"} {
			if value, ok := payload[key]; ok {
				id, parseErr := parseResourceIDFromAny(value)
				if parseErr != nil {
					return nil, fmt.Errorf(errno.ErrInvalidDatasetID)
				}
				ids = append(ids, id)
			}
		}

		if values, ok := payload["ids"].([]interface{}); ok {
			for _, value := range values {
				id, parseErr := parseResourceIDFromAny(value)
				if parseErr != nil {
					return nil, fmt.Errorf(errno.ErrInvalidDatasetID)
				}
				ids = append(ids, id)
			}
		}
		return ids, nil
	}

	if errors.Is(err, io.EOF) {
		return nil, nil
	}

	var payloadList []interface{}
	if listErr := c.ShouldBindBodyWith(&payloadList, binding.JSON); listErr == nil {
		ids := make([]int64, 0, len(payloadList))
		for _, value := range payloadList {
			id, parseErr := parseResourceIDFromAny(value)
			if parseErr != nil {
				return nil, fmt.Errorf(errno.ErrInvalidDatasetID)
			}
			ids = append(ids, id)
		}
		return ids, nil
	}

	return nil, fmt.Errorf("invalid request: %w", err)
}

func uniquePositiveIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func GetRowPermissionDatasetID(c *gin.Context) int64 {
	if datasetID, exists := c.Get(RowPermissionDatasetIDKey); exists {
		return datasetID.(int64)
	}
	return 0
}

func GetRowPermissionDatasetIDs(c *gin.Context) []int64 {
	if datasetIDs, exists := c.Get(RowPermissionDatasetIDsKey); exists {
		return datasetIDs.([]int64)
	}
	return nil
}
