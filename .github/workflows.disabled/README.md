# GitHub Actions (disabled)

Workflows here are **not** run by GitHub. Only files under `.github/workflows/` are active.

They were moved aside to avoid Actions minutes cost while the project is in hobby / local-only mode.

## Re-enable later

```bash
mv .github/workflows.disabled/*.yml .github/workflows/
```

Then turn branch protection status checks back on in the repo settings if you use them (`ci`, `gate`).

## Local checks (same as CI)

```bash
make lint
make test
make size bench gate   # gate needs Docker for DB fixtures
```
