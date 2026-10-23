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
	for _, q := range []string{
		`INSERT INTO user (id, name, current) VALUES ('u1', 'Ada', 1)`,
		`INSERT INTO workspace (id, name) VALUES ('ws1', 'WS')`,
		`INSERT INTO workspace_to_user (id, workspace_id, user_id, current) VALUES ('wtu1', 'ws1', 'u1', 1)`,
	} {
		_, err := db.Exec(q)
		require.NoError(t, err)
	}

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
