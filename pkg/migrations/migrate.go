package migrations

import (
	"embed"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // DB driver
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql" // migrate option
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/*sql
var migrationsFS embed.FS

const (
	dbURLPrefix = "mysql://"
)

func MigrateDB(dbURL string) error {
	migrationSource, err := iofs.New(migrationsFS, "sql")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}
	defer migrationSource.Close()

	migration, err := migrate.NewWithSourceInstance("iofs", migrationSource, dbURLPrefix+dbURL)
	if err != nil {
		return err
	}
	defer migration.Close()

	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// Migrate looks at the currently active migration version,
// then migrates either up or down to the specified version.
// Useful for testing migrations
func Migrate(dbURL string, version uint) error {
	migrationSource, err := iofs.New(migrationsFS, "sql")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}
	defer migrationSource.Close()

	migration, err := migrate.NewWithSourceInstance("iofs", migrationSource, dbURLPrefix+dbURL)
	if err != nil {
		return err
	}
	defer migration.Close()

	err = migration.Migrate(version)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
