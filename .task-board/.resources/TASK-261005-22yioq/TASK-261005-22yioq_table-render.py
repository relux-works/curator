#!/usr/bin/env python3
"""GFM render check for TASK-261005-22yioq rework 1, finding 2.

Renders the CIP-0002 operator-control table with markdown-it-py gfm-like and
asserts each of the four repaired rows has exactly 2 cells and the second cell
keeps the full enum + trailing semantics (i.e. no text was split into a
dropped third cell).
"""
import sys
from pathlib import Path

from markdown_it import MarkdownIt

text = Path("cips/CIP-0002-project-context-in-managed-launches.md").read_text()
start = text.index("| Control | Semantics and default |")
block = text[start:text.index("Existing launches remain", start)]
html = MarkdownIt("gfm-like").render(block)

import re
rows = re.findall(r"<tr>(.*?)</tr>", html, re.S)
body = rows[1:]  # skip header
assert len(body) == 7, f"expected 7 body rows, got {len(body)}"
needles = [
    ("off | admitted", "New registrations default"),
    ("enabled | disabled", "Exact item/closure approval"),
    ("deny | review", "Fleet ceilings cannot be overridden"),
    ("off | private", "separate from approved repository knowledge"),
]
checked = 0
for row in body:
    cells = re.findall(r"<td>(.*?)</td>", row, re.S)
    assert len(cells) == 2, f"row split into {len(cells)} cells: {row[:120]}"
    plain = re.sub(r"<[^>]+>", "", cells[1])
    for enum, tail in needles:
        if enum in plain:
            assert tail in plain, f"tail lost for {enum}: {plain[:120]}"
            checked += 1
assert checked == 4, f"only {checked}/4 repaired semantics found"
print(f"rows=7/7 two-cell semantics=4/4")
