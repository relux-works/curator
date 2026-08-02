# BUG-260801-1iu1ln cycle-8 reviewer verdict

## Verdict

Changes requested. Route to implementation rework (to-dev). This is ordinary recoverable conformance-harness work, not a Stop-The-Line boundary.

Reviewed signed CocoaSkills commit 120be14d31e02ad6c734a3f1a3659d05880933cd on branch task/BUG-260801-1iu1ln-lifecycle-observed-traces in the dedicated clean worktree. HEAD and required base ba250bfc4dfe104a160eadd5b5f4e340693bf892 have valid ECDSA signatures, the exact merge base is the required base, the worktree is clean, no tag points at HEAD, and the submitted restricted surfaces remain unchanged.

## Acceptance-blocking finding

The new atomic-publication tree witness is fail-open for mutations of the moved root itself. tests/protocol_lifecycle_observations.py:182-199 replaces the staged root ctime with the final live root ctime unconditionally, while the publication wrapper at lines 1158-1204 compares only the staged tree with the final tree after the wrapped CocoaSkills seam returns. Captured os callables bypass the attribute observer, so root-only mutations can be restored while every compared field remains normative.

Two independent Darwin probes wrapped the real cache_posix._rename_noreplace seam and acted after each real publication:

1. A captured os.open plus os.fchmod changed and restored the live cache root mode. It executed 42 times, but the complete publish-complete-immutable-entry-under-home-lock case remained exactly equal to the normative vector and reported publication=atomic-complete-directory.
2. A captured os.rename transiently moved the live cache entry to a sibling name and moved it back, making the live name absent between calls. It executed 42 times, but the same complete case again remained exactly normative.

The first survives because the only persistent evidence is additional root ctime drift, which line 196 discards as if all root ctime change came from the legitimate rename. The second also changes the destination parent, but that parent is outside the staged-to-live tree comparison. This violates the required observed binding for atomic, complete, immutable publication.

## Positive verification

- Canonical lifecycle, exhaustive scalar mutation, literal/lossy classification, process-path, identity, rollback and unknown-field gate: 417 passed in 48.35s.
- Four submitted cycle-8 regressions: 4 passed in 131.90s.
- Strict mypy: no issues in 68 source files.
- compileall, exact-base diff check and clean worktree check: exit 0.
- Private-failure expansion covers project, config, manager-home and source roots around the exact private-build phase, excluding only stable lock coordination state; submitted Skillfile, io.open and ctime regressions pass.

## Required next handoff

1. Add fail-closed evidence for the live publication root and its destination parent across and after the no-replace handoff, without treating arbitrary final root ctime as legitimate rename evidence.
2. Retain exact regressions proving complete-case inequality for captured root fchmod restore and transient live-name removal/restore. Include dir-fd forms; cover equivalent supported Windows semantics where applicable.
3. Preserve the four cycle-8 regressions, the inherited sabotage set, canonical 32 cases and exhaustive scalar/classification gate.
4. Re-run full exact-root, related suites, strict mypy/compileall/diff/release/build/Twine/signature/clean-tree gates and produce a new signed commit/evidence artifact.

Reviewer modified no CocoaSkills code, PR/main state, tags, releases, pins, claims, schemas, CI, changelog or packaging metadata.