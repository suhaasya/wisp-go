# Accessibility (P0 keyboard flows)

Wisp aims for **keyboard-operable P0 flows** with visible focus, semantic roles on major surfaces, and labels on form controls. This is not a full WCAG audit.

## Focus

- Global **`:focus-visible`**: 2px accent outline, 1px offset (`web/src/styles/app.css`, matches design mockup).
- Prefer theme components in `web/src/ui/` so focus styles stay consistent.

## Semantics (implemented)

| Surface | Roles / labels |
| --- | --- |
| Workspace tabs | `tablist` / `tab` / `tabpanel`, `aria-selected`, `aria-controls` |
| Schema sidebar tree | `role="tree"`, group `treeitem` + `aria-expanded` |
| Data grid | `role="grid"`, `aria-label`, row/column counts |
| Command palette | `role="dialog"`, `aria-modal`, labelled input |
| Modals (`Modal`, shortcut help) | `role="dialog"`, `aria-modal`, `aria-label` / `aria-labelledby` |
| Settings / connection form | Section labels, `aria-label` on segmented controls |

## Motion and text scaling

- **`prefers-reduced-motion: reduce`**: animations and transitions minimized in `app.css`.
- **`html`**: `font-size: 100%` and `text-size-adjust: 100%` so OS text scaling applies.

## P0 keyboard checklist

Record **Pass/Fail** when validating a release (keyboard only, no pointer). Automated coverage in `e2e/tests/` where noted.

| Flow | Steps | Auto |
| --- | --- | --- |
| Open command palette | `Mod+K`, type filter, Esc close | `e2e/tests/keyboard-palette.spec.ts` (Chromium ✓) |
| Shortcut cheat sheet | `Mod+/`, Esc close | `e2e/tests/dialog.spec.ts` (Chromium ✓) |
| Connections list | Tab to card, Enter connect (needs profile) | Manual |
| New connection form | Tab through General fields; section tabs | Manual |
| Workspace tab switch | Alt+1…9 or activate tab in tablist | `e2e/tests/workspace-tabs-a11y.spec.ts` (Chromium ✓; bench tablist) |
| New query tab | Alt+T | Manual |
| Close tab | Alt+W | Manual |
| Schema sidebar | Tree focus, expand group, open table | Manual |
| Grid browse | Scroll, cell focus (where implemented) | Manual |
| Filter bar | Tab to conditions, apply | Manual |
| Review / commit SQL | Mod+S on data tab | Manual |
| SQL editor run | Mod+Enter / Shift+Mod+Enter | Manual |
| Settings | Navigate sections, change theme/density | Manual |
| Destructive confirm | Tab to confirm, Esc cancel | Manual |

## Safari / Firefox

- Tab memory meter hidden without `performance.memory` (see `docs/browser-support.md`).
- Re-run the checklist in Firefox and Safari before 1.0; file issues for engine-specific focus or shortcut gaps.
