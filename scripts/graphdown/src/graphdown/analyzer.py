from __future__ import annotations

from collections import defaultdict, deque


def analyze(graph):
    stats = graph.stats
    stats.files = len(graph.nodes)
    stats.links = len(graph.edges)

    broken = 0
    resolved = 0
    selfl = 0
    for e in graph.edges:
        if e.status == "broken":
            broken += 1
        elif e.status == "resolved":
            resolved += 1
        if e.source == e.target:
            selfl += 1

    stats.broken_links = broken
    stats.resolved_links = resolved
    stats.self_links = selfl

    # orphans: nodes with no incoming resolved links from other nodes?
    incoming = defaultdict(int)
    for e in graph.edges:
        if e.status == "resolved":
            incoming[e.target] += 1
    orphans = 0
    node_ids = {n.id for n in graph.nodes}
    for n in graph.nodes:
        if incoming.get(n.id, 0) == 0:
            # check if any edge comes into it? also consider if it's only referenced from outside? treat as orphan if no incoming
            orphans += 1
    stats.orphan_files = orphans
    stats.components = 0
    return graph
