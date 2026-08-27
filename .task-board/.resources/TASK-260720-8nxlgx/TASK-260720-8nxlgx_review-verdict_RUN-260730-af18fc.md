# TASK-260720-8nxlgx review verdict — RUN-260730-af18fc

## Verdict

**ACCEPTED.**

The two acceptance-blocking findings from `RUN-260730-e46794` are genuinely
closed. I independently reproduced both closures on native Windows and then ran
twenty further adversarial boundary cases that the committed focused suite does
not cover — every one fails closed without exposing a receipt, an artifact path,
or candidate bytes. All acceptance criteria are met and every required gate is
green on both the macOS host and the native Windows host.

Six non-blocking observations are recorded below. None of them admits candidate
bytes or weakens the trust boundary; the two worth acting on later belong to
follow-up scope, not to this task.

The reviewer changed no CocoaSkills product source or repository test file.

## Exact candidate

- Base/HEAD: `15860e3f309888845b9271a257fb95f7c2825b56`
- `origin/main` is the same commit; the worktree is zero commits behind.
- Branch: `task/TASK-260720-8nxlgx-protected-windows-build-cache`
- Dependency `TASK-260720-2jfnz6` is `done`.
- Product delta is exactly three files, unstaged and uncommitted:

```text
28b977d4bba4e2c75ab646344a95a0cb338cc2420aa332bd3195dffb1482e6f6  src/csk/builds/cache.py        (M, +4 lines)
43a413ab71944976f20009016a176d74853cd1262230ed618f67f159f706cd52  src/csk/builds/cache_windows.py (new)
611aefdc0a4d2574d85b14b426e56ed0d090fecdde487f5e4d6f646c15874111  tests/test_build_cache_windows.py (new)
```

All three hashes match the developer's rework evidence and match the native
Windows snapshot byte-for-byte. The `cache.py` change is only the
`os.name == "nt"` selector branch in `cache_for_manager_home`; no portable
receipt or cache-key semantics are touched.

## Prior findings — independently confirmed closed

### R1 — late hard links

`_open_protected_child_file` now revalidates through `_revalidate_file_child`,
which queries `FILE_STANDARD_INFO.NumberOfLinks` on the retained handle **and**
on the independently reopened selected path, after identity, DACL, and
attribute validation, and then once more on both. Because those checks run in
the context manager's exit path, a link inserted after byte verification turns
the pending `return` into `_UntrustedState`. `_validate_source_unchanged`
applies the same retained-plus-reopened query to the publication source at
every final stability boundary.

Confirmed by the committed regressions
(`test_windows_late_lookup_hard_links_fail_final_validation` for both receipt
and artifact, `test_windows_late_publication_source_hard_link_is_rejected`) and
by my own boundary probe, which found no admission at any externally reachable
point.

### R2 — inheritable untrusted mutation ACEs

`_validate_manager_home_security` now rejects an untrusted allow ACE carrying
mutating rights when it is effective on the home **or** inheritable through
`OBJECT_INHERIT_ACE` / `CONTAINER_INHERIT_ACE`, including under
`INHERIT_ONLY_ACE`, and no longer counts inherit-only manager ACEs toward the
manager's effective home grant.

I extended this well past the committed `(F)` cases. Six *partial* inheritable
untrusted grants — write-only, delete-only, `WDAC`-only, `WO`-only,
append-only, and inherit-only write — each produce `untrusted-provenance` on
lookup, `cache_boundary_untrusted` on publication, and **zero** cache roots
created. I also verified the intended positive side: an inheritable untrusted
*read* grant on the home is tolerated, and every object created below it
(`builds`, `go-v1`, staging, quarantine, entry, `bin`, receipt, artifact)
carries **zero** `S-1-1-0` ACEs, so the protected DACL genuinely strips
inheritance rather than merely masking it.

## Independent adversarial probe — 20 cases, all fail closed

Attached as `TASK-260720-8nxlgx_reviewer-boundary-probe.py` and
`TASK-260720-8nxlgx_reviewer-profile-probe.py`; raw native output attached as
`TASK-260720-8nxlgx_reviewer-native-probe-output.log`.

| Case | Result |
| --- | --- |
| Manager home, inheritable untrusted `(W)` / `(DE)` / `(WDAC)` / `(WO)` / `(AD)` / `(IO)(W)` (6 cases) | `untrusted-provenance` + `cache_boundary_untrusted`, no roots created |
| Escaping junction substituted for the live `<key>` entry | no HIT, no bytes |
| Escaping junction substituted for the entry's `bin`, behind a validly sealed artifact | no HIT, no bytes |
| `bin` replaced by a regular file | `artifact directory is not a directory` |
| Receipt replaced by a directory | `cache receipt is not a regular file` |
| Pre-existing permissive entry holding a **self-consistent receipt for foreign bytes** under the correct key | lookup refuses; foreign receipt never leaked |
| Same entry under publication | foreign entry quarantined, result `published`, live artifact holds the honest bytes — never `reused-winner` of attacker bytes |
| Alternate data stream smuggled onto an otherwise exactly sealed artifact | `cache artifact has alternate data streams` |
| Untrusted `Everyone:(F)` on the driver root / sealed entry / sealed `bin` (3 cases) | each rejected by the exact-DACL comparison |
| Real `icacls /setowner` owner drift on the sealed artifact | `owner does not match the current manager principal` |
| Read-only lookup on a pristine home, then `publish` with no guard | tree stays `()`; `cache_lock_required` |
| Inheritable untrusted read ACE on the home | publication succeeds, HIT, zero inherited untrusted ACEs anywhere below |

## Acceptance criteria

| Criterion | Status |
| --- | --- |
| Only the manager principal and trusted OS administrators may mutate the boundary and descendants | Met. Every cache object requires a protected DACL matching exactly the manager / `S-1-5-18` / `S-1-5-32-544` profile; any extra ACE is untrusted. |
| Permissive or unverifiable roots, escaping reparse points, special files, hard-linked receipt/artifact files, ownership or DACL drift, containment races force a miss or fail closed and never admit candidate bytes | Met. Committed suite plus 20 independent cases above. |
| Exact receipt, input, artifact path, hash, size, concurrent-winner, immutability, dry-run rules match POSIX | Met. `test_windows_lookup_binds_exact_receipt_path_hash_and_size` covers all five binding cases; concurrent publication yields exactly `["published", "reused-winner"]`; differing bytes raise `CacheConflictError`; `dry_run_outcome` comes from the shared `CacheInspection`. |
| If protection cannot be established reliably, persistent reuse is disabled rather than opened | Met. `_protection_supported()` gates both entry points; on a POSIX host `inspect` is `UNSUPPORTED` and `quarantine` raises `cache_protection_unsupported`. A planted reparse point at `builds` disables reuse permanently rather than opening it (see N3). |
| Windows CI exercises positive and negative backend cases | Met without a workflow change: `.github/workflows/ci.yml` already runs `python -m pytest -v` on `windows-latest` for Python 3.11–3.14, so the new file is collected there. |
| Full pytest plus strict mypy pass without non-Windows import regressions | Met — see the ledger below. |

## Validation ledger

Every command below was run by me, standalone.

### macOS host

- `python -m pytest -q tests/test_build_cache_posix.py tests/test_build_cache_windows.py` → **43 passed, 22 skipped**, exit 0
- `python -m pytest -q` → **854 passed, 80 skipped**, exit 0
- `python -m mypy` (strict, `python_version = "3.11"`) → **Success: no issues found in 65 source files**, exit 0
- `uvx ruff check src/csk/builds/cache.py src/csk/builds/cache_windows.py tests/test_build_cache_windows.py` → **All checks passed!**, exit 0
- `git diff --check` → exit 0; `test ! -e uv.lock` → exit 0
- Non-Windows import surface: module imports cleanly, `_api()` cache stays unprimed
  (`hits=0, misses=0, currsize=0`), `_supported=False`, `inspect` →
  `UNSUPPORTED` with `receipt_bytes=None`, `quarantine` →
  `cache_protection_unsupported`, and `cache_for_manager_home` still returns
  `PosixBuildCache`.
- Same import check under **Python 3.11.14**, the declared floor and CI floor → clean.

### Native Windows

Remote file hashes were verified equal to the local candidate before every run.

- `python -m pytest -q tests\test_build_cache_windows.py` → **23 passed**, exit 0
- `python -m mypy --strict --platform win32 --follow-imports skip src\csk\builds\cache_windows.py`
  → **Success: no issues found in 1 source file**, exit 0
- Reviewer boundary probe (18 cases + 1 diagnostic) → all boundaries held, exit 0
- Reviewer profile probe (2 cases) → both held, exit 0
- `powershell -ExecutionPolicy Bypass -NoProfile ... python -m pytest -q`
  → **790 passed, 144 skipped**, exit 0
- One full-suite invocation failed on a pre-existing unrelated host
  PowerShell policy issue; see the evidence-integrity note below.

### Architecture fit

The backend mirrors `cache_posix` closely: same `_MissingState` /
`_UntrustedState` / `_CorruptState` taxonomy mapped onto the same
`CacheEntryStatus` and `BuildCacheError` codes, same layout, same
staging/quarantine namespaces, same guard discipline, and lazy `ctypes`
resolution so import stays safe off Windows. `cache.py` gains only the selector
branch. Neither backend is referenced from `ARCHITECTURE.md` or `docs/`, so this
task introduces no documentation regression relative to its accepted POSIX
sibling.

The manager-home precondition is satisfiable in production: a directory created
under `%USERPROFILE%` inherits exactly `<user>:F`, `SYSTEM:F`,
`Administrators:F`, which passes `_validate_manager_home_security`. The elevated
token caveat — new objects owned by `BUILTIN\Administrators` instead of the
manager SID — is already recorded as a platform finding in `LOGBOOK.md` entry
1351 and fails closed.

## Evidence-integrity note on the developer ledger

The developer's reported native full-suite **result** is correct; the
*mechanism* recorded next to it is not reproducible on the current host.

The rework evidence attributes **790 passed, 144 skipped** to process-only
`PSExecutionPolicyPreference=Bypass`. With exactly that workaround I got
**1 failed, 789 passed, 144 skipped**, exit 1:
`tests/test_shell_init.py::test_powershell_hook_activates_and_restores_on_every_prompt`
fails with PowerShell `UnauthorizedAccess`. Every execution-policy scope on the
host reports `Undefined` (effectively `Restricted`), and the environment
variable does not reach the `powershell.EXE -NoProfile -File` child the test
spawns.

Run instead under `powershell -ExecutionPolicy Bypass -NoProfile`, the same
test passes and the full suite reproduces the developer's exact numbers:

- single test → **1 passed**, exit 0
- full suite → **790 passed, 144 skipped** in 302s, exit 0

This is **not a regression from this task**: the failing test is unmodified,
imports nothing from the delta, and the delta's other two files are new. The
correction is only that the ledger should name
`powershell -ExecutionPolicy Bypass` as the working invocation rather than the
process-only environment variable.

## Non-blocking observations

### N1 — residual post-validation hard-link window (diagnostic)

With `_revalidate_child` instrumented so a hard link lands *after* the
artifact's own final retained-plus-reopened link check but before
`_inspect_entry` returns, `inspect` returns `hit` for a multiply linked
artifact.

Not acceptance-blocking, for four independent reasons:

1. Reaching it required replacing an internal function inside the manager
   process. That is code injection, not an attacker capability; at that point
   the process is already lost.
2. The accepted POSIX backend is weaker here. `cache_posix._validate_regular_file`
   checks `st_nlink` at open and `_hash_file` rechecks it via
   `_stable_file_state` immediately after hashing — and then
   `_inspect_open_entry` returns with **no** further `st_nlink` check at all.
   Windows now rechecks twice on two independent references at a strictly later
   point. The acceptance criterion makes POSIX the parity reference.
3. The bytes returned were already read and hash-verified against the receipt;
   a link appearing afterwards cannot retroactively alter them. `CacheInspection`'s
   own docstring already requires the consumer to inspect again before relying
   on the path, so the architecture assumes this window exists.
4. Only the manager principal or a trusted administrator holds rights to create
   the link, and the acceptance criteria place both inside the trust boundary.

No TOCTOU check can hold a Windows link count across a function return.
Recorded as a known limitation rather than a defect.

### N2 — publication-source validation reuses the manager-home validator

`_validate_publication_source_handle` calls `_validate_manager_home_security()`
on the artifact **source file**, which requires the manager to hold
`FILE_WRITE_DATA | FILE_APPEND_DATA | FILE_DELETE_CHILD | READ_CONTROL | WRITE_DAC`
on it. Confirmed natively: a source sealed read/execute — the equivalent of the
POSIX mode `0500` source that `cache_posix._open_publication_source` accepts —
is rejected as:

```text
cache_publication_invalid: publication artifact source is not private, singly
linked, owner-controlled regular state: manager home does not grant the manager
principal required control
```

Fails closed, so not a security defect. But it is stricter than POSIX for a
case POSIX allows, and the message says "manager home" while validating a build
artifact. Worth a dedicated source-profile validator with its own diagnostic;
relevant to the consumers `TASK-260720-11yhth` and `TASK-260720-2x6mjn`, whose
build outputs must satisfy it.

### N3 — a junction at `<home>/builds` disables the cache with no repair path

`_open_or_create_mutable_directory` retries, then calls `_move_aside`, which
opens the object and raises `_UntrustedState` on the reparse point — so the
junction is never quarantined and every later publication fails
`cache_boundary_untrusted`. This satisfies the criteria ("fail closed",
"persistent reuse is disabled rather than opened") and refusing to delete a
reparse point is the safer choice, but it means an operator must clear it by
hand. `test_windows_reparse_escape_and_unverifiable_boundary_fail_closed`
asserts exactly this behaviour, so it is deliberate.

### N4 — a failure inside `_create_stage` leaks an orphaned stage

`stage_exists = True` is set only after `_create_stage` returns, so an error
partway through staging leaves a partially sealed `entry-<32hex>` directory in
`.builds-staging` with no cleanup. POSIX sets its cleanup flag immediately
after `mkdir` and therefore leaks less. Disk-only: stage names are fresh
32-hex, nothing enumerates staging contents, and the staging root's own profile
is unaffected.

### N5 — cleanup failure can mask the original error code

`_publish_locked`'s `finally` guards `_remove_stage` with `except OSError`, but
`_remove_stage` can also raise `_UntrustedState` or `BuildCacheError`. A
failing cleanup therefore replaces the original error code
(`cache_boundary_untrusted` in place of the real cause). Diagnostic-only.

### N6 — cosmetic

- Dead constants: `_ERROR_ACCESS_DENIED`, `_ERROR_INVALID_HANDLE`,
  `_SECURITY_CHANGE_ACCESS` are each defined once and never used.
- `str()` wrappers around the already-`str` `GO_V1_DRIVER` and
  `build_input.artifact_path`.
- `_revalidate_child` for the entry runs twice: explicitly at the end of
  `_inspect_entry` and again on the context manager's exit.

## Acceptance evidence for the commit-owning mover

This reviewer run supplies no `commit_ack`. The candidate is verified at base
`15860e3f309888845b9271a257fb95f7c2825b56` with the exact three-file delta and
hashes recorded above, unstaged and uncommitted in
`/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-8nxlgx/worktree`.
The commit-owning mover should commit exactly that scope and make the final
`done` transition with `commit_ack=scope_committed`.
