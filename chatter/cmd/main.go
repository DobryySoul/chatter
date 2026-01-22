package main

import (
	"chatter/internal/app"
	"chatter/internal/config"
	"chatter/pkg/logger"
	"chatter/pkg/telemetry"
	"context"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	logger.InitLogger()

	shutdown, err := telemetry.Init(ctx, logger.Logger)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize telemetry", zap.Error(err))
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			logger.Logger.Error("Failed to shutdown telemetry", zap.Error(err))
		}
	}()

	cfg := config.Load()
	app.Run(ctx, cfg, logger.Logger)
}
