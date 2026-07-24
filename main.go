package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/nocmok/go-ledger/internal/config"
	"github.com/nocmok/go-ledger/internal/migrate"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	config, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("error loading config: %w", err))
	}

	if err := migrate.Migrate("migrations", config.DBConfig); err != nil {
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

	e := echo.New()

	eConfig := echo.StartConfig{
		Address: fmt.Sprintf(":%d", config.ServerConfig.Port),
	}
	if err := eConfig.Start(ctx, e); err != nil {
		panic(fmt.Errorf("error starting server: %w", err))
	}

	<-ctx.Done()
}
