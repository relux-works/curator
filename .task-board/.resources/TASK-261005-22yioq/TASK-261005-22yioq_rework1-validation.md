# TASK-261005-22yioq rework 1 — validation

Answers review-verdict-rev1 findings [P2]-1 (research links) and [P2]-2 (table pipes). Scope check: worktree diff vs rev1 candidate tree fc65f0a shows ONLY the 5 evidence lines (one per CIP) and 4 pipe escapes in CIP-0002 (lines 77, 80, 81, 82); README and all other lines byte-identical.

## Fixes
- Finding 1: all 5 CIPs now carry 2 commit-pinned Markdown links each (source draft + _evidence.md companion) to curator @ fae2ff9cab17a031c26a2b4c776afd8dc5e8b4f6; 10/10 links present. All 10 files verified present at fae2ff9c and byte-identical to 3d9aa987 (sha256 SAME x10). Stated short pin updated 3d9aa987 -> fae2ff9c to match link targets.
- Finding 2: CIP-0002 operator-control table enums escaped (`off \| admitted`, `enabled \| disabled`, `deny \| review`, `off \| private`).

## Commands and exit codes (all run in worktree, python via project .venv at curator-spec/.venv)
- `.venv/bin/python -B tools/validate.py` → EXIT 0 — validated 73 schemas and 1294 vector files
- `python3 -B /tmp/cip-table-regression.py` (named regression test: GFM width scan over CIP-0002..0006, 29 tables / 193 rows + 4-enum assertion) → EXIT 0 — inconsistent=0, enums=4/4 escaped
- `.venv/bin/python -B /tmp/cip-table-render.py` (markdown-it-py gfm-like render of operator-control table) → EXIT 0 — rows=7/7 two-cell, semantics=4/4
- Narrowing mutant (unescape exactly line-80 pipe in /tmp copy) → probe flags exactly line 80 (1/1); delete-variant (restore all 4 defects) → flags exactly 77,80,81,82 (4/4); `python3 -B /tmp/cip-mutant.py` → EXIT 0
- Research URL reachability: 10/10 HTTP 200 after redirects (curl, per-URL --max-time 30)
- Privacy grep over 5 CIPs + README (personal paths, ~user, private hosts, administrator, employer/ReluxWorks): 0 matches on all 5 patterns
- `git diff --check` → EXIT 0

## Not run
- Go tests / full `make validate`: docs-only change; reviewer instruction for this task says do not run go tests. Python unittest suite (672 tests, ~8 min) not rerun: no Python code changed since rev1 green.
