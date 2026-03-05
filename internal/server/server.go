package server

import (
	"context"
	"net/http"
	"time"

	"github.com/F3dosik/Hofermart/internal/config"
	"go.uber.org/zap"
)

type Server struct {
	config  *config.Config
	handler http.Handler
	logger  *zap.SugaredLogger
}

func New(cfg *config.Config, handler http.Handler, logger *zap.SugaredLogger) *Server {
	return &Server{
		config:  cfg,
		handler: handler,
		logger:  logger,
	}
}

func (s *Server) Run(ctx context.Context) {
	s.logger.Infow("Launching the service with config:",
		"serviceAddr", s.config.ServiceAddress,
		"accrualAddr", s.config.AccrualAddress,
		"databaseURI", s.config.DatabaseURI,
		"logLevel", s.config.LogLevel,
	)

	srv := &http.Server{
		Addr:    s.config.ServiceAddress,
		Handler: s.handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatalw("server failed", "error", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.logger.Errorw("graceful shutdown failed", "error", err)
	}
}
