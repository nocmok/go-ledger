package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/nocmok/go-ledger/internal/config"
	"github.com/nocmok/go-ledger/internal/migrate"
)

var repo Repository

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("ledger"),
		postgres.WithUsername("ledger"),
		postgres.WithPassword("ledger"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("error starting postgres container: %v", err)
	}
	defer func() {
		if err := container.Terminate(context.Background()); err != nil {
			log.Printf("failed to terminate postgres container: %v", err)
		}
	}()

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("error getting container host: %v", err)
	}
	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		log.Fatalf("error getting mapped port: %v", err)
	}

	dbConfig := config.DBConfig{
		Host:     host,
		Port:     mappedPort.Num(),
		Name:     "ledger",
		User:     "ledger",
		Password: "ledger",
	}

	if err := migrate.Migrate("../../migrations", dbConfig); err != nil {
		log.Fatalf("error applying migrations: %v", err)
	}

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("error building connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("error creating connection pool: %v", err)
	}
	defer pool.Close()

	repo = NewRepository(pool)

	code := m.Run()
	os.Exit(code)
}

func assertJSONEqual(t *testing.T, want, got json.RawMessage) {
	t.Helper()

	var wantVal, gotVal any
	if err := json.Unmarshal(want, &wantVal); err != nil {
		t.Fatalf("error unmarshaling expected metadata: %v", err)
	}
	if err := json.Unmarshal(got, &gotVal); err != nil {
		t.Fatalf("error unmarshaling actual metadata: %v", err)
	}
	if !reflect.DeepEqual(wantVal, gotVal) {
		t.Errorf("expected metadata %s, got %s", want, got)
	}
}

func assertLedgersEqual(t *testing.T, want, got Ledger) {
	t.Helper()

	if got.ID != want.ID {
		t.Errorf("expected ID %s, got %s", want.ID, got.ID)
	}
	if got.Name != want.Name {
		t.Errorf("expected name %q, got %q", want.Name, got.Name)
	}
	assertJSONEqual(t, want.Metadata, got.Metadata)
}

func TestRepository_Create_ReturnsSpecifiedData(t *testing.T) {
	ctx := context.Background()

	idempotencyKey := uuid.New()
	name := "primary ledger"
	metadata := json.RawMessage(`{"owner":"alice"}`)

	got, err := repo.Create(ctx, idempotencyKey, name, metadata)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if got.ID == uuid.Nil {
		t.Errorf("expected a non-nil ID")
	}
	if got.Name != name {
		t.Errorf("expected name %q, got %q", name, got.Name)
	}
	assertJSONEqual(t, metadata, got.Metadata)
}

func TestRepository_Get_FindsCreatedLedger(t *testing.T) {
	ctx := context.Background()

	idempotencyKey := uuid.New()
	name := "receivables"
	metadata := json.RawMessage(`{"currency":"USD"}`)

	created, err := repo.Create(ctx, idempotencyKey, name, metadata)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	assertLedgersEqual(t, created, got)
}

func TestRepository_Get_UnknownID_ReturnsNotFoundError(t *testing.T) {
	ctx := context.Background()

	_, err := repo.Get(ctx, uuid.New())
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("expected error to be pgx.ErrNoRows, got %v", err)
	}
}

func TestRepository_Create_DoubleInsertSameIdempotencyKey_ReturnsSameObject(t *testing.T) {
	ctx := context.Background()

	idempotencyKey := uuid.New()

	first, err := repo.Create(ctx, idempotencyKey, "first name", json.RawMessage(`{"n":1}`))
	if err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}

	second, err := repo.Create(ctx, idempotencyKey, "second name", json.RawMessage(`{"n":2}`))
	if err != nil {
		t.Fatalf("second Create returned error: %v", err)
	}
	assertLedgersEqual(t, first, second)

	third, err := repo.Create(ctx, idempotencyKey, "third name", json.RawMessage(`{"n":3}`))
	if err != nil {
		t.Fatalf("third Create returned error: %v", err)
	}
	assertLedgersEqual(t, first, third)
}
