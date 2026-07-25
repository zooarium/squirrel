package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"squirrel/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
)

func TestHookLogsCreateAndUpdate(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:audit_hook?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	var buf strings.Builder
	client.Use(Hook(slog.New(slog.NewJSONHandler(&buf, nil))))

	ctx := context.Background()
	cat, err := client.Category.Create().SetAppID(1).SetUserID(1).SetName("groceries").Save(ctx)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := client.Category.UpdateOne(cat).SetName("groceries-2").Save(ctx); err != nil {
		t.Fatalf("update: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 log lines, got %d: %q", len(lines), buf.String())
	}

	var created map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &created); err != nil {
		t.Fatalf("unmarshal create line: %v", err)
	}
	if created["action"] != "OpCreate" || created["entity_type"] != "Category" {
		t.Fatalf("unexpected create log line: %v", created)
	}
	if int(created["entity_id"].(float64)) != cat.ID {
		t.Fatalf("want entity_id %d, got %v", cat.ID, created["entity_id"])
	}
	if int(created["app_id"].(float64)) != 1 {
		t.Fatalf("want app_id 1, got %v", created["app_id"])
	}

	var updated map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &updated); err != nil {
		t.Fatalf("unmarshal update line: %v", err)
	}
	if updated["action"] != "OpUpdateOne" || updated["entity_type"] != "Category" {
		t.Fatalf("unexpected update log line: %v", updated)
	}
}
