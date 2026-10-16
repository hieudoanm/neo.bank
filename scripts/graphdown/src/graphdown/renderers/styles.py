from __future__ import annotations

from ..models import Edge, Graph

# Light palette used by the DOT output, which is normally rendered on white.
NODE_COLOR = "#fff2a8"
MISSING_NODE_COLOR = "#f5b942"
BROKEN_EDGE_COLOR = "#cc0000"
SELF_EDGE_COLOR = "#999999"
EDGE_COLOR = "#333333"

# Obsidian-inspired dark palette shared by the SVG and HTML outputs.
BACKGROUND = "#20202b"
ACCENT = "#a78bfa"
ACCENT_HIGHLIGHT = "#c4b5fd"
LINK_COLOR = "#625879"
TEXT_COLOR = "#d8d1e8"
MISSING_COLOR = "#7c6f9e"
BROKEN_COLOR = "#c2607a"


def edge_color(
    edge: Edge,
    *,
    resolved: str = EDGE_COLOR,
    broken: str = BROKEN_EDGE_COLOR,
    self_link: str = SELF_EDGE_COLOR,
) -> str:
    if edge.status != "resolved":
        return broken
    if edge.source == edge.target:
        return self_link
    return resolved


def edge_colors(
    graph: Graph,
    *,
    resolved: str = EDGE_COLOR,
    broken: str = BROKEN_EDGE_COLOR,
    self_link: str = SELF_EDGE_COLOR,
) -> dict[tuple[str, str], str]:
    """Map every source/target pair to one colour, preferring broken."""
    colors: dict[tuple[str, str], str] = {}
    for edge in graph.edges:
        key = (edge.source, edge.target)
        if colors.get(key) != broken:
            colors[key] = edge_color(
                edge, resolved=resolved, broken=broken, self_link=self_link
            )
    return colors


def placeholder_colors(
    graph: Graph,
    missing: str = MISSING_NODE_COLOR,
    broken: str = BROKEN_EDGE_COLOR,
) -> dict[str, str]:
    """Colour endpoints that are not nodes in the model.

    Links to directories and missing files still resolve to an edge, but no
    node exists for them; renderers that refuse unknown endpoints need one.
    """
    node_ids = {node.id for node in graph.nodes}
    colors: dict[str, str] = {}
    for edge in graph.edges:
        for endpoint in (edge.source, edge.target):
            if endpoint in node_ids:
                continue
            color = broken if edge.status != "resolved" else missing
            if colors.get(endpoint) != broken:
                colors[endpoint] = color
    return colors
