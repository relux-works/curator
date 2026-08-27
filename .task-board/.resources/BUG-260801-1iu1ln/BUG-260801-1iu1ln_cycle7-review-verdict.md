# BUG-260801-1iu1ln cycle-7 reviewer verdict

## Verdict

Changes requested. Route to implementation rework (`to-dev`). This is ordinary recoverable conformance-harness work, not a Stop-The-Line boundary.

Reviewed signed CocoaSkills commit `a0046fdfbd37ecce4c5d6d0e21152628c2d2432f` on branch `task/BUG-260801-1iu1ln-lifecycle-observed-traces` in the dedicated worktree. HEAD and required base `ba250bfc4dfe104a160eadd5b5f4e340693bf892` have valid ECDSA signatures, the exact merge-base is the required base, the worktree is clean, no tag points at HEAD, and the restricted release/pin/schema-v7/CI/version surfaces have no diff.

## Acceptance-blocking findings

### 1. Captured and file-object I/O aliases bypass the persistent mutation observer

The reusable observer at `tests/protocol_lifecycle_observations.py:165` records selected `Path` methods and replaces attributes on the live `os` module at lines 431-445. It does not observe `io.open`/file-object writes or function references captured before those replacements. The live-publication binding at lines 993-1000 and the dry-run/planning/private-failure users therefore remain fail-open to ordinary alternate I/O spellings.

A direct helper probe installed the observer over a temporary persistent root, corrupted and restored a file through `io.open(..., "r+b")`, and returned exact final bytes with `events=[]`.

A combined product-seam probe then used `io.open` plus captured `fchmod` and `utime` references to write, fsync, restore and timestamp-restore the real observed surfaces. The mutations executed, but every complete case still equaled its normative vector:

```text
mutation_counts={
  publication: 42,
  project-dry-run: 1,
  global-dry-run: 1,
  planning: 12,
  private-failure: 1
}
complete_case_equal={
  publication: true,
  project-dry-run: true,
  global-dry-run: true,
  planning: true,
  private-failure: true
}
```

The publication sabotage ran after the real POSIX no-replace rename and changed/restored a live artifact descendant. Project/global upgrade wrappers changed/restored their config files. The planning wrapper changed/restored `Skillfile.json` once for each of the twelve gates. The private-build failure wrapper changed/restored the persistent-generation witness.

A separate Darwin descriptor-relative probe captured `os.open`, `os.write`, `os.ftruncate`, `os.fchmod`, and `os.utime` before observer installation, used the captured callables with `dir_fd` after live exposure, and performed 42 artifact corruptions/restorations. The complete publication case still equaled the vector and reported:

```text
publication=atomic-complete-directory
result=published
manager_home_lock=true
complete_case_equal=true
```

This directly violates the required observed-trace binding for publication immutability, dry-run effects, planning gates, and private-build failure effects. Passing final snapshots do not repair the gap because the forbidden transient operation is the behavior being asserted.

### 2. The private-build failure protected surface set is incomplete

The failure watcher at `tests/protocol_lifecycle_observations.py:3477-3518` covers a synthetic generation file and the eight enumerated downstream effect paths, then installs the observer only over that tuple at lines 3678-3683. It excludes the existing project `Skillfile.json` and does not fail closed over the full project/config/source persistent roots.

An independent probe wrapped the real `installer._build_private_misses` failure path and used the observer-visible `os.open`/`os.write`/`os.ftruncate` descriptor path to mutate, fsync and restore the project `Skillfile.json`. The failure-path mutation executed once, yet `second-build-failure-preserves-persistent-state` remained byte-for-byte equal to the normative case. This survivor does not rely on the alias gap: the operations passed through the currently patched `os` attributes, but the target was outside `watched`.

## Positive verification

- Canonical 32 cases, all 378 scalar-leaf mutations, literal/lossy classification, identity, rollback and unknown-field helpers: 417 passed, 446 deselected, exit 0.
- All six committed cycle-7 regressions: 6 passed, 857 deselected, exit 0.
- Representative inherited seams covering omitted planning validation, guardless GC, exact recovery project identity, and post-restore rollback corruption: 4 passed, 859 deselected, exit 0.
- Focused product regressions for runtime build-root currentness, global transitive upgrade fetching, and recovery ordering: 3 passed, exit 0.
- Missing-transitive and unrelated-extra all-project fetch probes both produced `deduplicated=false`; the exact nonempty fetch-closure repair is sound.
- Strict configured mypy: no issues in 68 source files.
- Compileall, exact-base diff check, restricted-surface diff, signature, merge-base, clean-tree and no-tag checks: exit 0.

## Required next handoff

1. Replace the attribute-only mutation tracing assumption with a fail-closed mechanism that covers file-object writes and module/callable aliases used by CocoaSkills, including captured descriptor-relative callables on Darwin and Linux.
2. Expand private-build failure observation to every persistent project, manager/config and source surface that must remain untouched, with explicit allowances only for operation-private staging.
3. Retain regression probes proving complete-case inequality for the `io.open`/captured-callable survivors and the previously unwatched `Skillfile.json` mutation, while preserving all existing cycle-7 and inherited sabotage cases.
4. Re-run the canonical/scalar/classification gate, the full authenticated exact-root suite, related regressions, strict mypy/compileall/diff/release/build/Twine/signature/clean-tree gates; update LOGBOOK and producer evidence in a new signed commit.

Reviewer modified no CocoaSkills code, PR/main state, tags, releases, pins, claims, schemas, CI, changelog or packaging metadata.