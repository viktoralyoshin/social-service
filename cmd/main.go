package main

import (
	"social-service/internal/app"
	"social-service/internal/config"

	"github.com/viktoralyoshin/utils/pkg/logger"
)

func main() {
	cfg := config.Load()
	logger.Setup(cfg.Env)

	app.Start(cfg)
}
