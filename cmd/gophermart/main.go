package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/database"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/flags"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/logger"
	loyalty_system "github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/loyalty-system"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/router"
	runner2 "github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/runner"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()
	log, err := logger.New("debug")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	params := flags.Init(
		flags.WithAddr(),
		flags.WithDatabase(),
		flags.WithAccrual(),
	)

	db, err := sql.Open("pgx", params.DatabaseAddress)
	if err != nil {
		log.Sugar().Errorf("error while init db: %s", err.Error())
		os.Exit(1)
	}
	dbManager, err := database.New(ctx, db)
	if err != nil {
		log.Sugar().Errorf("error while init db: %s", err.Error())
		os.Exit(1)
	}

	server := &http.Server{Addr: params.ServerRunAddr, Handler: router.New(dbManager, log.Sugar())}
	loyaltyPointsSystem := loyalty_system.New(params.AccrualSystemAddress, dbManager, log.Sugar())

	runner := runner2.New(server, loyaltyPointsSystem, log.Sugar())
	if err = runner.Run(ctx); err != nil {
		log.Sugar().Errorf("error while running runner: %s", err.Error())
		return
	}
}
