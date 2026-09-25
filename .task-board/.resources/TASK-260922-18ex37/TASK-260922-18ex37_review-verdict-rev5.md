# TASK-260922-18ex37 review verdict — revision 5: CHANGES REQUESTED

Candidate tree 51a665cc on base ab34556e (reviewer: claude-opus-5-5 low).

## Finding F1 (BLOCKING, only finding)
`.github/workflows/ci.yml.merged.tmp` is in the candidate (untracked file in worktree picked up by the snapshot). It is a stale merge artefact that differs from the real ci.yml (25+/23-). Fix: `rm .github/workflows/ci.yml.merged.tmp` in the worktree and re-handoff. No other change needed. repeat-of: none (new mechanism — refresh merge scratch file).

## Verified (hold for rev6)
- Trunk overlap since c278af4f: only 07878da8 (3v8k23) touches candidate paths (gate-selftest.sh, ci.yml, docs/ci-gates.md). Every added line of 07878da8 in those three files is present verbatim in the candidate: 181/181, 21/21, 10/10 lines, 0 missing. 5dddbb57 touches none of the candidate paths.
- SPEC_PIN = dcc7f015e2d97edf2d52928afb6fd79ec8129e8b in ci.yml; no dced9b8 left in .github; no 5p8b0z row in .github; CHANGELOG.md identical to trunk; no root TASK-*/BUG-*, .temp/, test/, ledger/ entries.
- Hosted gate run 36186939097: success on all lanes (Test ubuntu/macos/windows, Race ubuntu/macos, Gate self-test x3, Lint, Naming, Interop conformance).
- Content otherwise per rev4 acceptance (not re-derived; rev4 ACCEPTED on content).
