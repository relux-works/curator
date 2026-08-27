# Verify integrated csk schema v6 implementation

## Description
Perform the final clean integration verification for the independent Python schema 6 and go-v1 implementation and attach reproducible evidence mapped to every story criterion and protocol vector cluster.

## Scope
Work from the landed dependency handoffs and the immutable curator-spec rc.4 conformance root. This is a verification task, not a new design task. Run the repository-supported full test, strict typing, packaging metadata, cross-platform CI, and conformance commands from a clean task worktree. Make only narrow integration corrections; route substantive defects to the owning task. Record that cocoaskills has no separate lint command unless one is added before execution, and run any declared lint gate that exists at that time.

## Acceptance Criteria
python -m pytest -v passes with the rc.4 conformance root and no unexpected skips; python -m mypy passes in strict mode; python -m build and python -m twine check dist/* pass; git diff --check passes; and the Linux, macOS, and Windows CI matrix with the real Go fixture is green. A task-scoped report records exact cocoaskills and curator-spec SHAs, commands, logs, platform evidence, and a matrix mapping every STORY-260720-1uv5gi acceptance criterion plus every required rejection, lifecycle, cache, dry-run, rollback, status, and launch case to passing evidence. No package-produced code runs during install verification and no release evidence is fabricated.
