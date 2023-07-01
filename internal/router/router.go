package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/kontik-pk/go-musthave-diploma-tpl/internal/database"
	"github.com/kontik-pk/go-musthave-diploma-tpl/internal/handlers"
	"go.uber.org/zap"
)

func New(dbManager *database.Manager, log *zap.SugaredLogger) *chi.Mux {
	handler := handlers.New(dbManager, log)
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", handler.Register)
		r.Post("/api/user/login", handler.Login)
	})
	r.Group(func(r chi.Router) {
		r.Use(handler.BasicAuth)
		r.Post("/api/user/orders", handler.LoadOrder)
		r.Post("/api/user/balance/withdraw", handler.Withdraw)
		r.Get("/api/user/orders", handler.GetOrders)
		r.Get("/api/user/withdrawals", handler.GetWithdrawals)
		r.Get("/api/user/balance", handler.GetBalance)
	})

	return r
}
