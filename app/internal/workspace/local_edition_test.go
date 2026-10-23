package workspace

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"selectDb/internal/db"
	"selectDb/internal/db/generated"
	"selectDb/internal/edition"
	"selectDb/internal/graph"
	"selectDb/internal/server"
	"selectDb/internal/system"

	"modernc.org/sqlite"
)

// newLocalWorkspace opens the local edition's database under a scratch home,
// with its one user signed in, as CheckForLogin leaves it.
func newLocalWorkspace(t *testing.T) (*Workspace, string) {
	t.Helper()

	edition.Local = true
	t.Cleanup(func() { edition.Local = false })
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	for _, name := range []string{"sqlite3", "sqlite3-mw"} {
		if !slices.Contains(sql.Drivers(), name) {
			sql.Register(name, &sqlite.Driver{})
		}
	}
	if err := db.Init(server.LocalDomain); err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.GetDB().Close()
		_ = db.Init("")
	})
	t.Cleanup(graph.ClearOpenWorkspace)

	q := generated.New(db.NewLiveDB())
	system.New(q, nil, nil, nil).CheckForLogin()
	u, err := q.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("no local user after CheckForLogin: %v", err)
	}
	w := New(q, nil)
	w.PullFunc = func(context.Context, string) error {
		t.Error("the local edition pulled from a server")
		return nil
	}
	return w, u.ID
}

// A folder with content of its own, so opening it seeds no sample.
func newFolder(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "weekly.sql"), []byte("select 1;"), 0o600); err != nil {
		t.Fatalf("write query: %v", err)
	}
	return dir
}

func TestLocalEditionCreatesWorkspaceWithoutServer(t *testing.T) {
	w, _ := newLocalWorkspace(t)
	dir := newFolder(t, "reports")

	state, err := w.CreateWorkspaceInFolder(dir, "Reports")
	if err != nil {
		t.Fatalf("CreateWorkspaceInFolder: %v", err)
	}
	if state.Status != Ready {
		t.Fatalf("status = %q, want %q", state.Status, Ready)
	}

	cfg, err := graph.ReadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if cfg.Server != server.LocalDomain {
		t.Errorf("config server = %q, want %q", cfg.Server, server.LocalDomain)
	}
	ws, err := w.Queries.GetWorkspaceByID(context.Background(), cfg.WorkspaceID)
	if err != nil {
		t.Fatalf("workspace row: %v", err)
	}
	if ws.Name != "Reports" {
		t.Errorf("name = %q, want %q", ws.Name, "Reports")
	}
}

// A folder a teammate made against a server opens as the local user's own, and
// its config is left naming that server: it is in git, and so is theirs.
func TestLocalEditionAdoptsFolderFromServer(t *testing.T) {
	w, userID := newLocalWorkspace(t)
	dir := newFolder(t, "analytics")
	if err := graph.WriteWorkspaceConfig(dir, "api.us-east.select-db.com", "ws-remote"); err != nil {
		t.Fatalf("write config: %v", err)
	}

	for range 2 {
		state, err := w.OpenFolder(dir)
		if err != nil {
			t.Fatalf("OpenFolder: %v", err)
		}
		if state.Status != Ready {
			t.Fatalf("status = %q, want %q", state.Status, Ready)
		}
	}

	member, err := w.hasMemberRow(context.Background(), userID, "ws-remote")
	if err != nil || !member {
		t.Fatalf("membership = %v, %v; want a row", member, err)
	}
	cfg, err := graph.ReadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if cfg.Server != "api.us-east.select-db.com" {
		t.Errorf("config server rewritten to %q", cfg.Server)
	}
}
