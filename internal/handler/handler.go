package handler

import (
	"net/http"

	"github.com/F3dosik/Hofermart/internal/middleware"
	"github.com/F3dosik/Hofermart/internal/middleware/gzip"
	"github.com/F3dosik/Hofermart/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	router         chi.Router
	userService    service.UserService
	orderService   service.OrderService
	balanceService service.BalanceService
	secretKey      string
	logger         *zap.SugaredLogger
}

func New(userService service.UserService, orderService service.OrderService,
	balanceService service.BalanceService, secretKey string, logger *zap.SugaredLogger) *Handler {
	h := &Handler{
		router:         chi.NewRouter(),
		userService:    userService,
		orderService:   orderService,
		balanceService: balanceService,
		logger:         logger,
		secretKey:      secretKey,
	}
	h.setupMiddleware()
	h.setupRoutes()
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) setupMiddleware() {
	h.router.Use(gzip.WithCompression(h.logger))
	h.router.Use(middleware.WithLogging(h.logger))
}

func (h *Handler) setupRoutes() {
	h.router.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireJSON(h.logger))
			r.Post("/register", h.register)
			r.Post("/login", h.login)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(h.logger, h.secretKey))
			r.With(middleware.RequirePlainText(h.logger)).
				Post("/orders", h.uploadOrder)
			r.Get("/orders", h.getOrders)
			r.Get("/balance", h.getBalance)
			r.With(middleware.RequireJSON(h.logger)).
				Post("/balance/withdraw", h.withdraw)
			r.Get("/withdrawals", h.getWithdrawals)
		})
	})
}
