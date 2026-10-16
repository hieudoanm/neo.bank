# Graphdown

Builds a dependency graph of files (especially Markdown links) in the repo/docs.

## Usage

By default `graphdown scan <dir>` writes its artefacts to a `.graphdown/`
directory **inside the scan root**, and names every file `graphdown.*`
(`graphdown.json`, `graphdown.md`, `graphdown.dot`, `graphdown.svg`,
`graphdown.html`). For this repo the convenience script scans `docs/`, so the
outputs land in `docs/.graphdown/`.

```bash
# Scan docs and generate outputs (default output dir: docs/.graphdown/)
./scripts/graphdown/graphdown-docs.sh

# Fail (non-zero exit) when broken local links are found
./scripts/graphdown/graphdown-docs.sh --check
```

Outputs: `docs/.graphdown/graphdown.json`, `graphdown.md`, `graphdown.dot`,
`graphdown.svg`, `graphdown.html` (interactive).

Pass `--output <dir>` to write somewhere else, e.g.
`graphdown scan ./docs --output ./build/graph`.

`graphdown.svg` is rendered with matplotlib and NetworkX (headless `Agg` backend — no Graphviz executable required) and uses the same Obsidian-inspired dark theme as `graphdown.html`: dark canvas, purple file nodes, muted links and light labels. It keeps direction, so links carry arrowheads. `graphdown.dot` stays light-themed and available as DOT source for anyone who does want to run `dot` themselves.

### Interactive graph

`graphdown.html` is an interactive PyVis/vis-network visualization: drag nodes, zoom, and hover a node or edge to see its full path or original link. It is written as a single self-contained file (the vis-network library is inlined) and can be regenerated from `graphdown.json` without rescanning:

```bash
uv run --project scripts/graphdown graphdown render docs/.graphdown/graphdown.json --format html
```

Purple file nodes are sized by how many connections they have, links are muted and drawn without arrowheads (the view is undirected, so reciprocal links are drawn once), and hovering shows the path and connection count. Links whose target is a directory or a missing file appear as muted/rose placeholder nodes, since vis-network requires every edge endpoint to be a node. The optional PyVis control panel loads Bootstrap from a CDN; the graph itself renders offline.

Both `graphdown.svg` and `graphdown.html` share the Obsidian palette from `renderers/styles.py`; only the DOT output keeps the light palette.

## Ignored paths

Paths excluded by `.gitignore` — nested files, `.git/info/exclude`, and global excludes — are omitted from the graph. The `Ignored files` statistic in `graphdown.md` reports how many candidate files were dropped, and links whose target is ignored are dropped rather than flagged as broken.

Use `--no-gitignore` to include them:

```bash
./scripts/graphdown/graphdown-docs.sh --no-gitignore
```

When the scan root is not inside a git work tree (or git is unavailable) no paths are ignored.
