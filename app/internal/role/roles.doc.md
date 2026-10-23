# Roles

A role is a named set of [permissions](/docs/workspace/permissions/). The rules
SELECT checks are those of the roles assigned to you, so a role does nothing
until you hold it. On your own machine, that makes roles guardrails you set on
yourself.

## Creating a role

Open **Settings → Roles** and click **New role**. Give it a name, unique within
the workspace, then tick the rules it carries. You can also **duplicate** an
existing role and adjust it.

Then click **Add user** at the top of the role and add yourself.

## Holding several roles

Rules from all the roles you hold are combined, and a deny always wins over an
allow. See [Allow and deny](/docs/workspace/permissions/#allow-and-deny), and
[Example setup](/docs/workspace/permissions/#example-setup) for a read-only
production database and a masked column.

> [!NOTE]
> Roles have no hierarchy or inheritance. Each role is a flat set of permissions.
