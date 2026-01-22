package main

import (
	"chatter/internal/app"
	"chatter/internal/config"
	"chatter/pkg/logger"
)

func main() {
	logger.InitLogger()

	cfg := config.Load()
	app.Run(cfg, logger.Logger)
}
