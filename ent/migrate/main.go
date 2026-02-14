//go:build ignore

package main

import (
	"context"
	"log"
	"os"

	"vyaya/ent/migrate"

	atlasmigrate "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx := context.Background()
	// Create a local migration directory able to understand Atlas migration file format for replay.
	dir, err := atlasmigrate.NewLocalDir("ent/migrate/migrations")
	if err != nil {
		log.Fatalf("failed creating atlas migration directory: %v", err)
	}
	// Migrate diff options.
	opts := []schema.MigrateOption{
		schema.WithDir(dir),                         // provide migration directory
		schema.WithMigrationMode(schema.ModeReplay), // provide migration mode
		schema.WithDialect(dialect.SQLite),          // Ent dialect to use
		schema.WithFormatter(atlasmigrate.DefaultFormatter),
	}
	if len(os.Args) != 2 {
		log.Fatalln("migration name must be provided as argument")
	}
	// Generate "diff" between schema and current database state.
	err = migrate.NamedDiff(ctx, "sqlite3://?mode=memory&cache=shared&_fk=1", os.Args[1], opts...)
	if err != nil {
		log.Fatalf("failed generating migration: %v", err)
	}
}
