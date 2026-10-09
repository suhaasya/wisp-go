# Browser support (Wisp 1.0)

Wisp targets **current** versions of Chrome, Edge, Firefox, and Safari on desktop. Mobile layouts are out of scope.

## Supported browsers

| Browser | Automation | JS heap / DOM (CI) | Whole-tab memory |
| --- | --- | --- | --- |
| Chrome / Edge (Chromium) | Playwright `chromium` + tabbench CDP | `make gate` tabbench | Chrome Task Manager (manual each release) |
| Firefox | Playwright `firefox` + `npm --prefix tools/tabbench run browsers` | Same scripts; CDP metrics skipped | Firefox about:processes (manual) |
| Safari | Playwright `webkit` (WebKit) + manual Safari | No `performance.memory`; tab meter hidden in UI | Safari Activity Monitor (manual) |

## Memory notes

- **JavaScript heap** and **DOM node** budgets are enforced in Chromium via `tools/tabbench/tabbench.mjs` and `make gate`.
- **`performance.memory`** is Chromium-only. The status bar **Tab** meter is hidden when that API is unavailable (Firefox, Safari).
- **Whole-tab** footprint (PRD target ≤ 40 MB, cap 70 MB) includes browser fixed cost; measure manually in each engine’s task manager before release and record below.

### Manual whole-tab log (fill per release)

| Browser | Version | OS | Whole tab (MB) | Date | Notes |
| --- | --- | --- | --- | --- | --- |
| Chrome | | | | | |
| Edge | | | | | |
| Firefox | | | | | |
| Safari | | | | | |

## Running cross-browser checks

**One-time:** install Playwright browsers (large download; run once, not in parallel):

```bash
rm -rf ~/Library/Caches/ms-playwright/__dirlock   # if a prior install was interrupted
cd e2e && npm ci && npx playwright install chromium firefox webkit
```

**E2E:** needs a live Wisp on the port in `instance.json` (or set `WISP_BASE` + `WISP_LAUNCH_URL`):

```bash
make build
./bin/wisp --no-browser --port 19287   # or use your dev instance on 8787

cd e2e && npx playwright test --project=chromium --project=firefox --project=webkit

On macOS, the `chromium` project uses **installed Google Chrome** by default (avoids a broken partial Playwright Chromium cache). Override with `PW_CHROMIUM_CHANNEL` or reinstall browsers after removing `~/Library/Caches/ms-playwright/__dirlock`.

# Tabbench browser smoke (no CDP on Firefox/WebKit)
WISP_BASE=http://127.0.0.1:19287 WISP_LAUNCH_URL='…' npm --prefix tools/tabbench run browsers
```

## Known differences

- **Clipboard**: requires a secure context and user gesture; behaviour varies slightly by engine.
- **Shortcuts**: Wisp uses **Alt** for tab/window actions the browser owns (see `docs/accessibility.md`). **Mod+K**, **Mod+S**, **Mod+Enter** follow the mockup.
- **SQL editor**: CodeMirror 6; lazy-loaded chunk size gated in `make gate`.
