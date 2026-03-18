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
	engine                  *gin.Engine
	app                     *app.Application
	db                      *gorm.DB
	permMiddleware          *middleware.PermissionMiddleware
	auditHandler            *handler.AuditHandler
	userHandler             *handler.UserHandler
	orgHandler              *handler.OrgHandler
	permHandler             *handler.PermHandler
	embeddedHandler         *handler.EmbeddedHandler
	roleHandler             *handler.RoleHandler
	roleMenuHandler         *handler.RoleMenuHandler
	menuHandler             *handler.MenuHandler
	mapHandler              *handler.MapHandler
	authHandler             *handler.AuthHandler
	datasourceHandler       *handler.DatasourceHandler
	datasetHandler          *handler.DatasetHandler
	chartHandler            *handler.ChartHandler
	visualHandler           *handler.VisualizationHandler
	watermarkHandler        *handler.WatermarkHandler
	systemParamHandler      *handler.SystemParamHandler
	systemVariableHandler   *handler.SystemVariableHandler
	licenseHandler          *handler.LicenseHandler
	msgCenterHandler        *handler.MsgCenterHandler
	shareHandler            *handler.ShareHandler
	ticketHandler           *handler.TicketHandler
	geoHandler              *handler.GeoHandler
	staticHandler           *handler.StaticHandler
	exportHandler           *handler.ExportHandler
	engineHandler           *handler.EngineHandler
	driverHandler           *handler.DriverHandler
	templateHandler         *handler.TemplateHandler
	syncHandler             *handler.SyncHandler
	frontendCompatHandler   *handler.FrontendCompatHandler
	permissionCompatHandler *handler.PermissionCompatHandler
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
	userService := service.NewUserService(userRepo, userRoleRepo, userPermRepo)
	userImportService := service.NewUserImportService(userService)
	userHandler := handler.NewUserHandler(userService, userImportService)
	// Role module initialization (must be before OrgService as it depends on roleRepo)
	roleRepo := repository.NewRoleRepository(db)
	roleService := service.NewRoleService(roleRepo, userRepo, userRoleRepo)
	roleHandler := handler.NewRoleHandler(roleService)

	// Organization module initialization
	orgRepo := repository.NewOrgRepository(db)
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

	authService := service.NewAuthService(userRepo, jwtInstance)
	authHandler := handler.NewAuthHandler(authService)

	datasourceRepo := repository.NewDatasourceRepository(db)
	syncRepo := repository.NewSyncRepository(db)
	datasourceService := service.NewDatasourceService(datasourceRepo)
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
	datasetService := service.NewDatasetServiceWithPermission(datasetRepo, rowPermService)
	if application != nil && application.Config != nil {
		calciteCfg := application.Config.Integration.Calcite
		datasetService.SetCalciteConfig(
			calciteCfg.Address,
			time.Duration(calciteCfg.TimeoutSec)*time.Second,
			calciteCfg.MaxRetries,
		)
	}
	datasetHandler := handler.NewDatasetHandler(datasetService)

	chartRepo := repository.NewChartRepository(db)
	chartService := service.NewChartService(chartRepo)
	chartHandler := handler.NewChartHandler(chartService)

	visualRepo := repository.NewVisualizationRepository(db)
	visualService := service.NewVisualizationService(visualRepo)
	visualHandler := handler.NewVisualizationHandler(visualService)
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

	// Static module initialization
	staticRepo := repository.NewStaticRepository(db)
	storeRepo := repository.NewStoreRepository(db)
	typefaceRepo := repository.NewTypefaceRepository(db)
	staticService := service.NewStaticService(staticRepo, storeRepo, typefaceRepo)
	staticHandler := handler.NewStaticHandler(staticService)

	// Permission middleware initialization
	resourcePermRepo := repository.NewResourcePermissionRepository(db)
	resourcePermService := service.NewResourcePermissionService(resourcePermRepo, adminChecker)
	datasetService.SetResourcePermissionService(resourcePermService)
	exportPermService := service.NewExportPermissionService(resourcePermService, nil)
	permMiddleware := middleware.NewPermissionMiddleware(resourcePermService, exportPermService, adminChecker)
	permissionCompatHandler := handler.NewPermissionCompatHandler(menuService, permService, roleMenuService, resourcePermService)

	// Export module initialization
	exportRepo := repository.NewExportRepository(db)
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
		engine:                  engine,
		app:                     application,
		db:                      db,
		permMiddleware:          permMiddleware,
		auditHandler:            auditHandler,
		userHandler:             userHandler,
		orgHandler:              orgHandler,
		permHandler:             permHandler,
		embeddedHandler:         embeddedHandler,
		roleHandler:             roleHandler,
		roleMenuHandler:         roleMenuHandler,
		menuHandler:             menuHandler,
		mapHandler:              mapHandler,
		authHandler:             authHandler,
		datasourceHandler:       datasourceHandler,
		datasetHandler:          datasetHandler,
		chartHandler:            chartHandler,
		visualHandler:           visualHandler,
		watermarkHandler:        watermarkHandler,
		systemParamHandler:      systemParamHandler,
		systemVariableHandler:   systemVariableHandler,
		licenseHandler:          licenseHandler,
		msgCenterHandler:        msgCenterHandler,
		shareHandler:            shareHandler,
		ticketHandler:           ticketHandler,
		geoHandler:              geoHandler,
		staticHandler:           staticHandler,
		exportHandler:           exportHandler,
		engineHandler:           engineHandler,
		driverHandler:           driverHandler,
		templateHandler:         templateHandler,
		syncHandler:             syncHandler,
		frontendCompatHandler:   frontendCompatHandler,
		permissionCompatHandler: permissionCompatHandler,
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
	handler.RegisterSystemParamRoutes(r.engine, r.systemParamHandler)
	handler.RegisterLicenseRoutes(r.engine, r.licenseHandler)
	handler.RegisterMsgCenterRoutes(r.engine, r.msgCenterHandler)
	handler.RegisterTicketRoutes(r.engine, r.ticketHandler)
	handler.RegisterVisualizationRoutes(r.engine.Group(""), r.visualHandler)
	handler.RegisterCompatibilityBridgeRoutes(r.engine, r.userHandler, r.orgHandler, r.datasourceHandler, r.datasetHandler, r.chartHandler)
	handler.RegisterFrontendCompatRoutes(r.engine, protected, r.frontendCompatHandler)
}

func (r *Router) registerAPIRoutes() {
	api := r.engine.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		handler.RegisterAuditRoutes(api, r.auditHandler)
		handler.RegisterUserRoutes(api, r.userHandler)
		handler.RegisterOrgRoutes(api, r.orgHandler)
		handler.RegisterPermRoutes(api, r.permHandler)
		handler.RegisterEmbeddedRoutes(api, r.embeddedHandler)
		handler.RegisterRoleRoutes(api, r.roleHandler)
		handler.RegisterRoleMenuRoutes(api, r.roleMenuHandler)
		handler.RegisterMenuRoutes(api, r.menuHandler)
		handler.RegisterPermissionCompatRoutes(api, r.permissionCompatHandler)
		handler.RegisterMapRoutes(api, r.mapHandler)
		handler.RegisterDatasourceRoutes(api, r.datasourceHandler)
		handler.RegisterSyncRoutes(api, r.syncHandler)
		r.registerDatasetRoutes(api)
		handler.RegisterChartRoutes(api, r.chartHandler)
		r.registerVisualizationRoutes(api)
		handler.RegisterWatermarkRoutes(api, r.watermarkHandler)
		handler.RegisterSystemParamRoutes(api, r.systemParamHandler)
		handler.RegisterSystemVariableRoutes(api, r.systemVariableHandler)
		handler.RegisterLicenseRoutes(api, r.licenseHandler)
		handler.RegisterMsgCenterRoutes(api, r.msgCenterHandler)
		handler.RegisterShareRoutes(api, r.shareHandler)
		handler.RegisterTicketRoutes(api, r.ticketHandler)
		handler.RegisterGeoRoutes(api, r.geoHandler)
		handler.RegisterStaticRoutes(api, r.staticHandler)
		handler.RegisterExportRoutes(api, r.exportHandler)
		handler.RegisterEngineRoutes(api, r.engineHandler)
		handler.RegisterDriverRoutes(api, r.driverHandler)
		handler.RegisterTemplateRoutes(api, r.templateHandler)
		handler.RegisterCompatibilityBridgeRoutes(api, r.userHandler, r.orgHandler, r.datasourceHandler, r.datasetHandler, r.chartHandler)
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

func (r *Router) registerDatasetRoutes(api *gin.RouterGroup) {
	datasetGroup := api.Group("/dataset")
	{
		datasetGroup.POST("/tree", r.datasetHandler.Tree)
		datasetGroup.POST("/fields", r.datasetHandler.Fields)
		datasetGroup.POST("/preview", r.datasetHandler.Preview)
		if r.permMiddleware != nil {
			datasetGroup.POST("/previewWithPerm", r.permMiddleware.CheckDatasetView(), r.datasetHandler.PreviewWithPermission)
		} else {
			datasetGroup.POST("/previewWithPerm", r.datasetHandler.PreviewWithPermission)
		}
	}
}

func (r *Router) registerVisualizationRoutes(api *gin.RouterGroup) {
	visualGroup := api.Group("/dataVisualization")
	{
		visualGroup.POST("/tree", r.visualHandler.Tree)
		if r.permMiddleware != nil {
			visualGroup.POST("/findById", r.permMiddleware.CheckDashboardView(), r.visualHandler.FindByID)
			visualGroup.POST("/updateCanvas", r.permMiddleware.CheckDashboardEdit(), r.visualHandler.UpdateCanvas)
			visualGroup.POST("/deleteLogic/:id", r.permMiddleware.CheckDashboardEdit(), r.visualHandler.DeleteLogic)
		} else {
			visualGroup.POST("/findById", r.visualHandler.FindByID)
			visualGroup.POST("/updateCanvas", r.visualHandler.UpdateCanvas)
			visualGroup.POST("/deleteLogic/:id", r.visualHandler.DeleteLogic)
		}
		visualGroup.POST("/list", r.visualHandler.List)
		visualGroup.POST("/saveCanvas", r.visualHandler.SaveCanvas)
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
