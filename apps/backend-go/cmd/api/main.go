package main

import (
	"os"
	"os/signal"
	"syscall"

	"dataease/backend/internal/app"
	scheduler "dataease/backend/internal/job"
	"dataease/backend/internal/job/jobs"
	"dataease/backend/internal/pkg/cache"
	"dataease/backend/internal/pkg/database"
	"dataease/backend/internal/pkg/logger"
	httptransport "dataease/backend/internal/transport/http"

	"go.uber.org/zap"
)

func main() {
	application, err := app.Init()
	if err != nil {
		logger.Fatal("Failed to initialize application", zap.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Application initialized",
		zap.String("name", application.Name),
		zap.String("version", application.Version),
	)

	db, err := database.Init(&application.Config.Database)
	if err != nil {
		logger.Fatal("Failed to connect database", zap.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() { _ = database.Close() }()

	redisClient, err := cache.Init(&application.Config.Redis)
	if err != nil {
		logger.Fatal("Failed to connect redis", zap.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() { _ = cache.Close() }()

	if err := database.AutoMigrate(db); err != nil {
		logger.Fatal("Failed to migrate database", zap.String("error", err.Error()))
		os.Exit(1)
	}

	if err := database.SeedDefaults(db); err != nil {
		logger.Fatal("Failed to seed database", zap.String("error", err.Error()))
		os.Exit(1)
	}

	database.CleanupStaleMenuData(db)

	if err := database.SeedDemoData(db); err != nil {
		logger.Fatal("Failed to seed demo data", zap.String("error", err.Error()))
		os.Exit(1)
	}

	jobRegistry, err := jobs.NewRegistry(application.Config.Scheduler)
	if err != nil {
		logger.Fatal("Failed to build scheduled job registry", zap.String("error", err.Error()))
		os.Exit(1)
	}

	jobScheduler := scheduler.NewScheduler()
	jobScheduler.SetRedis(redisClient)
	if err := jobScheduler.AddRegistry(jobRegistry, nil); err != nil {
		logger.Fatal("Failed to register scheduled jobs", zap.String("error", err.Error()))
		os.Exit(1)
	}
	jobScheduler.Start()
	defer jobScheduler.Stop()

	logger.Info("Scheduled job foundation initialized",
		zap.Int("registered_jobs", len(jobScheduler.Entries())),
		zap.Bool("sample_job_enabled", application.Config.Scheduler.SampleJobEnabled),
	)

	go func() {
		if err := httptransport.Start(application, db); err != nil {
			logger.Fatal("Failed to start HTTP server", zap.String("error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	jobScheduler.Stop()
	_ = logger.Sync()
}
