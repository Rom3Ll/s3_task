package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"s3_task/internal/app"
	"s3_task/internal/logger"
)

func main() {
	logger.Init()
	defer logger.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		logger.Fatal("run", zap.Error(err))
	}
}
