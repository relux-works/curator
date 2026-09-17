# Evidence — TASK-260916-y4sa6s: spec-signer-allowlist-and-update-confirmation (E1)

Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-ioemse/worktree`
(branch `task-board/story/STORY-260916-ioemse`, base = `origin/main` `23dafa7`).
Role: doc-writer (normative spec text + conformance). No implementation code touched.

Finding read first: `docs/security-audit-2026-09.md` E1 (High) + Appendix B
(E1 confirmed: no signer verification anywhere; `profile update` prints only
`updated profile <name> (lock <hash>)`; `latest` is `*`).

## What changed, per file

Normative spec text:

- `decisions/0012-context-packages-and-semver-locks.md` — new
  `## Amendment (2026-09-17, E1)` in the Erratum style (original quoted
  verbatim, why-incomplete, evidence, added rule + normative home), with the
  two rules and the `latest` residual; Decision 8 cross-referenced as the
  update counterpart (§9.2). The three original passages carry
  `[Amendment 2026-09-17, item N]` markers; nothing else in the decision body
  changed.
- `protocol/environments.md`
  - §1.1: three new diagnostics rows (`context_source_unsigned`,
    `context_source_signer_rejected`, `context_source_signers_missing`).
  - §1.3: "record, not a signature" now points at the §1.4 check and §12
    posture; lock content unchanged.
  - §1.4: `latest`-stays-`*` residual with the strict-tag cross-reference;
    new "Signer verification" step (tag OR commit verifies before lock
    entry, every git selection incl. exact tag/revision, three mutually
    exclusive failures, fail-closed with old lock standing, never silently
    select lower, path/local never verified, empty allowlist admits none).
  - §9.2: update synopsis gains `[--confirm-system-delta]`; step 3 prints
    the delta lines, then surfacing rows, then the allowlist warning, then
    the confirmation gate; closed `lock-delta` line grammar (added/removed/
    moved with pins, `-` for absent version, one LF); exact trigger
    (system-module inventory = sorted path/selector/bytes; MCP
    command/args/env_names with set comparison); warn-first revisions A
    (warn `profile_update_system_delta`, proceed) and B (refuse
    `profile_update_confirmation_required` unless flagged); per-run flag,
    no knob pre-confirm, `--all` confirms the run and stops at the first
    refusal, reinstall runs the same gate.
  - §9.7: two new diagnostics rows.
  - §12: per-source posture rows (`enforced` + verified signer,
    `unconfigured`, `required-missing`) + machine `require_source_signers`;
    re-verify-from-local-state without fetching (`unknown` when unprovable);
    pin-fails ⇒ non-current; unconfigured/required-missing stay current.
  - §12.1: `source_signers.<source>` + `require_source_signers` rows with
    defaults `{}` / `false`; closed entry prose (ssh key types, 40
    uppercase hex fingerprints, no cross-fields, absent ≠ empty).
  - §12.2: lockable set grows to ten keys; `require_source_signers` locks
    only to `true`; locked `source_signers` merges per source (system wins
    named sources with the §1 warning, machine adds unnamed ones).
  - §13: new conformance surfaces + revision A/B conformance sentences.
- `profiles/manager.md` §1: locked-key list + direction sentence +
  per-source-merge exception to whole-replacement (E2 precedent).
- `cli/curator.md`: `profile update` row gains the flag + gate clause
  (`--allow-lossy` precedent for per-operation flags).
- `CHANGELOG.md`: Unreleased → Added entry "E1: …" naming the finding, both
  knobs, both rollout revisions, and the `latest` residual.

Schemas + generator (spec-repo tooling, in scope per the brief):

- `schemas/v1/manager-config-v2.schema.json`: closed `$defs/sourceSigner`
  (ssh/gpg if/then/else, key-type allowlist, fingerprint grammar) +
  `source_signers` (default `{}`) + `require_source_signers` (default
  `false`). `context-lock-v1` untouched.
- `schemas/v1/system-config-v2.schema.json`: `source_signers` by reference,
  `require_source_signers` as `{"enum": [true]}`, both appended to the
  `locked` enum in §12.2 order.
- `tools/generate-vectors/manager_config.go`, `system_config.go`: defaults,
  every-knob fixtures, 15 manager + 4 system schema examples, 4 new
  manager-config vectors.
- `tools/generate-vectors/manager_config_test.go`, `main_test.go`: knob
  list, closed-def pin, direction pins.

Conformance:

- `conformance/v1/vectors/environments-source-signers.json` (new, static
  like `environments-env-passthrough.json`): 17 verification cases (ssh
  tag, gpg commit, either-suffices, unsigned/wrong-signer/invalid refused,
  no-allowlist accepted, require refused/accepted, path never verified,
  no-silent-fallback, empty-allowlist ×2, exact tag, revision + 2
  negatives), 6 merge cases, 5 posture cases, 17 delta cases (lines,
  trigger shapes, flag, 2 negatives), 2 `--all` cases.
- `tools/validate.py`: `MANAGER_CONFIG_KNOB_DEFAULT_PATHS` +2, directional
  `require_source_signers` gate, new `validate_environments_source_signers_vectors`
  (independent recomputation of verdicts, merges, rows, lines, triggers,
  both revisions; allowlist maps validated against the real schema).
- `tools/test_validate.py`: new `SourceSignersVectorTests` (32 tests),
  ten-key §12.2 pin, direction + default-drift tests.
- Regenerated: `conformance/v1/manifest.json` (+20 files),
  `schema-cases/index.json`, config schema-cases,
  `vectors/manager-config-v2.json`, `release/1.0.0-rc.9.json` pin. Every
  other existing vector is byte-identical (verified: only
  `manager-config-v2.json` + the new file differ under `vectors/`).

## Validation transcript

Shell: `bash` in the worktree. Python via a venv outside the tree
(`/tmp/curator-spec-venv`, `jsonschema==4.25.1` per `requirements-dev.txt`;
no venv was present and `.venv` is not gitignored, so the tree stays clean).

- `PATH=/tmp/curator-spec-venv/bin:$PATH python3 tools/validate.py`
  → `validated 60 schemas and 1087 vector files`, exit 0.
- `go test ./tools/...` → `ok .../tools/generate-vectors 2.400s`, exit 0.
- Regenerate idempotence: `go run ./tools/generate-vectors -root .` twice;
  `manifest.json` + `1.0.0-rc.9.json` hashes unchanged; vector churn confined
  to `manager-config-v2.json` + new file. (A literal `make regenerate-check`
  diffs against HEAD, so it trivially shows this uncommitted revision; the
  idempotence run is the cleanliness proof.)
- Targeted: `test_validate.SourceSignersVectorTests`,
  `ManagerConfigVectorTests`, `SystemConfigV2SchemaTests` → 82 tests, OK.
- Full: `python -B -m unittest discover -s tools -p 'test_*.py'` →
  `Ran 325 tests in 416.188s / OK`, exit 0.

## Deliberately out of scope

Implementation (`TASK-260916-1zgucp`), registry-side provenance, proposals
0014–0018, E2 (landed), tags/releases. Within the revision: the lock schema
(`context-lock-v1`) is unchanged per the brief.

## Findings for the board (spec gaps observed, not edited)

1. The settled MCP trigger is exactly `command`/`args`/`env_names`. An `http`
   declaration's `url` change therefore never triggers confirmation, and an
   MCP `environments` selector change never does either. Both are
   security-relevant mutations of the same declaration; the board may want a
   follow-up widening the trigger. No vector pins the silent behavior — the
   gate recomputes exactly the specified triple.
2. `env status` names the verified signer only when the manager's local
   source state reproduces the verification (`unknown` otherwise, staying
   current). If the shipped manager retains no tag/commit objects, every
   `enforced` row reports `unknown` — honest but weak. A retention rule for
   the local source state would strengthen this; it is out of this brief.
3. Reinstall-as-update runs the confirmation gate but `profile install`
   takes no `--confirm-system-delta`, so a triggered reinstall under
   revision B refuses and the operator completes it via `profile update
   --confirm-system-delta`. This is the fail-closed reading of the existing
   "re-resolves exactly as `profile update`" sentence; the manager task
   should confirm the UX.

---

# Revision 2 (rework per `TASK-260916-y4sa6s_review-verdict-rev1.md` R1–R5 plus adopted R6–R7)

Worktree and base unchanged (curator-spec Story worktree on branch
`task-board/story/STORY-260916-ioemse`, base `origin/main` `23dafa7`).
Every verdict "Pass" row stays byte-identical except where a correction
below touches it. Line numbers are post-edit.

## Per-correction file:line

- **R1 — MCP trigger = complete canonical declaration.**
  `protocol/environments.md:1834-1843`: the moved-`mcp` trigger is now a
  difference in the CCJ-1 bytes (`registry.md` §1) of the declaration
  object as the lock/materialization reads it — closed field list
  `transport`, `command`, `args`, `env_names`, `url` of an `http`
  declaration, the `environments` selector — with no field narrowed out
  and array order significant. `tools/validate.py:5273-5322` (`e1_check_snapshots`
  closed six-key declaration shape with per-transport consistency;
  `e1_mcp_changed` at :5325 compares `ccj1_bytes`). Vectors
  `conformance/v1/vectors/environments-source-signers.json:985-1071`:
  `changed-mcp-url` + `changed-mcp-url-confirmed` (http `url`-only change;
  A warning / B refusal / flagged acceptance), `changed-mcp-selector` +
  `changed-mcp-selector-confirmed` (selector-only change; same triple);
  `mcp-env-names-reorder-silent` retargeted to
  `mcp-env-names-reorder-triggers` (:1073, now triggers — CCJ-1 preserves
  array order, no set-comparison exception survives). All 13 pre-existing
  MCP snapshots gained the `url`/`environments` keys. Tests:
  `tools/test_validate.py:2215,2291-2322` (reorder/url/selector silenced
  rejected; URL-difference removal flips the verdict). §13 case list
  updated (`protocol/environments.md:2784-2791`).
- **R2 — reinstall takes the flag; revision-A migration hint.** The
  sentence "`profile install` takes no `--confirm-system-delta`" is
  deleted. `protocol/environments.md:1741-1746` (§9.1 reinstall path),
  `:1865-1871` (§9.2 identical per-invocation semantics, no
  pre-confirmation), `cli/curator.md:30` (install synopsis +
  gate clause). New `reinstall_cases` family pinned by vectors :1371-1418
  and the gate (`tools/validate.py:5596-5619`, reuses the §9.2
  recomputation). Revision A's `profile_update_system_delta` MUST carry
  the migration hint `revision B refuses with
  profile_update_confirmation_required unless --confirm-system-delta is
  given` (§9.2 :1847-1853, §9.7 diagnostics row :2147, §13 conformance
  sentence :2802-2805); the hint is pinned in every `revision_a`
  expectation (`e1_revision_outcomes`, `tools/validate.py:5385-5395`,
  `E1_HINT_DELTA` :4920) including `--all` profiles and the revision-a
  negative-claim comparison, with negative tests at
  `tools/test_validate.py:2324-2352` (hint dropped, hint without trigger,
  reinstall refusal flipped, reinstall mislabelled; the repaired-silence
  test now also repairs the hint so the rejection is genuine).
- **R3 — fingerprint grammar exact.**
  `schemas/v1/manager-config-v2.schema.json:370-375`: `fingerprint` keeps
  `pattern: ^[0-9A-F]{40}$` and gains `minLength`/`maxLength` 40, so the
  41-character trailing-newline form is rejected
  (`system-config-v2.schema.json:58` takes this grammar by reference —
  one fix covers both). Generated negative schema cases
  `invalid-source-signers-fingerprint-newline` in both configurations
  (`tools/generate-vectors/manager_config.go:213`,
  `system_config.go:103`; 41 characters, trailing `\n`, registered in
  `schema-cases/index.json` and `manifest.json`, exercised by
  `validate_schemas` through the real `Draft202012Validator` entry).
  Explicit real-entry test
  `tools/test_validate.py:3311` (40-char control accepted, 41-char
  newline form rejected, both schemas). §12.1 prose now says "exactly 40
  uppercase hex characters" (:2677); §13 names the exact-length cases.
- **R4 — revision selection needs commit evidence.**
  `tools/validate.py:5094-5098`: the candidate checker refuses
  tag-only evidence (`tag_signature` non-null) for an exact `revision`
  selection; tag-OR-commit semantics for tag selections unchanged. The
  pinned artifact is the negative test
  `tools/test_validate.py:2354-2361`
  (`test_revision_selection_tag_only_evidence_rejected`), which replays
  the verdict's exact forgery (commit signature moved to `tag_signature`)
  and asserts the gate raises. No vector case carries that name on
  purpose: a malformed fixture raises in the gate by design, so it cannot
  live in the published vector (which must pass); §1.4 prose already
  stated the rule and is unchanged.
- **R5 — curator LOGBOOK delta removed.** Ran `git checkout -- LOGBOOK.md`
  in the curator Story worktree
  (`/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260916-ioemse/worktree`);
  `git status --short` there is now empty, so the runtime's curator Change
  Request delta for this spec task is empty again. Findings live in this
  evidence and task notes only. No LOGBOOK.md was edited in revision 2.
- **R6 — SSH key identity ignores the comment.** `protocol/environments.md:2677-2686`
  (§12.1): identity is key type plus base64 material; the trailing
  comment is not part of it. Oracle `e1_signer_identity`
  (`tools/validate.py:5024-5034`) used for allowlist matching, the
  verified-signer membership, and `signers_seen`, which now reports
  canonical identities (comment-free; 7 existing expectations updated).
  Vectors :419-475: `ssh-same-key-different-comment-accepted`
  (same material under another comment accepted) and
  `ssh-different-material-rejected` (same comment, different material
  refused). Tests `tools/test_validate.py:2363-2390`. The Go generator
  never compares signer keys (it only emits fixtures), so no generator
  change was needed — verified by inspection.
- **R7 — update-confirmation posture.** `protocol/environments.md:2564-2570`
  (§12 status sentence) and the new row paragraph :2606-2614: `env status`
  reports the active update-confirmation revision with its behaviour;
  informative, never non-current, no pre-confirmation knob. Spelling
  decision: `A-warning` / `B-flip` — the revision letters plus the
  behaviour words the §9.2 rollout labels use ("warning release" /
  "flip release"), per the brief's "use the spelling the rollout section
  uses". Pinned by the new `confirmation_posture_cases` family (vectors
  :1420-1434, gate `tools/validate.py:5622-5633`,
  `e1_expected_confirmation_posture` :5396) with negative tests at
  `tools/test_validate.py:2392-2405`.

Also updated: `CHANGELOG.md:10-50` (Unreleased E1 entry extended:
canonical-declaration trigger, reinstall flag, migration hint, ssh
identity, confirmation revision, exact-length fingerprint, new vector
families).

## Validation transcript (revision 2)

Shell `/bin/bash`, workdir the curator-spec Story worktree,
`PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH"`,
`set -o pipefail`. The full Python suite (~340 tests, ~6 min) ran as
bounded split calls instead of one `make validate` invocation; every gate
below ran green with exit code 0, quoted verbatim:

- `python3 tools/validate.py` → `validated 60 schemas and 1089 vector
  files`, exit 0 (1089 = 1087 + the 2 new R3 schema cases).
- `go test ./tools/...` → `ok
  github.com/relux-works/curator-spec/tools/generate-vectors 1.032s`,
  exit 0.
- `python3 -B -m unittest test_implementation_coverage test_release_gate
  test_verify_release_commit test_verify_release_merge_policy` (from
  `tools/`) → `Ran 78 tests ... OK`, exit 0.
- `test_validate.py` in three class groups (from `tools/`) →
  `Ran 110 tests ... OK`, `Ran 75 tests ... OK` (incl. all 46
  `SourceSignersVectorTests`), `Ran 76 tests ... OK`, exit 0 each.
  Total: 78 + 110 + 75 + 76 = 339 tests, 0 failures (325 in rev1 + 14
  new: 13 E1 gate negatives + 1 R3 real-entry test).
- Generator idempotence: `go run ./tools/generate-vectors -root .`
  twice; `manifest.json`, `schema-cases/index.json`,
  `release/1.0.0-rc.9.json` hashes unchanged (`sha256sum -c` OK).
- Baselined regenerate-check (reviewer's `GIT_INDEX_FILE` form, temp
  index staged from the candidate tree, generator re-run, then `git diff
  --exit-code -- conformance/v1 release/1.0.0-rc.*.json`) → exit 0.
  The literal `make regenerate-check` was not claimed green: it diffs
  against HEAD and is baseline-sensitive on this uncommitted tree (exit 2
  there, as the verdict established) — the baselined form above is the
  zero-drift proof.
- `git diff --check` → clean, exit 0.
- Scope: `git status --short` lists 140 paths — spec, schema, vector,
  manifest/index, CHANGELOG, and spec-repo tooling only;
  `schemas/v1/context-lock-v1.schema.json` unchanged (0 diff lines);
  under `vectors/` only `environments-source-signers.json` (new) and the
  regenerated `manager-config-v2.json` differ from `origin/main` — every
  other pre-existing vector byte-identical. Rev2's regenerated delta is
  exactly the 2 new R3 schema-case files plus index/manifest/rc.9 pin
  updates; no existing generated instance content changed in rev2.

## Honest curator-delta statement

The curator Story worktree (`curator/.temp/STORY-260916-ioemse/worktree`)
has zero modified/untracked paths after the R5 revert; the only curator
artifact of this run is board state (notes/resources/checklist), not
repository bytes. No implementation code touched anywhere.

---

# Revision 3 (rework per `TASK-260916-y4sa6s_review-verdict-rev2.md`: R1 remainder — optional-field absence in MCP snapshots)

Review verdict rev2 read first; R2–R7 stay closed and byte-identical except
where this correction touches them. One focused correction; the only files
this session modified are the E1 vector, its manifest/rc.9 pins,
`tools/validate.py` and `tools/test_validate.py`. No normative prose, schema,
CHANGELOG, CLI, generator, or §13 change. Line numbers are post-edit.

## The gap

`e1_check_snapshots` required every MCP snapshot to carry exactly the six
keys with both optional arrays always present, so the fixtures could not
express field absence that §9.2 compares byte-for-byte — a comparator
padding absent optional arrays with `[]` survived the whole E1 gate (0
cases with absent optional fields).

## Per-file file:line

- `conformance/v1/vectors/environments-source-signers.json`
  - All 23 pre-existing MCP snapshots rewritten to presence-preserving
    real declaration objects: no `null` padding of transport-irrelevant
    fields, `env_names`/`environments` present only when set (empty
    `environments: []` became absent — it is schema-invalid under
    `agent-mcp-v1`, `minItems: 1`; absent keeps meaning every adapter per
    §2.2/§9.2). Key order and compact single-line style preserved;
    verdicts untouched.
  - 8 new delta cases (A warning + migration hint / B refusal / flagged
    acceptance each, same shape as the rev2 url/selector pairs):
    `absent-selector-to-present` (:1095) + `-confirmed`,
    `present-selector-to-absent` (:1139) + `-confirmed` (selector absent
    against `["codex_cli"]`, mirroring the reviewer's reproduction),
    `absent-env-names-to-empty` (:1183) + `-confirmed`,
    `empty-env-names-to-absent` (:1227) + `-confirmed` (absent against
    explicit `[]`). URL-only, selector-only and array-order cases kept.
- `tools/validate.py`
  - `E1_DELTA_CASES` (:4985–4992): the 8 new names; inventory stays exact.
  - `e1_check_snapshots` (:5261): new required `mcp_validator` parameter;
    the six-key shape guard is gone. Per-transport consistency checks
    stay (presence-based: stdio carries no `url` key, http carries no
    `command`/`args` keys); optional arrays checked only when present;
    every snapshot is then validated as the `server` of a minimal
    manifest through the real `Draft202012Validator` agent-mcp-v1 entry
    (registry-backed, same pattern as the detector gate), preserving
    actual field presence. `e1_mcp_changed` (CCJ-1 comparison) is
    byte-identical — the settled canonical-byte rule is untouched.
  - `mcp_validator` construction (:5466) and the three call sites
    (:5545 delta, :5603 `--all`, :5636 reinstall).
- `tools/test_validate.py`: `test_absent_optional_padding_narrowing_fails_on_new_cases`
  (:2406) — the required narrowing test: the CCJ-1 rule triggers on all
  four new scenarios with trigger + migration hint + B refusal pinned;
  the absent-padding comparator returns False on both env_names
  absent/present cases; padding those fixtures and running the published
  gate raises on `section 9.2 delta rule` (asserted message), proving
  the suite catches the narrowed implementation.
- Regenerated: `conformance/v1/manifest.json` (E1 vector hash
  `sha256:fb59e44c…d252d46f8d9c`, verified equal to the file bytes),
  `release/1.0.0-rc.9.json` (manifest pin
  `sha256:8c36fc67…d5ddf4ac9eb13521`, verified equal to the manifest
  bytes). No schema-case or index content change.

## No §9.2 prose change — existing sentence quoted

Item 4 allowed a clarifying half-sentence only if needed; it is not
needed. `protocol/environments.md:1837-1843` already states the rule the
new fixtures exercise: "Any byte difference triggers … and no field is
narrowed out; absent (an `http` declaration carries no `command` or
`args`) differs from present." The frozen `agent-mcp-v1` schema is
unchanged, as required.

## Validation transcript (revision 3)

Shell `/bin/bash`, workdir the curator-spec Story worktree,
`PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH"`.
Full Python suite ran as bounded split calls; every gate below ran green
with exit code 0, quoted verbatim:

- `python3 tools/validate.py` → `validated 60 schemas and 1089 vector
  files`, exit 0.
- `go test ./tools/...` → `ok
  github.com/relux-works/curator-spec/tools/generate-vectors 0.758s`,
  exit 0.
- `python3 -B -m unittest test_implementation_coverage test_release_gate
  test_verify_release_commit test_verify_release_merge_policy` (from
  `tools/`) → `Ran 78 tests ... OK`, exit 0.
- `test_validate.py` in four class groups (from `tools/`) → `Ran 44
  tests ... OK`, `Ran 66 tests ... OK`, `Ran 76 tests ... OK` (incl. all
  47 `SourceSignersVectorTests`), `Ran 76 tests ... OK`, exit 0 each.
  Total: 78 + 44 + 66 + 76 + 76 = 340 tests, 0 failures (339 in rev2 + 1
  new narrowing test).
- Generator: `make regenerate` twice; `manifest.json`, `rc.9.json`,
  `schema-cases/index.json` hashes unchanged on the second run
  (`sha256sum -c` OK). Baselined regenerate-check (disposable copy of
  the candidate tree committed as its own baseline, literal `make
  regenerate-check` there) → exit 0. The literal in-worktree
  `make regenerate-check` was not claimed green: it diffs against HEAD
  and is baseline-sensitive on this uncommitted tree — the baselined
  form above is the zero-drift proof.
- `git diff --check` → clean, exit 0.
- Probes through the real gate entry (throwaway, `/tmp`, quoted here):
  re-adding `environments: []` to a new snapshot is refused with `an
  mcp snapshot is a declaration valid under agent-mcp-v1 (...)`;
  re-adding `"url": null` is refused with `a stdio declaration carries
  no url`; absent-selector snapshots in delta fixtures: 26 (was 0);
  absent-`env_names` snapshots: 4 (was 0).
- Scope: `git status --short` lists the same 140 paths as rev2 — spec,
  schema, vector, manifest/index, CHANGELOG, and spec-repo tooling only;
  no untracked files; `tools/__pycache__` removed after the runs. E1
  case count is now 65 (19 verification + 6 merge + 5 posture + 29 delta
  + 2 `--all` + 2 reinstall + 2 confirmation posture).
- Patch: `TASK-260916-y4sa6s_spec-patch_rev3.patch`
  (`git add -N . && git diff origin/main`), 11577 lines,
  `git patch-id --stable`
  `af964a62a8b2998fb54fff35226ecaa00ddf1063`,
  sha256 `77336b0d29a22ed83453d622dc3333a57104b24d0576c6d5ebe3f92e9699cd85`.
  It contains the 8 new cases and the gate/test changes; the old six-key
  guard message is gone (0 occurrences).

## Honest curator-delta statement (revision 3)

The curator Story worktree
(`curator/.temp/STORY-260916-ioemse/worktree`) `git status --short` is
empty (verified this session); no LOGBOOK.md touched anywhere. No
implementation code touched anywhere.
