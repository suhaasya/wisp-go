# Compatibility matrix (generated)

Last updated: pending first CI matrix run.

| Cell | Result |
| --- | --- |
| _no report yet_ | Run nightly or `make e2e-matrix` with docker compose |

## Notes

- PR CI runs cells tagged `pr` (pg17, mysql84, mariadb114 × TCP + SSH).
- Nightly runs all cells in `e2e/matrix/manifest.mjs`.
- Regenerate this file: `node tools/gen-compatibility.mjs e2e/matrix-report.json`.
- Socket transports: follow-up (not in CI yet).
