package migrate_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/nocmok/go-ledger/internal/config"
	"github.com/nocmok/go-ledger/internal/migrate"
)

// repoRoot resolves the repository root from this file's location, since
// migrate.Migrate reads migrations from the relative path "migrations".
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine caller for repo root lookup")
	}
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func TestMigrate_AppliesAllMigrations(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("ledger"),
		postgres.WithUsername("ledger"),
		postgres.WithPassword("ledger"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("error starting postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("failed to terminate postgres container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("error getting container host: %v", err)
	}
	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("error getting mapped port: %v", err)
	}

	dbConfig := config.DBConfig{
		Host:     host,
		Port:     int(mappedPort.Num()),
		Name:     "ledger",
		User:     "ledger",
		Password: "ledger",
	}

	t.Chdir(repoRoot(t))

	if err := migrate.Migrate(dbConfig); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("error building connection string: %v", err)
	}

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		t.Fatalf("error connecting to database: %v", err)
	}
	defer conn.Close(ctx)

	var version int
	var dirty bool
	if err := conn.QueryRow(ctx, "select version, dirty from schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatalf("error reading schema_migrations: %v", err)
	}
	if dirty {
		t.Fatalf("migration left database in a dirty state at version %d", version)
	}
	if version <= 0 {
		t.Fatalf("expected a positive migration version, got %d", version)
	}
}
