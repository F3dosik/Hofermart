package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/F3dosik/Hofermart/internal/client"
	"github.com/F3dosik/Hofermart/internal/config"
	"github.com/F3dosik/Hofermart/internal/db"
	"github.com/F3dosik/Hofermart/internal/handler"
	"github.com/F3dosik/Hofermart/internal/logger"
	"github.com/F3dosik/Hofermart/internal/repository"
	"github.com/F3dosik/Hofermart/internal/server"
	"github.com/F3dosik/Hofermart/internal/service"
	"github.com/F3dosik/Hofermart/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("configuration loading error: %v", err)
	}

	mode := logger.Mode(cfg.LogLevel)
	baseLogger, sugarLogger := logger.NewLogger(mode)
	defer func() { _ = baseLogger.Sync() }()

	if err := db.RunMigrations(cfg.DatabaseURI, sugarLogger); err != nil {
		sugarLogger.Fatalw("migrations failed: ", "error", err)
	}

	pool, err := db.NewPool(ctx, cfg.DatabaseURI)
	if err != nil {
		sugarLogger.Fatalw("database connection failed", "error", err)
	}
	defer pool.Close()

	repo := repository.New(pool)

	jobChan := make(chan *worker.ScheduleJob, 2*cfg.WorkerCount)

	userService := service.NewUserService(repo, cfg.JWTSecret)
	orderService := service.NewOrderService(repo, jobChan)
	balanceService := service.NewBalanceService(repo)
	h := handler.New(userService, orderService, balanceService, cfg.JWTSecret, sugarLogger)

	client := client.NewAccrual(cfg.AccrualAddress)
	schedule := worker.NewScheduler(jobChan)
	worker := worker.New(repo, client, schedule, jobChan, cfg.WorkerCount, cfg.PollInterval, cfg.MaxDelay, sugarLogger)

	go schedule.Run(ctx)
	if err := worker.LoadPendingOrders(ctx); err != nil {
		sugarLogger.Fatalw("failed to restore pending orders", "error", err)
	}
	go worker.Run(ctx)

	srv := server.New(cfg, h, sugarLogger)
	srv.Run(ctx)
}
