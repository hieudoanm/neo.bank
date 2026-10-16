from __future__ import annotations

#: base name shared by every artefact written by ``graphdown scan``
BASENAME = "graphdown"

#: output formats produced by ``graphdown scan``
OUTPUT_EXTENSIONS = ("json", "dot", "md", "svg", "html")

#: default output directory name, created inside the scan root
OUTPUT_DIRNAME = ".graphdown"

#: output artefacts produced by ``graphdown scan``; never part of the graph
OUTPUT_NAMES = frozenset(f"{BASENAME}.{ext}" for ext in OUTPUT_EXTENSIONS)
