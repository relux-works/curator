# Brief — TASK-260910-s9jz1g: passphrase-protected signing key (R6, service)

Story `STORY-260910-9484i4` (registry-key-management, `EPIC-260910-16qce1`),
wave 4 of the 2026-09 security-audit remediation; single leaf. Rules:
`remediation-registry-producer-rules.md` (attached; `main` is now `c7ef32c`,
the CI protocol-suite pin is already `dced9b8` — do not move it). Role: developer.
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260910-9484i4/worktree`.

## Finding (read it first)
`docs/security-audit-2026-09.md` R6 (Low, operational): `keys.py` stores the
signing key as plain PKCS8 PEM beside `registry.db`; a volume leak yields key
and history together; profile §7 permits an external key provider but none
is supported.

## Settled decisions (do not reopen)
- Passphrase source is ONE environment variable, `CSK_REGISTRY_KEY_PASSPHRASE`
  (never a CLI flag — argv is world-readable). When set, `genkey` (and the
  rotation commands that write a staged/active key) write encrypted PKCS8
  PEM (`BestAvailableEncryption`), and every `load_key` path decrypts with
  it; when unset, behaviour is unchanged (plain PEM written and read).
  Loading an encrypted PEM without the variable, or with a wrong passphrase,
  fails closed at startup / command start with one clear diagnostic naming
  the variable (no traceback, no partial start); loading a plain PEM while
  the variable is set succeeds (migration path) but logs a warning that the
  key is unencrypted.
- KMS hook point: the key loading is funnelled through one provider seam
  (a small interface / factory in `keys.py`, default = file-backed PEM,
  optional passphrase) and `SECURITY.md` documents it as the integration
  point for an external provider — no external provider is implemented.
- No change to the public-key material, `key_id`, rotation semantics or
  the protocol; no spec edits.

## Deliverable
1. `signing.load_key(pem, passphrase=None)` / `export_key_pem(key,
   passphrase=None)` and the `keys.py` provider seam above; `genkey`,
   rotation staging and activation honour the variable.
2. Tests: encrypted PEM round-trip (generate → export encrypted → load with
   the passphrase → same public key); wrong passphrase and missing variable
   fail closed with the named diagnostic; plain PEM still loads; rotation
   with an encrypted key; `serve` startup with an encrypted key.
3. Docs: `README.md` (operator setup: variable, rotation of the passphrase,
   how to encrypt an existing key), `SECURITY.md` (threat model note: what
   the passphrase does and does not protect; the provider seam as the KMS
   hook point), `compose.yaml` example wiring the secret without baking it
   into the image; `CHANGELOG.md` Unreleased entry "R6: …".

## Out of scope
An actual KMS/HSM provider, key rotation redesign, R4–R8 tasks, spec edits.

## Checklist and handoff
Tick the checklist items you satisfy; attach `TASK-260910-s9jz1g_results.md`
(per-AC file:line, transcripts: `python -m pytest -q` with
`CURATOR_CONFORMANCE_ROOT`, `python -m mypy`), then
`task-board handoff TASK-260910-s9jz1g --role developer`; the runtime runs
the hosted gate once at handoff.
