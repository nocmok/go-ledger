package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nocmok/go-ledger/internal/config"
)

func Migrate(config config.DBConfig) (err error) {
	url := fmt.Sprintf("pgx5://%s:%s@%s:%d/%s", config.User, config.Password, config.Host, config.Port, config.Name)
	m, err := migrate.New("file://migrations", url)
	if err != nil {
		return err
	}
	defer func() {
		srcErr, dbErr := m.Close()
		err = errors.Join(err, srcErr, dbErr)
	}()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
