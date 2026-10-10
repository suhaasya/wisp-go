# Connections

Wisp supports PostgreSQL, MySQL, and MariaDB over common deployment patterns.

## TCP host and port

Default ports: PostgreSQL **5432**, MySQL/MariaDB **3306**. Use **SSL mode** (PostgreSQL) or **Use TLS** (MySQL) when the server requires encryption.

## Connection URL

Paste a URL (`postgresql://…`, `mysql://…`) into the URL field; Wisp parses host, port, database, and SSL parameters. User and password still use secure storage when you save the profile.

## TLS

- **PostgreSQL:** `sslmode` (disable, require, verify-full, etc.) matches libpq behaviour.
- **MySQL/MariaDB:** enable TLS and optionally provide CA / client cert paths.

Certificate files are read from your machine; they are not uploaded anywhere.

## SSH tunnel and jump hosts

Use **SSH** when the database is reachable only through a bastion:

1. Enable SSH on the connection.
2. Set bastion host, user, and auth (key file or agent).
3. Optional **jump host** chain for multi-hop setups.

Wisp opens a local forwarded port and connects the database driver through it.

## Unix socket and named pipe

- **PostgreSQL on Linux/macOS:** socket path instead of TCP when the server listens locally.
- **Windows:** named pipe path for local SQL Server–style layouts where supported by the driver configuration.

## Cloud providers

There is no special “provider” mode: use the hostname and TLS settings from your vendor (RDS, Cloud SQL, Azure, etc.). For IP allowlists, connect from a machine whose IP is permitted, or use SSH to a host inside the network.

## Saved profiles

Connections are saved as **profiles** with environment tags (e.g. production). Tags drive colour cues and read-only defaults where you configure them in **Data safety** settings.

## Troubleshooting connects

See [troubleshooting.md](troubleshooting.md) for TLS, SSH, and firewall issues.
