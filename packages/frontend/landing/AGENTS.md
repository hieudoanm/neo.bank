# AGENTS.md - packages/frontend/landing

## Overview

Static landing page for NEO Bank. A single `index.html` + `styles.css` +
`assets/` — no framework, no build step, no JavaScript except a tiny year-injection snippet. Deployed to GitHub
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

## Footer

The footer follows the shared contract: `© {year} {repo-name}` (lowercase
repo name, dynamic year via `<span id="year">` + the year script), a
one-line tagline/license descriptor, and a right-aligned `<nav
aria-label="Footer">` with exactly two links — `GitHub` (repo root) and
`README` (`<repo>#readme`), both `target="_blank" rel="noreferrer"`.

## Features & CTA sections

`#features` and `#install` follow the shared template: centered
`.section-head` (h2 + one supporting paragraph), a `.grid-3` card grid
(`article.card` = `.card-icon` + h3 + p), and a closing CTA section on the
alt background (`.section-alt`, `.container-narrow`) ending in a
`.cta-row` with a primary `btn-lg` action plus at most one ghost button.
