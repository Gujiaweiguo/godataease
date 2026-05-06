package handler

import (
	"strings"

	"dataease/backend/internal/service"
	transportmiddleware "dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

func buildPermissionScope(c *gin.Context) (service.PermissionMutationScope, error) {
	isAdmin := isAdminRequest(c)
	scope := service.PermissionMutationScope{
		ActorID:  int64(transportmiddleware.GetUserID(c)),
		OrgID:    transportmiddleware.GetOrgID(c),
		Username: transportmiddleware.GetUsername(c),
	}
	if isAdmin {
		scope.OrgID = 0
		return scope, nil
	}
	if scope.ActorID > 0 && scope.OrgID <= 0 {
		return scope, service.ErrInvalidOrgContext
	}
	return scope, nil
}

func isAdminRequest(c *gin.Context) bool {
	return strings.EqualFold(strings.TrimSpace(transportmiddleware.GetRole(c)), "admin")
}
