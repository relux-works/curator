# TASK-260924-m28s6b — Skillfile lock replay on a fresh machine (THE ONLY CURRENT INSTRUCTION)

Control root: curator; your Story worktree (STORY-260924-3eywt2; it carries the checkpointed default-on leaf TASK-260924-1aa9wb — schema 2
is the default reader path, no switch). Read `campaign-producer-rules.md` (CHANGELOG policy: do NOT edit CHANGELOG.md; put the entry in
the results resource; no repo files for results; artifacts only in $TMPDIR) and the binding
`skillfile-lock-replay-addendum-20260924.md` + `skillfile-default-on-handoff-20260924.md`. Landing gate = hosted CI at handoff.
Today (`internal/closure/resolve.go` ~300-414) a missing local snapshot with an existing lock fails `source_snapshot_unavailable`. Implement
the accepted §3 rule exactly: re-materialize from the declared source (git/repository: fetch EXACTLY the locked revision; path: read the
current bytes), accept only when package identity AND content_sha256 equal the lock, else `source_snapshot_changed`;
`source_snapshot_unavailable` only when the source cannot be reached; lock bytes unchanged; never re-resolve a tag/branch.
`Skillfile.lock.json` is a committed file: prove Curator never adds it to .gitignore / never treats it as machine-private; docs say to
commit it (like package-lock.json).
Rows (production entry, local bare repos / fixtures, no network): fresh machine (empty snapshot store) × {git tag, repository, path} through
install AND update with an existing lock; path content changed → source_snapshot_changed; unreachable source → source_snapshot_unavailable;
moved tag after locking → still the locked revision (no re-resolution); lock bytes identical before/after. Narrowing mutants in a
disposable copy (accept on identity only; re-resolve the tag; path accepted without hash check) → killed. Bounded runs. Attach results
(board resource), check DoD, `task-board handoff TASK-260924-m28s6b --role developer`. A `run_wrote_outside_worktree … policy warn` block
is a warning — verify status `to-review`.
