package db

import (
	"context"
	"fmt"
	"log/slog"

	"squirrel/ent"
	"squirrel/ent/migrate"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"           // postgres driver
	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

// NewClient creates a new ent.Client for the configured driver.
//
// Supported drivers:
//   - "sqlite3": uses path to open a file-backed SQLite database.
//   - "postgres": uses dsn to open a Postgres connection.
//
// Auto-migration runs for both drivers. The returned *entsql.Driver is the
// same connection used by the client; keep it around for readiness pings
// (ent.Client exposes no driver accessor of its own).
func NewClient(driver, path, dsn string) (*ent.Client, *entsql.Driver, error) {
	switch driver {
	case "postgres":
		return newPostgresClient(dsn)
	case "sqlite3", "":
		return NewSQLiteClient(path)
	default:
		return nil, nil, fmt.Errorf("unsupported database driver: %q", driver)
	}
}

// NewSQLiteClient creates a new ent.Client for SQLite.
func NewSQLiteClient(path string) (*ent.Client, *entsql.Driver, error) {
	slog.Info("opening sqlite connection", "path", path)
	drv, err := entsql.Open(dialect.SQLite, fmt.Sprintf("file:%s?cache=shared&_fk=1&_journal_mode=WAL&_busy_timeout=5000", path))
	if err != nil {
		slog.Error("failed to open sqlite connection", "path", path, "error", err)
		return nil, nil, fmt.Errorf("failed opening connection to sqlite: %w", err)
	}
	client, err := migrateAndReturn(ent.NewClient(ent.Driver(drv)))
	if err != nil {
		return nil, nil, err
	}
	return client, drv, nil
}

// newPostgresClient creates a new ent.Client for Postgres.
func newPostgresClient(dsn string) (*ent.Client, *entsql.Driver, error) {
	if dsn == "" {
		return nil, nil, fmt.Errorf("postgres driver requires a non-empty DSN")
	}

	slog.Info("opening postgres connection")
	drv, err := entsql.Open(dialect.Postgres, dsn)
	if err != nil {
		slog.Error("failed to open postgres connection", "error", err)
		return nil, nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}
	client, err := migrateAndReturn(ent.NewClient(ent.Driver(drv)))
	if err != nil {
		return nil, nil, err
	}
	return client, drv, nil
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

// Ping verifies the database connection is alive, for use by readiness
// checks. Works uniformly across sqlite3/postgres since it's a plain
// connection ping, not a query.
func Ping(ctx context.Context, drv *entsql.Driver) error {
	return drv.DB().PingContext(ctx)
}
