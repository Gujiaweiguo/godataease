package handler

import (
	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

// CompatRouteMapping records the canonical path for a legacy compat route.
type CompatRouteMapping struct {
	LegacyPath      string `json:"legacyPath"`
	CanonicalPath   string `json:"canonicalPath"`
	Bucket          string `json:"bucket"`
	MigrationStatus string `json:"migrationStatus"`
}

const (
	bucketFrontendMigration = "FRONTEND_MIGRATION"
	statusPending           = "pending"
)

// compatRouteMappings is the registry of compat routes tracked for frontend migration.
var compatRouteMappings = []CompatRouteMapping{
	{LegacyPath: "/user/org/option", CanonicalPath: "/api/user/options", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/user/list", CanonicalPath: "/api/user/list", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/user/create", CanonicalPath: "/api/user/create", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/user/edit", CanonicalPath: "/api/user/update", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/user/delete/:id", CanonicalPath: "/api/user/delete/:id", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/user/options", CanonicalPath: "/api/user/options", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/user/resetPwd/:uid", CanonicalPath: "/api/user/resetPwd/:id", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/org/create", CanonicalPath: "/api/org/create", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/org/update", CanonicalPath: "/api/org/update", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/org/delete/:orgId", CanonicalPath: "/api/org/delete/:orgId", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
	{LegacyPath: "/org/list", CanonicalPath: "/api/org/list", Bucket: bucketFrontendMigration, MigrationStatus: statusPending},
}

type CompatibilityBridgeHandler struct{}

func GetCompatRouteMappings() []CompatRouteMapping {
	return append([]CompatRouteMapping(nil), compatRouteMappings...)
}

func (h *CompatibilityBridgeHandler) GetCompatRouteMappings(c *gin.Context) {
	response.Success(c, GetCompatRouteMappings())
}

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
func RegisterCompatibilityBridgeRoutes(r gin.IRouter, user *UserHandler, org *OrgHandler, datasourceHandler *DatasourceHandler, datasetHandler *DatasetHandler, chartHandler *ChartHandler, permMiddleware *middleware.PermissionMiddleware, menuAuthMiddlewares ...*middleware.MenuAuthMiddleware) {
	compatibilityBridgeHandler := &CompatibilityBridgeHandler{}
	if user != nil && org != nil {
		if basePathProvider, ok := r.(interface{ BasePath() string }); ok && basePathProvider.BasePath() == "/api" {
			adminGroup := r.Group("/admin")
			adminGroup.GET("/compat-route-mappings", compatibilityBridgeHandler.GetCompatRouteMappings)
		}
	}

	var menuAuthMiddleware *middleware.MenuAuthMiddleware
	if len(menuAuthMiddlewares) > 0 {
		menuAuthMiddleware = menuAuthMiddlewares[0]
	}
	if datasourceHandler != nil {
		registerDatasourceCompatRoutes(r, datasourceHandler, permMiddleware, menuAuthMiddleware)
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
