# THE ONLY CURRENT INSTRUCTION — TASK-261007-2uet3u: curator v0.15.0-rc.4 release notes (developer, docs)
The operator approved the rc.4 tag (2026-10-07). The orchestrator tags main right after this lands. Write the consolidated `## v0.15.0-rc.4 — 2026-10-07` section of CHANGELOG.md.
**Source of truth:** `git log v0.15.0-rc.3..origin/main`. Read it in full; board-record commits ("Record … board state") carry no product change.
**Do:**
1. Turn `## Unreleased` into the rc.4 section, and leave a fresh empty `## Unreleased` above it. Group the entries as Added / Changed / Fixed / Security / Known issues, in the style of the rc.3 section.
2. Reconcile against the history. Include everything that landed after rc.3, at least:
   - the atomic v1→v2 profile and identity migration plus the v2 writer flip (TASK-261003-1uzji7): content identities are now WRITTEN as v2, and v1 stays readable;
   - the opaque NUL gate scoped to v1 identities (TASK-261006-3ptm2x), with its operator actions: v2 re-pin, re-issuing bare v1 revocations from the v2 digest, and versioned verdict-cache and pin carriers;
   - the 2026-10 inline audit fixes N1–N5 (issue #106): the gc fix, the expanded-snapshot budget, the restore mode, the symlink-backup restore, and "a trust pin no longer waives strict findings";
   - Go 1.26/1.27 qualification in the go-v1 driver.
   Docs-only research (CIP drafts in .research/) goes under a short "Documentation" line, or is omitted. Never claim anything that is not in the history.
3. **Known issues**, stated plainly:
   - the Windows broker real-Git flake (TASK-260930-fp8vx7) is still a documented risk;
   - the security-posture revision B and Codex seed revision B flips are NOT in rc.4; they are held for a later release;
   - the board-close gap for the Go qualification task is a tooling issue only (spm#537); no product impact.
4. Change ONLY CHANGELOG.md. Never edit LOGBOOK.md. There are no version constants (GoReleaser ldflags set the version).
**Evidence:** in the outcome, list the commit subjects you reconciled against and map each entry to its commits.
Then `task-board handoff TASK-261007-2uet3u --role developer` and END YOUR TURN.
