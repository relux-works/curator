#!/usr/bin/env python3
"""Regression probe for TASK-261005-22yioq rework 1, finding 2.

Scans GFM tables in the five published CIPs and fails when any body row has
more unescaped pipes than the table header declares (GFM treats `|` as a cell
separator even inside backticks). Also asserts the four fixed enum cells in
CIP-0002 render as single cells containing a literal pipe.
"""
import re
import sys
from pathlib import Path

CIPS = sorted(Path("cips").glob("CIP-000[2-6]-*.md"))
assert len(CIPS) == 5, f"expected 5 CIPs, found {len(CIPS)}"

# Split a GFM table row on unescaped pipes (backslash-escapes don't split).
CELL_SPLIT = re.compile(r"(?<!\\)\|")


def tables_of(path):
    """Yield (start_line, header_cells, [(lineno, row_cells)]) per table."""
    lines = path.read_text().splitlines()
    i = 0
    while i < len(lines):
        if lines[i].lstrip().startswith("|") and i + 1 < len(lines) and re.match(
            r"^\s*\|?[\s:|-]+\|[\s:|-]*\|?\s*$", lines[i + 1]
        ):
            header = [c for c in CELL_SPLIT.split(lines[i].strip().strip("|"))]
            rows = []
            j = i + 2
            while j < len(lines) and lines[j].lstrip().startswith("|"):
                rows.append((j + 1, CELL_SPLIT.split(lines[j].strip().strip("|"))))
                j += 1
            yield (i + 1, header, rows)
            i = j
        else:
            i += 1


def main():
    bad = []
    n_tables = n_rows = 0
    for path in CIPS:
        for start, header, rows in tables_of(path):
            n_tables += 1
            for lineno, cells in rows:
                n_rows += 1
                if len(cells) != len(header):
                    bad.append(f"{path.name}:{lineno}: {len(cells)} cells, header has {len(header)}")
    print(f"tables={n_tables} rows={n_rows} inconsistent={len(bad)}")
    for b in bad:
        print("BAD", b)

    # Narrow assertion: the four repaired enum cells each hold a literal pipe.
    text = Path("cips/CIP-0002-project-context-in-managed-launches.md").read_text()
    for enum in ("off \\| admitted", "enabled \\| disabled", "deny \\| review", "off \\| private"):
        assert f"`{enum}`" in text, f"missing escaped enum `{enum}`"
    print("enums=4/4 escaped")

    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
