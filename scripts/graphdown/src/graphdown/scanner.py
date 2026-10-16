from __future__ import annotations

import os
import subprocess
from dataclasses import dataclass
from fnmatch import fnmatch
from pathlib import Path

from .constants import OUTPUT_NAMES


@dataclass(frozen=True)
class ScanResult:
    """Candidate files kept for the graph and paths dropped by `.gitignore`."""

    files: list[Path]
    ignored: list[str]


def git_ignored(root: Path, rel_paths: list[str]) -> set[str]:
    """Return the subset of ``rel_paths`` that git considers ignored.

    Paths are interpreted relative to ``root`` and honour every ``.gitignore``
    that applies to it (nested files, ``.git/info/exclude``, global excludes).
    If git is unavailable or ``root`` is outside a work tree, no path is
    reported as ignored.
    """
    if not rel_paths:
        return set()
    try:
        proc = subprocess.run(
            ["git", "-C", str(root), "check-ignore", "--stdin", "-z"],
            input="\0".join(rel_paths).encode("utf-8"),
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            check=False,
        )
    except (OSError, ValueError):
        return set()
    if proc.returncode not in (0, 1):  # 1: none of the given paths are ignored
        return set()
    return {p for p in proc.stdout.decode("utf-8", "replace").split("\0") if p}


def _included(rel: str, patterns: list[str]) -> bool:
    return any(fnmatch(rel, pattern) for pattern in patterns)


def _excluded(rel: str, patterns: list[str]) -> bool:
    """True when ``rel`` or one of its parent directories matches a pattern."""
    parts = rel.split("/")
    for i in range(len(parts)):
        prefix = "/".join(parts[: i + 1])
        if any(fnmatch(prefix, pattern) for pattern in patterns):
            return True
    return False


def scan_files(
    root: Path,
    include: list[str] | None = None,
    exclude: list[str] | None = None,
    extensions: list[str] | None = None,
    respect_gitignore: bool = True,
) -> ScanResult:
    """Discover candidate files below ``root``.

    Files ignored by git are skipped when ``respect_gitignore`` is true, which
    keeps vendored trees such as ``.venv/`` out of the graph. Output artefacts
    named ``graphdown.*`` are skipped as well. The returned :class:`ScanResult`
    carries both the kept files and the ignored paths that were dropped.
    """
    root = root.resolve()
    if extensions is None:
        extensions = [".md", ".mdx"]
    allowed = {ext.lower() for ext in extensions}

    found: dict[Path, str] = {}
    for dirpath, dirnames, filenames in os.walk(root, followlinks=False):
        for fname in filenames:
            if fname in OUTPUT_NAMES:
                continue
            fpath = Path(dirpath) / fname
            if fpath.suffix.lower() not in allowed:
                continue
            resolved = fpath.resolve()
            try:
                rel = resolved.relative_to(root)
            except ValueError:
                continue
            found[resolved] = rel.as_posix()

    ignored: list[str] = []
    if respect_gitignore:
        ignored = sorted(git_ignored(root, list(found.values())))
        drop = set(ignored)
        found = {p: rel for p, rel in found.items() if rel not in drop}

    if include:
        found = {p: rel for p, rel in found.items() if _included(rel, include)}
    if exclude:
        found = {p: rel for p, rel in found.items() if not _excluded(rel, exclude)}

    return ScanResult(files=sorted(found), ignored=ignored)
