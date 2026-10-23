# The Workspace Folder

A workspace is a folder on your machine. Any folder: a repository you cloned, a
directory of `.sql` files you have had for years, or an empty one you just made.
SELECT opens it the way an editor does, and reads what is already in it.

There is no managed location and nothing hidden in your application data
directory. Your files stay where you put them.

## Opening a folder

With no folder open, the app shows the folders you opened recently where your
files would be. Open one from there, or with the
button in the top left corner -- `Cmd+O`.

![The app with no folder open: recent folders where the files would be, and the workspace button in the top left corner.](/shots/folder.start.light.webp)

What happens next depends on what is in the folder.

**Already a workspace.** It holds a `select.config.json`, so SELECT builds the
tree. A workspace a teammate made opens the same way, and so does one whose
config names a Select server: it becomes yours on this machine.

**Not a workspace yet.** SELECT offers to make one, named after the folder.
Creating it adds that one file and changes nothing else; an empty folder also
gets a small sample database to start from.

![Opening a folder that is not a workspace yet: its path, the workspace name filled in from the folder, and what creating one will add.](/shots/folder.setup.light.webp)

## `select.config.json`

This is the file that makes a folder a workspace. It sits at the root and
names the workspace:

```json select.config.json
{
  "$comment": "Managed by SELECT. Do not edit by hand.",
  "version": 1,
  "server": "local",
  "workspaceId": "4f0c2a9e-..."
}
```

Do not edit it by hand. SELECT writes it when you create a workspace and reads
it every time you open the folder.

> [!TIP]
> **Commit it.** That is the point of the file. A teammate clones the
> repository, opens the folder, and lands in the same workspace without being
> told anything.

The workspace's **name** is deliberately not in it. The name is kept on your
machine, so renaming it renames it for you alone, and a folder you open for the
first time is named after its directory.

## Sharing a workspace with your team

1. One person creates the workspace in a folder and commits
   `select.config.json` along with their queries.
2. Everyone else clones the repository and opens the folder.

That is all: there is no account and nobody to invite. Git moves the files, and
everything that is not a file stays on each machine. See
[Git](/docs/workspace/git/) for what lives where.

## Large folders

SELECT reads your `.gitignore` and skips the folders it excludes when building
the tree and watching for changes, so opening a repository with a `node_modules`
or a `target` in it stays fast. Only folders are skipped, never files: `.env` is
gitignored by default and SELECT still reads your `$VARIABLES` from it. See
[.gitignore](/docs/workspace-files/gitignore/).

## Deleting a workspace

Deleting a workspace closes the folder, forgets the workspace on this machine,
and deletes `select.config.json`. **Your files are not touched.** The folder,
the queries and the databases in it stay exactly as they were, and you can
create a new workspace in the same folder afterwards.

In a repository, the deleted `select.config.json` is a change like any other:
your teammates only lose it if you commit it.

## Folders SELECT will not open

Your home directory and a filesystem root are refused. Both mean indexing
everything you own, and both are mis-clicks rather than intentions. Pick a
folder inside them instead.
