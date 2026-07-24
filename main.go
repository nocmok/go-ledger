package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/nocmok/go-ledger/internal/config"
	"github.com/nocmok/go-ledger/internal/ledger"
	"github.com/nocmok/go-ledger/internal/migrate"
	"github.com/nocmok/go-ledger/internal/model"
)

func errorHandlingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ec *echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				ec.JSON(http.StatusInternalServerError, model.Error{
					Type:    model.ErrorTypeInternalError,
					Message: "internal error",
				})
			}
		}()
		err := next(ec)
		if err == nil {
			return nil
		}
		if err, ok := errors.AsType[*model.Error](err); ok {
			return ec.JSON(http.StatusBadRequest, err)
		}
		return ec.JSON(http.StatusInternalServerError, model.Error{
			Type:    model.ErrorTypeInternalError,
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

	ledgerRepository := ledger.NewRepository(pgxpool)
	ledgerService := ledger.NewService(ledgerRepository)

	e := echo.New()

	e.Use(errorHandlingMiddleware)

	type statusResponse struct {
		Status string `json:"status"`
	}

	e.GET("/ready", func(ec *echo.Context) error {
		return ec.JSON(http.StatusOK, statusResponse{Status: "ok"})
	})

	e.GET("/live", func(ec *echo.Context) error {
		return ec.JSON(http.StatusOK, statusResponse{Status: "ok"})
	})

	e.POST("/ledgers", func(ec *echo.Context) error {
		type Headers struct {
			IdempotencyKey *uuid.UUID `header:"Idempotency-Key"`
		}
		headers := Headers{}
		if err := echo.BindHeaders(ec, &headers); err != nil {
			return err
		}
		if headers.IdempotencyKey == nil {
			return &model.Error{
				Type:    model.ErrorTypeInvalidRequest,
				Message: "idempotency key required",
			}
		}
		type CreateLedgerRequest struct {
			Name     *string          `json:"name"`
			Metadata *json.RawMessage `json:"metadata"`
		}
		body := CreateLedgerRequest{}
		if err := ec.Bind(&body); err != nil {
			return err
		}
		errorDetails := []model.ErrorDetail{}
		if body.Name == nil {
			errorDetails = append(errorDetails, model.ErrorDetail{Field: "name", Message: "name required"})
		}
		if body.Metadata == nil {
			errorDetails = append(errorDetails, model.ErrorDetail{Field: "metadata", Message: "metadata required"})
		}
		if len(errorDetails) > 0 {
			return &model.Error{
				Type:    model.ErrorTypeInvalidRequest,
				Message: "invalid request",
				Details: errorDetails,
			}
		}
		ledger, err := ledgerService.Create(
			ec.Request().Context(),
			*headers.IdempotencyKey,
			*body.Name,
			*body.Metadata,
		)
		if err != nil {
			return err
		}
		ec.JSON(http.StatusCreated, ledger)
		return nil
	})

	eConfig := echo.StartConfig{
		Address:         fmt.Sprintf(":%d", config.ServerConfig.Port),
		GracefulTimeout: 5 * time.Second,
		HideBanner:      true,
	}
	if err := eConfig.Start(ctx, e); err != nil {
		panic(fmt.Errorf("error starting server: %w", err))
	}
}
