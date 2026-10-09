# GitHub Actions (disabled)

Only files under `.github/workflows/` are active.

**Active:** `ci.yml`, `gate.yml`, `release.yml`

**Still disabled:** `nightly.yml` (placeholder for LW-048)

## Re-enable nightly

```bash
mv .github/workflows.disabled/nightly.yml .github/workflows/
```

## Local checks (same as CI)

```bash
make lint
make test
make size bench gate   # gate needs Docker for DB fixtures
```
