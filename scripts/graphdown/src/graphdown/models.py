from __future__ import annotations

from pydantic import BaseModel, Field
from typing import Literal, Optional


class Node(BaseModel):
    id: str
    path: str
    extension: str


class Edge(BaseModel):
    source: str
    target: str
    raw_link: str
    line: int
    fragment: Optional[str] = None
    status: Literal["resolved", "broken", "outside_root"] = "resolved"


class Stats(BaseModel):
    files: int = 0
    ignored_files: int = 0
    links: int = 0
    broken_links: int = 0
    resolved_links: int = 0
    self_links: int = 0
    orphan_files: int = 0
    components: int = 0


class Graph(BaseModel):
    version: int = 1
    root: str = "."
    nodes: list[Node] = Field(default_factory=list)
    edges: list[Edge] = Field(default_factory=list)
    stats: Stats = Field(default_factory=Stats)
