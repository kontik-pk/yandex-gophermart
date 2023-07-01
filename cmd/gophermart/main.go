package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kontik-pk/go-musthave-diploma-tpl/internal/database"
	"github.com/kontik-pk/go-musthave-diploma-tpl/internal/flags"
	"github.com/kontik-pk/go-musthave-diploma-tpl/internal/logger"
	loyalty_system "github.com/kontik-pk/go-musthave-diploma-tpl/internal/loyalty-system"
	"github.com/kontik-pk/go-musthave-diploma-tpl/internal/router"
	runner2 "github.com/kontik-pk/go-musthave-diploma-tpl/internal/runner"
	server "github.com/kontik-pk/go-musthave-diploma-tpl/internal/server"
	"os"
)

const logLevel = "info"

func main() {
	ctx := context.Background()
	log, err := logger.New(logLevel)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	params := flags.Init(
		flags.WithAddr(),
		flags.WithDatabase(),
		flags.WithAccrual(),
	)

	db, err := sql.Open("pgx", params.Database.ConnectionString)
	if err != nil {
		log.Sugar().Errorf("error while init db: %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Sugar().Errorf("error while closing db: %s", err.Error())
			os.Exit(1)
		}
	}()
	dbManager, err := database.New(ctx, db)
	if err != nil {
		log.Sugar().Errorf("error while init db: %s", err.Error())
		os.Exit(1)
	}

	appServer := server.New(params.Server.Address, router.New(dbManager, log.Sugar()))
	loyaltyPointsSystem := loyalty_system.New(params.AccrualSystem.Address, dbManager, log.Sugar())

	runner := runner2.New(appServer, loyaltyPointsSystem, log.Sugar())
	if err = runner.Run(ctx); err != nil {
		log.Sugar().Errorf("error while running runner: %s", err.Error())
		return
	}
}
