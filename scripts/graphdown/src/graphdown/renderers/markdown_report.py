from __future__ import annotations

from pathlib import Path
from ..models import Graph


def write_report(graph: Graph, path: Path) -> None:
    lines = []
    lines.append("# Graphdown Report")
    lines.append("")
    lines.append(f"- Files: {graph.stats.files}")
    lines.append(f"- Ignored files: {graph.stats.ignored_files}")
    lines.append(f"- Links: {graph.stats.links}")
    lines.append(f"- Resolved links: {graph.stats.resolved_links}")
    lines.append(f"- Broken links: {graph.stats.broken_links}")
    lines.append(f"- Self links: {graph.stats.self_links}")
    lines.append(f"- Orphan files: {graph.stats.orphan_files}")
    lines.append("")
    if graph.stats.broken_links > 0:
        lines.append("## Broken Links")
        for e in graph.edges:
            if e.status == "broken":
                lines.append(f"- [{e.source}] -> `{e.raw_link}` (line {e.line})")
        lines.append("")
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")
