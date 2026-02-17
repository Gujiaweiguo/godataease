package http

import (
	"fmt"
	"net/http"
	"os"
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

type Router struct {
	engine       *gin.Engine
	app          *app.Application
	db           *gorm.DB
	auditHandler *handler.AuditHandler
}

func NewRouter(application *app.Application, db *gorm.DB) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Recovery())
	engine.Use(requestLogger())
	engine.Use(metricsMiddleware())

	auditLogRepo := repository.NewAuditLogRepository(db)
	loginFailureRepo := repository.NewLoginFailureRepository(db)
	auditLogDetailRepo := repository.NewAuditLogDetailRepository(db)
	auditService := service.NewAuditService(auditLogRepo, loginFailureRepo, auditLogDetailRepo)
	auditHandler := handler.NewAuditHandler(auditService)

	middleware.SetAuditService(auditService)

	return &Router{
		engine:       engine,
		app:          application,
		db:           db,
		auditHandler: auditHandler,
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

	api := r.engine.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		handler.RegisterAuditRoutes(api, r.auditHandler)
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func Start(application *app.Application, db *gorm.DB) error {
	router := NewRouter(application, db)
	router.RegisterRoutes()

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting HTTP server", zap.String("port", port))

	return router.Engine().Run(":" + port)
}
