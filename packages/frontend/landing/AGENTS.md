# AGENTS.md - packages/frontend/landing

## Overview

Static landing page for NEO Bank. A single `index.html` + `styles.css` +
`assets/` — no framework, no build step, no JavaScript. Deployed to GitHub
Pages as-is at `https://hieudoanm.github.io/neo.bank/`.

## Conventions

- **Fonts:** Inter (sans) loaded via the Google Fonts `<link>` in `index.html`; mono stack `ui-monospace, 'SF Mono', SFMono-Regular, Menlo, Consolas, monospace`. Tokens: `--font-sans` / `--font-mono`. Do not add other typefaces.

- Keep it dependency-free: edit `index.html`/`styles.css` directly.
- Asset URLs are relative (`styles.css`, `assets/…`, `favicon.svg`) so the
  page works under the `/neo.bank/` Pages base path and from `file://`.
  Raster assets live in `assets/`; root-level files are the standard
  metadata set (`robots.txt`, `sitemap.xml`, `humans.txt`, `404.html`,
  `favicon.*`, `og.png`).
- Light theme only; design follows the "AI unslop" rules (see
  `hieudoanm/notes/ai-unslop-design.md` in the notes repo): no filler
  badges, monospace only for code, one accent colour, real UI
  screenshots instead of illustrations.
- Keep it a landing page — the interactive app lives in
  `packages/frontend/app`.

## Commands

There are none — serve statically:

```bash
python3 -m http.server 8000 -d packages/frontend/landing
```
