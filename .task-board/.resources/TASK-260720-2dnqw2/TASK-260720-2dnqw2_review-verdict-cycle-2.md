# TASK-260720-2dnqw2 review verdict — cycle 2

## Verdict

ACCEPTED.

The cycle-2 candidate closes both demonstrated false-accept classes from the
first review, matches the rc.5 canonical metadata contract, stays within the
task's architectural boundary, and passes all independent validation gates.

## Reviewed candidate

- Product repository: `/Users/iv/Developer/Wildberries/cocoaskills`
- Task worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2dnqw2/worktree`
- Branch: `task/TASK-260720-2dnqw2-canonical-build-metadata`
- Recorded base and unchanged worktree HEAD:
  `97a0ed870782b48eebc5a9c25a9cfa8fea5ff245`
- Local `main` and `origin/main` also resolve to that SHA.
- Reviewed every changed and new source/test file:
  `src/csk/audit_registry.py`, `src/csk/builds/__init__.py`,
  `src/csk/builds/toolchain.py`, `src/csk/protocol_json.py`,
  `src/csk/builds/metadata.py`, `src/csk/install_marker.py`,
  `tests/test_protocol_json.py`, and `tests/test_build_metadata.py`.
- Reviewer run `RUN-260730-2565d2` is not goal-bound and had no operator
  directives at either checkpoint.

The reviewer did not edit, stage, commit, or discard candidate code. The
candidate remains unstaged and uncommitted for the commit-owning mover.

## rc.5 provenance and exact identities

The caller-supplied suite root was:

`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`

Its `manifest.json` independently hashes to the required non-release candidate
identity:

`sha256:b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`

Relevant manifest entries were independently compared with the actual suite
files and all matched:

- `expected/build-driver/build-input.ccj.json`: 869 bytes,
  `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`
- `expected/build-driver/receipt.ccj.json`: 1120 bytes,
  `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`
- `expected/build-driver/marker.json`: 1339 bytes,
  `sha256:feae3ffbe4e6c9bed17a6f077702c523bf6b0c7783edcef9716fddaa3d62502b`
- `vectors/build-drivers.json`: 75049 bytes,
  `sha256:f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea`

The focused suite reconfirmed the required
`policy.execution_policy=manager-worker-v1`, the exact portable cache key and
receipt hash, the two exact schema-invalid non-alias identities, canonical
receipt storage, marker-v1 compatibility, and deterministic marker-v2 state.

## Cycle-1 finding closure

### Strict UTF-8 boundary

`protocol_json` now decodes bytes strictly as UTF-8 before calling
`json.loads`; UTF-8 BOMs and UTF-16/UTF-32 inputs with or without BOMs are
rejected. The behavior is covered through the generic protocol reader, CCJ-1
reader, audit-registry reader, and install-marker reader.

### Native target/toolchain binding

Build metadata now reuses the trusted toolchain architecture-to-tuning mapping
and normalized Go-version parser. Unsupported architectures, wrong tuning
variables, malformed/non-normalized Go identities, and Go-version targets that
differ from the native target fail before canonicalization or key derivation.

Independent replay of the exact cycle-2 regression selection:

`14 passed in 0.07s`, exit 0.

## Independent validation ledger

All commands were run directly from the candidate worktree with the explicit
caller-supplied `CURATOR_CONFORMANCE_ROOT`.

1. Focused rc.5 tests:
   `184 passed in 0.37s`, exit 0.
2. Strict type check:
   `Success: no issues found in 61 source files`, exit 0.
3. Full rc.5 test suite:
   `849 passed, 6 skipped in 82.76s`, exit 0. The six skips are the existing
   platform-conditional cases.
4. Isolated sdist and universal-wheel build:
   exit 0; both new metadata modules are included.
5. Twine validation of both isolated artifacts:
   `PASSED`, exit 0.
6. `git diff --check`:
   exit 0 with no output.
7. Post-build `git status` showed exactly the same candidate source/test file
   set as before validation.

## Architecture and scope

- Typed go-v1 input/receipt logic is contained under `src/csk/builds`.
- Install-marker schemas 1 and 2 have a dedicated model module.
- Shared CCJ-1 behavior is centralized narrowly in `protocol_json`; the
  audit-registry refactor preserves its signature-envelope ownership.
- Cache-directory layout, filesystem trust, compiler execution, installer
  mutation, status/repair/GC, receipt v2, marker v3, external repository
  models, and conformance-claim emission are not implemented here.
- No physical cache, lock, receipt, quarantine, or manager-home paths enter the
  logical input, receipt, or marker models.

No acceptance-blocking findings remain. Route the accepted branch to `done`.
No `commit_ack` is supplied by this reviewer.
