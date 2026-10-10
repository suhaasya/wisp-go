# Release checklist (1.0)

Use this list for release candidates and the final **1.0.0** tag. Manual steps are intentional; automation lives in CI and GoReleaser ([release.md](release.md)).

## Before tagging

- [ ] **Version** — bump `VERSION` / tag `v1.0.0`; update [CHANGELOG.md](../CHANGELOG.md) (move Unreleased → 1.0.0).
- [ ] **Licence and name** — product name **Wisp**; [LICENSE](../LICENSE) (Apache 2.0) in release artifacts / README.
- [ ] **P0 bugs** — triage GitHub Issues; zero open P0 for 1.0 (label or milestone).
- [ ] **Beta feedback** — review issues filed via **Settings → About & feedback** during beta window.

## Build and gates (local)

```bash
make lint
make test
make web
make size bench gate
```

- [ ] `make gate` — no metric **over cap** (warn above target is OK per AGENTS.md).
- [ ] Record numbers in the release announcement (binary MB, download MB, bundle KB from `bench-results/`).

## CI (on RC branch or main)

- [ ] **ci** workflow green on commit to tag.
- [ ] **gate** job green.
- [ ] **e2e** — Chromium, Firefox, WebKit (`make e2e` or CI).
- [ ] **e2e-matrix** — full LW-046 matrix on RC (`WISP_MATRIX_TAG=nightly` or CI nightly), not only PR subset.

## Artifacts

- [ ] Push annotated tag `v1.0.0` → **release** workflow.
- [ ] `MINISIGN_SECRET_KEY` set → `.minisig` on archives ([release.md](release.md)).
- [ ] Spot-check install script on macOS/Linux and install.ps1 on Windows.
- [ ] `wisp --update` from previous release (optional smoke).

## Docs and web

- [ ] [docs/user/README.md](user/README.md) matches shipped behaviour.
- [ ] [website/index.html](../website/index.html) budget numbers match [PRD](PRD.md) / `internal/bench/budgets.go`.
- [ ] [docs/compatibility.md](compatibility.md) regenerated after nightly matrix if needed.

## Announcement

- [ ] GitHub Release notes from CHANGELOG.
- [ ] Landing page deployed (static host of `website/` or GitHub Pages).
- [ ] Screenshots in release post (not stored in repo; capture from running app or design mockup reference).

## Post-release

- [ ] Monitor issues for 48h; hotfix `v1.0.1` if P0 found.

### Sign-off

| Role | Name | Date |
| --- | --- | --- |
| Engineering | | |
| QA (matrix + e2e) | | |
