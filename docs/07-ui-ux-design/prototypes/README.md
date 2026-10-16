# NEO Bank — Design Prototypes

Static HTML prototypes for the NEO Bank interface, with scripts to capture them as images. They implement the tokens, components, and layouts defined in [`../../07-ui-ux-design.md`](../../07-ui-ux-design.md) and exist to validate the design system in a real browser before production code is written.

## 1. What this is

- **Static, dependency-free HTML.** Each page is a single self-contained file you can open directly in a browser — no build step, no server.
- **Styling via CDN.** Tailwind CSS v4 (browser build) and DaisyUI v5 are loaded from a CDN and configured in an inline `@theme` block that mirrors the design tokens. **Prototype-only** — production uses a compiled, purged stylesheet (see [`../../08-tech-stack/frontend/styles/tailwindcss.md`](../../08-tech-stack/frontend/styles/tailwindcss.md)).
- **Both themes.** Every page ships with `neo-light` and `neo-dark` and a theme switcher, so every screen can be reviewed in both modes.
- **Reproducible screenshots.** A Playwright script renders every page in both themes to `images/light/` and `images/dark/`.

## 2. Structure

```text
07-ui-ux-design/prototypes/
├── html/              # HTML prototypes (numbered 00-09)
├── images/
│   ├── light/         # neo-light PNG screenshots
│   └── dark/          # neo-dark PNG screenshots
├── screenshot.py      # Renders every HTML file in both themes
├── screenshot.sh      # Runs setup + screenshot.py (uv + Playwright)
├── pyproject.toml     # Python dependencies (managed with uv)
└── README.md          # This file
```

## 3. Viewing

Open any file in `html/` directly:

```sh
open html/01-landing.html
```

Or serve the folder if your browser restricts local resources:

```sh
python3 -m http.server 8080 --directory html
```

## 4. Files

| HTML                                                           | PNG                      | Screen             | Notes                                                                            |
| -------------------------------------------------------------- | ------------------------ | ------------------ | -------------------------------------------------------------------------------- |
| [`html/00-components.html`](html/00-components.html)           | `00-components.png`      | Component showcase | The front-end core: colour tokens, buttons, badges, alerts, form fields, states. |
| [`html/01-landing.html`](html/01-landing.html)                 | `01-landing.png`         | Landing            | Public marketing page: hero, features, security, FAQ, CTA, footer.               |
| [`html/02-sign-in.html`](html/02-sign-in.html)                 | `02-sign-in.png`         | Login              | Unauthenticated layout, email + password, error and focus states.                |
| [`html/03-register.html`](html/03-register.html)               | `03-register.png`        | Register           | Password strength meter, confirm-password field.                                 |
| [`html/04-dashboard.html`](html/04-dashboard.html)             | `04-dashboard.png`       | Dashboard          | Sidebar navigation shell, balance card, quick actions, recent transactions.      |
| [`html/05-transfer.html`](html/05-transfer.html)               | `05-transfer.png`        | Transfer (step 1)  | Find the recipient: account-ID search and verified result.                       |
| [`html/06-transfer-amount.html`](html/06-transfer-amount.html) | `06-transfer-amount.png` | Transfer (step 2)  | Amount entry, optional description, and confirmation summary.                    |
| [`html/07-transfer-status.html`](html/07-transfer-status.html) | `07-transfer-status.png` | Transfer Status    | Success result, plus an error-state example.                                     |
| [`html/08-transactions.html`](html/08-transactions.html)       | `08-transactions.png`    | Transaction History | Full history list, status badges, pagination.                                   |
| [`html/09-profile.html`](html/09-profile.html)                 | `09-profile.png`         | Profile            | User info, account settings (theme toggle, notifications), change password, delete account. |

The authenticated screens (`04`–`09`) and the component showcase (`00`) share a sidebar app shell. The transfer flow is a three-step wizard: find the recipient (`05`), enter the amount and confirm (`06`), then view the result (`07`).

Each screen is captured in **both** themes under its own name: `images/light/<n>-<name>.png` is the `neo-light` capture and `images/dark/<n>-<name>.png` is the `neo-dark` capture.

## 5. Generating screenshots

Screenshots use [Playwright](https://playwright.dev/python/) via `uv`. Run the wrapper script, which installs dependencies and the browser as needed:

```sh
./docs/07-ui-ux-design/screenshot.sh
```

Or run the steps manually:

```sh
cd docs/07-ui-ux-design
uv sync
uv run playwright install chromium
uv run screenshot.py
```

The script:

1. Reads every `html/*.html` file (sorted by name).
2. Opens it in headless Chromium, sets `localStorage['neo-theme']`, and reloads.
3. Takes a full-page screenshot at a 1920×1080 viewport.
4. Writes `images/light/<name>.png` and `images/dark/<name>.png`.

The `.venv/` directory and `uv.lock` are created by `uv` and are gitignored.

## 6. Theme switcher

The floating control in the bottom-right corner is a **prototype affordance**, not part of the product UI. It sets `data-theme` on `<html>`, persists the choice in `localStorage` under the key `neo-theme`, and falls back to `prefers-color-scheme`. A small inline script runs before paint to avoid a flash of the wrong theme. See the Themes & Implementation Mapping section of [`design-tokens.md`](../system/design-tokens.md#themes--implementation-mapping) for the production theming rules.

## 7. Tokens

The prototypes use the design tokens directly through utility classes generated from the `@theme` block (for example `bg-surface`, `text-text-primary`, `bg-primary-muted`, `text-success-fg`). The canonical, light/dark table lives in the Colors section of [`design-tokens.md`](../system/design-tokens.md#colors) — if a token changes there, update `@theme` and the `[data-theme]` blocks in these files to match.

## 8. Keeping prototypes in sync

These files are illustrative. When the design doc changes:

1. Update the token values in the `@theme` block and the `[data-theme='neo-light']` / `[data-theme='neo-dark']` blocks.
2. Update the affected screens.
3. Regenerate screenshots with `./screenshot.sh` and commit the HTML and PNG changes together.
4. Keep the pages consistent with the component rules in [`../../07-ui-ux-design.md`](../../07-ui-ux-design.md) — in particular, status colour pairs are always used together (`bg-success-bg` with `text-success-fg`) to preserve contrast.
