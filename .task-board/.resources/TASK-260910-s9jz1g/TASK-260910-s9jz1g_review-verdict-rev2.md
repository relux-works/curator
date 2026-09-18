# TASK-260910-s9jz1g — review verdict revision 2

Verdict: accepted. F1 and F2 are resolved; no blocking findings.

Reviewed candidate tree `0756023966f1719b5d0cf48287d448d1876f83d5`, base `c7ef32c75cc8dfda1abe38647af282a03175e8d3`.
Published patch SHA256 verified: `80227991ea00da7775d2eb255642a26d39d61523e883ae8eb9abc4b6e0a196e5`.
All candidate files matched the managed worktree byte-for-byte; tests and attacks ran only in /tmp/s9-review2. No candidate edits or new worktree files.
Read campaign rules, both briefs, producer results, both patches, rev2 hosted validation, audit R6, and pinned profile §7.

| Item | Evidence / disposition |
|---|---|
| Encryption and unchanged public identity | signing.py:100,114 uses password and BestAvailableEncryption; test_key_passphrase.py:30 round-trip retains public pin/key ID. |
| All production loads and writes | keys.py:76,101 provider load/store; repository search found signing load/export calls only in this provider. initialize:151, prepare:202, activate:223 use provider stores (3/3 active/staged write sites). |
| F1 activation | keys.py:210-226 re-exports through provider and removes stage; atomic helper :281 uses temporary 0600 file, fsync, os.replace. Installed CLI unset→genkey→prepare→set→activate produces encrypted active PEM, matching staged public key, stage removed, mode 0600. Encrypted→encrypted inverse also passes (2/2 shapes). |
| F2 empty | keys.py:107-118 distinguishes absence from empty; provider also refuses explicit empty bytes (:77,:102). Installed CLI empty genkey, rotation prepare/activate/cancel and serve refuse before key/file-content mutations. Empty encrypted loads refuse too. |
| Missing/wrong and plain migration | keys.py:79-99 names variable once, never interpolates secret; plain PEM warns. cli.py:509 catches before server launch; app.py:754 loads before Store. Tests :91,:106,:119,:228,:249 and CLI probes cover refusals. |
| Provider/KMS seam | keys.py:45 Protocol and :121 factory; SECURITY.md:35 documents future hook, no external implementation. This remains a file-oriented seam; no claim that a non-exportable HSM signer is implemented. |
| Serve and rotation tests | test_key_passphrase.py:147,185,209,348,364,378,394,411 exercise encrypted rotation and app/CLI startup, including listener-not-called refusals. |
| Docs | README.md:155 setup, encryption migration, secret rotation; SECURITY.md:23 limits protection to offline copies without secret; compose.yaml:20 deployment env wiring; CHANGELOG.md:36 R6 entry. |
| Scope/rev1 comparison | Reconstructed rev1 from published patch; only keys.py, tests and corresponding README/SECURITY/CHANGELOG corrections differ. Other four changed files byte-identical. No spec, pin or unrelated task edits. |
| Validation | Independent full pytest and strict mypy transcripts below; git diff --check exit 0. No configured standalone linter. |

## Independent validation

Shell zsh, `set -o pipefail`; CPython 3.14.6, dedicated venv /tmp/s9-review2/venv installed editable from /tmp/s9-review2/source (archive of exact candidate).
Conformance checkout HEAD verified `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`, matching unchanged .github/workflows/ci.yml:30.
Initial pytest/mypy attempts preceded completion of pip install and returned missing-module exit 1; these were setup failures, then rerun after installation completed successfully.

Commands from disposable source:
```
CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 /tmp/s9-review2/venv/bin/python -m pytest -q
/tmp/s9-review2/venv/bin/python -m mypy
```

### pytest (exit 0)
```text
........................................................................ [ 38%]
........................................................................ [ 76%]
.............................................                            [100%]
=============================== warnings summary ===============================
../../../../tmp/s9-review2/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/s9-review2/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/s9-review2/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/s9-review2/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
189 passed, 2 warnings in 72.06s (0:01:12)

```

### mypy (exit 0)
```text
Success: no issues found in 14 source files

```

### probe (exit 0)
```text
genkey passed 0
prepare-key-rotation passed 0
activate-key-rotation passed 0
prepare-key-rotation refused 1
serve refused 1
prepare-key-rotation refused 1
serve refused 1
prepare-key-rotation refused 1
serve refused 1
genkey passed 0
prepare-key-rotation passed 0
activate-key-rotation passed 0
prepare-key-rotation refused 1
serve refused 1
prepare-key-rotation refused 1
serve refused 1
prepare-key-rotation refused 1
serve refused 1
genkey refused 1
CLI probes PASS: both activation shapes; 12 encrypted-load refusals; empty genkey; no key bytes or secret in diagnostics
genkey passed 0
prepare-key-rotation passed 0
activate-key-rotation refused 1
cancel-key-rotation refused 1
prepare-key-rotation refused 1
serve refused 1
Empty configured secret: staged rotation activate/cancel/prepare and serve refuse without file changes: 4/4

```

## Narrowing mutants

2/2 selected mutants caught, each in its own disposable copy; this is bounded targeted coverage, not an exhaustive mutation claim.

1. Narrow activation re-encryption to already-encrypted stages; plain stages use old os.replace path. `tests/test_key_passphrase.py::test_activation_encrypts_plain_staged_key` exits 1 at line 202 (encrypted-header assertion).
2. Narrow empty-secret refusal to direct env/helper/explicit-provider paths by letting default factory map empty to None. `tests/test_key_passphrase.py::test_empty_passphrase_at_load_time_fails_closed` exits 1 at line 85 (DID NOT RAISE KeyPassphraseError). Production load path is therefore tested, beyond helper-only refusal.

Baseline full suite ran against untouched source concurrently with isolated mutants. The CLI probe script and mutation harness are attached separately for reproducibility; synthetic private-key bytes printed by pytest assertion rewriting are not needed in this verdict.

## Accepted existing evidence and limits

Read TASK-260910-s9jz1g_change-request_rev2-validation.log: hosted run 35342421184, exit 0, all six Python 3.11/3.14 × macOS/Linux/Windows tests, mypy, distribution and Docker builds successful. Those hosted checks were accepted from attached evidence, not rerun. Independently reran pytest/mypy locally plus CLI probes and selected mutants.
No live-network positive serve probe was added: existing tests verify the app endpoint and uvicorn handoff, and independent subprocess probes verify startup refusal.

Nonblocking documentation provenance caveat retained from revision 1: R6 CHANGELOG says dced9b8 whereas actual CI pin is 47c3c8c. The corrected review brief explicitly requires preserving unrelated rev1 content. Pinned §7 states custody/rotation obligations but does not literally mention external providers; the seam requirement comes from settled task scope. Neither discrepancy changes this implementation or validation contract. No protocol changes or spec edits were made.

Run goal query returned no active goal (not goal-bound). Acceptance must use accept_cr revision=2, routing to integrating; reviewer does not mark done or commit.
