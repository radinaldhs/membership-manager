package main

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/app"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

func main() {
	logger := logging.CreateDefaultLogger(logging.LevelInfo.String())
	cfg, err := config.Load()
	if err != nil {
		logger.Critical(fmt.Sprintf("Load config error: %s", err.Error()))
		os.Exit(1)
	}

	a, err := app.New(cfg)
	if err != nil {
		logger.Critical(fmt.Sprintf("App creation error: %s", err.Error()))
		os.Exit(1)
	}

	go func() {
		a.Start(context.Background())
	}()

	stopAppOnOsSignal(a)
}
