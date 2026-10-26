"""
Schema construction and shared utilities used across lint rule modules.
"""
from __future__ import annotations

from sqlglot import exp
from sqlglot.dialects.dialect import Dialect as SqlglotDialect
from sqlglot.errors import TokenError
from sqlglot.optimizer.scope import Scope, ScopeType, traverse_scope
from sqlglot.schema import MappingSchema

# Map Go dialect names → sqlglot dialect names
DIALECT_MAP = {
    "postgresql": "postgres",
    "postgres":   "postgres",
    "mysql":      "mysql",
    "sqlite":     "sqlite",
}


def sqlglot_dialect_name(go_dialect: str) -> str:
    """Map a Go dialect name onto sqlglot's. An unknown name passes through so
    sqlglot raises, rather than being parsed as some other dialect in silence.
    """
    name = (go_dialect or "").lower()
    return DIALECT_MAP.get(name, name)


# Dialects whose grammar accepts OFFSET only as part of a LIMIT clause, so
# OFFSET on its own does not parse. PostgreSQL takes the two independently.
_OFFSET_NEEDS_LIMIT = {"mysql", "sqlite"}


def offset_requires_limit(sg_dialect: str) -> bool:
    """Whether OFFSET without LIMIT is a syntax error rather than a style choice."""
    return sg_dialect in _OFFSET_NEEDS_LIMIT


def unquoted_identifier_case(sg_dialect: str) -> str | None:
    """
    The case an unquoted identifier is folded to, or None if it is left alone.

    sqlglot decides this to parse at all, so asking it keeps one answer per
    dialect rather than a list here that has to be remembered.
    """
    try:
        strategy = SqlglotDialect.get_or_raise(sg_dialect or "postgres").NORMALIZATION_STRATEGY
    except Exception:
        return None
    name = getattr(strategy, "name", "")
    if name == "LOWERCASE":
        return "lower"
    if name in ("UPPERCASE", "CASE_INSENSITIVE_UPPERCASE"):
        return "upper"
    return None


def tokenize(sql: str, sg_dialect: str) -> list:
    """Tokenize with the dialect's own tokenizer. The generic one knows only
    double quotes, so MySQL backticks and SQLite brackets arrive as UNKNOWN
    tokens and disappear from every scan that looks for an identifier.
    """
    if not sg_dialect:
        raise ValueError("tokenize requires a dialect name")

    # Dialect.tokenize binds the tokenizer to the dialect. Building the class
    # directly leaves it on the generic settings, which differ in more than
    # quoting: MySQL allows an identifier to start with a digit, so 2fa reads
    # as the number 2 followed by fa.
    dialect = SqlglotDialect.get_or_raise(sg_dialect)
    try:
        return list(dialect.tokenize(sql))
    except TokenError:
        # A quote or a block comment the caret sits inside is unterminated,
        # which is the normal state of one being typed. Closing it costs
        # nothing: the added characters land past the caret, so no earlier
        # token moves.
        closers = {
            **dialect.tokenizer_class._IDENTIFIERS,
            **dialect.tokenizer_class._QUOTES,
            **{k: v for k, v in dialect.tokenizer_class._COMMENTS.items() if v},
        }
        for closer in dict.fromkeys(closers.values()):
            try:
                return list(dialect.tokenize(sql + closer))
            except TokenError:
                continue
        raise


def pos(node: exp.Expression) -> tuple[int, int]:
    """Where a node starts. Use span to mark a range. Falls back to (1, 0)."""
    line, col, _, _ = span(node)
    return line, col


def span(node: exp.Expression) -> tuple[int, int, int, int]:
    """
    Return (start_line, start_col, end_line, end_col) covering node in the source.

    The end column comes from the token's own length, so a quoted identifier
    includes its quotes. Falls back to (1, 0, 1, 0) when nothing is tagged.
    """
    if not isinstance(node, exp.Expression):
        return 1, 0, 1, 0

    for candidate in (node, node.this):
        if isinstance(candidate, exp.Expression):
            extent = _token_extent(candidate)
            if extent is not None:
                return extent

    return _subtree_extent(node)


def _token_extent(node: exp.Expression) -> tuple[int, int, int, int] | None:
    """The span of the token sqlglot consumed to build node, if it kept one."""
    m = node.meta
    if not m or "line" not in m:
        return None
    token_len = m["end"] - m["start"] + 1
    start_col = max(0, m["col"] - token_len)
    return m["line"], start_col, m["line"], start_col + token_len


def _subtree_extent(node: exp.Expression) -> tuple[int, int, int, int]:
    """
    The span of node's tagged descendants, first token to last.

    sqlglot tags the token it consumed, so a node built from a keyword keeps
    none: OFFSET, HAVING and every operator carry theirs on the operands below.
    """
    starts, ends = [], []
    for descendant in node.walk():
        extent = _token_extent(descendant)
        if extent is not None:
            starts.append((extent[0], extent[1]))
            ends.append((extent[2], extent[3]))
    if not starts:
        return 1, 0, 1, 0

    start_line, start_col = min(starts)
    end_line, end_col = max(ends)
    return start_line, start_col, end_line, end_col


def build_schema(schema_dict: dict, dialect: str) -> MappingSchema | None:
    """
    Build a sqlglot MappingSchema from:
      { schema_name: { table_name: [col, ...] | { col: type, ... } } }
    Returns None if schema_dict is empty.
    """
    if not schema_dict:
        return None
    normalised: dict[str, dict[str, dict[str, str]]] = {}
    for schema_name, tables in schema_dict.items():
        normalised[schema_name] = {}
        for table_name, cols in tables.items():
            if isinstance(cols, list):
                normalised[schema_name][table_name] = {c: "TEXT" for c in cols}
            else:
                normalised[schema_name][table_name] = cols
    try:
        return MappingSchema(normalised, dialect=dialect)
    except Exception:
        return None


def collect_virtual_names(stmt: exp.Expression) -> set[str]:
    """
    Return the set of names that are virtual (CTEs + subquery aliases) so rule
    modules can skip them during unknown-table validation.
    sqlglot 25+ stores the WITH clause under "with_" internally; use find() to
    be version-agnostic.
    """
    virtuals: set[str] = set()
    with_clause = stmt.find(exp.With)
    if with_clause:
        for cte in with_clause.expressions:
            virtuals.add(cte.alias.lower())
    return virtuals


def cte_name_for(source) -> str:
    """The CTE name a scope is, when it is one (handling both direct and recursive CTEs)."""
    expression = getattr(source, "expression", None)
    if expression is None:
        return ""
    p = expression.parent
    if isinstance(p, exp.CTE):
        return p.alias
    if isinstance(p, exp.SetOperation) and isinstance(p.parent, exp.CTE):
        return p.parent.alias
    return ""


def _cte_lookup_key(ident: exp.Expression | str | None, dialect: SqlglotDialect) -> str:
    if isinstance(ident, exp.Identifier):
        return dialect.normalize_identifier(ident.copy()).name
    if isinstance(ident, str):
        return dialect.normalize_identifier(exp.Identifier(this=ident, quoted=False)).name
    return ""


def traverse_scopes(stmt: exp.Expression, sg_dialect: str = "postgres") -> list[Scope]:
    """Run sqlglot's traverse_scope and bind FROM/JOIN tables to CTEs using
    the dialect's identifier normalization rules.

    sqlglot's _traverse_tables matches table names against scope.sources by
    exact string equality, which misses a CTE referenced in a different case
    on PostgreSQL (folds unquoted identifiers to lower case) or SQLite
    (case-insensitive identifiers).
    """
    scopes = list(traverse_scope(stmt))
    try:
        dialect = SqlglotDialect.get_or_raise(sg_dialect or "postgres")
    except Exception:
        return scopes

    for scope in scopes:
        if not scope.cte_sources:
            continue
        norm_ctes: dict[str, Scope] = {}
        for cte_alias, cte_scope in scope.cte_sources.items():
            expr = getattr(cte_scope, "expression", None)
            p = expr.parent if expr is not None else None
            cte_node = p if isinstance(p, exp.CTE) else (p.parent if isinstance(p, exp.SetOperation) and isinstance(p.parent, exp.CTE) else None)
            alias_arg = cte_node.args.get("alias") if isinstance(cte_node, exp.CTE) else None
            cte_ident = (
                alias_arg.this
                if isinstance(alias_arg, exp.TableAlias) and isinstance(alias_arg.this, exp.Identifier)
                else cte_alias
            )
            key = _cte_lookup_key(cte_ident, dialect)
            if key:
                norm_ctes[key] = cte_scope

        if not norm_ctes:
            continue

        changed = False
        for source_key, source in list(scope.sources.items()):
            if (
                isinstance(source, exp.Table)
                and not source.db
                and not source.catalog
                and not source.args.get("pivots")
            ):
                tbl_ident = source.this if isinstance(source.this, exp.Identifier) else source.name
                matched = norm_ctes.get(_cte_lookup_key(tbl_ident, dialect))
                if matched is not None:
                    scope.sources[source_key] = matched
                    changed = True
        if changed:
            scope._selected_sources = None

    return scopes
