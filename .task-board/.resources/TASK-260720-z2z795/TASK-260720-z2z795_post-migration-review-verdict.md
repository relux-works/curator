# TASK-260720-z2z795 post-migration review verdict

Date: 2026-07-30  
Role: reviewer  
Run: `RUN-260730-9dff10`

## Verdict

**Changes requested → `to-dev`.**

This is ordinary implementation rework. No external blocker, product decision,
approval, or human-only architecture decision is required.

The required pre-verdict goal query returned:

```text
Active Goal: none (run is not goal-bound)
```

The review did not modify product or test files, stage, commit, push, publish,
use SSH, or claim the old PR #7 CI as current-byte evidence.

## Reviewed provenance and scope

- Prerequisite `TASK-260720-1pvfj5` is accepted and `done`.
- Original accepted base:
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Current committed task head and local task-branch remote:
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`.
- The accepted base remains an ancestor of current head.
- Clean canonical main checkout:
  `97a0ed870782b48eebc5a9c25a9cfa8fea5ff245`, equal to local
  `origin/main`.
- Worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`
- Current rework remains unstaged and limited to:
  - `src/csk/locking.py`
  - `src/csk/transactions.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`
- Current SHA-256:
  - `locking.py`:
    `c272ff852f44a965c8da6e15afb396b43cd4ce7f6fbcc9522350fbf8f4760bb0`
  - `transactions.py`:
    `415a2418474582fbc9094c7f6e157212fd1e67543ccd77c905182f587e77cb52`
  - `test_locking.py`:
    `b668a726eb685d5a4882790c2283cb035697adbe189ed17cb73c572bcd42eb11`
  - `test_transactions.py`:
    `f578bbc4f400b6003e754e2ab58401af2875a287312ef055567fe9dd7b67606c`

## Prior migration finding closed

The legacy-to-stable transition now fails closed, stable v1 records use a
non-integer legacy PID guard, acquisition rechecks the canonical descriptor and
published token, and legacy `.lock.stale-*` witnesses are reserved from
transaction namespaces. The new migration set, prior lock-integrity set,
earlier reviewer set, and contract-targeted set all pass independently.

## Material finding 1: physical lock aliases do not share canonical identities

`canonical_project_identity()` and `_canonical_manager_home()` use
`Path.resolve(strict=False)` plus `os.path.normcase`
(`src/csk/locking.py:64-78`). On macOS, `normcase` is a no-op and `resolve`
preserves the caller's case and Unicode spelling even when the filesystem maps
both spellings to the same inode. `ProjectLock` hashes this textual identity
into its lock filename (`src/csk/locking.py:227-237`), while process-wide
project/build/home ordering is keyed by the same textual manager-home identity
(`src/csk/locking.py:239-279`, `362-421`, `429-480`).

On the current default case-insensitive, normalization-insensitive macOS
filesystem, deterministic barrier probes produced:

```text
home_samefile=true
project_samefile=true
home_id_equal=false
project_id_equal=false
duplicate_project_acquired=true
home_acquired_while_build_held=true
```

and for NFC/NFD aliases:

```text
samefile=true
identity_equal=false
duplicate_project_acquired=true
```

Two operations can therefore enter planning for the same physical project at
once. Worse, a manager-home lock acquired through a case alias succeeds while
the same physical home's build lock is held through the other spelling. This
directly violates canonical project locking and the required
project → optional build → home hierarchy.

The inverse risk exists on Windows: unconditional `normcase` can merge distinct
names in a per-directory case-sensitive tree. The accepted Go reference has
explicit per-directory Windows case-sensitivity handling; the Python identity
layer does not.

Required rework:

1. Derive project and manager-home identities from physical filesystem lookup
   semantics, not textual `resolve` plus platform-wide `normcase`.
2. Existing macOS case and NFC/NFD aliases must map to one identity and one
   process-order namespace, while distinct names on a case-sensitive volume
   remain distinct.
3. Windows must respect per-directory case sensitivity for both existing paths
   and missing suffixes.
4. Add real concurrent regressions proving that case/normalization aliases
   cannot acquire duplicate project locks and that an aliased build lock blocks
   home acquisition. Add the inverse Windows case-sensitive regression.

## Material finding 2: a journal filename is not bound to its embedded transaction ID

`_load_journal_record(transaction_id)` selects a record by the caller/scanned
filename, decodes its embedded `journal.transaction_id`, and validates the
journal, but never compares the two IDs
(`src/csk/transactions.py:766-791`). Subsequent saves and removal derive paths
from the embedded ID (`src/csk/transactions.py:712-758`, `1170-1223`).

A deterministic probe created a legitimate crash state
(`phase=committing`, target `state=backed_up`), renamed the canonical
`real-id.json` record to `alias-id.json` without changing its canonical bytes,
and called home-wide recovery. Observed:

```text
before: embedded_id=real-id filename=alias-id.json live=absent
after:  live_text=desired
error:  recover transaction alias-id: transaction journal disappeared: real-id
journals: [alias-id.json]
```

Recovery installed the desired target before discovering that its next durable
save path did not exist. A mismatched/corrupt record therefore causes target
mutation and remains permanently recoverable on every scan. This violates
durable transaction identity and fail-closed recovery.

Required rework:

1. Bind every loaded `.json` and `.json.delete` filename ID to the embedded
   transaction ID before `_resume()` or any target/sidecar mutation.
2. Reject a mismatch as `TransactionCorruptionError` while preserving the
   record, sidecars, targets, and lock witness byte-for-byte.
3. Add prepared, committing/backed-up, rolling-back, cleanup, and removal-tomb
   mismatch regressions.

## Material finding 3: valid read-only targets cannot be staged or cleaned up

The staging path applies each final mode before the staged tree is usable:

- directories are `chmod`ed to the desired mode before descendants are created
  (`src/csk/transactions.py:244-259`);
- files are `chmod`ed before `_copy_staging_file()` reopens them for append
  (`src/csk/transactions.py:249-258`, `262-303`).

Public-API probes with otherwise valid regular targets produced:

```text
0444 desired file:
  PermissionError opening .desired for append
0555 desired directory with a child:
  PermissionError creating .desired/payload
```

Preparation cleaned its journal and sidecars, but it cannot represent those
valid generic targets. A second probe started from a live `0555` directory,
successfully installed the desired replacement, then failed permanently while
removing the read-only backup tomb:

```text
commit_returned=false
live_new=new
error=PermissionError removing .backup.delete/old
journals=[readonly-cleanup.json]
sidecars=[...backup.delete]
```

The cleanup journal remains, and recovery repeats the same permission failure.
This is not a special-file case: `digest_path()` and the staging manifest accept
these regular files/directories and bind their modes into the desired digest.

Required rework:

1. Use transaction-owned writable construction modes, then durably finalize
   exact desired modes only after contents/descendants are complete.
2. Make mode finalization crash-recoverable; a crash in the temporary-mode
   window must remain distinguishable from foreign mutation.
3. Make journal-owned cleanup able to remove read-only directory trees without
   losing manifest/digest ownership or leaving an unrecoverable permission
   window.
4. Add `0444` file and nested `0555` directory coverage across
   prepare/commit, crash recovery, rollback, and committed-sidecar cleanup.

## Independent validation ledger

- Tool readiness:
  - Git `2.50.1`
  - ripgrep `15.2.0`
  - Python `3.11.14`
  - pytest `8.3.4`
  - mypy `2.3.0`
  - Ruff `0.16.0`
  - build `1.2.2.post1`
  - task-board `0.23.0`
- New migration regressions: `14 passed`, exit 0.
- Prior lock-integrity regressions: `14 passed`, exit 0.
- Earlier reviewer regressions: `14 passed`, exit 0.
- Contract-targeted ordering/recovery/rollback/concurrency/preimage/order
  regressions: `13 passed`, exit 0.
- Focused pytest: `88 passed, 1 skipped in 5.51s`, exit 0.
- Strict mypy: no issues in `57 source files`, exit 0.
- Ruff lint: clean, exit 0.
- Ruff format: four files already formatted, exit 0.
- Full pytest: `658 passed, 20 skipped in 91.97s`, exit 0.
- Package build: sdist and wheel built successfully, exit 0.
- `git diff --check`: exit 0 before and after validation.
- Post-validation status, scope, and all four hashes are unchanged.
- The focused skip is the real-Windows file/tree durability test.

The earlier GitHub Windows matrix covers only committed head `5f8bfbbd...`,
not the current unstaged rework. No current-byte real-Windows result is claimed.

`logbook` is not installed and task-board exposes no logbook command. Findings
are persisted in this task-scoped outcome, raw evidence, and board notes.
