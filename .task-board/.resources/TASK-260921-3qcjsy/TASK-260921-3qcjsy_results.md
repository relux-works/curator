# TASK-260921-3qcjsy results — adopt decisions 0017 and 0018

Adopted 2026-09-21 per operator decision (STORY-260921-3z0fgr).
Both documents carry Status adopted with a dated adoption note;
every option is adopted and every open question is resolved to a
recorded choice in an Adoption choices table in each document.
Normative text is amended in the same revision; `make validate`
is green (see Validation).

## 0017 environment credential modes — choices for operator confirmation

Options 1–3 adopted as recommended (verbatim):

- O1: `isolation.<profile>.<env-id> = shared|isolated` stays the
  canonical v1 knob, described as credential-store sharing, not a
  sandbox; `credential_mode`, the rename, and a per-run flag are
  deferred.
- O2: registry-owned strategies per environment × GOOS exactly as
  listed — codex_cli keyring-preferred → `auth.json` file-link; pi
  file-link with the native root corrected to `~/.pi/agent`;
  claude_code Linux file-link; claude_code macOS isolated-by-default
  with shared refused until the experiment passes; opencode ambient
  only; `keychain-shared` and copy-at-provision rejected.
- O3: fix-first manager repairs, then an explicit inspect → plan →
  apply migration under the manager lock — never silent inside
  `resolve --repair` — with no secret copies at any step.

Open questions resolved (fail-closed / least-privilege where the
draft gives no recommendation):

| # | Adopted choice |
|---|---|
| 1 | macOS claude_code shared store: **unsupported until the experiment is recorded**. `environment_shared_unsupported` stands; the §7.4 residual is answered negatively (a 2.1.273 managed-home login wrote a file under `CLAUDE_CONFIG_DIR`, not a suffixed Keychain item). Reconsideration gate: an authorized disposable-account experiment proving (a) which store is read first, (b) whether the file is rewritten in place on refresh, (c) when, if ever, a suffixed Keychain item is written. The manager never exports Keychain secrets to JSON. |
| 2 | `isolated` = **bounded store separation**: the managed home does not share the declared credential stores. Not account separation (ambient auth unchanged) and never filesystem isolation. Stated in §7.4 and manager §12.4. |
| 3 | Pi root precedence per option 2: native root is **`~/.pi/agent`**. Migration inventories both `~/.pi/auth.json` and `~/.pi/agent/auth.json`; operator chooses on conflict with preservation. The 2 B `~/.pi/auth.json` is a manager-created artefact: unlink the recorded link, never delete the native file. |
| 4 | Codex keyring identity **assumed operator-global** (CODEX_HOME-independent) until proven otherwise. Probe: disposable-account scratch-CODEX_HOME login-state observation (logged-in ⇒ global; not-logged-in ⇒ per-home), secrets never exported. Consequences: under native `keyring` or `auto` storage, `isolated` is `environment_isolated_unsupported` — `isolated` for `codex_cli` is available under `file` storage only (under `auto` the file-link is inert on a keyring host, so the home would authenticate through the operator-global keyring); a future revision may admit `auto` only where the manager proves the effective store is `file`, probe-gated. Sharing is defined by the native effective storage only (a diverged managed `config.toml` changes nothing; the manager never realigns configs); any selector outside the verified `file`/`keyring`/`auto` set (including `ephemeral`) fails closed with `environment_credential_unsupported`. |
| 5 | Fields named. Config: `isolation.<profile>.<env-id>` stays the **only** config field. Marker: each passthrough record carries `isolation`, `strategy`, `source_role` (`native`/`managed`), `backend` (`file`/`keychain`/`ambient`), `backend_version`, `provenance` (`provisioned`/`repaired`/`migrated`); linkless strategies record without a path. A schema-1 marker records `path` and `strategy` only and is never rewritten to add the record. Publication under the manager-home mutation lock via temp + atomic rename with journal protection; rollback restores the preceding marker. Backups/discovery never follow auth symlinks and never archive credential bytes. Marker-schema enactment is a named follow-up spec revision. |
| 6 | Fleet enforcement of `isolated`: **not in this adoption — follow-up**. The system schema still locks only toward `shared`; enforcement needs a reviewed policy revision. |
| 7 | Copy consent: **none authorized — copy stays refused**. No consent shape exists; `secret_material_waivers` waives a finding, never extraction. Any future copy needs operator-owned profile/env, source/destination roles, purpose, reason, expiry, plus an explicit no-copy-boundary revision. |

## 0018 curator run permission interface — choices for operator confirmation

Adopted **with the 2026-09-16 amendment**: config-driven mode
(launcher-global default, per-profile setting, CLI override,
built-in default `yolo` for interactive launches), lockable
fleet-wide force-`native`; headless, CI, and tracked silence
resolves `native`; legacy policy/lock transport fails closed.
Options 1–6 adopted as amended.

| # | Adopted choice |
|---|---|
| 1 | `--yolo` ships **in the same increment** as `--permissions`, as an exact alias of the yolo mode. `-d`/`--danger` stay rejected. |
| 2 | Tracked-yolo `ax` admission: **both** — an ax-owned permission representation **and** validated natively-mapped admission, carried by a versioned `ax` capability. Until then: tracked + effective `yolo` (any level) is refused (`permission_mode_tracked_unsupported`), no fallback to untracked, no bypass spelling in the composed `ax` document. The carrying version is fixed by the implementing revision (follow-up). |
| 3 | Unknown future native policy forms: **refused (fail closed)** with a versioned-capability token. Provider grammar closed per pinned tool release; unknown forms ⇒ `usage` (exit 2), never a resolved-policy claim. Grammar distinguishes prompt text from flags (launcher SPEC owns the parsing rule). |
| 4 | When native stored settings relax the posture beneath `native`: the launcher prints the **effective-native-policy stderr line** (warning naming the relaxation + source; best-effort detection; never claims beyond what it inspected) and records it in the **launch record** — the tracked `ax` launch document (launcher-SPEC-owned extension key) when tracked, the stderr provenance stream when untracked (no persistent record). Exact spelling/key: launcher SPEC revision (follow-up). |
| 5 | Negative tests drive the **real `curator run` entry with a fake tool**; rows named: per-environment mappings; per-conflict refusals (`=`/separate, aliases, `exec` placement); precedence (flag/profile/global/default-interactive/default-headless); v1-file-with-member + unknown-value rejections; lock refusals from every level; tracked refusals from every level; headless/CI-silence ⇒ `native` (`source=default-headless`); legacy-fragment would-be-`yolo` refusals from every level including the flag; one narrowing mutant per bound. Launcher follow-up implements; no new spec vectors in this revision. |
| 6 | Draft's proposal adopted: each mapping **re-verified per tool release**; capability table keys (environment, tool release) to a grammar version. On drift the **`yolo` mapping fails closed first** (refuse `yolo` for unverified releases; `native` forwards verbatim with no claims). Exact token/member encoding: to be fixed by the implementing revision (follow-up). |
| 7 | Fail-closed rule **normative now** (item 5): unestablished transport ⇒ would-be `yolo` refused (`permission_policy_unsupported`); `native` proceeds. Token value + exact fragment member names: to be fixed by the implementing revision (follow-up). Non-interactive markers: closed set {`CI`, `GITHUB_ACTIONS`} fixed by this adoption; additions by later spec revision only; versioned in launcher SPEC §4.6, mirrored in environments §10.1. Operator-confirmation item: the set is thin (Jenkins/Azure set other variables, not `CI`); the TTY prong is the primary detector, so silence does not fail open by default. |

## Normative amendments in this revision (curator-spec)

- `protocol/environments.md`: §7.4 (Pi root, keyring assumption,
  credential-record content for the follow-up marker revision +
  migration rules, residual answered negatively; codex `isolated`
  file-only; non-empty-directory repair ⇒
  `environment_credential_conflict`),
  §7.7 + §10.4 (`environment_credential_conflict`,
  `environment_credential_unsupported`), §8.4.1 (conflict row; §8.2
  unchanged — schema-1 records path+strategy only),
  §10.1 (repair conflict rule, never-silent migration, permission-mode
  resolution + headless detector + transport fail-closed rule), §10.2
  (transport requirement), §12.1 (new `permissions.<profile>` knob),
  §12.2 (`permissions` lockable toward `native` only), §13 (vector
  clause).
- `profiles/manager.md`: §1 (locked list + direction), §12.3
  (profile-remove entries), §12.4 (conflict rule, bounded-store
  framing, diagnostics), §12.5 (repair mirror, fragment shaping,
  conflict row).
- `schemas/v1/manager-config-v2.schema.json`,
  `schemas/v1/system-config-v2.schema.json`,
  `tools/generate-vectors/{manager_config,system_config}.go`,
  `tools/validate.py`, `tools/test_validate.py`,
  `tools/generate-vectors/{manager_config_test,main_test}.go`:
  `permissions` knob grammar (native/yolo per profile, default
  absent), native-only system direction, regenerated vectors (56
  manager-config-v2 cases), 5 new schema cases, gate cross-checks.
- `cli/curator.md` (informative permission-mode summary + examples),
  `UNRESOLVED_QUESTIONS.md` (Adopted decisions index),
  `CHANGELOG.md` (Unreleased entry), `COMPATIBILITY.md` (system
  schema-2 key list + four direction rules),
  `conformance/README.md` (direction-case note).
- Frozen v1 protocol schemas untouched; rc.5–rc.8 byte-frozen
  metadata untouched (rc.9 repinned by regeneration, pins only).

## Follow-up leaves per repository (one-line ACs for the orchestrator)

curator:

- F-C1 envprofile credential-link repairs — implement fix-first
  repairs per §7.4/§10.1 (unlink only a recorded symlink still
  targeting the declared store; `environment_credential_conflict`
  on regular file / unexpected target; never remove bytes). AC:
  shared→isolated leaves no stale link and regular-file-at-link
  refuses with the conflict diagnostic naming the path.
- F-C2 explicit credential migration step — inspect → plan → apply
  under the manager lock (inventory old marker, link targets, both
  Pi roots; preserve effective mode; operator resolves
  isolated→shared conflicts out of band; no secret copies). AC: a
  Pi wrong-target home migrates to `~/.pi/agent` with bytes
  preserved and mode intact, printing the plan before applying.
- F-C3 production-entry tests on temporary stores — cover both
  0017 hazards, the migration, and every repair refusal, each with
  a narrowing mutant proving the bound. AC: `go test`
  ./internal/envprofile/... passes with a mutant-killed row per
  refusal (conflict admitted ⇒ test fails).

curator-agent-launcher:

- F-L1 permission interface (SPEC + implementation + tests) —
  SPEC §§3/4.1–4.7/6, `curator-run-defaults-v2`,
  internal/{cli,mapping,composition,execution}, choice-5 negative
  rows with narrowing mutants, Q4 line + launch record, versioned
  provider-capability table; the launcher SPEC owns the choice-6
  capability encoding (0018 Compatibility). AC: `curator run` resolves
  flag>profile>global>built-in default, maps/refuses per the
  tables, and the fake-tool suite passes with every refusal
  mutation-killed.

curator-spec (follow-up spec revisions):

- F-S1 marker-schema extension for credential records (0017
  choice 5) — schema members + cases + vectors + gates, including
  the first spec vector rows pinning
  `environment_credential_conflict` /
  `environment_credential_unsupported` (no spec vector pins them
  today). AC: a marker carrying the named record validates, the
  conflict/unsupported vector rows exist, and `make validate`
  is green.
- F-S2 fragment permission-transport members + minimum token
  (0018 choice 7) — member names, token value, cases + vectors +
  gates. AC: a fragment carrying policy + lock engagement
  validates and legacy fragments still refuse would-be `yolo`.
- F-S3 fleet `isolated` enforcement policy revision (0017
  choice 6). AC: reviewed policy text merged; no knob
  reinterpretation.

agent-session-manager-spec:

- F-A1 ax-owned permission representation + versioned capability
  (0018 choice 2). AC: tracked yolo admitted only through the
  versioned capability with D5/D3.6 satisfied.

Operator-run experiments (not leaves; gate recorded choices):

- E1 macOS claude_code shared-store experiment (0017 choice 1
  a/b/c) on a disposable account; secrets never exported.
- E2 Codex keyring CODEX_HOME probe (0017 choice 4) on a
  disposable account; secrets never exported.

## Validation evidence

- `python tools/validate.py` → `validated 62 schemas and 1124
  vector files`, exit 0 (rerun after changes).
- `go test ./tools/...` → ok, exit 0 (rerun after changes).
- `python -B -m unittest test_validate.ManagerConfigVectorTests
  test_validate.SystemConfigV2SchemaTests` → 53 tests OK (rerun).
- `python -B -m unittest discover -s tools -p 'test_*.py'` → 579
  tests OK (576 baseline + 3 new gate tests), exit 0 (rerun after
  changes; ~19.6 min).
- `go run ./tools/generate-vectors -root .` run twice with
  identical bytes (reproducibility substance of
  regenerate-check; the make target itself cannot pass with
  uncommitted work by construction).
- `gofmt -l tools/` empty, `go vet ./tools/...` clean.
- Baseline (pre-change): validate.py green (62/1119), go test
  green, full unittest 576 tests OK.

## Notes and anomalies

- The vector generator was SIGKILLed (exit 137) on its first two
  executions this session after ~76 s with no writes; a pristine
  HEAD build and later retries completed in <1 s (exit 0).
  Attributed to transient environmental first-run behavior, not
  the change: the modified binary reproduces the pristine corpus
  byte-identical on unmodified content.
- `unittest discover` transiently rewrites one vector +
  `manifest.json` + `release/1.0.0-rc.9.json` per repinning case
  (`run_main_with_vector`, restored byte-identical in `finally`).
  Never run gates concurrently and never revert those files
  mid-suite.
- Directory-at-entry stays repair-re-linkable (pinned by
  `passthrough-directory-at-entry-detached` with
  `repair_relinks: true`); the adopted conflict rule therefore
  covers regular files and unexpected symlink targets only, with
  repair removing only an empty directory.
- System `python3` lacks `jsonschema`; gates ran via the
  project-local `.venv` (uv). CI installs requirements into
  system python instead.
- No `decisions/README.md` exists; the decisions index is the
  Filed proposals / Adopted decisions list in
  `UNRESOLVED_QUESTIONS.md`.

## Revision 2 (2026-09-21) — rework-1: F1–F3 + N1–N4

Verdict rev1 CHANGES_REQUESTED; schemas, generator, vectors, gates,
index, CHANGELOG, COMPATIBILITY accepted as correct and in scope.
Revision 2 fixes exactly the three blocking prose findings plus N1–N4,
with no schema, vector, or generator change (changed-path set still
exactly the rev1 160; F3 grep
`not an adoption|adopting revision amends|stays open question` over
`decisions/001[78]*.md` returns nothing).

- F1 (schema-1 marker prose): §8.2 bullet restored to baseline
  (strategy only); §7.4 "Credential record" reworded as the content
  of the marker revision choice 5 defines (follow-up), with "a
  schema-1 marker records `path` and `strategy` only and is never
  rewritten to add the record"; same sentence added to 0017 choice 5
  and this row 5; 0017 Compatibility now lists §8.2 as unchanged.
- F2 (codex `auto` isolation): `isolated` for `codex_cli` is
  available under `file` storage only; `auto` joins `keyring` in
  `environment_isolated_unsupported` — stated in §7.4 (passthrough
  table row, rule paragraph with the fail-closed rationale, matrix
  row), manager §12.4 (paragraph + table row), 0017 choice 4, and
  this row 4. A future revision may admit `auto` only where the
  manager proves the effective store is `file`, probe-gated (named
  as that revision's refinement; no new leaf).
- F3 (proposal-era sentences): 0017:54-56 and 0018:60-63 reworded to
  the adopted state; 0018:186 (fragment untouched; members are
  choice 7 / F-S2), 0018:276 (choice 7 follow-up), 0018:242 (closed
  marker set choice 7 fixes, item 5 points at choice 7); 0018:190
  "this proposal" → "this decision" (same class). The remaining
  "adopting revision" (item 5, fragments predating this adoption)
  is coherent as adopted text.
- N1: non-empty-directory repair names
  `environment_credential_conflict` (§7.4).
- N2: F-S1 now includes the first spec vector rows pinning
  `environment_credential_conflict` /
  `environment_credential_unsupported`.
- N3: 0018 row 7 marks the thin {CI, GITHUB_ACTIONS} set as an
  operator-confirmation item (TTY prong primary).
- N4: 0018 Compatibility states the launcher SPEC owns the choice-6
  capability encoding (F-L1); F-L1 mirrors it.

Validation reruns on the revision-2 tree (all green, own runs;
`make validate` covered as equivalent bounded chunks — validate.py +
full 579-test unittest + go test — plus gofmt/vet):

- `.venv/bin/python tools/validate.py` → `validated 62 schemas and
  1124 vector files`, exit 0.
- `go test ./tools/...` → ok; `gofmt -l tools/` empty;
  `go vet ./tools/...` clean.
- unittest chunk A (test_implementation_coverage, test_release_gate,
  test_verify_release_commit, test_verify_release_merge_policy):
  78 OK.
- unittest test_validate fast classes in three chunks: 83 + 196 +
  172 = 451 OK (one mixed invocation also named 7 classes from
  other files and reported 7 loader errors for those names only;
  the 7 classes had already passed in chunk A; no test failed).
- `WriteNofollowVectorTests` 12 OK; `DotfileManagersVectorTests`
  20 OK; `ReadFailureVectorTests` 9 + 9 = 18 OK (second half are
  in-memory mutation checks, milliseconds each — verified genuine
  by verbose rerun).
- Total 579/579, equal to rev1 and the reviewer's rerun count.
- Post-run `git status` path set is exactly the rev1 160 (transient
  per-case corpus rewrites restored byte-identical); whole-tree
  diff of the worktree against HEAD + rev1 patch applied in a
  scratch clone shows exactly the 4 intended prose files differing
  (`decisions/0017*`, `decisions/0018*`, `protocol/environments.md`,
  `profiles/manager.md`) — no schema, vector, or generator change.
- N2 re-verified on the revision-2 tree: no file under
  `conformance/`, `release/`, or `schemas/` names
  `environment_credential_conflict` /
  `environment_credential_unsupported`.
