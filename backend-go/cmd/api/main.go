package main

import (
	"os"
	"os/signal"
	"syscall"

	"dataease/backend/internal/app"
	"dataease/backend/internal/pkg/database"
	"dataease/backend/internal/pkg/logger"
	httptransport "dataease/backend/internal/transport/http"
)

func main() {
	application, err := app.Init()
	if err != nil {
		logger.Fatal("Failed to initialize application", logger.L().String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Application initialized",
		logger.L().String("name", application.Name),
		logger.L().String("version", application.Version),
	)

	db, err := database.Init(&application.Config.Database)
	if err != nil {
		logger.Fatal("Failed to connect database", logger.L().String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Close()

	go func() {
		if err := httptransport.Start(application, db); err != nil {
			logger.Fatal("Failed to start HTTP server", logger.L().String("error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	logger.Sync()
}
