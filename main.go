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
	"github.com/nocmok/go-ledger/internal/account"
	"github.com/nocmok/go-ledger/internal/config"
	"github.com/nocmok/go-ledger/internal/currency"
	"github.com/nocmok/go-ledger/internal/db"
	"github.com/nocmok/go-ledger/internal/ledger"
	"github.com/nocmok/go-ledger/internal/migrate"
	"github.com/nocmok/go-ledger/internal/model"
	"github.com/nocmok/go-ledger/internal/transaction"
)

func errorHandlingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ec *echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				ec.JSON(http.StatusInternalServerError, model.Error{Message: "internal error"})
			}
		}()
		if err := next(ec); err != nil {
			return ec.JSON(http.StatusInternalServerError, model.Error{Message: "internal error"})
		}
		return nil
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

	txManager := db.NewTxManager(pgxpool)
	ledgerRepository := ledger.NewRepository(pgxpool)
	ledgerService := ledger.NewService(ledgerRepository)
	accountRepository := account.NewRepository(pgxpool)
	accountService := account.NewService(accountRepository, ledgerService)
	transactionRepository := transaction.NewRepository()
	transactionService := transaction.NewService(txManager, transactionRepository)

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
		var headers struct {
			IdempotencyKey *uuid.UUID `header:"Idempotency-Key"`
		}
		if err := echo.BindHeaders(ec, &headers); err != nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "malformed idempotency key"})
		}
		if headers.IdempotencyKey == nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "idempotency key required"})
		}
		var body struct {
			Name     *string          `json:"name"`
			Metadata *json.RawMessage `json:"metadata"`
		}
		if err := ec.Bind(&body); err != nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "malformed body"})
		}
		details := []model.ErrorDetail{}
		if body.Name == nil {
			details = append(details, model.ErrorDetail{Field: "name", Message: "name required"})
		}
		if body.Metadata == nil {
			details = append(details, model.ErrorDetail{Field: "metadata", Message: "metadata required"})
		}
		if len(details) > 0 {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "invalid request", Details: details})
		}
		l, err := ledgerService.Create(
			ec.Request().Context(),
			*headers.IdempotencyKey,
			*body.Name,
			*body.Metadata,
		)
		if err != nil {
			return err
		}
		ec.JSON(http.StatusCreated, l)
		return nil
	})

	e.GET("/ledgers/:ledgerId", func(ec *echo.Context) error {
		var params struct {
			LedgerId uuid.UUID `param:"ledgerId"`
		}
		if err := ec.Bind(&params); err != nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "malformed ledger id"})
		}
		l, err := ledgerService.Get(ec.Request().Context(), params.LedgerId)
		if err != nil {
			if errors.Is(err, ledger.ErrNotFound) {
				return ec.JSON(http.StatusNotFound, model.Error{Message: "not found"})
			}
			return err
		}
		return ec.JSON(http.StatusOK, l)
	})

	e.POST("/ledgers/:ledgerId/accounts", func(ec *echo.Context) error {
		var headers struct {
			IdempotencyKey *uuid.UUID `header:"Idempotency-Key"`
		}
		if err := echo.BindHeaders(ec, &headers); err != nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "malformed idempotency key"})
		}
		if headers.IdempotencyKey == nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "idempotency key required"})
		}
		var body struct {
			LedgerId         *uuid.UUID         `param:"ledgerId"`
			Name             *string            `json:"name"`
			Currency         *currency.Currency `json:"currency"`
			OverdraftAllowed *bool              `json:"overdraftAllowed"`
			Metadata         *json.RawMessage   `json:"metadata"`
		}
		if err := ec.Bind(&body); err != nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "malformed params or body"})
		}
		details := []model.ErrorDetail{}
		if body.LedgerId == nil {
			details = append(details, model.ErrorDetail{Field: "ledgerId", Message: "ledgerId required"})
		}
		if body.Name == nil {
			details = append(details, model.ErrorDetail{Field: "name", Message: "name required"})
		}
		if body.Currency == nil {
			details = append(details, model.ErrorDetail{Field: "currency", Message: "currency reqiured"})
		}
		if body.OverdraftAllowed == nil {
			details = append(details, model.ErrorDetail{Field: "overdraftAllowed", Message: "overdraftAllowed reqiured"})
		}
		if body.Metadata == nil {
			details = append(details, model.ErrorDetail{Field: "metadata", Message: "metadata required"})
		}
		if len(details) > 0 {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "invalid request", Details: details})
		}
		a, err := accountService.Create(
			ec.Request().Context(),
			*headers.IdempotencyKey,
			*body.LedgerId,
			*body.Name,
			*body.Currency,
			*body.OverdraftAllowed,
			*body.Metadata,
		)
		if err != nil {
			if errors.Is(err, account.ErrLedgerNotFound) {
				return ec.JSON(http.StatusConflict, model.Error{Message: "invalid ledger id"})
			}
			return err
		}
		return ec.JSON(http.StatusCreated, a)
	})

	e.GET("/ledgers/:ledgerId/accounts/:accountId", func(ec *echo.Context) error {
		var params struct {
			LedgerId  uuid.UUID `param:"ledgerId"`
			AccountId uuid.UUID `param:"accountId"`
		}
		if err := ec.Bind(&params); err != nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "malformed params"})
		}
		a, err := accountService.Get(ec.Request().Context(), params.LedgerId, params.AccountId)
		if err != nil {
			if errors.Is(err, account.ErrNotFound) {
				return ec.JSON(http.StatusNotFound, model.Error{Message: "not found"})
			}
			return err
		}
		return ec.JSON(http.StatusOK, a)
	})

	e.POST("/ledgers/:ledgerId/transactions", func(ec *echo.Context) error {
		var headers struct {
			IdempotencyKey *uuid.UUID `header:"Idempotency-Key"`
		}
		if err := echo.BindHeaders(ec, &headers); err != nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "malformed idempotency key"})
		}
		if headers.IdempotencyKey == nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "idempotency key required"})
		}
		var body struct {
			LedgerId *uuid.UUID         `param:"ledgerId"`
			Currency *currency.Currency `json:"currency"`
			Entries  []struct {
				AccountId uuid.UUID `json:"accountId"`
				Amount    int64     `json:"amount"`
			} `json:"entries"`
			Metadata json.RawMessage `json:"metadata"`
		}
		if err := ec.Bind(&body); err != nil {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "malformed params or body"})
		}
		details := []model.ErrorDetail{}
		if body.Currency == nil {
			details = append(details, model.ErrorDetail{Field: "currency", Message: "currency required"})
		}
		if body.Entries == nil {
			details = append(details, model.ErrorDetail{Field: "entries", Message: "entries required"})
		}
		if body.Metadata == nil {
			details = append(details, model.ErrorDetail{Field: "metadata", Message: "metadata required"})
		}
		if len(details) > 0 {
			return ec.JSON(http.StatusBadRequest, model.Error{Message: "invalid request", Details: details})
		}
		entries := make([]transaction.Entry, 0, len(body.Entries))
		for _, entry := range body.Entries {
			entries = append(entries, transaction.Entry{
				AccountId: entry.AccountId,
				Amount:    entry.Amount,
			})
		}
		t, err := transactionService.CreateTransaction(
			ec.Request().Context(),
			*headers.IdempotencyKey,
			*body.LedgerId,
			*body.Currency,
			entries,
			body.Metadata,
		)
		if err != nil {
			return err
		}
		return ec.JSON(http.StatusCreated, t)
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
