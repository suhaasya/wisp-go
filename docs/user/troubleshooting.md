# Troubleshooting

## Browser does not open

- Run `wisp` from a terminal and read stderr.
- Check another browser or paste the URL printed in the terminal (loopback with token).
- Ensure nothing else blocked the port; use `wisp stop` and retry.

## “Could not connect”

| Symptom | Things to check |
| --- | --- |
| Timeout | Host, port, firewall, VPN |
| TLS errors | CA file, `sslmode`, server cert hostname |
| Auth failed | User, password, pg_hba / MySQL user host |
| SSH failed | Key permissions, bastion user, jump chain order |
| Socket | Correct socket path and permissions |

Use **Test** on the connection form before save.

## Slow grid or query

- Large tables without suitable indexes make sorts/filters slow on the server side.
- Approximate counts and wide rows increase work; narrow filters help.

## Update failed

- Install [minisign](https://github.com/jedisct1/minisign) for scripted installs.
- `--update` needs network access to GitHub releases and a signature matching the embedded public key.

## High memory

- Close unused connection tabs.
- Lower **Go memory limit** in Settings if you set a custom cap.
- For gate expectations see the [landing page](../../website/index.html) or PRD budgets.

## Still stuck

**Settings → About & feedback → Send feedback** with OS, browser version, database engine, and redacted connection shape (no passwords).
