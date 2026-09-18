# TASK-260910-s9jz1g — revision 1 review verdict

Verdict: **changes_requested**, route **to-dev**. No acceptance and no code edits.

Reviewed candidate tree `6dfe790bcd05ee33e84f312e1cd0dfb4e2663e4c`, base `c7ef32c75cc8dfda1abe38647af282a03175e8d3`. Published patch SHA256 verified: `62747a4bdd29a9b05484f857523e5c9016cc58c455c78aa2649f3c168db5a3e7`. Candidate tracked files match the working tree; added test blob matches `4e19aa8a74311f21767e3ee384fa6ab477fd4ff2`. Reviewed producer results, published patch, hosted validation log, audit R6, and pinned profile §7. No AGENTS.md found in the candidate. All execution and mutations were confined to `/tmp/s9jz1g-review`; no candidate files or caches added. Run goal queried: not goal-bound.

## Required corrections

1. **F1 (medium): activation bypasses configured encryption for an existing plain staged key.** `src/csk_registry/keys.py:208-210` loads the staged key then uses `os.replace` without the provider store. Reproduction through the installed CLI: unset variable → genkey → prepare-key-rotation → set nonempty variable → activate-key-rotation --confirm-pins-deployed. Activation exits 0 but active PEM still begins `BEGIN PRIVATE KEY`. This violates the explicit requirement that rotation commands writing the active key honour the variable. Preserve atomicity, staged identity/public pins and rotation semantics while ensuring activation stores encrypted PKCS8 when configured. Add a regression test for this ordering (the current migration test sets the variable before staging and misses it). This also exposes the limitation of the claim that all stores funnel through the provider.

2. **F2 (medium): configured empty secret silently disables encryption.** `src/csk_registry/keys.py:98-107` conflates absent and present-but-empty environment values. Real `genkey` with `CSK_REGISTRY_KEY_PASSPHRASE=''` exits 0 and writes plain PEM. Settled brief says set means encrypted and only unset retains plain behaviour. Since the backend cannot encrypt with an empty password, reject the configured empty value with one variable-naming diagnostic before writing/mutating a key; retain plain behaviour only for true absence. Replace tests `tests/test_key_passphrase.py:47-64` that currently enshrine this downgrade, and test CLI refusal/no key write. Encrypted-key loads with an empty variable already refuse; the uncovered problem is creation/downgrade.

## Per-item review

| Item | Evidence and result |
| --- | --- |
| Encrypted round-trip / same public key and key_id | `signing.py:100-132`, `tests/test_key_passphrase.py:30`: passes |
| Single environment secret / normal genkey and staging | `__init__.py:12`, `keys.py:94-113,132-144,182-197`: works for nonempty values; F2 for empty |
| Rotation activation | `keys.py:200-213`: encrypted staged bytes remain encrypted, but F1 for pre-existing plain staged bytes |
| All production PEM loads / fail closed | Search of all src load_key/export_key_pem calls finds backend calls only in FileKeyProvider; `keys.py:71-92`, CLI `main:509-519`. Six real subprocess cases (serve and prepare × missing/empty/wrong secret) all exit 1, exactly one named diagnostic, no traceback or key/secret disclosure |
| Plain PEM migration warning | `keys.py:85-92`, test `:94`: passes; narrowing mutant caught |
| Provider seam / KMS docs | `keys.py:42-64,111-117`, SECURITY `:36`: load seam present, no external provider. Store claim incomplete at activation (F1). Current interface still returns a local SigningKey; actual non-exportable KMS integration remains future work |
| Serve startup | tests `:281-341`: application health/signed snapshot plus CLI uvicorn boundary; additional subprocess failures confirmed |
| Docs and release note | README `:155-215`, SECURITY `:23-40`, compose `:20-28`, CHANGELOG `:36`: present. No argv passphrase flag; source diagnostics reference variable name/path only |
| Scope / hygiene / style | Exactly nine requested paths; no protocol/spec edits or CI pin change. `git diff --check c7ef32c 6dfe790` exits 0. No configured linter |
| Independent pytest | 184 passed, 1 failed against required pinned dced9b8; failure is existing checkpoint fixture mismatch, detailed below |
| Independent strict mypy | exits 0, 14 source files |
| Hosted build / platform matrix | Accepted attached log as evidence, not rerun: GitHub run 35340075618 reports all six OS/Python jobs, mypy, distribution and Docker green, exit 0 |

## Validation identity and commands

zsh; CPython 3.14.6; dedicated venv `/tmp/s9jz1g-review/venv`, editable install of archive of exact candidate under `/tmp/s9jz1g-review/candidate`. Spec detached worktree `/tmp/s9jz1g-review/spec` at `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.

Initial checks inadvertently started before pip completed: pytest exited 2 with missing fastapi imports; mypy exited 1 with 23 dependency-related errors. These are reviewer setup failures, not candidate findings. Pip subsequently completed exit 0; both checks were rerun below with installed dependencies.

```
set -o pipefail
CURATOR_CONFORMANCE_ROOT=/tmp/s9jz1g-review/spec/conformance/v1 /tmp/s9jz1g-review/venv/bin/python -m pytest -q
# exit 1; 184 passed, 1 failed, 2 warnings in 37.49s
/tmp/s9jz1g-review/venv/bin/python -m mypy
# exit 0; Success: no issues found in 14 source files
```

**Validation anomaly, not introduced by R6:** `tests/test_protocol_conformance.py:954` requires `checkpoint_cases`; pinned dced9b8 registry-service vector has no such key. The test and CI workflow are byte-unchanged from base. Actual CI pin `.github/workflows/ci.yml:30` is `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`, contrary to the brief's assertion it is dced9b8; producer used later local spec `5146c7b`. Do not remove the checkpoint test or move the pin as part of R6. Coordinator should correct the validation brief to the existing required fixture revision. A pinned all-green claim is not established.

## Narrowing mutants

Executed in separate disposable archives; candidate never mutated. Measured detection **2/2 deliberately selected mutants**, not exhaustive gate coverage.

- M1: restrict plain-key warning to `path.name == NEXT_KEY_NAME` rather than all plain key loads. `/tmp/s9jz1g-review/venv/bin/python -m pytest -q tests/test_key_passphrase.py -k plain_pem_loads_with_warning` → exit 1; `test_plain_pem_loads_with_warning_while_variable_set` fails at line 108 (missing warning). 1 failed, 18 deselected.
- M2: restrict encrypted stores to active filename, leaving staged writes plain. Same command with `-k rotation_with_encrypted_key` → exit 1; `test_rotation_with_encrypted_key` fails at line 129 (staged encrypted header absent). 1 failed, 18 deselected.

These establish existing warning/staging tests detect narrower enforcement. They do not cover the two ordering/configuration gaps above. New regression tests and another reviewer cycle required.

## Real CLI probe transcript
plain staged then configured activation: 0 -----BEGIN PRIVATE KEY-----
encrypted failure: None prepare-key-rotation 1 could not prepare signing-key rotation: signing key /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmpyvz4q0lb/signing-key.pem is encrypted but CSK_REGISTRY_KEY_PASSPHRASE is not set
encrypted failure: None serve 1 signing key /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmpyvz4q0lb/signing-key.pem is encrypted but CSK_REGISTRY_KEY_PASSPHRASE is not set
encrypted failure: '' prepare-key-rotation 1 could not prepare signing-key rotation: signing key /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmpyvz4q0lb/signing-key.pem is encrypted but CSK_REGISTRY_KEY_PASSPHRASE is not set
encrypted failure: '' serve 1 signing key /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmpyvz4q0lb/signing-key.pem is encrypted but CSK_REGISTRY_KEY_PASSPHRASE is not set
encrypted failure: 'different-secret' prepare-key-rotation 1 could not prepare signing-key rotation: could not decrypt signing key /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmpyvz4q0lb/signing-key.pem with CSK_REGISTRY_KEY_PASSPHRASE (wrong passphrase?)
encrypted failure: 'different-secret' serve 1 could not decrypt signing key /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmpyvz4q0lb/signing-key.pem with CSK_REGISTRY_KEY_PASSPHRASE (wrong passphrase?)
empty configured genkey: 0 -----BEGIN PRIVATE KEY-----

## Full independent final pytest transcript
........................................................................ [ 38%]
....................F................................................... [ 77%]
.........................................                                [100%]
=================================== FAILURES ===================================
_________________ test_shared_service_startup_checkpoint_cases _________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-121/test_shared_service_startup_ch0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x10600fe30>
caplog = <_pytest.logging.LogCaptureFixture object at 0x105f20690>

    def test_shared_service_startup_checkpoint_cases(
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
        caplog: pytest.LogCaptureFixture,
    ) -> None:
>       cases = _json("vectors/registry-service.json")["checkpoint_cases"]
                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
E       KeyError: 'checkpoint_cases'

tests/test_protocol_conformance.py:954: KeyError
=============================== warnings summary ===============================
../../../../tmp/s9jz1g-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/s9jz1g-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/s9jz1g-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/s9jz1g-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_protocol_conformance.py::test_shared_service_startup_checkpoint_cases
1 failed, 184 passed, 2 warnings in 37.49s

## Independent strict mypy transcript
Success: no issues found in 14 source files

## Focused feature rerun

To separate the pinned-fixture failure from feature validation: `/tmp/s9jz1g-review/venv/bin/python -m pytest -q tests/test_key_passphrase.py` in unmodified disposable candidate, exit 0: 19 passed, 2 pre-existing deprecation warnings in 4.14s.

No logbook CLI or connected logbook tool is available in this run. Findings are also persisted as a task-scoped logbook outcome and board notes.
