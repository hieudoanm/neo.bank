from __future__ import annotations

import json
from pathlib import Path
from .models import Graph


def graph_to_json(graph: Graph) -> str:
    data = graph.model_dump(mode="json")
    # deterministic sort
    data["nodes"].sort(key=lambda x: x["id"])
    data["edges"].sort(key=lambda x: (x["source"], x["target"], x["line"]))
    return json.dumps(data, indent=2, ensure_ascii=False) + "\n"


def write_json(graph: Graph, path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        f.write(graph_to_json(graph))
