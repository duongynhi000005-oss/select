# Query Execution

Run a query with `Cmd+Enter`. SELECT executes the SQL against one or more associated databases and displays results in a table below the editor.

## Limits

Two **workspace** settings (configured in **Settings → Workspace**, and kept on
your machine rather than in the folder) control execution bounds:

| Setting                  | Default | Max   | Description                            |
|--------------------------|---------|-------|----------------------------------------|
| **statement_timeout_ms** | 30000   | none  | Max query execution time (milliseconds)|
| **max_result_size_mb**   | 100     | 250   | Max result size before truncation (MB) |

When a timeout fires, SELECT cancels the query on the database side. If the result exceeds the size limit, it is truncated and a notice is shown.

Set `statement_timeout_ms` to `0` to disable the timeout entirely. This is useful for long-running migrations or bulk operations.

### Long-running statements

For a migration, a bulk insert or an index rebuild, keep `statement_timeout_ms`
high enough to cover the expected duration, or set it to `0`. There is no
progress indicator for DDL statements that return no rows.

## Result caching

Results are cached on your machine. Switching tabs and coming back shows the cached result instantly without re-executing. Running the query again replaces the cache.
