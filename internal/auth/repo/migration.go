package repo

import (
	"embed"

	"github.com/SlayerSv/payments/pkg/migrations"
)

//go:embed migrations
var migrationsFS embed.FS

func StartMigrations(dbURL string) {
	migrations.RunMigrations(dbURL, migrationsFS, "migrations")
}
