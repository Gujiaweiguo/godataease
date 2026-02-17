package main

import (
	"dataease/backend/internal/app"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/transport/http"
	"os"
	"os/signal"
	"syscall"
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

	go func() {
		if err := http.Start(application); err != nil {
			logger.Fatal("Failed to start HTTP server", logger.L().String("error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	logger.Sync()
}
