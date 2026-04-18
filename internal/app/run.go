package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	appcfg "s3_task/internal/config"
	"s3_task/internal/httpserver"
	"s3_task/internal/logger"
	"s3_task/internal/s3client"
)

func Run(ctx context.Context) error {
	cfg, err := appcfg.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	s3c, err := s3client.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("s3 client: %w", err)
	}

	srv := httpserver.New(cfg, s3c)
	mux := http.NewServeMux()
	srv.Register(mux, "web")

	httpSrv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("http listening", zap.String("addr", cfg.HTTPAddr))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown", zap.Error(err))
	}
	return nil
}
