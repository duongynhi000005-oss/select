package syncer

import (
	"database/sql/driver"
	"testing"

	"selectDb/internal/db/generated"
	"selectDb/internal/edition"
	"selectDb/internal/graph"

	"github.com/stretchr/testify/require"
)

// A tracked write in the local edition reaches the graph and nothing else: a
// queued commit would wait forever for a server.
func TestLocalEditionQueuesNoCommit(t *testing.T) {
	edition.Local = true
	t.Cleanup(func() { edition.Local = false })

	db := newTestDB(t)
	seedCurrentUser(t, db, "u1", "Ada")
	seedWorkspace(t, db, "ws1", "WS")
	seedWorkspaceToUser(t, db, "wtu1", "ws1", "u1", true)

	queries := generated.New(db)
	s := &Syncer{Queries: queries, Graph: graph.New(queries)}
	require.NoError(t, s.CommitMutation(&QueueMutationParams{
		Query:     "INSERT INTO role (id, workspace_id, name) VALUES (:id, :workspace_id, :name)",
		TableName: "role",
		Operation: "insert",
		Args: []driver.NamedValue{
			{Name: "id", Value: "r1"},
			{Name: "workspace_id", Value: "ws1"},
			{Name: "name", Value: "Reader"},
		},
	}))

	var queued int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM mutation_commit`).Scan(&queued))
	require.Zero(t, queued)
}
