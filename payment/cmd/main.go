package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/horizoonn/factory-platform/payment/internal/app"
	"github.com/horizoonn/factory-platform/payment/internal/config"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

func main() {
	cfg := config.NewConfigMust()

	if err := logger.InitWithConfig(cfg.Logger()); err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}

	exitCode := 0
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Warn(context.Background(), "failed to sync logger", zap.Error(err))
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	a := app.New(cfg)

	if err := a.Run(ctx); err != nil {
		logger.Error(ctx, "run payment app failed", zap.Error(err))
		exitCode = 1
		return
	}
}
