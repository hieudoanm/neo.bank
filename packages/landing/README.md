# @neo-bank/landing

Static landing page for [NEO Bank](https://github.com/hieudoanm/neo.bank) —
"Instant transfers, no waiting". Plain HTML/CSS, no build step.

Deployed to GitHub Pages at <https://hieudoanm.github.io/neo.bank/>.

## Files

| File | Purpose |
| ---- | ------- |
| `index.html` | Single-page markup |
| `styles.css` | All styling |
| `assets/` | Images and SVG logos referenced by the page |
| `favicon.svg`, `favicon.ico` | Browser icons |
| `og.png` | Social share image (1200×630) |
| `robots.txt`, `sitemap.xml`, `humans.txt` | Crawler/credit metadata |
| `404.html` | GitHub Pages 404 page |

## Preview

```bash
python3 -m http.server 8000 -d packages/landing
```
