from __future__ import annotations

from pathlib import Path
from ..models import Graph
from .styles import NODE_COLOR, edge_color, placeholder_colors


def _esc(s: str) -> str:
    return s.replace("\\", "\\\\").replace('"', '\\"')


def to_dot(graph: Graph, direction: str = "TB", title: str | None = None) -> str:
    lines = []
    lines.append("digraph graphdown {")
    lines.append("  rankdir=%s;" % direction)
    lines.append("  splines=polyline;")
    lines.append("  nodesep=0.5;")
    lines.append("  ranksep=0.5;")
    if title:
        lines.append(f'  label="{_esc(title)}";')
    for n in graph.nodes:
        style = f'shape=note, style=filled, fillcolor="{NODE_COLOR}"'
        lines.append(f'  "{_esc(n.id)}" [label="{_esc(Path(n.id).name)}", {style}];')
    for nid, color in sorted(placeholder_colors(graph).items()):
        style = f'shape=ellipse, style=filled, fillcolor="{color}"'
        lines.append(f'  "{_esc(nid)}" [label="{_esc(Path(nid).name)}", {style}];')
    for e in graph.edges:
        lines.append(
            f'  "{_esc(e.source)}" -> "{_esc(e.target)}"'
            f' [color="{edge_color(e)}", arrowsize=0.6];'
        )
    lines.append("}")
    return "\n".join(lines) + "\n"


def write_dot(
    graph: Graph, path: Path, direction: str = "TB", title: str | None = None
) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        f.write(to_dot(graph, direction, title))
