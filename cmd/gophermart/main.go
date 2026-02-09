package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/F3dosik/Hofermart/internal/config"
	"github.com/F3dosik/Hofermart/internal/logger"
	"github.com/F3dosik/Hofermart/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration loading error: %v", err)
	}

	mode := logger.Mode(cfg.LogLevel)
	baseLogger, sugarLogger := logger.NewLogger(mode)
	defer func() { _ = baseLogger.Sync() }()
	server := server.NewServer(cfg, sugarLogger)
	server.Run(ctx)
}
