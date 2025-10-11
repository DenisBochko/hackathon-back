package main

import (
	"context"
	"fmt"
	"hackathon-back/internal/docs"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"hackathon-back/internal/app"
	"hackathon-back/internal/config"
	_ "hackathon-back/internal/docs" // DO NOT REMOVE MFK
	"hackathon-back/pkg/logger"
)

// @title Hackathon API
// @version 1.0
// @description API для Hackathon
// @host localhost:8080
// @BasePath /api/
func main() {
	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoadConfig()
	config.MustPrintConfig(cfg)

	docs.SwaggerInfo.Title = cfg.App.ServiceName
	docs.SwaggerInfo.Version = cfg.App.Version
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port)
	docs.SwaggerInfo.BasePath = cfg.HTTPServer.BasePath

	loggerCfg := &logger.Config{
		Level:      cfg.Level,
		FormatJSON: cfg.FormatJSON,
		Rotation: logger.Rotation{
			File:       cfg.Rotation.File,
			MaxSize:    cfg.Rotation.MaxSize,
			MaxBackups: cfg.Rotation.MaxBackups,
			MaxAge:     cfg.Rotation.MaxAge,
		},
	}

	log := logger.MustSetupLogger(loggerCfg)

	errors := make(chan error)

	application := app.MustNew(cfg, log)
	defer func() {
		close(errors)

		if err := application.Shutdown(); err != nil {
			log.Error("Failed to shutdown application", zap.Error(err))
		}

		if err := log.Sync(); err != nil {
			log.Warn("Failed to sync logger", zap.Error(err))
		}

		log.Info("Application has shutdown")
	}()

	go func() { errors <- application.Run() }()

	select {
	case err := <-errors:
		if err != nil {
			log.Error("Server error, shutting down...", zap.Error(err))
		}
	case <-ctx.Done():
		log.Info("Received stop signal, shutting down...")
	}
}
