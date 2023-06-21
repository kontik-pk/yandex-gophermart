package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/database"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/handlers"
	"go.uber.org/zap"
)

func New(dbManager *database.Manager, log *zap.SugaredLogger) *chi.Mux {
	handler := handlers.New(dbManager, log)
	r := chi.NewRouter()
	r.Post("/api/user/register", handler.Register)
	r.Post("/api/user/login", handler.Login)
	r.Post("/api/user/orders", handler.BasicAuth(handler.LoadOrder))
	r.Post("/api/user/balance/withdraw", handler.BasicAuth(handler.Withdraw))
	r.Get("/api/user/orders", handler.BasicAuth(handler.GetOrders))
	r.Get("/api/user/withdrawals", handler.BasicAuth(handler.GetWithdrawals))
	r.Get("/api/user/balance", handler.BasicAuth(handler.GetBalance))
	return r
}
