from __future__ import annotations

from pathlib import Path
from typing import Optional

import typer

from .analyzer import analyze
from .constants import BASENAME, OUTPUT_DIRNAME
from .markdown_parser import extract_links
from .models import Edge, Graph, Node
from .renderers.dot import write_dot
from .renderers.markdown_report import write_report
from .renderers.pyvis_graph import write_html
from .renderers.svg import write_svg
from .resolver import normalize_relative
from .scanner import git_ignored, scan_files
from .serializers import write_json

app = typer.Typer(name="graphdown")

EXTERNAL_PREFIXES = ("http://", "https://", "ftp://", "mailto:", "tel:")


@app.command()
def scan(
    root: Path = typer.Argument(
        ..., exists=True, file_okay=False, dir_okay=True, readable=True
    ),
    output: Optional[Path] = typer.Option(
        None, help=f"Output directory (default: <root>/{OUTPUT_DIRNAME})"
    ),
    format: Optional[str] = typer.Option(
        None, help="Comma-separated formats: json,dot,svg,md,html"
    ),
    check: bool = typer.Option(False, help="Exit non-zero if broken links found"),
    direction: str = typer.Option("TB"),
    gitignore: bool = typer.Option(
        True,
        "--gitignore/--no-gitignore",
        help="Skip files and folders excluded by .gitignore",
    ),
):
    root_r = root.resolve()
    outdir = (output or (root_r / OUTPUT_DIRNAME)).resolve()
    formats = (
        ["json", "md", "dot"]
        if format is None
        else [f.strip() for f in format.split(",")]
    )

    graph = Graph(root=root_r.as_posix())

    try:
        out_rel = outdir.relative_to(root_r)
    except ValueError:
        out_rel = None

    scan = scan_files(root_r, respect_gitignore=gitignore)
    for fp in scan.files:
        try:
            rel = fp.relative_to(root_r)
        except ValueError:
            continue
        if out_rel is not None and (rel == out_rel or out_rel in rel.parents):
            continue
        nid = rel.as_posix()
        graph.nodes.append(Node(id=nid, path=nid, extension=fp.suffix))

    graph.stats.ignored_files = len(scan.ignored)
    graph.nodes.sort(key=lambda n: n.id)
    node_set = {n.id for n in graph.nodes}

    parsed: list[tuple[str, Path, list[dict]]] = []
    targets: set[str] = set()
    for n in graph.nodes:
        fp = root_r / Path(n.path)
        try:
            text = fp.read_text(encoding="utf-8", errors="ignore")
        except OSError:
            continue
        links = extract_links(text, fp)
        parsed.append((n.id, fp, links))
        for lk in links:
            raw = lk["raw_link"]
            if not raw or raw.startswith("#") or raw.startswith(EXTERNAL_PREFIXES):
                continue
            rel_t, _frag, status = normalize_relative(root_r, fp, raw)
            if status == "resolved":
                targets.add(rel_t)

    # Paths excluded by .gitignore are not part of the repository, so links to
    # them are dropped from the graph instead of being reported as broken.
    ignored_targets = git_ignored(root_r, sorted(targets)) if gitignore else set()

    for source, fp, links in parsed:
        for lk in links:
            raw = lk["raw_link"]
            if not raw or raw.startswith("#") or raw.startswith(EXTERNAL_PREFIXES):
                continue
            rel_t, frag, status = normalize_relative(root_r, fp, raw)
            if status == "outside_root":
                graph.edges.append(
                    Edge(
                        source=source,
                        target=rel_t,
                        raw_link=raw,
                        line=lk["line"],
                        fragment=frag,
                        status="broken",
                    )
                )
                continue
            if rel_t in ignored_targets:
                continue
            tid = rel_t
            if tid not in node_set and not (root_r / tid).exists():
                graph.edges.append(
                    Edge(
                        source=source,
                        target=tid,
                        raw_link=raw,
                        line=lk["line"],
                        fragment=frag,
                        status="broken",
                    )
                )
                continue
            graph.edges.append(
                Edge(
                    source=source,
                    target=tid,
                    raw_link=raw,
                    line=lk["line"],
                    fragment=frag,
                    status="resolved",
                )
            )

    graph = analyze(graph)
    outdir.mkdir(parents=True, exist_ok=True)

    if "json" in formats:
        write_json(graph, outdir / f"{BASENAME}.json")
    if "md" in formats:
        write_report(graph, outdir / f"{BASENAME}.md")
    if "dot" in formats:
        write_dot(graph, outdir / f"{BASENAME}.dot", direction=direction)
    if "svg" in formats:
        write_svg(graph, outdir / f"{BASENAME}.svg", direction=direction)
    if "html" in formats:
        write_html(graph, outdir / f"{BASENAME}.html")

    if check and graph.stats.broken_links > 0:
        raise typer.Exit(code=1)
    raise typer.Exit(code=0)


@app.command()
def render(
    graph_json: Path = typer.Argument(..., exists=True, file_okay=True),
    format: str = typer.Option("svg", help="Output format: svg, dot, html"),
    direction: str = typer.Option("TB"),
    title: Optional[str] = typer.Option(None),
):
    import json

    data = json.loads(graph_json.read_text())
    from .models import Graph

    g = Graph(**data)
    outdir = graph_json.parent
    if format == "dot":
        write_dot(g, outdir / f"{BASENAME}.dot", direction=direction, title=title)
    elif format == "svg":
        write_svg(g, outdir / f"{BASENAME}.svg", direction=direction, title=title)
    elif format == "html":
        write_html(g, outdir / f"{BASENAME}.html", title=title)
    else:
        typer.echo(f"Unsupported format: {format}", err=True)
        raise typer.Exit(code=1)


@app.command()
def report(graph_json: Path = typer.Argument(..., exists=True, file_okay=True)):
    import json

    data = json.loads(graph_json.read_text())
    from .models import Graph

    g = Graph(**data)
    write_report(g, graph_json.parent / f"{BASENAME}.md")


def main():
    app()


if __name__ == "__main__":
    main()
