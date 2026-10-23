# SSH Tunnels

When your database is behind a bastion or jump host, SELECT can open an **SSH tunnel** to reach it.

## Setup

In the connection form, switch the connection mode to **DSN + SSH tunnel**. This reveals the SSH configuration fields:

![The connection form in tunnel mode: the DSN, then the tunnel's own user, host, port and authentication method.](/shots/dbform.ssh.light.webp)

| Field           | Description                                              |
|-----------------|----------------------------------------------------------|
| **Host**        | Bastion/jump host hostname or IP (not the database host) |
| **Host key**    | Public key of the SSH server (looked up in `~/.ssh/known_hosts` when empty) |
| **Port**        | SSH port, defaults to `22`                               |
| **User**        | System user on the SSH host                              |
| **Auth method** | `agent`, `key_file`, `password`, or `private_key`        |
| **Password**    | SSH password (when auth method is password)              |
| **Private key** | Full PEM key content (when auth method is private_key)   |
| **Key file**    | Path to a private key on disk (`key_file` auth; only the path is stored) |

> [!IMPORTANT]
> The **Host** field refers to the SSH jump host, not the database. The database host goes in the DSN.

## Authentication methods

SELECT runs on your machine, so it can authenticate without storing any secret:

| Method        | Stored | Notes                                                                 |
|---------------|--------|----------------------------------------------------------------------|
| `agent`       | nothing | Default. Talks to the running agent (unix `SSH_AUTH_SOCK`; the OpenSSH named pipe on Windows). Encrypted and hardware-backed keys (e.g. YubiKey) work, since the agent does the signing. |
| `key_file`    | path only | Reads the key at connect time; prompts for a passphrase if encrypted (held in memory only). |
| `password`    | value/`$var` | Prefer an `.env` `$variable` over a raw value.                  |
| `private_key` | value/`$var` | Raw PEM or an `.env` `$variable`; prefer the variable.           |

## Host key verification

A pinned host key lets SELECT verify, on every connection, that it is reaching the real bastion and not an impostor (a man-in-the-middle). When a host key is pinned, a mismatch aborts the connection.

SELECT pins automatically when possible: at connect time it looks the host up in `~/.ssh/known_hosts` and pins that key, reusing the trust your own `ssh` client already established, so there is nothing to fill in. If the host is **not** in `known_hosts`, there is no key to pin, so the connection proceeds **unverified**. To get verification, either `ssh` to the host once so it lands in `known_hosts`, or paste its key into the **Host key** field. Get the key by running:

```bash
ssh-keyscan bastion.example.com
```

The output looks like this (lines starting with `#` are comments, ignore them):

```
# bastion.example.com:22 SSH-2.0-OpenSSH_8.7
bastion.example.com ecdsa-sha2-nistp256 AAAAE2VjZHNh...
bastion.example.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5...
```

Paste any of the non-comment lines into the host key field. SELECT will lock the connection to that key type automatically.

## How it works

SELECT opens a **local port-forward** through the SSH host to the database. The DSN connects to the forwarded local port transparently.

Connections are **pooled and kept alive** to avoid re-establishing tunnels on every query. If the tunnel drops, SELECT detects it via keepalive and reconnects on the next query.
