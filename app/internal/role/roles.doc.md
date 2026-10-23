# Roles

A role is a named set of [permissions](/docs/workspace/permissions/). It does
nothing until you hold it: the rules SELECT checks are the rules of every role
assigned to you.

With no server and no other users, the person a role is assigned to is you.
That makes roles guardrails: a production database you can read but never
write, a column you would rather never see on screen.

## Creating a role

Open **Settings → Roles** and click **New role**. Give it a name, unique within
the workspace, then tick the rules it carries. You can also **duplicate** an
existing role and adjust it.

A new role applies to nobody yet. Click **Add user** at the top of the role and
add yourself.

## Holding several roles

Rules from all the roles you hold are combined, and a deny always wins over an
allow. See [Allow and deny](/docs/workspace/permissions/#allow-and-deny).

## Examples

- **Read-only production**: allow `SELECT` and `SEE` on the production
  database, and nothing else. Writes and schema changes there are refused,
  while every other database stays as open as it was.
- **Masked column**: allow `SELECT` and `SEE` on the database, then deny `SEE`
  on `customers.email`. Its cells come back as `*****`, and a query that
  filters on it is refused. The allow matters: once a role you hold has any rule
  on a database, whatever it does not allow there is refused.
- **No accidental deletes**: allow `SELECT`, `SEE`, `INSERT` and `UPDATE` on a
  database, but not `DELETE`.

> [!NOTE]
> Roles have no hierarchy or inheritance. Each role is a flat set of permissions.
