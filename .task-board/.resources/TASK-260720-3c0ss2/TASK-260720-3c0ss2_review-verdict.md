# TASK-260720-3c0ss2 — review verdict

## Verdict

**Changes requested → `to-dev`.**

The exact `curator-build-source-v1` framing, frozen snapshot recheck boundary,
root-marker behavior, and fresh-install context selection are correct, but the
installed-tree up-to-date branch does not enforce build-root isolation. This is
ordinary implementation rework, not a stop-the-line boundary.

The reviewer run is not goal-bound (`task-board spawn goal
RUN-260729-d29f5f` reported `Active Goal: none`). No `commit_ack` was supplied.

## Blocking finding

### P1 — A marker-consistent installed build root is accepted as up-to-date

Candidate locations:

- `src/csk/installer.py:765-769` returns `up-to-date` before
  `whitelist.copy_context` can apply `plan.spec.build_roots`.
- `src/csk/installer.py:923-953` validates marker metadata and compares the
  installed tree against the marker's own marker-excluding content hash, but
  never proves that declared build roots are absent from either the physical
  tree or the marker `files` list.

Reproduction:

1. Install a schema-6 skill with build root `assets/build-tool`.
2. Seed the installed context and marker with the build-root files and
   recomputed `content_sha256`, matching what the pre-exclusion installer at
   this task's schema-6 base could write.
3. Run install again with the candidate.

Observed:

```text
second_errors=[]
... skill-build ... up-to-date
stale_build_root_exists=True
stale_build_file_in_marker=True
```

This violates the acceptance requirement that declared build roots remain
statically invisible on cache/up-to-date paths. It also contradicts the
producer outcome's claim that the up-to-date path refreshes context.

Required rework:

1. Make context currentness fail closed when any declared build-root subtree is
   present in the installed context. Reject stale marker `files` entries below
   a build root as well, so the next real install rebuilds a sanitized tree.
2. Add a regression that seeds a marker-consistent pre-exclusion schema-6
   installed tree, then proves the candidate does not report it up-to-date and
   removes the build-root subtree.
3. Re-run focused rc.5/task pytest, strict mypy, full pytest, and diff hygiene.

## Positive evidence

- Base/worktree provenance is sound: clean local `main` and `origin/main` are
  both `dd76b570f88339fd1d659c02950e68b17f6ba834`; reflog records the
  fast-forward before worktree creation. Dependency `TASK-260720-z9j4c9` is
  accepted/done at that base.
- The authoritative rc.5 full fixture matches exactly through both code paths:
  in-memory and filesystem snapshot digests are
  `sha256:27cdcac0734aa3e069e95a10341e89b118a07c60002516e7b401e95477f01332`;
  selected context is exactly `["SKILL.md", "assets/prompt.md"]`.
- Domain prefix, unsigned UTF-8 ordering, `F` records, uint64 big-endian
  lengths, binary/empty files, structural collision separation, invalid and
  duplicate paths, platform collisions, links, special files, root marker,
  mode/mtime non-inputs, and pre/post-callback mutation checks are implemented
  and covered.
- Legacy `content_sha256` has no body change; the added build-source algorithm
  is separate.
- Independent focused gate: **151 passed**, exit 0.
- Independent strict mypy: **57 source files clean**, exit 0.
- Independent full gate with `CURATOR_CONFORMANCE_ROOT` and
  `CURATOR_SCHEMA_V6_ROOT`: **702 passed, 1 skipped**, exit 0.
- `git diff --check` exited 0. Tests did not alter candidate source.

## Evidence anomaly

`TASK-260720-3c0ss2_results.md` transcribes two shared-vector digests
incorrectly. The authoritative values, already used correctly by the candidate
tests and implementation, are:

- empty snapshot:
  `sha256:3a518980ed122b2139e46152d9c4dda7426a42572f3235cde8cbe781566f5753`
- binary/empty/root-marker vector:
  `sha256:68008c9a1131c1295d78f4f7d184c3df5f7382a88d8d40333be7cf02b2ee4de9`

Update the producer outcome during rework so its evidence matches the accepted
vector.

The finding and required rework are also recorded in `LOGBOOK.md` under
2026-07-30 0312.
