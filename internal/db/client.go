package db

import (
	"context"
	"fmt"
	"log/slog"

	"squirrel/ent"
	"squirrel/ent/migrate"

	"entgo.io/ent/dialect"
	_ "github.com/lib/pq"           // postgres driver
	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

// NewClient creates a new ent.Client for the configured driver.
//
// Supported drivers:
//   - "sqlite3": uses path to open a file-backed SQLite database.
//   - "postgres": uses dsn to open a Postgres connection.
//
// Auto-migration runs for both drivers.
func NewClient(driver, path, dsn string) (*ent.Client, error) {
	switch driver {
	case "postgres":
		return newPostgresClient(dsn)
	case "sqlite3", "":
		return NewSQLiteClient(path)
	default:
		return nil, fmt.Errorf("unsupported database driver: %q", driver)
	}
}

// NewSQLiteClient creates a new ent.Client for SQLite.
func NewSQLiteClient(path string) (*ent.Client, error) {
	slog.Info("opening sqlite connection", "path", path)
	client, err := ent.Open(dialect.SQLite, fmt.Sprintf("file:%s?cache=shared&_fk=1", path))
	if err != nil {
		slog.Error("failed to open sqlite connection", "path", path, "error", err)
		return nil, fmt.Errorf("failed opening connection to sqlite: %w", err)
	}

	return migrateAndReturn(client)
}

// newPostgresClient creates a new ent.Client for Postgres.
func newPostgresClient(dsn string) (*ent.Client, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres driver requires a non-empty DSN")
	}

	slog.Info("opening postgres connection")
	client, err := ent.Open(dialect.Postgres, dsn)
	if err != nil {
		slog.Error("failed to open postgres connection", "error", err)
		return nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	return migrateAndReturn(client)
}

// migrateAndReturn runs auto-migration and returns the ready client.
func migrateAndReturn(client *ent.Client) (*ent.Client, error) {
	// Run the auto migration tool to keep schema in sync at startup.
	slog.Info("running auto migration")
	if err := client.Schema.Create(context.Background(), migrate.WithGlobalUniqueID(true)); err != nil {
		slog.Error("failed to create schema resources", "error", err)
		if cerr := client.Close(); cerr != nil {
			slog.Error("failed to close client after schema creation failure", "error", cerr)
		}
		return nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	slog.Info("database initialization completed successfully")
	return client, nil
}
