package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nocmok/go-ledger/internal/config"
	"github.com/nocmok/go-ledger/internal/migrate"
)

func main() {
	config, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("error loading config: %w", err))
	}

	if err := migrate.Migrate(config.DBConfig); err != nil {
		panic(fmt.Errorf("failed to run migration: %w", err))
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", config.DBConfig.User, config.DBConfig.Password, config.DBConfig.Host, config.DBConfig.Port, config.DBConfig.Name)
	pgxpoolConfig, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		panic(fmt.Errorf("error parsing pgxpool config: %w", err))
	}
	pgxpoolConfig.MaxConns = int32(config.DBConfig.MaxConn)
	pgxpoolConfig.MaxConnIdleTime = time.Minute
	pgxpool, err := pgxpool.NewWithConfig(ctx, pgxpoolConfig)
	if err != nil {
		panic(fmt.Errorf("error creating connection pool: %w", err))
	}
	defer pgxpool.Close()

	<-ctx.Done()
}
