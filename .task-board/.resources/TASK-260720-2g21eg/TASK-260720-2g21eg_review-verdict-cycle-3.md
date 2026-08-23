# TASK-260720-2g21eg review verdict — cycle 3

## Verdict

Changes requested. Route to `to-dev`.

Cycle 3 closes the previously reported `.pth` execution, five-probe, and
diagnostic-precedence defects, and every requested regression gate is green.
The implementation still leaves executable Python runtime code outside the
worker identity and exposes the fixed hidden mode through the actual installed
CLI argument surface. These are in-scope implementation defects with viable
rework paths, not a human-only or external stop-the-line boundary.

## Candidate reviewed

- CocoaSkills worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`
- Base/HEAD, clean local `main`, and `origin/main`:
  `495ad021847529ce5a544dba415ca2fe19949539`
- Both declared dependency tasks were `done` with accepted outcome handoffs.
- Accepted rc.5 conformance root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Candidate hashes:
  - `src/csk/builds/go_v1.py`:
    `6e04f58712fd07a45a97d681838d02a89583636a94914bd5ae93bc4d415f0571`
  - `src/csk/cli.py`:
    `9e0724b53e6fbcd86f967f611c19336130620d9a7fb7500b9dc4d879dc35a92c`
  - `tests/test_builds_go_v1.py`:
    `dcdea7a9ae43c45b3801945940d9eb50fd85093a0624232115114c1388667a0c`
  - `tests/test_builds_go_v1_fixture.py`:
    `07e33bb2661a36ef5b033d25bde25b6d0d5f2c7374e7cb43ea79bfc9a661c570`
- Installed cycle-3 wheel:
  `sha256:8de8b5bb6152489957e171722f5bbd73c00eccdd2124872b6e929de8125bf5c9`;
  installed `go_v1.py` and `cli.py` matched the candidate hashes.

No CocoaSkills source, test, index, or commit was modified during review.

## Blocking findings

### 1. Importable standard-library code is outside the worker identity

`_StartupIdentity` records the site/standard-library root inode and mode plus
selected immediate `*.pth`, `sitecustomize.py`, `usercustomize.py`,
`pyvenv.cfg`, `python._pth`, and `pybuilddir.txt` files
(`go_v1.py:1412-1450`, `go_v1.py:2247-2342`). It does not record the bytes or
physical identities of ordinary importable files under the standard-library
root. Its mutation guard watches the root and selected hooks, not those module
files.

`worker_runtime_proof` is produced only after the launcher has imported
`csk.cli`, `go_v1`, and their standard-library dependencies. The validator then
accepts every loaded module whose path is below `stdlib_root` without matching
that module to any identity entry or digest (`go_v1.py:4458-4479`,
`go_v1.py:4540-4558`). `json`, which `go_v1` imports, is one concrete reachable
example. Thus mutable Python code can execute before the ready proof while the
manager still reports the same worker identity.

Independent isolated reproduction:

1. Build a canonical fake installed-manager prefix matching the implementation's
   launcher/interpreter/site/stdlib layout.
2. Resolve `_ManagerIdentity` and create a valid worker runtime proof containing
   the reachable `stdlib/json/__init__.py`.
3. Install `_IdentityMutationGuard`, replace that module's bytes in place, close
   the guard, resolve identity again, and validate the same runtime proof.

Observed:

- standard-library module SHA-256 changed;
- `manager_identity_unchanged=true`;
- `startup_identity_unchanged=true`;
- `mutation_guard_accepted_change=true`;
- runtime proof was accepted both before and after the change;
- the module was absent from `startup.hooks`.

This leaves `identity-verified-manager-owned-worker`,
`pre-launch-worker-identity-verification`, and
`post-exec-identity-reverification` unsatisfied. It also contradicts Protocol
Core 4.2.1's requirement that every mutable component added by an interpreter
implementation be treated as worker TCB.

Required rework:

- Bind every mutable/importable runtime component the fixed Python worker can
  execute, including the standard-library module bytes and any permitted
  `python*.zip`, or replace this launch with an equivalently closed
  self-contained manager-owned worker.
- Carry those identities through pre-launch verification, the in-session proof,
  retained mutation protection, and post-exec re-verification.
- Add an installed-worker negative that mutates a standard-library module
  actually imported before `worker_runtime_proof`; require
  `build_execution_worker_identity_invalid`, no compiler start, no result, and
  complete worker-domain teardown.

### 2. The actual installed CLI lets a user select the hidden worker mode

`cli.main()` dispatches to `run_worker()` whenever its ordinary external
invocation has actual `sys.argv[1:] == ["__csk-go-worker-v1"]`, before the
public parser (`cli.py:26-34`). Consequently the literal user command

```text
cycle3/venv/bin/csk __csk-go-worker-v1 < /dev/null
```

selected the worker protocol, emitted a 150-byte length-framed
`build_execution_worker_protocol_invalid` response, and exited 3. This is not a
parser rejection; it proves the user-supplied argument selected the hidden
branch.

The existing test checks only that programmatic `cli.main([WORKER_MODE])` enters
the parser, then explicitly asserts that the actual `sys.argv` path calls the
worker (`tests/test_builds_go_v1.py:1035-1050`). That does not cover the
installed command surface specified by Protocol Core 4.2.1: the fixed hidden
mode must not be user-visible and no user option may select it.

Required rework:

- Gate the hidden re-execution on manager-owned, authenticated launch context so
  a direct installed-CLI invocation cannot enter an accepted worker session.
- The worker must authenticate its manager parent independently of values the
  caller can self-assert in the first request; the current request supplies its
  own secret and expected identity.
- Add a subprocess-level installed-entry-point negative for the literal hidden
  argument. It must be rejected as a user command without entering the worker
  protocol, starting Go, or publishing anything, while the one manager-created
  re-execution remains accepted.

## Independent verification

All commands used the candidate `src` tree and the exact source-matching
cycle-3 installed wheel.

| Gate | Result |
| --- | --- |
| Accepted-root focused pytest over build source, source identity, toolchain, metadata, go-v1 unit tests, and native fixture | exit 0; 289 passed, 4 skipped |
| Full accepted-root pytest | exit 0; 962 passed, 6 skipped |
| Separate installed-wheel native macOS Go fixture | exit 0; 6 passed |
| `python -m mypy src/csk` strict | exit 0; no issues in 62 source files |
| `python -m tabnanny` on the four task files | exit 0 |
| Trailing-whitespace check on the four task files | none |
| `git diff --check` | exit 0 |
| Candidate/source-matching wheel hashes | unchanged after review |
| Standard-library identity/mutation-guard/runtime-proof probe | exit 0; reproduced |
| Literal installed hidden-mode invocation | entered worker; exit 3 with framed worker failure |

The focused tests cover the exact source-aware argv, five-per-operation native
probe calls, fixed environment, named vector cases, graph rejection surface,
output verification, and never-run fixture invariant. Their green result does
not exercise either open boundary above.

## Routing

This is ordinary implementation rework. Preserve this verdict and its probe
artifacts, route the task to `to-dev`, then require a new independent reviewer
cycle. Do not use `blocked`; no external input or human-only architecture
decision is required.
