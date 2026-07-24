package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/nocmok/go-ledger/internal/config"
	"github.com/nocmok/go-ledger/internal/migrate"
)

func errorHandlingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	type ErrorType string
	const (
		ErrorTypeInternalError  ErrorType = "INTERNAL_ERROR"
		ErrorTypeInvalidRequest ErrorType = "INVALID_REQUEST"
	)
	type ErrorDetail struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}
	type Error struct {
		Type    ErrorType     `json:"type"`
		Message string        `json:"message"`
		Details []ErrorDetail `json:"details"`
	}
	return func(context *echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				context.JSON(http.StatusInternalServerError, Error{
					Type:    ErrorTypeInternalError,
					Message: "internal error",
				})
			}
		}()
		err := next(context)
		if err == nil {
			return nil
		}
		return context.JSON(http.StatusInternalServerError, Error{
			Type:    ErrorTypeInternalError,
			Message: "internal error",
		})
	}
}

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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

	e.Use(errorHandlingMiddleware)

	type statusResponse struct {
		Status string `json:"status"`
	}

	e.GET("/ready", func(context *echo.Context) error {
		return context.JSON(http.StatusOK, statusResponse{Status: "ok"})
	})

	e.GET("/live", func(context *echo.Context) error {
		return context.JSON(http.StatusOK, statusResponse{Status: "ok"})
	})

	eConfig := echo.StartConfig{
		Address:         fmt.Sprintf(":%d", config.ServerConfig.Port),
		GracefulTimeout: 5 * time.Second,
	}
	if err := eConfig.Start(ctx, e); err != nil {
		panic(fmt.Errorf("error starting server: %w", err))
	}
}
