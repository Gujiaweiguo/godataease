package http

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"dataease/backend/internal/app"
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
	engine                *gin.Engine
	app                   *app.Application
	db                    *gorm.DB
	auditHandler          *handler.AuditHandler
	userHandler           *handler.UserHandler
	orgHandler            *handler.OrgHandler
	permHandler           *handler.PermHandler
	embeddedHandler       *handler.EmbeddedHandler
	roleHandler           *handler.RoleHandler
	menuHandler           *handler.MenuHandler
	mapHandler            *handler.MapHandler
	authHandler           *handler.AuthHandler
	datasourceHandler     *handler.DatasourceHandler
	datasetHandler        *handler.DatasetHandler
	chartHandler          *handler.ChartHandler
	visualHandler         *handler.VisualizationHandler
	systemParamHandler    *handler.SystemParamHandler
	licenseHandler        *handler.LicenseHandler
	msgCenterHandler      *handler.MsgCenterHandler
	shareHandler          *handler.ShareHandler
	ticketHandler         *handler.TicketHandler
	geoHandler            *handler.GeoHandler
	staticHandler         *handler.StaticHandler
	exportHandler         *handler.ExportHandler
	engineHandler         *handler.EngineHandler
	driverHandler         *handler.DriverHandler
	templateHandler       *handler.TemplateHandler
	frontendCompatHandler *handler.FrontendCompatHandler
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
	userHandler := handler.NewUserHandler(userService)

	// Organization module initialization
	orgRepo := repository.NewOrgRepository(db)
	orgService := service.NewOrgService(orgRepo)
	orgHandler := handler.NewOrgHandler(orgService)

	// Permission module initialization
	permRepo := repository.NewPermRepository(db)
	permService := service.NewPermService(permRepo)
	permHandler := handler.NewPermHandler(permService)

	// Embedded module initialization
	embeddedRepo := repository.NewEmbeddedRepository(db)
	embeddedService := service.NewEmbeddedService(embeddedRepo)
	embeddedHandler := handler.NewEmbeddedHandler(embeddedService)

	// Role module initialization
	roleRepo := repository.NewRoleRepository(db)
	roleService := service.NewRoleService(roleRepo)
	roleHandler := handler.NewRoleHandler(roleService)

	// Menu module initialization
	menuRepo := repository.NewMenuRepository(db)
	menuService := service.NewMenuService(menuRepo)
	menuHandler := handler.NewMenuHandler(menuService)

	// Map module initialization
	areaRepo := repository.NewAreaRepository(db)
	mapService := service.NewMapService(areaRepo)
	mapHandler := handler.NewMapHandler(mapService)

	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)

	datasourceRepo := repository.NewDatasourceRepository(db)
	datasourceService := service.NewDatasourceService(datasourceRepo)
	datasourceHandler := handler.NewDatasourceHandler(datasourceService)

	datasetRepo := repository.NewDatasetRepository(db)
	datasetService := service.NewDatasetService(datasetRepo)
	datasetHandler := handler.NewDatasetHandler(datasetService)

	chartRepo := repository.NewChartRepository(db)
	chartService := service.NewChartService(chartRepo)
	chartHandler := handler.NewChartHandler(chartService)

	visualRepo := repository.NewVisualizationRepository(db)
	visualService := service.NewVisualizationService(visualRepo)
	visualHandler := handler.NewVisualizationHandler(visualService)

	systemParamRepo := repository.NewSystemParamRepository(db)
	systemParamService := service.NewSystemParamService(systemParamRepo, auditService)
	systemParamHandler := handler.NewSystemParamHandler(systemParamService)

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

	// Export module initialization
	exportRepo := repository.NewExportRepository(db)
	exportService := service.NewExportService(exportRepo)
	exportHandler := handler.NewExportHandler(exportService)

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

	frontendCompatHandler := handler.NewFrontendCompatHandler()

	return &Router{
		engine:                engine,
		app:                   application,
		db:                    db,
		auditHandler:          auditHandler,
		userHandler:           userHandler,
		orgHandler:            orgHandler,
		permHandler:           permHandler,
		embeddedHandler:       embeddedHandler,
		roleHandler:           roleHandler,
		menuHandler:           menuHandler,
		mapHandler:            mapHandler,
		authHandler:           authHandler,
		datasourceHandler:     datasourceHandler,
		datasetHandler:        datasetHandler,
		chartHandler:          chartHandler,
		visualHandler:         visualHandler,
		systemParamHandler:    systemParamHandler,
		licenseHandler:        licenseHandler,
		msgCenterHandler:      msgCenterHandler,
		shareHandler:          shareHandler,
		ticketHandler:         ticketHandler,
		geoHandler:            geoHandler,
		staticHandler:         staticHandler,
		exportHandler:         exportHandler,
		engineHandler:         engineHandler,
		driverHandler:         driverHandler,
		templateHandler:       templateHandler,
		frontendCompatHandler: frontendCompatHandler,
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
	handler.RegisterCompatibilityBridgeRoutes(r.engine, r.userHandler, r.orgHandler, r.datasourceHandler, r.datasetHandler, r.chartHandler)
	handler.RegisterFrontendCompatRoutes(r.engine, r.frontendCompatHandler)

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
		handler.RegisterMenuRoutes(api, r.menuHandler)
		handler.RegisterMapRoutes(api, r.mapHandler)
		handler.RegisterDatasourceRoutes(api, r.datasourceHandler)
		handler.RegisterDatasetRoutes(api, r.datasetHandler)
		handler.RegisterChartRoutes(api, r.chartHandler)
		handler.RegisterVisualizationRoutes(api, r.visualHandler)
		handler.RegisterSystemParamRoutes(api, r.systemParamHandler)
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
