package handler

import (
	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterCompatibilityBridgeRoutes registers Java-era API compatibility routes.
//
// P4 Legacy Compat Contract classification (C1 policy lock):
//
//	PERMANENT SHIM (保留 shim):
//	  /api/login/localLogin — external system SSO dependency (handled elsewhere)
//	  /de2api/*             — plugin/external system dependency
//	  /xpackComponent/*     — enterprise plugin dependency
//
//	FRONTEND MIGRATION (C2 中执行):
//	  /user/org/option      — migrate frontend to canonical org endpoint
//	  /user/list            — migrate frontend to canonical user endpoint
//	  /user/create          — migrate frontend to canonical user endpoint
//	  /user/edit            — migrate frontend to canonical user endpoint
//	  /user/delete/:id      — migrate frontend to canonical user endpoint
//	  /user/options         — migrate frontend to canonical user endpoint
//	  /user/byCurOrg        — migrate frontend to canonical user endpoint
//	  /user/resetPwd/:uid   — migrate frontend to canonical user endpoint
//	  /org/create           — migrate frontend to canonical org endpoint
//	  /org/update           — migrate frontend to canonical org endpoint
//	  /org/delete/:orgId    — migrate frontend to canonical org endpoint
//	  /org/list             — migrate frontend to canonical org endpoint
//
//	DUAL-SUPPORT TRANSITION (C1 keep, C3 migrate):
//	  /datasource/*         — datasource CRUD compatibility (C3 migrate to canonical)
//	  /datasetTree/*        — dataset tree compatibility (C3 migrate to canonical)
//	  /datasetData/*        — dataset data compatibility (C3 migrate to canonical)
//	  /chartData/*          — chart data compatibility (C3 migrate to canonical)
//	  /chart/*              — chart CRUD compatibility (C3 migrate to canonical)
//	  /datasetField/*       — dataset field compatibility (C3 migrate to canonical)
func RegisterCompatibilityBridgeRoutes(r gin.IRouter, user *UserHandler, org *OrgHandler, datasourceHandler *DatasourceHandler, datasetHandler *DatasetHandler, chartHandler *ChartHandler, permMiddleware *middleware.PermissionMiddleware) {
	if datasourceHandler != nil {
		registerDatasourceCompatRoutes(r, datasourceHandler, bridgeGetCurrentUserID, bridgeGetCurrentUsername)
	}

	if datasetHandler != nil {
		registerDatasetCompatRoutes(r, datasetHandler, permMiddleware)
	}

	if chartHandler != nil {
		registerChartCompatRoutes(r, chartHandler, datasetHandler, permMiddleware)
	}

	if user != nil {
		userGroup := r.Group("/user")
		{
			userGroup.POST("/list", user.ListUsers)
			userGroup.POST("/create", user.CreateUser)
			userGroup.POST("/edit", user.UpdateUser)
			userGroup.POST("/update", user.UpdateUser)
			userGroup.POST("/delete/:id", user.DeleteUser)
			userGroup.GET("/options", user.GetUserOptions)
			userGroup.GET("/org/option", user.GetUserOptions)
			userGroup.GET("/defaultPwd", user.GetDefaultPassword)
			userGroup.POST("/enable", user.SwitchEnable)
			userGroup.POST("/resetPwd/:uid", middleware.AuditLog(middleware.AuditConfig{
				ActionType:   audit.ActionTypeUserAction,
				ActionName:   "RESET_USER_PASSWORD",
				ResourceType: audit.ResourceTypeUser,
			}), user.ResetPasswordCompat)
		}
	}

	if org != nil {
		orgGroup := r.Group("/org")
		{
			orgGroup.POST("/create", org.CreateOrg)
			orgGroup.POST("/update", org.UpdateOrg)
			orgGroup.POST("/delete/:orgId", org.DeleteOrg)
			orgGroup.GET("/list", org.ListOrgs)
			orgGroup.GET("/info/:orgId", org.GetOrgByID)
			orgGroup.GET("/tree", org.GetOrgTree)
			orgGroup.GET("/checkName", org.CheckOrgName)
			orgGroup.POST("/updateStatus", org.UpdateOrgStatus)
			orgGroup.GET("/children/:parentId", org.GetChildOrgs)
		}
	}
}

func bridgeGetCurrentUserID(c *gin.Context) int64 {
	if uid, exists := c.Get("user_id"); exists {
		switch v := uid.(type) {
		case int64:
			return v
		case uint64:
			return int64(v)
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 1
}

func bridgeGetCurrentUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if s, ok := username.(string); ok {
			return s
		}
	}
	return defaultAdminCredential
}
