from __future__ import annotations

import json
from pathlib import Path

from pyvis.network import Network

from ..models import Graph
from .styles import (
    ACCENT,
    ACCENT_HIGHLIGHT,
    BACKGROUND,
    BROKEN_COLOR,
    LINK_COLOR,
    MISSING_COLOR,
    TEXT_COLOR,
    placeholder_colors,
)

#: node radius grows with the number of connections, like Obsidian
_MIN_SIZE = 5
_MAX_SIZE = 20
_SIZE_STEP = 1.5

#: force-directed layout; the fixed seed keeps repeat runs reproducible
_OPTIONS = json.dumps(
    {
        "layout": {"randomSeed": 42},
        "nodes": {
            "shape": "dot",
            "borderWidth": 0,
            "borderWidthSelected": 0,
            "color": {
                "background": ACCENT,
                "border": ACCENT,
                "highlight": {
                    "background": ACCENT_HIGHLIGHT,
                    "border": ACCENT_HIGHLIGHT,
                },
                "hover": {"background": ACCENT_HIGHLIGHT, "border": ACCENT_HIGHLIGHT},
            },
            "font": {
                "color": TEXT_COLOR,
                "size": 14,
                "face": "sans-serif",
                "strokeWidth": 0,
            },
            "scaling": {"min": _MIN_SIZE, "max": _MAX_SIZE},
        },
        "edges": {
            "color": {
                "color": LINK_COLOR,
                "highlight": ACCENT,
                "hover": ACCENT_HIGHLIGHT,
                "inherit": False,
            },
            "width": 1,
            "selectionWidth": 1,
            "hoverWidth": 1,
            "smooth": {"enabled": False},
        },
        "physics": {
            "enabled": True,
            "solver": "forceAtlas2Based",
            "forceAtlas2Based": {
                "gravitationalConstant": -45,
                "centralGravity": 0.01,
                "springLength": 100,
                "springConstant": 0.08,
                "damping": 0.4,
                "avoidOverlap": 0.5,
            },
            "minVelocity": 0.75,
            "stabilization": {"enabled": True, "iterations": 1000, "fit": True},
        },
        "interaction": {
            "hover": True,
            "tooltipDelay": 150,
            "hideEdgesOnDrag": False,
            "hideNodesOnDrag": False,
            "navigationButtons": False,
            "keyboard": True,
        },
    }
)


def _size(connections: int) -> float:
    return min(_MIN_SIZE + connections * _SIZE_STEP, _MAX_SIZE)


def _theme(html: str) -> str:
    """Apply the dark page chrome PyVis's template cannot express.

    Only the canvas colour is configurable through ``Network``, so the white
    page background and light border are patched in afterwards. Both calls are
    safe no-ops if a future template changes shape.
    """
    page = (
        "<style>"
        f"body {{ background-color: {BACKGROUND}; color: {TEXT_COLOR}; margin: 0; }}"
        f"h1 {{ color: {TEXT_COLOR}; }}"
        "</style></head>"
    )
    return (
        html.replace("</head>", page, 1)
        .replace("border: 1px solid lightgray;", "border: none;", 1)
        .replace("<h1></h1>", "", 2)
    )


def _color_spec(color: str) -> dict[str, object]:
    """Turn a single colour into a full vis colour object."""
    return {
        "background": color,
        "border": color,
        "highlight": {"background": ACCENT_HIGHLIGHT, "border": ACCENT_HIGHLIGHT},
        "hover": {"background": ACCENT_HIGHLIGHT, "border": ACCENT_HIGHLIGHT},
    }


def _connections(graph: Graph) -> dict[str, int]:
    """Count unique neighbours per endpoint, treating links as undirected.

    Obsidian's graph view is undirected, so reciprocal links count once.
    """
    counts: dict[str, int] = {node.id: 0 for node in graph.nodes}
    seen: set[tuple[str, str]] = set()
    for edge in graph.edges:
        if edge.source == edge.target:
            continue
        pair = (
            (edge.source, edge.target)
            if edge.source <= edge.target
            else (edge.target, edge.source)
        )
        if pair in seen:
            continue
        seen.add(pair)
        counts[pair[0]] = counts.get(pair[0], 0) + 1
        counts[pair[1]] = counts.get(pair[1], 0) + 1
    return counts


def write_html(graph: Graph, path: Path, title: str | None = None) -> None:
    """Write a self-contained interactive HTML graph styled like Obsidian.

    Consumes the canonical graph model, so it can be regenerated from
    ``graphdown.json`` without rescanning the source directory. Nodes are sized
    by the number of connections and links are drawn without arrowheads to match
    Obsidian's undirected graph view.
    """
    # vis-network needs an explicit node for every edge endpoint; keep the
    # Obsidian palette there too (muted for directories, rose for broken links)
    placeholders = placeholder_colors(graph, missing=MISSING_COLOR, broken=BROKEN_COLOR)
    connections = _connections(graph)

    net = Network(
        height="100vh",
        width="100%",
        directed=False,
        bgcolor=BACKGROUND,
        font_color=TEXT_COLOR,
        heading=title or "",
        cdn_resources="in_line",
        select_menu=False,
        filter_menu=False,
    )
    for node in sorted(graph.nodes, key=lambda n: n.id):
        degree = connections.get(node.id, 0)
        net.add_node(
            node.id,
            label=Path(node.id).stem,
            title=f"{node.id}\nConnections: {degree}",
            size=_size(degree),
        )
    # drop pyvis's per-node default colour so the Obsidian palette in
    # options.nodes is what the browser actually uses
    for node_options in net.nodes:
        node_options.pop("color", None)
    for nid, color in sorted(placeholders.items()):
        degree = connections.get(nid, 0)
        net.add_node(
            nid,
            label=Path(nid).stem,
            title=f"{nid}\nConnections: {degree}",
            size=_size(degree),
            color=_color_spec(color),
        )
    seen: set[tuple[str, str]] = set()
    for edge in sorted(graph.edges, key=lambda e: (e.source, e.target, e.raw_link)):
        pair = (
            (edge.source, edge.target)
            if edge.source <= edge.target
            else (edge.target, edge.source)
        )
        if pair in seen:  # undirected view: draw reciprocal links once
            continue
        seen.add(pair)
        # only override the colour for broken links; healthy edges inherit the
        # palette (and its highlight/hover colours) from options.edges
        options: dict[str, object] = {"title": edge.raw_link}
        if edge.status != "resolved":
            options["color"] = BROKEN_COLOR
        net.add_edge(edge.source, edge.target, **options)

    net.set_options(_OPTIONS)
    html = _theme(net.generate_html(notebook=False))
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(html, encoding="utf-8")
