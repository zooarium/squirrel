package db

import (
	"context"
	"fmt"
	"log/slog"

	"vyaya/ent"
	"vyaya/ent/migrate"

	_ "github.com/mattn/go-sqlite3"
)

// NewSQLiteClient creates a new ent.Client for SQLite.
func NewSQLiteClient(path string) (*ent.Client, error) {
	slog.Info("opening sqlite connection", "path", path)
	client, err := ent.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&_fk=1", path))
	if err != nil {
		slog.Error("failed to open sqlite connection", "path", path, "error", err)
		return nil, fmt.Errorf("failed opening connection to sqlite: %v", err)
	}

	// Run the auto migration tool if you want to keep it simple,
	// OR use the versioned migrations.
	// For versioned migrations, we typically use the migrate package.
	slog.Info("running auto migration")
	if err := client.Schema.Create(context.Background(), migrate.WithGlobalUniqueID(true)); err != nil {
		slog.Error("failed to create schema resources", "error", err)
		if cerr := client.Close(); cerr != nil {
			slog.Error("failed to close client after schema creation failure", "error", cerr)
		}
		return nil, fmt.Errorf("failed creating schema resources: %v", err)
	}

	slog.Info("database initialization completed successfully")
	return client, nil
}
