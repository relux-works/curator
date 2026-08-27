# TASK-260720-1ljev5 review verdict — cycle 1

Verdict branch: **changes_requested**

Route: **to-dev**

The lock/journal integration, marker-v1 runtime compatibility, grace handling,
receipt validation, and focused Unix tests are well covered. The change cannot
be accepted because the conservative-reference and protected-boundary
acceptance criteria still have unsafe gaps.

## Findings

### 1. P1 — An uncertain project can be forgotten, so a later GC can delete its referenced builds

`internal/scopes/gc.go:146-158` records an uncertainty but retains a consumer
only when at least one valid marker was read. `Collect` then rewrites
`consumers.json` before it checks `marked.uncertain`. This creates unsafe
two-pass behavior:

- If the consumer registry is unreadable or malformed, the first pass warns and
  skips the build sweep but rewrites the registry as empty. The next pass sees
  no uncertainty and may sweep every old unreferenced entry.
- If a project has only an invalid/unreadable marker or its skill scope cannot
  be read, the first pass warns but drops that project from the consumer
  registry. The next pass no longer visits the uncertain marker and may remove
  its build.

`readScope` compounds this at `internal/scopes/gc.go:190-214`: non-directory
members, symlinked skill directories, and non-`ENOENT` `Lstat` failures are
silently ignored. A symlinked `.agents/skills`, global, or hybrid root is also
followed by `os.ReadDir` without being reported as uncertainty. The existing
invalid-marker test always installs another valid marker first, so it does not
exercise the consumer-loss sequence.

Required rework:

- Do not rewrite an unreadable/invalid consumer registry.
- Preserve a registered consumer whenever its scope is uncertain; prune only
  after proving the scope is absent or valid and empty.
- Validate the consumer registry shape/version rather than accepting missing,
  null, or unknown fields as an empty registry.
- Detect and warn on symlink/reparse scope roots or members and on all
  non-absence marker metadata failures.
- Add two-consecutive-pass regressions for a corrupt registry, an invalid-only
  marker project, an unreadable installed-skill directory, and redirected scope
  roots. The second pass must still refuse the build sweep.

This finding violates the requirements to retain uncertain candidates
conservatively and to ensure invalid markers or untrusted roots cannot lead to
deletion.

### 2. P1 — Rename and deletion are not bound to the validated cache-root handle

`buildcache.Sweep` opens and holds a validated root handle, but uses it only to
list names. `sweepUnreferenced` reopens entries by pathname, closes those
handles after inspection, and calls `retireEntry(base, name)`.
`retireEntry` then uses pathname-based `os.CreateTemp`, `os.Rename`,
`os.RemoveAll`, and directory sync operations.

On Unix, an open directory descriptor does not pin later pathname resolution.
If the validated root pathname is exchanged before retirement, these operations
can act in the replacement target while the held descriptor still refers to
the original root. The comment that every removal resolves against the proven
boundary is therefore not true, and the current escape tests cover only lexical
names, an initially redirected root, and links inside an entry; they do not
cover a boundary exchange between validation and mutation.

Required rework:

- Perform retirement and resumable cleanup relative to the validated
  root/directory handle with no-follow semantics and object-identity
  revalidation appropriate to Unix and Windows.
- Add an adversarial boundary-swap race test proving that neither the candidate
  nor cleanup can touch a replacement directory outside the Curator cache
  root.

This finding violates the explicit requirements that sweep remain inside the
verified protected boundary and that deleting one entry cannot escape the
Curator cache root.

### 3. P2 — The submitted Windows evidence is not a passing or internally consistent gate

The attached `win-buildcache.log` exits 1. It shows
`TestAtomicPublicationConflictingRace` failing in this tree, while the attached
`win-base-buildcache.log` shows that same test passing on the accepted base.
That contradicts the implementation notes and `EXIT-CODES.md`, which say the
same failures were reproduced on both trees. The Windows reparse tests also
skip because the test account lacks symlink privilege, and the real
`internal/scopes` sweep integrations skip because the fixture cannot create
protected Windows state.

Required rework:

- Produce a task-scoped native Windows gate in which the task-owned changes and
  required reparse/race cases actually run and pass, or provide reproducible
  like-for-like baseline evidence that resolves the conflicting race result.
- Keep pre-existing failures separated from the task gate; the task acceptance
  criterion itself requires Windows, fault, and race tests to pass.

## Independent validation

- `go test ./internal/buildcache ./internal/scopes ./cmd/curator -count=1` —
  pass.
- `go test ./internal/install -run 'TestPostCommitMaintenance|TestMaintenanceFailureAfterCommitIsAWarning' -count=1`
  — pass.
- Attached producer evidence reports `go test ./... -count=1`, build, vet,
  lint, and focused race gates green on Darwin, but the Windows evidence above
  remains red/incomplete.

No project code was modified during review.
