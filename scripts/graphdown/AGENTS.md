# Build a Python CLI for Markdown Extraction and Graph Visualization

## Objective

Build a maintainable Python CLI that scans a directory, extracts relationships between files—focusing on Markdown links—and generates a machine-readable graph plus visual representations.

The tool should help developers and AI coding agents understand a repository's documentation structure, navigate references, detect broken links, and identify orphaned files.

## 1. Technology Stack

- Python 3.12+.
- `uv` for project and dependency management.
- `typer` for the CLI.
- `pydantic` for validating graph data models.
- `markdown-it-py` for parsing Markdown.
- `matplotlib` and `networkx` for generating SVG.
- `pyvis` for the interactive HTML visualization.
- `pytest` for testing.
- `ruff` for linting and formatting.
- `mypy` or an equivalent static type checker.

Use type annotations throughout. Avoid `Any` unless strictly necessary. Prefer small, focused modules and pure functions where practical.

## 2. Graph Extraction

Recursively scan a user-specified directory for files.

### Markdown link extraction

Extract links in these forms:

- Inline links: `[text](./other.md)`
- Relative links: `[text](../../docs/architecture.md)`
- Reference-style links: `[text][architecture]` and `[architecture]: ./architecture.md`
- Autolinks and relevant Markdown links inside images, if appropriate.

Handle URL-encoded paths, optional titles, fragments, and query strings correctly.

Focus on local file references. Ignore external HTTP(S), email, and other non-file URLs by default.

### Path resolution

- Resolve relative links against the directory containing the source file.
- Normalize paths and represent them relative to the scan root using POSIX separators.
- Distinguish filesystem paths from URL fragments such as `guide.md#installation`.
- Detect missing targets and links that resolve outside the scan root.
- Support configurable file extensions, including `.md`, `.mdx`, `.txt`, `.py`, `.ts`, `.tsx`, `.js`, `.jsx`, `.json`, and `.yaml`.
- Parse Markdown links in Markdown files. Treat other files as graph nodes even if their contents are not parsed.
- Avoid following symbolic links by default to prevent cycles and unintended traversal.
- Define and document behavior for directory links and links without explicit extensions.

Every extracted edge should retain its original link, source line number, target path, optional fragment, and resolution status.

### Ignored paths

- Honour `.gitignore`: when the scan root is inside a git work tree, files and directories excluded by git ignore rules (nested `.gitignore` files, `.git/info/exclude`, and global excludes) are omitted from the graph.
- Do not report links whose target is ignored as broken; drop them so ignored paths never appear in the graph.
- Report how many candidate files were ignored, and provide a way to opt out (e.g. `--no-gitignore`).
- Degrade gracefully when git is unavailable or the scan root is outside a work tree.

### Graph analysis

Calculate:

- Total files and links.
- Incoming and outgoing link counts.
- Broken links.
- Orphan files, defined as files with no incoming links from other files.
- Self-links.
- Connected components.
- Most referenced files.

Document how the root document and files without incoming links are treated.

## 3. Canonical Graph Data Model

Use Pydantic models to define and validate a versioned JSON schema.

Example:

    {
      "version": 1,
      "root": ".",
      "nodes": [
        {
          "id": "README.md",
          "path": "README.md",
          "extension": ".md"
        }
      ],
      "edges": [
        {
          "source": "README.md",
          "target": "docs/setup.md",
          "raw_link": "./docs/setup.md",
          "line": 12,
          "fragment": null,
          "status": "resolved"
        }
      ],
      "stats": {
        "files": 1,
        "links": 1,
        "broken_links": 0
      }
    }

Treat this as an illustrative schema; implement accurate statistics and include every discovered target as a node when appropriate, including missing targets when useful for visualizing broken links.

Use normalized relative paths as stable node IDs. Sort nodes and edges deterministically. Ensure repeated runs on an unchanged directory produce identical JSON, excluding any intentionally variable metadata.

Write JSON as UTF-8 with predictable indentation and a trailing newline.

## 4. Graph Visualization

Implement a separate visualization module that consumes the canonical graph model rather than rescanning files.

### Graphviz DOT

Generate a DOT representation with:

- Directed edges from source file to linked file.
- Configurable layout direction: top-to-bottom (`TB`) or left-to-right (`LR`).
- File names as node labels, with optional shortened labels for long paths.
- Optional grouping of nodes by directory using Graphviz clusters.
- Different visual styles for Markdown files, other files, missing targets, and orphan files.
- A legend explaining node and edge styles.
- Escaping of all user-controlled labels and identifiers.

Avoid using full absolute paths as visible labels unless explicitly requested.

### SVG output

Generate SVG with `matplotlib` and `networkx`, rendering the canonical graph model through the headless `Agg` backend.

- Build a directed NetworkX graph from the model; links to directories or missing files become nodes automatically.
- Use `nx.spring_layout` with a fixed seed so repeat runs lay the graph out identically.
- Support configurable output paths, layout direction (`TB` or `LR`), and graph title.
- Use file name stems as labels and the Obsidian-inspired dark palette in `renderers/styles.py` (dark canvas, purple nodes, muted links, light labels), matching the HTML output.
- Keep the DOT output on the light palette, since it is normally rendered on a white background.
- Set `svg.hashsalt` and omit the creation timestamp so repeat runs are byte-identical.
- Do not require the Graphviz executable, Mermaid CLI, Puppeteer, Chrome, or a headless browser.

Support DOT-only output for users who want to inspect or render the graph themselves.

### Other outputs

Generate, with a shared `graphdown.*` base name, into `<scan-root>/.graphdown/` by default:

- `graphdown.json`: canonical machine-readable graph.
- `graphdown.md`: readable report of graph statistics, broken links, orphan files, and important relationships.
- `graphdown.dot`: optional Graphviz source.
- `graphdown.svg`: vector visualization.
- `graphdown.html`: interactive PyVis/vis-network visualization, Obsidian-inspired and self-contained in a single file.

Keep each output independently selectable. All artefacts use the same base name, so a plain file-name constant (e.g. `BASENAME = "graphdown"`) must drive both writing and the scanner's output-file exclusion set.

### PyVis HTML output

Generate a single self-contained HTML file with the `pyvis` package (`cdn_resources="in_line"`), consuming the canonical graph model so it can be regenerated from `graphdown.json` without rescanning.

- Style the view like Obsidian's graph: dark purple palette, borderless dot nodes, muted links without arrowheads, and a force-directed layout (`forceAtlas2Based`).
- Size nodes by their number of connections and show the path plus connection count as a hover tooltip.
- Treat the view as undirected: reciprocal links are drawn once.
- vis-network requires every edge endpoint to be a node, so materialise links to directories or missing files as placeholder nodes.
- Patch the PyVis page chrome (body background, canvas border) since only the canvas colour is configurable through `Network`.
- Seed the layout so repeated runs produce identical output.

## 5. CLI Interface

Use Typer to implement commands similar to:

    graphdown scan ./docs
    graphdown scan ./docs --output ./build/graph
    graphdown scan ./docs --format json
    graphdown scan ./docs --check
    graphdown render ./docs/.graphdown/graphdown.json --format svg
    graphdown render ./docs/.graphdown/graphdown.json --format dot
    graphdown render ./docs/.graphdown/graphdown.json --format html
    graphdown report ./docs/.graphdown/graphdown.json
    graphdown --help

Requirements:

- Support an explicit scan root and output directory. Default the output directory to `<scan-root>/.graphdown`.
- Name every artefact `graphdown.*`.
- Support include/exclude glob patterns and extension filters.
- Allow JSON, Markdown, DOT, SVG, and HTML outputs to be selected independently.
- Make `--check` return a non-zero exit code when broken local links are detected.
- Print concise progress and useful error messages to stderr.
- Reserve stdout for machine-readable output when using stdout mode.
- Handle nonexistent paths, permission errors, invalid JSON, and rendering failures gracefully.
- Do not overwrite unrelated files without explicit user intent.
- Ensure output files are excluded from scanning if they are created inside the scan root.

Choose sensible defaults and document them.

## 6. Architecture

Use a clean package structure similar to:

    src/graphdown/
      __init__.py
      cli.py
      constants.py
      scanner.py
      markdown_parser.py
      resolver.py
      models.py
      analyzer.py
      serializers.py
      renderers/
        __init__.py
        dot.py
        svg.py
        pyvis_graph.py
        styles.py
        markdown_report.py
      exceptions.py

Keep responsibilities separate:

- Scanner: discovers files.
- Markdown parser: extracts links and source locations.
- Resolver: normalizes and validates targets.
- Models: defines graph structures.
- Analyzer: calculates graph statistics.
- Serializers: reads and writes JSON.
- Renderers: produces DOT, SVG, HTML, and Markdown reports.
- CLI: orchestrates the workflow.

The visualization layer must be able to regenerate outputs from an existing JSON file without reading the original source directory.

## 7. Testing

Write automated tests covering:

- Inline and reference-style Markdown links.
- Nested relative paths and URL-encoded paths.
- Fragments and query strings.
- Missing targets and external URLs.
- Path traversal outside the root.
- Symbolic links and cyclic directory structures.
- Non-Markdown files as graph nodes.
- Empty directories and graphs.
- Duplicate links and self-links.
- Stable JSON serialization.
- DOT label escaping and missing-target styling.
- SVG and DOT generation.
- HTML generation, including the Obsidian theme and connection-based node sizing.
- CLI exit codes, stdout/stderr behavior, and error handling.

Use temporary directories and fixtures rather than depending on a real repository. Mock external processes where appropriate, and include an integration test for SVG rendering.

## 8. Developer Experience

Provide:

- A `pyproject.toml` with dependencies and tool configuration.
- A README with installation, usage, examples, and architecture.
- A Makefile with targets for formatting, linting, type checking, testing, and building.
- A small example directory and sample outputs.
- Clear documentation of supported Markdown syntax and known limitations.

Use `uv run` for development commands. Keep the implementation simple, maintainable, deterministic, and easy for coding agents to understand.

## 9. Implementation Workflow

1. Inspect the existing repository before making changes. Preserve existing conventions and avoid overwriting unrelated files.
2. Define the graph models and JSON schema.
3. Implement filesystem scanning, Markdown parsing, and link resolution.
4. Implement graph analysis and deterministic JSON serialization.
5. Implement DOT, SVG, and Markdown report generation.
6. Implement CLI commands and options.
7. Add unit tests and integration tests.
8. Run formatting, linting, type checking, and tests.
9. Test an end-to-end scan and render on a small fixture.
10. Update documentation with actual commands and verified results.

Do not claim tests pass unless they have been executed successfully. Report any environmental dependencies or unverified behavior.

## Acceptance Criteria

The implementation is complete when a user can scan a folder, generate deterministic graph JSON, render SVG from that JSON without rescanning the folder, inspect broken links and orphan files, and run automated tests successfully.
