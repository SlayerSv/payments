package migrations

import (
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5"
)

// RunMigrations — универсальная функция для всех сервисов
func RunMigrations(dbURL string, sqlFiles embed.FS, dirName string) {
	// dirName — это название папки, где лежат .sql (например, "migrations")
	d, err := iofs.New(sqlFiles, dirName)
	if err != nil {
		log.Fatalf("migrations: failed to create iofs: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		log.Fatalf("migrations: failed to create migrator: %v", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrations: failed to apply: %v", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("Database is up to date")
	} else {
		fmt.Println("Migrations applied successfully!")
	}
}
