from __future__ import annotations

from pathlib import Path
from urllib.parse import unquote


def normalize_relative(
    root: Path, source: Path, target: str
) -> tuple[str, str | None, str]:
    # strip fragment/query for filesystem
    frag = None
    qmark = None
    # simple split
    if "#" in target:
        target, frag = target.split("#", 1)
    if "?" in target:
        target, qmark = target.split("?", 1)  # not used in id, but preserve
    if target.startswith("mailto:") or target.startswith("tel:"):
        return target, frag, "external"
    if target.startswith(("http://", "https://", "ftp://")):
        return target, frag, "external"
    # resolve
    src_dir = source.parent if source.is_file() else source
    cand = (src_dir / unquote(target)).resolve()
    root_r = root.resolve()
    try:
        rel = cand.relative_to(root_r)
        return rel.as_posix(), frag, "resolved"
    except Exception:
        return target, frag, "outside_root"
