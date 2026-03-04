package db

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gsupert/internal/config"
	embeddedmigrations "gsupert/internal/migrations"
)

func RunMigrations(cfg *config.Config) error {
	sourceDriver, err := iofs.New(embeddedmigrations.Files, ".")
	if err != nil {
		return fmt.Errorf("initialize migration source: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", sourceDriver, buildPostgresURL(cfg))
	if err != nil {
		return fmt.Errorf("initialize migration runner: %w", err)
	}
	defer closeMigrator(migrator)

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("Database migrations are up to date")
			return nil
		}
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Printf("Database migrations applied successfully")
	return nil
}

func buildPostgresURL(cfg *config.Config) string {
	query := url.Values{}
	query.Set("sslmode", cfg.DBSSLMode)

	dbPath := cfg.DBName
	if !strings.HasPrefix(dbPath, "/") {
		dbPath = "/" + dbPath
	}

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:     fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort),
		Path:     dbPath,
		RawQuery: query.Encode(),
	}).String()
}

func closeMigrator(migrator *migrate.Migrate) {
	sourceErr, dbErr := migrator.Close()
	if sourceErr != nil {
		log.Printf("Migration source close warning: %v", sourceErr)
	}
	if dbErr != nil {
		log.Printf("Migration database close warning: %v", dbErr)
	}
}
