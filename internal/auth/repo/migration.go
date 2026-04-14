package repo

import (
	"embed"

	"github.com/SlayerSv/payments/internal/shared/migrations"
)

//go:embed migrations
var migrationsFS embed.FS

func StartMigrations(dbURL string) {
	migrations.RunMigrations(dbURL, migrationsFS, "migrations")
}
