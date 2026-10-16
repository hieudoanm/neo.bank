from __future__ import annotations

from pathlib import Path

import matplotlib

matplotlib.use("Agg")
matplotlib.rcParams["svg.hashsalt"] = "graphdown"

import matplotlib.pyplot as plt
import networkx as nx

from ..models import Graph
from .styles import (
    ACCENT,
    BACKGROUND,
    BROKEN_COLOR,
    LINK_COLOR,
    MISSING_COLOR,
    TEXT_COLOR,
    edge_colors,
    placeholder_colors,
)

#: fixed seed so repeated runs lay the graph out identically
_SEED = 42


def to_networkx(graph: Graph) -> nx.DiGraph:
    """Build a directed NetworkX graph from the canonical model."""
    g = nx.DiGraph()
    for node in graph.nodes:
        g.add_node(node.id)
    for edge in graph.edges:
        g.add_edge(edge.source, edge.target)
    return g


def write_svg(
    graph: Graph, path: Path, direction: str = "TB", title: str | None = None
) -> None:
    """Render the graph to SVG with matplotlib and NetworkX.

    Consumes the canonical graph model, so it can be regenerated from
    ``graphdown.json`` without rescanning the source directory. Uses the same
    Obsidian-inspired dark palette as the HTML output.
    """
    g = to_networkx(graph)
    positions = nx.spring_layout(g, seed=_SEED)
    if direction.upper() == "LR":
        positions = {nid: (y, x) for nid, (x, y) in positions.items()}
    labels = {nid: Path(nid).stem for nid in g.nodes}

    placeholders = placeholder_colors(graph, missing=MISSING_COLOR, broken=BROKEN_COLOR)
    pair_colors = edge_colors(graph, resolved=LINK_COLOR, broken=BROKEN_COLOR)

    fig, ax = plt.subplots(figsize=(12, 8))
    fig.patch.set_facecolor(BACKGROUND)
    ax.set_facecolor(BACKGROUND)
    nx.draw_networkx(
        g,
        positions,
        ax=ax,
        node_size=500,
        font_size=8,
        font_color=TEXT_COLOR,
        arrows=True,
        arrowsize=8,
        edgecolors="none",
        with_labels=True,
        labels=labels,
        node_color=[placeholders.get(nid, ACCENT) for nid in g.nodes],
        edge_color=[pair_colors.get((u, v), LINK_COLOR) for u, v in g.edges],
    )
    if title:
        ax.set_title(title, color=TEXT_COLOR)
    ax.axis("off")

    path.parent.mkdir(parents=True, exist_ok=True)
    # omit the creation timestamp so repeat runs are byte-identical
    fig.savefig(
        path,
        format="svg",
        bbox_inches="tight",
        facecolor=BACKGROUND,
        metadata={"Date": None},
    )
    plt.close(fig)
