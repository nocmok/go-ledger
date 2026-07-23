package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nocmok/go-ledger/internal/config"
)

func Migrate(path string, config config.DBConfig) (err error) {
	dbUrl := fmt.Sprintf("pgx5://%s:%s@%s:%d/%s", config.User, config.Password, config.Host, config.Port, config.Name)
	srcUrl := fmt.Sprintf("file://%s", path)
	m, err := migrate.New(srcUrl, dbUrl)
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
