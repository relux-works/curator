# TASK-260910-s9jz1g results — service-key-passphrase-support (R6)

Story `STORY-260910-9484i4`, worktree `.temp/STORY-260910-9484i4/worktree`,
base `main = c7ef32c`.

## Revision 2 (rework: F1 activation encryption, F2 empty-passphrase refusal)

Rework brief `TASK-260910-s9jz1g_rework-rev2.md` disposition of the two
round-1 medium findings. Everything else is unchanged from revision 1
(the rev1 record is preserved verbatim below; only the line numbers noted
here moved).

### F1 — activation honours the configured passphrase

- `src/csk_registry/keys.py:210` (`activate_rotation`): the active-key
  write now goes through the provider seam —
  `resolved.store(active_key_path(home), staged)` (`:223`) — instead of a
  raw `os.replace`, then removes the staged file and fsyncs the directory.
  A plain staged key activated while `CSK_REGISTRY_KEY_PASSPHRASE` is set
  is re-exported as encrypted PKCS8; the staged key material, public pins,
  keyring write order, and rotation semantics are unchanged (`store` is the
  same atomic 0600 replace used by `genkey`/staging).
- Tests: `tests/test_key_passphrase.py:185`
  (`test_activation_encrypts_plain_staged_key`: the reviewer's exact
  unset → genkey → prepare → set → activate sequence ends with an
  `ENCRYPTED PRIVATE KEY` active PEM, staged file gone, staged public pin
  preserved in the overlap set) and `:209`
  (`test_activation_keeps_encrypted_staged_key_loadable`: encrypted staged
  key activated under the same passphrase stays loadable). The pre-existing
  `:147` encrypted-rotation test covers the same inverse path end to end.
- Manual CLI reproduction (real console entry, `/tmp/r6-manual-check`):
  plain staged header `BEGIN PRIVATE KEY` → activate exit 0 → active
  header `BEGIN ENCRYPTED PRIVATE KEY`, mode `0600`, staged file removed.

### F2 — present-but-empty passphrase is an error, not "plain"

- `src/csk_registry/keys.py:33` `_EMPTY_PASSPHRASE_MESSAGE` (names the
  variable exactly once, carries no secret or key material);
  `key_passphrase_from_env()` (`:107`) returns `None` only for true
  absence and raises `KeyPassphraseError` for a present-but-empty value,
  so every key operation (`genkey`, rotation, `serve` startup with a key
  to load, and all other `load_active_key` paths via
  `default_key_provider()`) fails closed before any write or mutation;
  `FileKeyProvider.load` (`:77`) and `.store` (`:103`) guard an explicit
  empty passphrase the same way.
- Tests: replaced `:47-64` (`test_passphrase_from_env_shapes` now asserts
  unset→None / set→bytes only; new
  `test_empty_passphrase_rejected_before_any_write` at `:56` asserts the
  refusal at the env function, the provider factory, `initialize_key`
  with no file created, and `genkey` exit 1 with the variable named
  exactly once and no `Traceback`) and added the empty-at-load case
  (`test_empty_passphrase_at_load_time_fails_closed`, `:77`) plus
  `serve` with an empty variable
  (`test_serve_command_with_empty_passphrase_has_no_partial_start`,
  `:411`, asserting `uvicorn.run` is never reached).
- Manual CLI check: `export CSK_REGISTRY_KEY_PASSPHRASE=''` →
  `export-snapshot` exit 1, stderr is exactly the one diagnostic naming
  the variable, no traceback.
- Docs: `README.md:165` (activation writes encrypted PEM; new
  set-but-empty bullet), `SECURITY.md:32` (empty rejected; only true
  absence selects plain PEM), `CHANGELOG.md:36` (R6 entry now says
  activation and the empty refusal; pin reference untouched).

### Revision 2 validation transcripts (`set -o pipefail`, bash)

Interpreter: CPython 3.14.6, venv `/tmp/csk-venv-r6` (editable install of
this worktree; the shared `/tmp/csk-venv` still points at
`STORY-260910-stz5f0` and was not touched). Conformance root:
`CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1` (detached
curator-spec at `47c3c8c`, the CI pin in
`.github/workflows/ci.yml:30`) — this corrects both the rev1 producer
root (`5146c7b`) and the round-1 review root (`dced9b8`, which missed
`checkpoint_cases`).

- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1
  /tmp/csk-venv-r6/bin/python -m pytest tests/test_key_passphrase.py -q`
  → `23 passed, 2 warnings` (pre-existing fastapi/starlette
  deprecations), exit 0, 19.35 s.
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1
  /tmp/csk-venv-r6/bin/python -m pytest -q` → `189 passed, 2 warnings`,
  exit 0, 77.93 s. The `checkpoint_cases` conformance test that failed
  under the wrong round-1 root passes at the corrected pin.
- `/tmp/csk-venv-r6/bin/python -m mypy` → `Success: no issues found in
  14 source files`, exit 0.
- No linter is configured in this repo (CI jobs: test, typecheck, build,
  docker); style follows the surrounding code. `git diff --check`
  clean by construction (edits only); `git status` shows exactly the
  nine rev1 paths, no strays. Local `docker build` and the hosted gate
  (`scripts/remote-gate.sh`) were not run here: the runtime runs the
  hosted gate once at handoff.

### Revision 2 scope notes

- Passphrase secrecy: every diagnostic names only the variable (and the
  key path where relevant); the secret value is never formatted into a
  message, log, or argv — verified by the `count(...) == 1` / no-`Traceback`
  assertions and the manual transcripts above.
- Deliberately out of scope (unchanged): actual KMS/HSM provider (seam
  only), rotation redesign, R4/R5/R7/R8/P4, spec edits, tags/releases,
  CI pin moves. No `--passphrase` CLI flag. No protocol, envelope, or
  key-material change.
- Spec gaps: none found.

---

## Revision 1 record (preserved)

Story `STORY-260910-9484i4`, worktree `.temp/STORY-260910-9484i4/worktree`,
base `main = c7ef32c`. Interpreter: CPython 3.14.6 (venv
`/tmp/csk-venv-s9jz1g`, editable install of this worktree; see Finding 1).

### Per-file changes

- `src/csk_registry/__init__.py:12` — new `KEY_PASSPHRASE_ENV =
  "CSK_REGISTRY_KEY_PASSPHRASE"` constant next to the other env names.
- `src/csk_registry/signing.py:100` — `load_key(pem, passphrase=None)`;
  `:114` — `export_key_pem(key, passphrase=None)` (None = plain PKCS8 as
  before, otherwise `BestAvailableEncryption`). Thin pass-through; the
  passphrase/encryption decision matrix lives in `keys.py`.
- `src/csk_registry/keys.py` — provider seam + env wiring:
  `:34` `KeyPassphraseError(ValueError)`; `:42` `KeyProvider` Protocol
  (passphrase/load/store); `:66` `FileKeyProvider` (encrypted-PEM detection,
  fail-closed diagnostics, unencrypted-key warning, atomic store);
  `:98` `key_passphrase_from_env()` (empty = unset); `:111`
  `default_key_provider()` factory. `load_active_key` (`:128`),
  `initialize_key` (`:132`), `public_keys` (`:146`), `prepare_rotation`
  (`:182`), `activate_rotation` (`:200`), `cancel_rotation` (`:216`),
  `retire_public_key` (`:232`) all funnel through an optional
  keyword-only `provider` (default: env-backed file provider).
- `src/csk_registry/cli.py:511` — `main()` catches `KeyPassphraseError` and
  prints the single diagnostic to stderr with exit 1 (covers commands whose
  handlers do not catch `ValueError`, and `serve`, where `app_from_env`
  raises before `uvicorn.run`, so there is no partial start).
- `tests/test_key_passphrase.py` (new, 19 tests) — see AC mapping.
- `README.md:155` — "Passphrase-protected signing key" operator section
  (setup, semantics, encrypt-in-place snippet, passphrase-rotation snippet,
  staged-key note). Both snippets were executed end-to-end against the real
  console entry: plain→encrypted keeps mode 600 and the same `key_id`
  (`351df417a08acf88`) across rotation; old secret and unset variable each
  exit 1 with the single diagnostic.
- `SECURITY.md:23` — threat-model note (protects against offline volume/file
  copies made without the secret; not against live-process compromise, env
  leaks, or as a permissions/KMS substitute) and the `KeyProvider` seam as
  the KMS hook point.
- `compose.yaml:20` — commented secret wiring (host-env interpolation /
  env_file), nothing baked into the image; `docker compose config -q` passes.
- `CHANGELOG.md:36` — Unreleased `### Security` entry "R6: …".

### AC mapping

Task AC: "Encrypted PEM round-trip test passes" —
`tests/test_key_passphrase.py:30`
(`test_encrypted_pem_round_trip_preserves_public_key`: generate → export
encrypted → load with passphrase → same `public_pinned`/`key_id`, asserting
the `ENCRYPTED PRIVATE KEY` header).

Brief deliverables:

1. `signing.load_key(pem, passphrase=None)` (`signing.py:100`),
   `export_key_pem(key, passphrase=None)` (`signing.py:114`), provider seam
   (`keys.py:42,66,111`); `genkey` and rotation staging/activation honour
   the variable via the default provider (`keys.py:132,182,200`;
   activation moves the already-encrypted staged bytes and decrypts both
   keys on load).
2. Tests — round-trip (`:30`), plain-envelope unchanged (`:39`),
   env shapes incl. empty=unset (`:47`, `:59`), missing passphrase (`:67`),
   wrong passphrase (`:81`), plain PEM + warning (`:94`), genkey (`:112`),
   encrypted rotation (`:123`), plain→encrypted rotation migration (`:142`),
   single-diagnostic fail-closed on two command paths (`:162`, `:185`),
   explicit-provider override (`:200`), custom-provider funnel proof
   (`:213`), serve startup positive/negative via `app_from_env`
   (`:281`, `:296`) and via the real `serve` CLI with mocked uvicorn
   (`:309`, `:323`, asserting `uvicorn.run` is never reached).
3. Docs — README (`README.md:155`), SECURITY (`SECURITY.md:23`),
   compose (`compose.yaml:20`), CHANGELOG (`CHANGELOG.md:36`).

Gating argument: the pre-change backend behavior was probed directly
(`cryptography` 50.0.1: encrypted+None → `TypeError`, encrypted+wrong →
`ValueError("Incorrect password…")`, plain+password → `TypeError`, empty
password → `ValueError`). The tests assert the distinct `KeyPassphraseError`
type, the variable name occurring exactly once on stderr, absence of
`Traceback`, and exact PEM-header markers — all of which fail against the
unmapped backend behavior, so the suite gates the feature rather than its
own presence.

### Validation transcripts (rev1; story venv, `set -o pipefail`)

- `export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1;
  /tmp/csk-venv-s9jz1g/bin/python -m pytest -q` → `185 passed, 2 warnings`
  (warnings: pre-existing fastapi/starlette deprecations), exit 0, 39.09 s.
  Local spec checkout is `5146c7b` (contains `dced9b8`); no new vectors or
  schemas consumed, CI pin untouched.
- `/tmp/csk-venv-s9jz1g/bin/python -m mypy` → `Success: no issues found in
  14 source files`, exit 0.
- `/tmp/csk-venv-s9jz1g/bin/python -m build -q -o /tmp/csk-dist-s9jz1g` +
  `twine check` → both PASSED, exit 0.
- `docker compose config -q` → OK (compose edit is comment-only).
- No linter is configured in this repo (CI jobs: test, typecheck, build,
  docker; no ruff/flake8 config); "lint clean" is by style review against
  surrounding code. Local `docker build` and the hosted gate
  (`scripts/remote-gate.sh`) were not run here: the runtime runs the hosted
  gate once at handoff.

### Deliberately out of scope

Actual KMS/HSM provider (seam only), rotation redesign, R4/R5/R7/R8/P4,
spec edits, tags/releases. No `--passphrase` CLI flag by settled decision
(argv is world-readable). No protocol/envelope/key-material change.

### Spec gaps

None found.

### Findings

1. The shared `/tmp/csk-venv` is an editable install pointing at a *different*
   story worktree (`STORY-260910-stz5f0`), so its console entry and bare
   imports do not test this tree (pytest was unaffected via
   `pythonpath=["src"]`). I built `/tmp/csk-venv-s9jz1g` from this worktree
   for all evidence above and left the shared venv untouched. Reviewers
   should build their own venv per the campaign rules.
2. The brief states the CI protocol-suite pin "is already `dced9b8`", but
   `.github/workflows/ci.yml:30` pins `47c3c8c…`. Per the brief's own "do
   not move it" and the no-new-vectors rule, the pin was left untouched;
   flagging the discrepancy for the orchestrator.
