"""What the reference collector reports for the clauses that name a relation
somewhere other than a FROM."""
import server

SCHEMA = {
    "schemas": [
        {
            "name": "main",
            "tables": [
                {"name": "t1", "columns": [{"name": "c1", "type": "INTEGER"}]},
                {"name": "t2", "columns": [{"name": "c1", "type": "INTEGER"}]},
            ],
        },
    ],
}


def _relations(sql: str, sg_dialect: str = "postgresql") -> list[str]:
    response = server._dispatch({
        "action": "collect_references",
        "sql": sql,
        "dialect": sg_dialect,
        "schema": SCHEMA,
        "default_schema": "main",
    })
    return [relation["table"] for relation in response["relations"]]


class TestDeleteUsing:
    def test_it_reads_the_relation(self):
        assert _relations("DELETE FROM t1 USING t2 WHERE t1.c1 = t2.c1") == ["t1", "t2"]

    def test_it_reads_several(self):
        assert _relations("DELETE FROM t1 USING t2, (SELECT 1) s WHERE 1 = 1") == [
            "t1", "t2", "s"]

    def test_every_dialect(self):
        for sg_dialect in ("postgresql", "mysql", "sqlite"):
            assert _relations("DELETE FROM t1 USING t2 WHERE 1 = 1", sg_dialect) == [
                "t1", "t2"]


class TestUpdateFrom:
    def test_it_reads_the_relation(self):
        assert _relations("UPDATE t1 SET c1 = 1 FROM t2 WHERE t1.c1 = t2.c1") == [
            "t1", "t2"]


class TestCTECaseInsensitivity:
    def test_unquoted_cte_case_insensitive_postgres(self):
        response = server._dispatch({
            "action": "collect_references",
            "sql": "WITH A AS (SELECT c1 FROM t1) SELECT a.c1 FROM a",
            "dialect": "postgresql",
            "schema": {"main": {"t1": {"c1": "INT"}}},
            "default_schema": "main",
        })
        assert [r["table"] for r in response["relations"]] == ["t1", "A"]
        assert [r["schema"] for r in response["relations"]] == ["main", ""]
        assert [v["table"] for v in response["virtual_tables"]] == ["A"]

    def test_unquoted_cte_column_resolution(self):
        response = server._dispatch({
            "action": "collect_column_refs",
            "sql": "WITH A AS (SELECT c1 FROM t1) SELECT a.c1 FROM a",
            "dialect": "postgresql",
            "schema": {"main": {"t1": {"c1": "INT"}}},
            "default_schema": "main",
        })
        c1_refs = [c for c in response["column_refs"] if c["column"] == "c1"]
        assert len(c1_refs) == 2
        assert all(c["resolved"] for c in c1_refs)

    def test_recursive_cte_columns_resolved(self):
        response = server._dispatch({
            "action": "collect_column_refs",
            "sql": "WITH RECURSIVE R(n) AS (SELECT 1 UNION ALL SELECT r.n + 1 FROM r WHERE r.n < 3) SELECT R.n FROM r",
            "dialect": "postgresql",
            "schema": {"main": {"t1": {"c1": "INT"}}},
            "default_schema": "main",
        })
        n_refs = [c for c in response["column_refs"] if c["column"] == "n"]
        assert len(n_refs) == 3
        assert all(c["resolved"] for c in n_refs)

    def test_sqlite_type_mismatch_across_differently_cased_cte(self):
        response = server._dispatch({
            "action": "lint",
            "sql": "WITH Totals AS (SELECT c1 FROM t1) SELECT totals.c1 FROM totals WHERE totals.c1 = 'five'",
            "dialect": "sqlite",
            "schema": {"main": {"t1": {"c1": "INT"}}},
            "default_schema": "main",
        })
        mismatches = [d for d in response["diagnostics"] if d["rule_id"] == "type-mismatch-comparison"]
        assert len(mismatches) == 1
