package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dataease/backend/internal/app"
	pkgauth "dataease/backend/internal/pkg/auth"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/pkg/metrics"
	"dataease/backend/internal/repository"
	"dataease/backend/internal/service"
	"dataease/backend/internal/transport/http/handler"
	"dataease/backend/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// routeInfo stores information about a registered route
type routeInfo struct {
	method string
	path   string
	source string // e.g., "health", "auth", "user", etc.
}

// detectRouteConflicts checks for duplicate route registrations and returns warning messages
func detectRouteConflicts(routes []routeInfo) []string {
	seen := make(map[string]string)
	var conflicts []string

	for _, r := range routes {
		key := r.method + ":" + r.path
		if existing, ok := seen[key]; ok {
			conflicts = append(conflicts,
				fmt.Sprintf("Route conflict detected: %s %s already registered by '%s', duplicate registration by '%s'",
					r.method, r.path, existing, r.source))
		} else {
			seen[key] = r.source
		}
	}

	return conflicts
}

// collectRoutesFromEngine extracts all registered routes from a Gin engine
func collectRoutesFromEngine(engine *gin.Engine) []routeInfo {
	var routes []routeInfo

	for _, route := range engine.Routes() {
		// Determine source from route path for better identification
		source := determineRouteSource(route.Path)
		routes = append(routes, routeInfo{
			method: route.Method,
			path:   route.Path,
			source: source,
		})
	}

	return routes
}

// determineRouteSource identifies the likely source module of a route based on its path
func determineRouteSource(path string) string {
	// Health and readiness endpoints
	if path == "/health" || path == "/ready" || path == "/metrics" {
		return "system"
	}

	// API routes
	if strings.HasPrefix(path, "/api/") {
		parts := strings.Split(strings.TrimPrefix(path, "/api/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			return "api/" + parts[0]
		}
		return "api"
	}

	// Auth routes (login, logout, etc.)
	if strings.Contains(path, "/login") || strings.Contains(path, "/logout") || strings.Contains(path, "/auth") {
		return "auth"
	}

	// Compatibility bridge routes
	if strings.Contains(path, "/de2api/") || strings.Contains(path, "/compatible") {
		return "compatibility-bridge"
	}

	return "unknown"
}

type Router struct {
	engine                         *gin.Engine
	app                            *app.Application
	db                             *gorm.DB
	permMiddleware                 *middleware.PermissionMiddleware
	auditHandler                   *handler.AuditHandler
	userHandler                    *handler.UserHandler
	orgHandler                     *handler.OrgHandler
	permHandler                    *handler.PermHandler
	embeddedHandler                *handler.EmbeddedHandler
	roleHandler                    *handler.RoleHandler
	roleMenuHandler                *handler.RoleMenuHandler
	menuHandler                    *handler.MenuHandler
	mapHandler                     *handler.MapHandler
	authHandler                    *handler.AuthHandler
	datasourceHandler              *handler.DatasourceHandler
	datasetHandler                 *handler.DatasetHandler
	chartHandler                   *handler.ChartHandler
	visualHandler                  *handler.VisualizationHandler
	linkageHandler                 *handler.LinkageHandler
	linkJumpHandler                *handler.LinkJumpHandler
	outerParamsHandler             *handler.OuterParamsHandler
	visualizationBackgroundHandler *handler.VisualizationBackgroundHandler
	storeHandler                   *handler.StoreHandler
	watermarkHandler               *handler.WatermarkHandler
	systemParamHandler             *handler.SystemParamHandler
	systemVariableHandler          *handler.SystemVariableHandler
	licenseHandler                 *handler.LicenseHandler
	msgCenterHandler               *handler.MsgCenterHandler
	shareHandler                   *handler.ShareHandler
	ticketHandler                  *handler.TicketHandler
	geoHandler                     *handler.GeoHandler
	staticHandler                  *handler.StaticHandler
	subjectHandler                 *handler.SubjectHandler
	fontHandler                    *handler.FontHandler
	pdfTemplateHandler             *handler.PdfTemplateHandler
	exportHandler                  *handler.ExportHandler
	engineHandler                  *handler.EngineHandler
	driverHandler                  *handler.DriverHandler
	templateHandler                *handler.TemplateHandler
	syncHandler                    *handler.SyncHandler
	frontendCompatHandler          *handler.FrontendCompatHandler
	permissionCompatHandler        *handler.PermissionCompatHandler
	customGeoHandler               *handler.CustomGeoHandler
	dataPermissionHandler          *handler.DataPermissionHandler
	resourceGovernanceHandler      *handler.ResourceGovernanceHandler
	menuAuthMiddleware             *middleware.MenuAuthMiddleware
}

func NewRouter(application *app.Application, db *gorm.DB) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Recovery())
	engine.Use(requestLogger())
	engine.Use(metricsMiddleware())

	// Audit module initialization
	auditLogRepo := repository.NewAuditLogRepository(db)
	loginFailureRepo := repository.NewLoginFailureRepository(db)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(db)
	auditService := service.NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)
	auditHandler := handler.NewAuditHandler(auditService)

	middleware.SetAuditService(auditService)

	// User module initialization
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	userPermRepo := repository.NewUserPermRepository(db)
	middleware.SetRoleIDsResolver(func(userID uint64) ([]int64, error) {
		return userRoleRepo.GetRoleIDsByUserID(int64(userID))
	})
	userService := service.NewUserService(userRepo, userRoleRepo, userPermRepo)
	userService.SetAuditService(auditService)
	userImportService := service.NewUserImportService(userService)
	userHandler := handler.NewUserHandler(userService, userImportService)
	// Role module initialization (must be before OrgService as it depends on roleRepo)
	roleRepo := repository.NewRoleRepository(db)
	userService.SetRoleRepository(roleRepo)
	roleService := service.NewRoleService(roleRepo, userRepo, userRoleRepo)
	roleHandler := handler.NewRoleHandler(roleService)

	// Organization module initialization
	orgRepo := repository.NewOrgRepository(db)
	userService.SetOrgRepository(orgRepo)
	orgService := service.NewOrgService(orgRepo, auditService, userRepo, roleRepo)
	orgHandler := handler.NewOrgHandler(orgService)

	// Permission module initialization
	permRepo := repository.NewPermRepository(db)
	permService := service.NewPermService(permRepo)
	permHandler := handler.NewPermHandler(permService)

	// Embedded module initialization
	embeddedRepo := repository.NewEmbeddedRepository(db)
	embeddedService := service.NewEmbeddedService(embeddedRepo)
	embeddedHandler := handler.NewEmbeddedHandler(embeddedService)

	// Menu module initialization
	menuRepo := repository.NewMenuRepository(db)
	roleMenuRepo := repository.NewRoleMenuRepository(db)
	menuService := service.NewMenuServiceWithRoleFilter(menuRepo, roleMenuRepo)
	menuHandler := handler.NewMenuHandler(menuService)

	// RoleMenu module initialization
	roleMenuService := service.NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)
	roleMenuHandler := handler.NewRoleMenuHandler(roleMenuService)
	menuAuthMiddleware := middleware.NewMenuAuthMiddleware(roleMenuService, menuService)

	// Map module initialization
	areaRepo := repository.NewAreaRepository(db)
	mapService := service.NewMapService(areaRepo)
	mapHandler := handler.NewMapHandler(mapService)

	jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{})
	if application != nil && application.Config != nil {
		jwtInstance = pkgauth.NewJWT(&pkgauth.JWTConfig{
			Secret: application.Config.JWT.Secret,
			Expire: application.Config.JWT.Expire,
		})
	}

	authService := service.NewAuthService(userRepo, userRoleRepo, orgRepo, jwtInstance)
	authHandler := handler.NewAuthHandler(authService)
	userHandler.SetAuthService(authService)

	datasourceRepo := repository.NewDatasourceRepository(db)
	syncRepo := repository.NewSyncRepository(db)
	datasourceService := service.NewDatasourceService(datasourceRepo)
	datasourceService.SetUserRepository(userRepo)
	if application != nil && application.Config != nil {
		seatunnelCfg := application.Config.Integration.Seatunnel
		datasourceService.SetSeatunnelConfig(
			seatunnelCfg.Address,
			time.Duration(seatunnelCfg.TimeoutSec)*time.Second,
			seatunnelCfg.MaxRetries,
		)
	}
	datasourceHandler := handler.NewDatasourceHandler(datasourceService)
	syncService := service.NewSyncService(syncRepo, datasourceRepo, datasourceService)
	syncHandler := handler.NewSyncHandler(syncService)
	adminChecker := middleware.NewDefaultAdminChecker([]int64{1})

	// Dataset with row permission integration
	datasetRepo := repository.NewDatasetRepository(db)
	rowPermRepo := repository.NewRowPermissionRepository(db)
	columnPermRepo := repository.NewColumnPermissionRepository(db)
	rowPermService := service.NewRowPermissionService(rowPermRepo, columnPermRepo, userRoleRepo, adminChecker)
	rowPermService.SetDatasetFieldResolver(datasetRepo)
	datasetService := service.NewDatasetServiceWithPermission(datasetRepo, rowPermService)
	if application != nil && application.Config != nil {
		calciteCfg := application.Config.Integration.Calcite
		datasetService.SetCalciteConfig(
			calciteCfg.Address,
			time.Duration(calciteCfg.TimeoutSec)*time.Second,
			calciteCfg.MaxRetries,
		)
	}
	datasetService.SetUserRepository(userRepo)
	datasetService.SetDatasourceRepository(datasourceRepo)

	chartRepo := repository.NewChartRepository(db)
	chartService := service.NewChartService(chartRepo)
	chartService.SetRowPermissionService(rowPermService)
	chartService.SetColumnPermissionService(service.NewColumnPermissionService(columnPermRepo))
	datasetHandler := handler.NewDatasetHandler(datasetService, chartService)

	chartHandler := handler.NewChartHandler(chartService, datasetService)
	dataPermissionService := service.NewDataPermissionAdminService(rowPermRepo, columnPermRepo, chartService)
	dataPermissionHandler := handler.NewDataPermissionHandler(dataPermissionService)

	visualRepo := repository.NewVisualizationRepository(db)
	visualService := service.NewVisualizationService(visualRepo)
	visualHandler := handler.NewVisualizationHandler(visualService)

	linkageRepo := repository.NewLinkageRepository(db)
	linkageService := service.NewLinkageService(linkageRepo)
	linkageHandler := handler.NewLinkageHandler(linkageService)

	linkJumpRepo := repository.NewLinkJumpRepository(db)
	linkJumpService := service.NewLinkJumpService(linkJumpRepo)
	linkJumpHandler := handler.NewLinkJumpHandler(linkJumpService)

	outerParamsRepo := repository.NewOuterParamsRepository(db)
	outerParamsService := service.NewOuterParamsService(outerParamsRepo)
	outerParamsHandler := handler.NewOuterParamsHandler(outerParamsService)

	visualizationBackgroundRepo := repository.NewVisualizationBackgroundRepository(db)
	visualizationBackgroundHandler := handler.NewVisualizationBackgroundHandler(visualizationBackgroundRepo)

	favoriteRepo := repository.NewFavoriteRepository(db)
	storeHandler := handler.NewStoreHandler(favoriteRepo)

	watermarkRepo := repository.NewWatermarkRepository(db)
	watermarkService := service.NewWatermarkService(watermarkRepo)
	watermarkHandler := handler.NewWatermarkHandler(watermarkService)

	systemParamRepo := repository.NewSystemParamRepository(db)
	systemVariableRepo := repository.NewSystemVariableRepository(db)
	systemParamService := service.NewSystemParamService(systemParamRepo, auditService)
	systemVariableService := service.NewSystemVariableService(systemVariableRepo)
	systemParamHandler := handler.NewSystemParamHandler(systemParamService)
	systemVariableHandler := handler.NewSystemVariableHandler(systemVariableService)

	licenseRepo := repository.NewLicenseRepository(db)
	licenseService := service.NewLicenseService(licenseRepo)
	licenseHandler := handler.NewLicenseHandler(licenseService)

	msgCenterRepo := repository.NewMsgCenterRepository(db)
	msgCenterService := service.NewMsgCenterService(msgCenterRepo)
	msgCenterHandler := handler.NewMsgCenterHandler(msgCenterService)

	shareRepo := repository.NewShareRepository(db)
	shareService := service.NewShareService(shareRepo)
	shareHandler := handler.NewShareHandler(shareService)

	ticketRepo := repository.NewTicketRepository(db)
	ticketService := service.NewTicketService(ticketRepo)
	ticketHandler := handler.NewTicketHandler(ticketService)

	// Geo module initialization
	geoRepo := repository.NewGeoRepository(db)
	geoService := service.NewGeoService(geoRepo)
	geoHandler := handler.NewGeoHandler(geoService)

	customGeoRepo := repository.NewCustomGeoRepository(db)
	customGeoHandler := handler.NewCustomGeoHandler(customGeoRepo)

	// Static module initialization
	staticRepo := repository.NewStaticRepository(db)
	storeRepo := repository.NewStoreRepository(db)
	typefaceRepo := repository.NewTypefaceRepository(db)
	staticService := service.NewStaticService(staticRepo, storeRepo, typefaceRepo)
	staticHandler := handler.NewStaticHandler(staticService)

	subjectRepo := repository.NewSubjectRepository(db)
	subjectHandler := handler.NewSubjectHandler(subjectRepo)

	fontHandler := handler.NewFontHandler(typefaceRepo)
	pdfTemplateHandler := handler.NewPdfTemplateHandler()

	// Permission middleware initialization
	resourcePermRepo := repository.NewResourcePermissionRepository(db)
	roleService.SetResourcePermissionRepository(resourcePermRepo)
	resourcePermService := service.NewResourcePermissionService(resourcePermRepo, adminChecker)
	datasourceService.SetResourcePermissionService(resourcePermService)
	datasetService.SetResourcePermissionService(resourcePermService)
	visualService.SetResourcePermissionService(resourcePermService)
	resourceGovernanceAdminService := service.NewResourceGovernanceAdminService(datasourceService, datasetService, visualService)
	exportPermService := service.NewExportPermissionService(resourcePermService, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermService, exportPermService, adminChecker)
	permMiddleware.SetChartDatasetResolver(chartRepo)
	permMiddleware.SetVisualizationTypeResolver(visualService)
	permissionCompatHandler := handler.NewPermissionCompatHandler(menuService, permService, roleMenuService, resourcePermService)
	permissionCompatHandler.SetRoleService(roleService)
	resourceGovernanceHandler := handler.NewResourceGovernanceHandler(resourceGovernanceAdminService, adminChecker)

	// Export module initialization
	exportRepo := repository.NewExportRepository(db)
	datasetService.SetExportRepository(exportRepo)
	exportService := service.NewExportService(exportRepo)
	exportHandler := handler.NewExportHandler(exportService, exportPermService, adminChecker)

	// Engine module initialization
	engineRepo := repository.NewEngineRepository(db)
	engineService := service.NewEngineService(engineRepo)
	engineHandler := handler.NewEngineHandler(engineService)

	// Driver module initialization
	driverRepo := repository.NewDriverRepository(db)
	driverService := service.NewDriverService(driverRepo)
	driverHandler := handler.NewDriverHandler(driverService)

	// Template module initialization
	templateRepo := repository.NewTemplateRepository(db)
	templateService := service.NewTemplateService(templateRepo)
	templateHandler := handler.NewTemplateHandler(templateService)

	frontendCompatHandler := handler.NewFrontendCompatHandler(menuService, datasetService, datasourceService, visualService, userService, userRoleRepo.GetRoleIDsByUserID)

	return &Router{
		engine:                         engine,
		app:                            application,
		db:                             db,
		permMiddleware:                 permMiddleware,
		auditHandler:                   auditHandler,
		userHandler:                    userHandler,
		orgHandler:                     orgHandler,
		permHandler:                    permHandler,
		embeddedHandler:                embeddedHandler,
		roleHandler:                    roleHandler,
		roleMenuHandler:                roleMenuHandler,
		menuHandler:                    menuHandler,
		mapHandler:                     mapHandler,
		authHandler:                    authHandler,
		datasourceHandler:              datasourceHandler,
		datasetHandler:                 datasetHandler,
		chartHandler:                   chartHandler,
		visualHandler:                  visualHandler,
		linkageHandler:                 linkageHandler,
		linkJumpHandler:                linkJumpHandler,
		outerParamsHandler:             outerParamsHandler,
		visualizationBackgroundHandler: visualizationBackgroundHandler,
		storeHandler:                   storeHandler,
		watermarkHandler:               watermarkHandler,
		systemParamHandler:             systemParamHandler,
		systemVariableHandler:          systemVariableHandler,
		licenseHandler:                 licenseHandler,
		msgCenterHandler:               msgCenterHandler,
		shareHandler:                   shareHandler,
		ticketHandler:                  ticketHandler,
		geoHandler:                     geoHandler,
		customGeoHandler:               customGeoHandler,
		staticHandler:                  staticHandler,
		subjectHandler:                 subjectHandler,
		fontHandler:                    fontHandler,
		pdfTemplateHandler:             pdfTemplateHandler,
		exportHandler:                  exportHandler,
		engineHandler:                  engineHandler,
		driverHandler:                  driverHandler,
		templateHandler:                templateHandler,
		syncHandler:                    syncHandler,
		frontendCompatHandler:          frontendCompatHandler,
		permissionCompatHandler:        permissionCompatHandler,
		dataPermissionHandler:          dataPermissionHandler,
		resourceGovernanceHandler:      resourceGovernanceHandler,
		menuAuthMiddleware:             menuAuthMiddleware,
	}
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		)
	}
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())
		metrics.RecordRequest(c.Request.Method, c.FullPath(), status, duration)
	}
}

func (r *Router) RegisterRoutes() {
	r.registerRootRoutes()
	r.registerAPIRoutes()
	r.registerFrontendFallback()
}

func (r *Router) registerRootRoutes() {
	protected := r.engine.Group("")
	if r.app != nil && r.app.Config != nil {
		jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{
			Secret: r.app.Config.JWT.Secret,
			Expire: r.app.Config.JWT.Expire,
		})
		protected.Use(middleware.Auth(jwtInstance))
	}

	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "dataease-backend",
		})
	})

	r.engine.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ready": true,
		})
	})

	r.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	handler.RegisterAuthRoutes(r.engine, r.authHandler)
	if r.userHandler != nil {
		protected.POST("/api/user/switchLanguage", r.userHandler.SwitchLanguage)
	}
	handler.RegisterSystemParamRoutes(r.engine, r.systemParamHandler)
	handler.RegisterLicenseRoutes(r.engine, r.licenseHandler)
	handler.RegisterMsgCenterRoutes(r.engine, r.msgCenterHandler)
	handler.RegisterTicketRoutes(r.engine, r.ticketHandler)
	handler.RegisterVisualizationRoutes(r.engine.Group(""), r.visualHandler, r.permMiddleware)
	handler.RegisterLinkageRoutes(r.engine.Group(""), r.linkageHandler)
	handler.RegisterLinkJumpRoutes(r.engine.Group(""), r.linkJumpHandler)
	handler.RegisterOuterParamsRoutes(r.engine.Group(""), r.outerParamsHandler)
	handler.RegisterVisualizationBackgroundRoutes(r.engine.Group(""), r.visualizationBackgroundHandler)
	handler.RegisterStoreRoutes(r.engine.Group(""), r.storeHandler)
	handler.RegisterCompatibilityBridgeRoutes(r.engine, r.userHandler, r.orgHandler, r.datasourceHandler, r.datasetHandler, nil, nil)
	handler.RegisterCompatibilityBridgeRoutes(r.engine, nil, nil, nil, nil, r.chartHandler, r.permMiddleware)
	handler.RegisterFrontendCompatRoutes(r.engine, protected, r.frontendCompatHandler)
}

func (r *Router) registerAPIRoutes() {
	api := r.engine.Group("/api")
	userAPI := api
	roleAPI := api
	roleDe2API := r.engine.Group("/de2api")
	roleMenuAPI := api
	permissionCompatAPI := api
	permissionCompatDe2API := r.engine.Group("/de2api")
	dataPermissionAPI := api
	datasourceAPI := api
	datasourceDe2API := r.engine.Group("/de2api")
	datasetAPI := api
	datasetDe2API := r.engine.Group("/de2api")
	visualizationDe2API := r.engine.Group("/de2api")
	auditAPI := api
	exportAPI := api
	if r.app != nil && r.app.Config != nil {
		jwtInstance := pkgauth.NewJWT(&pkgauth.JWTConfig{
			Secret: r.app.Config.JWT.Secret,
			Expire: r.app.Config.JWT.Expire,
		})
		protectedDatasourceAPI := r.engine.Group("/api")
		protectedDatasourceAPI.Use(middleware.Auth(jwtInstance))
		datasourceAPI = protectedDatasourceAPI
		protectedUserAPI := r.engine.Group("/api")
		protectedUserAPI.Use(middleware.Auth(jwtInstance))
		userAPI = protectedUserAPI
		protectedRoleAPI := r.engine.Group("/api")
		protectedRoleAPI.Use(middleware.Auth(jwtInstance))
		roleAPI = protectedRoleAPI
		protectedRoleDe2API := r.engine.Group("/de2api")
		protectedRoleDe2API.Use(middleware.Auth(jwtInstance))
		roleDe2API = protectedRoleDe2API
		roleMenuAPI = protectedRoleAPI
		protectedPermissionCompatAPI := r.engine.Group("/api")
		protectedPermissionCompatAPI.Use(middleware.Auth(jwtInstance), r.menuAuthMiddleware.RequireMenuAuth("/system/permission"))
		permissionCompatAPI = protectedPermissionCompatAPI
		protectedPermissionCompatDe2API := r.engine.Group("/de2api")
		protectedPermissionCompatDe2API.Use(middleware.Auth(jwtInstance), r.menuAuthMiddleware.RequireMenuAuth("/system/permission"))
		permissionCompatDe2API = protectedPermissionCompatDe2API
		dataPermissionAPI = protectedRoleAPI
		protectedDatasourceDe2API := r.engine.Group("/de2api")
		protectedDatasourceDe2API.Use(middleware.Auth(jwtInstance))
		datasourceDe2API = protectedDatasourceDe2API

		protectedDatasetAPI := r.engine.Group("/api")
		protectedDatasetAPI.Use(middleware.Auth(jwtInstance))
		datasetAPI = protectedDatasetAPI

		protectedDatasetDe2API := r.engine.Group("/de2api")
		protectedDatasetDe2API.Use(middleware.Auth(jwtInstance))
		datasetDe2API = protectedDatasetDe2API

		protectedVisualizationDe2API := r.engine.Group("/de2api")
		protectedVisualizationDe2API.Use(middleware.Auth(jwtInstance))
		visualizationDe2API = protectedVisualizationDe2API

		protectedAuditAPI := r.engine.Group("/api")
		protectedAuditAPI.Use(middleware.Auth(jwtInstance))
		auditAPI = protectedAuditAPI

		protectedExportAPI := r.engine.Group("/api")
		protectedExportAPI.Use(middleware.Auth(jwtInstance))
		exportAPI = protectedExportAPI
	}
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		handler.RegisterAuditRoutes(auditAPI, r.auditHandler)
		handler.RegisterUserRoutes(userAPI, r.userHandler)
		handler.RegisterOrgRoutes(api, r.orgHandler)
		handler.RegisterPermRoutes(permissionCompatAPI, r.permHandler)
		handler.RegisterEmbeddedRoutes(api, r.embeddedHandler)
		handler.RegisterRoleRoutes(roleAPI, r.roleHandler)
		handler.RegisterRoleRoutes(roleDe2API, r.roleHandler)
		handler.RegisterRoleMenuRoutes(roleMenuAPI, r.roleMenuHandler)
		handler.RegisterMenuRoutes(api, r.menuHandler)
		handler.RegisterPermissionCompatRoutes(permissionCompatAPI, r.permissionCompatHandler)
		handler.RegisterPermissionCompatRoutes(permissionCompatDe2API, r.permissionCompatHandler)
		handler.RegisterResourceGovernanceRoutes(roleAPI, r.resourceGovernanceHandler)
		handler.RegisterDataPermissionRoutes(dataPermissionAPI, r.dataPermissionHandler)
		handler.RegisterMapRoutes(api, r.mapHandler)
		handler.RegisterDatasourceRoutes(datasourceAPI, r.datasourceHandler)
		handler.RegisterCompatibilityBridgeRoutes(datasourceDe2API, nil, nil, r.datasourceHandler, nil, nil, nil)
		handler.RegisterSyncRoutes(api, r.syncHandler)
		r.registerDatasetRoutes(datasetAPI)
		handler.RegisterCompatibilityBridgeRoutes(datasetAPI, nil, nil, nil, r.datasetHandler, nil, r.permMiddleware)
		handler.RegisterCompatibilityBridgeRoutes(datasetDe2API, nil, nil, nil, r.datasetHandler, nil, r.permMiddleware)
		r.registerChartRoutes(api)
		r.registerVisualizationRoutes(api)
		r.registerVisualizationDe2DetailRoute(visualizationDe2API)
		handler.RegisterWatermarkRoutes(api, r.watermarkHandler)
		handler.RegisterSystemParamRoutes(api, r.systemParamHandler)
		handler.RegisterSystemVariableRoutes(api, r.systemVariableHandler)
		handler.RegisterLicenseRoutes(api, r.licenseHandler)
		handler.RegisterMsgCenterRoutes(api, r.msgCenterHandler)
		handler.RegisterShareRoutes(api, r.shareHandler)
		handler.RegisterTicketRoutes(api, r.ticketHandler)
		handler.RegisterGeoRoutes(api, r.geoHandler)
		handler.RegisterCustomGeoRoutes(api, r.customGeoHandler)
		handler.RegisterStaticRoutes(api, r.staticHandler)
		handler.RegisterSubjectRoutes(api, r.subjectHandler)
		handler.RegisterFontRoutes(api, r.fontHandler)
		handler.RegisterPdfTemplateRoutes(api, r.pdfTemplateHandler)
		handler.RegisterExportRoutes(exportAPI, r.exportHandler)
		handler.RegisterEngineRoutes(api, r.engineHandler)
		handler.RegisterDriverRoutes(api, r.driverHandler)
		handler.RegisterTemplateRoutes(api, r.templateHandler)
		handler.RegisterCompatibilityBridgeRoutes(datasourceAPI, nil, nil, r.datasourceHandler, nil, nil, nil)
		api.GET("/panel/view/getComponentInfo/:dvId", r.visualHandler.GetComponentInfo)
		handler.RegisterCompatibilityBridgeRoutes(api, r.userHandler, r.orgHandler, nil, nil, nil, r.permMiddleware)
		handler.RegisterDatasetFieldDeleteRoutes(api.Group("/datasetField"), r.datasetHandler, r.chartHandler)
	}
}

func (r *Router) registerFrontendFallback() {
	frontendDir := os.Getenv("FRONTEND_DIST_PATH")
	if frontendDir == "" {
		frontendDir = "/opt/module/dataease2.0/frontend"
	}

	r.engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if mappedPath, handled := rewriteCompatibilityPath(path); handled {
			c.Request.URL.Path = mappedPath
			r.engine.HandleContext(c)
			return
		}

		if handleSpecialFrontendPath(c, path) {
			return
		}

		serveFrontendAsset(c, path, frontendDir)
	})
}

func rewriteCompatibilityPath(path string) (string, bool) {
	if strings.HasPrefix(path, "/de2api/") {
		mappedPath := strings.TrimPrefix(path, "/de2api")
		if mappedPath == "" {
			mappedPath = "/"
		}
		if !strings.HasPrefix(mappedPath, "/") {
			mappedPath = "/" + mappedPath
		}
		if !strings.HasPrefix(mappedPath, "/api/") {
			mappedPath = "/api" + mappedPath
		}
		return mappedPath, true
	}

	if strings.HasPrefix(path, "/login/websocket/") {
		return strings.TrimPrefix(path, "/login"), true
	}

	return "", false
}

func handleSpecialFrontendPath(c *gin.Context, path string) bool {
	if path == "/websocket" || path == "/websocket/" || strings.HasPrefix(path, "/websocket/") {
		c.Status(http.StatusNoContent)
		return true
	}

	switch path {
	case "/login", "/login/":
		c.Redirect(http.StatusFound, "/#/login")
		return true
	case "/admin-login", "/admin-login/":
		c.Redirect(http.StatusFound, "/#/admin-login")
		return true
	default:
		return false
	}
}

func serveFrontendAsset(c *gin.Context, path, frontendDir string) {
	if strings.HasPrefix(path, "/api/") || path == "/health" || path == "/ready" || path == "/metrics" {
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	cleanPath := strings.TrimPrefix(filepath.Clean(path), "/")
	if cleanPath == "" || cleanPath == "." {
		c.Header("Cache-Control", "no-store")
		c.File(filepath.Join(frontendDir, "index.html"))
		return
	}

	assetPath := filepath.Join(frontendDir, cleanPath)
	if info, err := os.Stat(assetPath); err == nil && !info.IsDir() {
		c.Header("Cache-Control", "no-store")
		c.File(assetPath)
		return
	}

	c.Header("Cache-Control", "no-store")
	c.File(filepath.Join(frontendDir, "index.html"))
}

func (r *Router) registerVisualizationDe2DetailRoute(api *gin.RouterGroup) {
	visualGroup := api.Group("/dataVisualization")
	if r.permMiddleware != nil {
		visualGroup.POST("/findById", r.permMiddleware.CheckVisualizationView(), r.visualHandler.FindByID)
		return
	}
	visualGroup.POST("/findById", r.visualHandler.FindByID)
}

func (r *Router) registerDatasetRoutes(api *gin.RouterGroup) {
	datasetGroup := api.Group("/dataset")
	{
		datasetGroup.POST("/tree", r.datasetHandler.Tree)
		datasetGroup.POST("/fields", r.datasetHandler.Fields)
		datasetGroup.POST("/preview", r.datasetHandler.Preview)
		if r.permMiddleware != nil {
			datasetGroup.POST("/previewWithPerm", r.permMiddleware.CheckDatasetView(), middleware.RowPermissionMiddleware(), r.datasetHandler.PreviewWithPermission)
		} else {
			datasetGroup.POST("/previewWithPerm", r.datasetHandler.PreviewWithPermission)
		}
		datasetGroup.POST("/save", r.datasetHandler.Save)
		datasetGroup.POST("/create", r.datasetHandler.Create)
		datasetGroup.POST("/rename", r.datasetHandler.Rename)
		datasetGroup.POST("/move", r.datasetHandler.Move)
		datasetGroup.POST("/delete/:id", r.datasetHandler.Delete)
		datasetGroup.POST("/perDelete/:id", r.datasetHandler.PerDelete)
		datasetGroup.POST("/get/:id", r.datasetHandler.GetDetail)
		datasetGroup.POST("/details/:id", r.datasetHandler.Details)
		datasetGroup.POST("/dsDetails", r.datasetHandler.DsDetails)
		datasetGroup.POST("/getSqlParams", r.datasetHandler.GetSQLParams)
		datasetGroup.GET("/barInfo/:id", r.datasetHandler.BarInfo)
		datasetGroup.POST("/getDatasetTotal", r.datasetHandler.GetDatasetTotal)
		datasetGroup.POST("/previewSql", r.datasetHandler.PreviewSQL)
		datasetGroup.POST("/enumValueObj", r.datasetHandler.EnumValueObj)
		datasetGroup.POST("/enumValueDs", r.datasetHandler.EnumValueDs)
		datasetGroup.POST("/enumValue", r.datasetHandler.EnumValue)
	}
}

func (r *Router) registerChartRoutes(api *gin.RouterGroup) {
	chartGroup := api.Group("/chart")
	{
		chartGroup.POST("/query", r.chartHandler.Query)
		if r.permMiddleware != nil {
			chartGroup.POST("/data", r.permMiddleware.CheckChartDataView(), middleware.RowPermissionMiddleware(), r.chartHandler.Data)
			chartGroup.POST("/getData", r.permMiddleware.CheckChartDataView(), middleware.RowPermissionMiddleware(), r.chartHandler.Data)
		} else {
			chartGroup.POST("/data", r.chartHandler.Data)
			chartGroup.POST("/getData", r.chartHandler.Data)
		}
		chartGroup.POST("/getChart/:id", r.chartHandler.GetChart)
		chartGroup.POST("/getDetail/:id", r.chartHandler.GetDetail)
		chartGroup.GET("/checkSameDataSet/:viewIdSource/:viewIdTarget", r.chartHandler.CheckSameDataSet)
		chartGroup.POST("/save", r.chartHandler.SaveFromMap)
		if r.permMiddleware != nil {
			chartGroup.POST("/listByDQ/:id/:chartId", r.permMiddleware.CheckDatasetView(), r.chartHandler.ListByDQ)
		} else {
			chartGroup.POST("/listByDQ/:id/:chartId", r.chartHandler.ListByDQ)
		}
		chartGroup.POST("/copyField/:id/:chartId", r.chartHandler.CopyField)
		chartGroup.POST("/deleteField/:id", r.chartHandler.DeleteField)
		chartGroup.POST("/deleteFieldByChart/:chartId", r.chartHandler.DeleteFieldByChart)
	}
	chartDataGroup := api.Group("/chartData")
	{
		if r.permMiddleware != nil {
			chartDataGroup.POST("/getData", r.permMiddleware.CheckChartDataView(), middleware.RowPermissionMiddleware(), r.chartHandler.Data)
		} else {
			chartDataGroup.POST("/getData", r.chartHandler.Data)
		}
		chartDataGroup.POST("/getFieldData/:fieldId/:fieldType", r.chartHandler.GetFieldData)
		chartDataGroup.POST("/getDrillFieldData/:fieldId", r.chartHandler.GetDrillFieldData)
		chartDataGroup.POST("/innerExportDetails", r.chartHandler.InnerExportDetails)
		chartDataGroup.POST("/innerExportDataSetDetails", r.chartHandler.InnerExportDataSetDetails)
	}
	datasetFieldGroup := api.Group("/datasetField")
	{
		datasetFieldGroup.POST("/listByDatasetGroup/:datasetId", r.datasetHandler.ListByDatasetGroup)
		datasetFieldGroup.GET("/listWithPermissions/:datasetId", r.datasetHandler.ListWithPermissions)
		datasetFieldGroup.POST("/save", r.datasetHandler.SaveField)
		datasetFieldGroup.POST("/getFunction", r.datasetHandler.GetFieldFunctions)
		datasetFieldGroup.POST("/multFieldValuesForPermissions", r.datasetHandler.MultFieldValuesForPermissions)
		datasetFieldGroup.POST("/copilotFields/:id", r.datasetHandler.CopilotFields)
		datasetFieldGroup.POST("/listByDsIds", r.datasetHandler.ListFieldsByDsIds)
	}
}

func (r *Router) registerVisualizationRoutes(api *gin.RouterGroup) {
	visualGroup := api.Group("/dataVisualization")
	{
		visualGroup.POST("/tree", r.visualHandler.Tree)
		if r.permMiddleware != nil {
			visualGroup.POST("/findById", r.permMiddleware.CheckVisualizationView(), r.visualHandler.FindByID)
			visualGroup.POST("/updateCanvas", r.permMiddleware.CheckVisualizationEdit(), r.visualHandler.UpdateCanvas)
			visualGroup.POST("/deleteLogic/:id", r.permMiddleware.CheckVisualizationEdit(), r.visualHandler.DeleteLogic)
			visualGroup.POST("/saveCanvas", r.permMiddleware.CheckVisualizationParentEdit(), r.visualHandler.SaveCanvas)
		} else {
			visualGroup.POST("/findById", r.visualHandler.FindByID)
			visualGroup.POST("/updateCanvas", r.visualHandler.UpdateCanvas)
			visualGroup.POST("/deleteLogic/:id", r.visualHandler.DeleteLogic)
			visualGroup.POST("/saveCanvas", r.visualHandler.SaveCanvas)
		}
		visualGroup.POST("/list", r.visualHandler.List)
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func Start(application *app.Application, db *gorm.DB) error {
	router := NewRouter(application, db)
	router.RegisterRoutes()

	routes := collectRoutesFromEngine(router.engine)
	conflicts := detectRouteConflicts(routes)
	for _, conflict := range conflicts {
		logger.Warn("Route registration conflict", zap.String("warning", conflict))
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting HTTP server", zap.String("port", port))

	return router.Engine().Run(":" + port)
}
