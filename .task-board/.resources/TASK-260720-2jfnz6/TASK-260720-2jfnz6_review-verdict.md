# TASK-260720-2jfnz6 review verdict

## Verdict

ACCEPTED.

The protected POSIX build-cache implementation satisfies the task acceptance
criteria, remains inside the assigned module boundary, and passes the focused,
strict-typing, full-regression, and diff-hygiene gates. No acceptance-blocking
finding remains.

## Reviewed candidate

- Product repository:
  `/Users/iv/Developer/intranet/cocoaskills`
- Task worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/worktree`
- Branch:
  `task/TASK-260720-2jfnz6-protected-posix-build-cache`
- Recorded base and current worktree HEAD:
  `495ad021847529ce5a544dba415ca2fe19949539`
- Local `main` and `origin/main` resolve to the same SHA, which contains the
  accepted canonical-build-metadata dependency.
- Reviewed source hashes:
  - `src/csk/builds/cache.py`:
    `d5936a40bc6628f37a7b5e485ff5651d11cd38ec939b27cb81e2cf4b9c6f0821`
  - `src/csk/builds/cache_posix.py`:
    `0c7d62fe823640178c8356018d3d0ce563c7a63db456dcc82fd12e97db054736`
  - `tests/test_build_cache_posix.py`:
    `908318074e28d65600275df6e7b26dff1a904723272981742e79e908b8c626b7`
- Reviewer run `RUN-260730-91bfce` is not goal-bound and had no operator
  directives at either checkpoint.

The reviewer did not edit, stage, commit, or discard candidate code. The three
task files remain untracked and unstaged for the commit-owning mover.

## Acceptance evidence

- The portable interface addresses cache state only by complete
  `GoBuildInput`, logical key derived from that input, canonical receipt bytes,
  and the manager-derived artifact-relative path.
- POSIX storage is isolated below the csk-specific
  `<manager-home>/builds/go-v1/<hex-key>` namespace; source snapshots under the
  existing cache namespace cannot collide with it.
- Lookup validates the effective-UID owner and protected modes before parsing
  candidate receipt bytes. Every live component is opened relative to held
  directory descriptors with `O_NOFOLLOW`; receipt and artifact files must be
  regular, singly linked, owner-controlled files with exact immutable modes.
- Exact directory contents, canonical receipt bytes/hash, the complete expected
  input and key, the derived artifact path, artifact size, and SHA-256 are
  checked on lookup. The same complete input and bytes are reverified in the
  protected stage and again on the selected live winner.
- Inspection creates, repairs, quarantines, and locks nothing. The tests prove
  a miss and untrusted/corrupt dry-run classification leave the complete tree
  unchanged.
- Publication copies from a verified private file descriptor into
  `.builds-staging`, outside the live namespace, then uses native Darwin
  `renameatx_np(RENAME_EXCL)` or Linux `renameat2(RENAME_NOREPLACE)` to select
  one complete directory winner. It never merges files into an entry.
- Existing corrupt or untrusted candidates are moved outside `builds` under
  the caller-held mutation guard and are never adopted, including the
  self-consistent forged-entry case.
- Identical concurrent winners are compared byte-for-byte and the staged loser
  is discarded. Different bytes for the same logical key raise
  `CacheConflictError`.
- Published entries are sealed as owner-only read/execute state; locked
  quarantine can still move sealed state for later GC. Windows ACLs, installer
  transactions, status, and GC implementation remain out of scope.
- The backend's `CacheMutationGuard.assert_held()` contract matches the
  separately developed manager-home lock witness.

## Independent validation ledger

All commands ran from the exact candidate worktree with Python 3.14.4,
pytest 9.1.1, and mypy 2.3.0. Python bytecode, pytest temporary state, mypy
cache, and the alternate Git index were directed outside the source worktree.

1. POSIX-focused pytest:
   `33 passed in 0.37s`, exit 0.
2. Strict `python -m mypy`:
   `Success: no issues found in 63 source files`, exit 0.
3. Full repository pytest with the accepted rc.5 conformance root:
   `882 passed, 6 skipped in 90.12s`, exit 0.
4. Alternate-index `git diff --check` over all three untracked task files:
   exit 0 with no findings.
5. The producer-built wheel was independently inspected and contains both
   `csk/builds/cache.py` and `csk/builds/cache_posix.py`; the attached producer
   evidence records successful isolated build and Twine checks.
6. Post-validation Git state is unchanged: exactly the three task files are
   untracked, with no staged changes.

## Threat-boundary probe

An additional diagnostic changed the manager-home mode to `0777` from inside
the publishing process after stage construction. Publication completed and a
subsequent inspection correctly classified the boundary as
`untrusted-provenance`. This requires arbitrary same-principal interference,
which the accepted contract explicitly places outside the install-time
invariant; cooperating writers are serialized by the caller-held manager-home
lock, and root/administrator is part of the TCB. The probe therefore does not
invalidate acceptance and no broader adversary model is silently claimed.

## Routing

Route the accepted verdict to `done`. No `commit_ack` is supplied by this
reviewer; the candidate remains for the commit-owning mover.
