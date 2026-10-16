from __future__ import annotations

from markdown_it import MarkdownIt
from pathlib import Path


def extract_links(text: str, source_path: Path) -> list[dict]:
    md = MarkdownIt()
    tokens = md.parse(text, {})
    links: list[dict] = []

    def walk(toks):
        for tok in toks:
            if tok.type == "link_open":
                href = tok.attrGet("href") or ""
                line = tok.map[0] + 1 if tok.map else 0
                links.append({"raw_link": href, "line": line})
            if tok.children:
                walk(tok.children)

    walk(tokens)
    return links
