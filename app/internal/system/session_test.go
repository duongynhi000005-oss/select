package system

import (
	"context"
	"database/sql"
	"path/filepath"
	"slices"
	"testing"

	"selectDb/internal/db"
	"selectDb/internal/db/generated"
	"selectDb/internal/edition"

	"modernc.org/sqlite"
)

// Every launch of the local edition signs in, and every one after the first
// finds the user the first one made.
func TestLocalEditionSignsInOneLocalUser(t *testing.T) {
	edition.Local = true
	t.Cleanup(func() { edition.Local = false })

	if !slices.Contains(sql.Drivers(), "sqlite3") {
		sql.Register("sqlite3", &sqlite.Driver{})
	}
	path := filepath.Join(t.TempDir(), "db.sqlite")
	if err := db.RunMigrationsAt(path); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	q := generated.New(conn)
	s := New(q, nil, nil, nil)
	s.CheckForLogin()
	first, err := q.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("no current user after the first launch: %v", err)
	}
	s.CheckForLogin()

	var users int
	if err := conn.QueryRow(`SELECT count(*) FROM user`).Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if users != 1 {
		t.Fatalf("users = %d, want 1", users)
	}
	again, err := q.GetCurrentUser(context.Background())
	if err != nil || again.ID != first.ID {
		t.Fatalf("current user = %v, %v; want %s", again.ID, err, first.ID)
	}
	if !first.Name.Valid || first.Name.String == "" {
		t.Errorf("local user has no name")
	}
}
