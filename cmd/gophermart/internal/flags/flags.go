package flags

import (
	"flag"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/models"
	"os"
)

const (
	defaultAddr string = "localhost:8080"
)

func WithDatabase() models.Option {
	return func(p *models.Params) {
		flag.StringVar(&p.DatabaseAddress, "d", "postgres://practicum:yandex@localhost:5432/postgres?sslmode=disable", "connection string for db")
		if envDBAddr := os.Getenv("DATABASE_URI"); envDBAddr != "" {
			p.DatabaseAddress = envDBAddr
		}
	}
}

func WithAddr() models.Option {
	return func(p *models.Params) {
		flag.StringVar(&p.ServerRunAddr, "a", defaultAddr, "address and port to run server")
		if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
			p.ServerRunAddr = envRunAddr
		}
	}
}

func WithAccrual() models.Option {
	return func(p *models.Params) {
		flag.StringVar(&p.AccrualSystemAddress, "r", "", "address and port to run server")
		if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
			p.AccrualSystemAddress = envAccrualAddr
		}
	}
}

func Init(opts ...models.Option) *models.Params {
	p := &models.Params{}
	for _, opt := range opts {
		opt(p)
	}
	flag.Parse()
	return p
}
