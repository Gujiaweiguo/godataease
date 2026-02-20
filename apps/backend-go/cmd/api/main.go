package main

import (
	"os"
	"os/signal"
	"syscall"

	"dataease/backend/internal/app"
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
	defer database.Close()

	go func() {
		if err := httptransport.Start(application, db); err != nil {
			logger.Fatal("Failed to start HTTP server", zap.String("error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	logger.Sync()
}
