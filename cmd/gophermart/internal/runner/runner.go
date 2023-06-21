package runner

import (
	"context"
	"fmt"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/loyalty-system"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (r *Runner) Run(ctx context.Context) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		r.log.Infof("Stopping server")
		if err := r.server.Shutdown(ctx); err != nil {
			r.log.Errorf("Error stopping server: %s", err)
		}
	}()

	go r.actualizeOrdersInfo(ctx)

	r.log.Infof("Starting server on addr: %s", r.server.Addr)
	if err := r.server.ListenAndServe(); err != nil {
		return fmt.Errorf("error while running server: %w", err)
	}
	return nil
}

func (r *Runner) actualizeOrdersInfo(ctx context.Context) {
	r.log.Infof("Starting actualize orders info")
	ticker := time.NewTicker(time.Second)
	errorsCounter := 0
	for {
		select {
		case <-ctx.Done():
			r.log.Infof("Stopping actualize orders info: context done")
			return
		case <-ticker.C:
			if err := r.loyaltyPointsSystem.UpdateOrdersInfo(); err != nil {
				r.log.Errorf("error while request to loyalty system: %s", err.Error())
				errorsCounter++
				if errorsCounter > 10 {
					r.log.Infof("Stopping actualize orders because of many errors")
					return
				}
			}
		}
	}
}

func New(server *http.Server, loyaltyPointsSystem *loyalty.LoyaltySystem, log *zap.SugaredLogger) *Runner {
	return &Runner{
		server:              server,
		log:                 log,
		loyaltyPointsSystem: loyaltyPointsSystem,
	}
}

type Runner struct {
	log                 *zap.SugaredLogger
	server              *http.Server
	loyaltyPointsSystem *loyalty.LoyaltySystem
}
