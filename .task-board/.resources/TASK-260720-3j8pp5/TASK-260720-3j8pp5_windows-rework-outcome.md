# TASK-260720-3j8pp5 Windows rework outcome

Date: 2026-07-30
Owner: orchestrator takeover after two nonproductive producer runs

## Scope

- `src/csk/builds/toolchain.py`
- `tests/test_builds_toolchain.py`

The rework replaces `DirEntry.stat(follow_symlinks=False)` with a fresh
`os.lstat()` by native path after collecting entry names. This avoids the
Windows `DirEntry` metadata snapshot mismatch seen in GitHub Actions while
preserving descriptor/path identity comparison and mutation detection.

The subprocess stdin assertion now accepts the native Windows CRLF emitted by
text-mode Python. A deterministic regression replaces `os.scandir()` entries
with entries whose `stat()` always raises, proving the scanner obtains physical
identity through `os.lstat()` instead.

## Provenance

- Worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3j8pp5/worktree`
- Branch: `task/TASK-260720-3j8pp5-toolchain-identity`
- Base/HEAD before rework:
  `d5d16bfcaa2fe43dc994b819c2659512c4fd8f0a`
- Product diff is exactly the two files listed above.
- `RUN-260730-f34eae` and `RUN-260730-7be130` were cancelled after producing
  no usable product changes; the orchestrator completed the narrow rework.

## Tool readiness

- task-board `0.23.0`
- git `2.50.1`
- uv `0.11.3`
- gh `2.83.2`

## Validation

- Focused pytest: `63 passed`
- Full pytest with both fixture roots: `768 passed, 1 skipped`
- Strict mypy: no issues in `58 source files`
- Package build: wheel and sdist succeeded
- Twine check: wheel and sdist passed
- `git diff --check`: clean
- Ruff safety lint (`E4,E7,E9,F`): passed

The repository-wide Ruff format/full-rule baseline contains unrelated existing
drift in these files. No bulk formatting or unrelated lint cleanup was applied.
The exact committed candidate still requires the GitHub macOS/Ubuntu/Windows
matrix and an independent exact-byte review before landing.
