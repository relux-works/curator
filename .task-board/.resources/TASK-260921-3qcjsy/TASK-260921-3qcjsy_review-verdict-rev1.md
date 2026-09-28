# TASK-260921-3qcjsy review verdict — revision 1 (CR-TASK-260921-3qcjsy-1)

Reviewer run RUN-260921-341588 (claude-opus-5/max), 2026-09-21.
Reviewed delta: base `802caee548ddc8b19408746d26c7972d39b39cc2` → candidate
tree `85565cd0babe1fa756770379efba06b34a8d6664` (160 paths). The story
worktree's working tree hashes to exactly that tree OID (temp-index
`git write-tree`), and every rerun below ran on a disposable clone
(`git archive` of the candidate tree, HEAD tree OID re-verified equal).

## Verdict: CHANGES REQUESTED → `to-dev`

The gate is green and the mechanical amendments (schemas, generator,
vectors, gates, index, CHANGELOG, COMPATIBILITY) are correct and in scope.
Three findings in the prose block acceptance: one normative statement the
frozen marker schema cannot satisfy (F1), one adoption choice that admits a
silently shared home against the fail-closed ruling (F2), and stale
proposal-era sentences that contradict `Status: adopted` (F3). Each is a
bounded text fix; no schema, vector, or generator change is needed for F1–F3.

## Blocking findings

### F1 — §8.2 schema-1 marker prose claims a record the frozen v1 marker schema rejects

- `protocol/environments.md:1887-1890` (§8.2 "Environment marker, schema 1")
  now says the schema-1 marker records the passthrough "credential record
  (effective mode, source role, backend, and provenance)";
  `protocol/environments.md:1480-1495` (§7.4 "Credential record") says
  "Every passthrough record carries" `isolation`, `strategy`, `source_role`,
  `backend`, `backend_version`, `provenance`, and that linkless strategies
  "record without a path".
- `schemas/v1/agent-environment-marker-v1.schema.json:68-80` (frozen v1)
  closes `passthrough.items` (`additionalProperties: false`) and requires
  `path` and `strategy`. §8.2 itself says readers "MUST NOT infer newer
  semantics from unknown fields", and Decision 0017 (Compatibility) says
  "Frozen v1 protocol schemas stay untouched; the marker-schema extension
  is a follow-up spec revision" (results.md F-S1).
- Reproduction (jsonschema 4.25.1 through `tools/validate.py`'s registry):
  `valid-passthrough-file-link.json` + `{isolation, source_role, backend,
  provenance}` on the entry → REJECTED ("Additional properties are not
  allowed"); `passthrough: [{"strategy": "ambient"}]` → REJECTED ("'path'
  is a required property").
- Fix: keep the §8.2 schema-1 bullet as before (strategy only) and word the
  §7.4 "Credential record" paragraph as the content of the marker revision
  the F-S1 follow-up defines (e.g. "From the marker revision Decision 0017
  choice 5 defines (follow-up), every passthrough record carries …; a
  schema-1 marker records `path` and `strategy` only and is never rewritten
  to add the record"); align the 0017 Compatibility sentence "§8.2
  (passthrough record)" and results.md accordingly.

### F2 — `isolated` for `codex_cli` under `auto` storage admits a silently shared home

- `protocol/environments.md:1460-1461` ("`isolated` remains available for
  `codex_cli` under `file` or `auto` storage"), `:1598` (matrix row),
  `profiles/manager.md:2764` (table row names `keyring` only), and Decision
  0017 choice 4 (`decisions/0017-environment-credential-modes.md:153`:
  "`auto` keeps file-link behavior (inert when the tool uses keyring)").
- On the document's own terms: the adopted assumption is that the Codex
  keyring identity is operator-global; under `auto` the tool uses the
  keyring when one is available (the text says the file-link is then
  inert). An `isolated` home under `auto` on a keyring host therefore links
  nothing and authenticates through the operator-global keyring — a
  silently shared home, which `protocol/environments.md:1442` forbids
  ("`environment_isolated_unsupported`, never a silently shared home").
  The brief's ruling for Q4 is the fail-closed reading; admitting `auto`
  is the fail-open one.
- Fix: `isolated` for `codex_cli` is available under `file` storage only;
  `auto` joins `keyring` in `environment_isolated_unsupported` (or is
  admitted only where the manager proves the effective store is `file`,
  probe-gated) — stated in §7.4 (paragraph and matrix row), manager §12.4
  (paragraph and table), 0017 choice 4, and results.md row 4.

### F3 — proposal-era sentences contradict `Status: adopted`

- `decisions/0017-environment-credential-modes.md:55` and
  `decisions/0018-curator-run-permission-interface.md:61`: "The verdict's
  FINAL recommended design (§3) is recorded below as a proposal, not an
  adoption."
- `decisions/0018-curator-run-permission-interface.md:184`: "(the adopting
  revision amends the fragment schema accordingly)" — `launch-env-fragment-v1`
  is untouched by this revision; the members are choice 7 / F-S2.
- `decisions/0018-curator-run-permission-interface.md:273`: "stays open
  question 7" — there is no open question 7; it is choice 7 (follow-up).
- `decisions/0018-curator-run-permission-interface.md:240`: "or an
  equivalent marker the adopting revision enumerates" — choice 7 fixed the
  closed set {`CI`, `GITHUB_ACTIONS`}; point item 5 at choice 7.
- Reproduction: `grep -n "not an adoption\|adopting revision amends\|stays
  open question" decisions/001[78]*.md`.
- Fix: reword each to the adopted state (AC: "amended normative text
  consistent").

## Non-blocking observations (fix in the same rework if cheap, else note)

- N1 `protocol/environments.md:1416-1418`: "a non-empty [directory] fails
  the repair instead of destroying its contents" names no diagnostic and is
  a bound the decision does not spell out (in its no-destroyed-bytes spirit;
  acceptable). Name the diagnostic (`environment_credential_conflict` fits
  "never removed"; `environment_repair_failed` is the generic).
- N2 No spec vector pins `environment_credential_conflict` or
  `environment_credential_unsupported` (the read-failure family has no
  regular-file-at-passthrough-entry case). 0017 defers production-entry
  tests to curator (F-C3); consider a spec-side row with F-S1.
- N3 0018 choice 7's closed marker set {`CI`, `GITHUB_ACTIONS`} is thin
  (Jenkins sets `BUILD_ID`/`JENKINS_URL`, Azure Pipelines `TF_BUILD`, not
  `CI`); the TTY prong is the primary detector, so silence does not fail
  open by default — an operator-confirmation item, not a defect.
- N4 results.md files the choice-6 capability encoding under F-L1 (launcher
  SPEC) while 0018's Compatibility calls it a "follow-up spec revision" —
  consistent if the launcher SPEC is the owning spec; say so.

## Verified (own reruns on the disposable clone)

- `python tools/validate.py` → `validated 62 schemas and 1124 vector files`,
  exit 0. `go test ./tools/...` → ok. `gofmt -l tools/` empty; `go vet
  ./tools/...` clean.
- `python -B -m unittest` split into bounded chunks (the single-call limit):
  `test_validate` fast classes 451 OK (312 s); other tool test files 78 OK (83 s); `WriteNofollowVectorTests` 12 OK (124 s); `DotfileManagersVectorTests` 20 OK (235 s); `ReadFailureVectorTests` 18 OK (473 s) — 579 tests, 0 failures, equal to the producer's 579; the clone's `git status` was clean afterwards (the per-case corpus rewrites restored byte-identical).
- Generator: `go run ./tools/generate-vectors -root .` then `git diff
  --exit-code -- conformance/v1 release/1.0.0-rc.{5..9}.json` → exit 0: every
  schema case, `index.json`, `manifest.json`, `vectors/manager-config-v2.json`
  and the rc.9 pins are byte-identical generator output, not hand edits.
- `release/1.0.0-rc.9.json` edit is allowed by the repository's own rules:
  rc.9 is the live suite pin regenerated at each revision (Makefile
  `regenerate-check` lists it as generated output; tag `v1.0.0-rc.12`
  message "Protocol version stays 1.0.0-rc.9 (release/1.0.0-rc.9.json
  regenerated at each revision)"; every recent landing repins it), while
  rc.5–rc.8 are byte-frozen and untouched here. No new release entry is
  needed for an Unreleased revision.
- Schema changes, enumerated and checked against the decision text:
  1. `manager-config-v2` `environments.permissions`: identifier-keyed map
     (same `common.schema.json#/$defs/identifier` grammar as `isolation`),
     values `native|yolo`, `default: {}` ← 0018 item 2 (`permissions.<profile>`,
     absent is a silent level) and §12.1 row.
  2. `system-config-v2` `environments.permissions`: values `native` only,
     and `locked` enum gains `environments.permissions` ← 0018 item 3
     (lockable only toward `native`; manager §1 list).
  3. No other schema changed; frozen v1 protocol schemas, `manager-config-v1`,
     `system-config-v1` untouched.
  Gates cross-check the §12.1 enum (`MANAGER_CONFIG_KNOB_ENUM_PATHS`) and
  the one-direction system enum (`SYSTEM_CONFIG_PERMISSIONS_ENUM_PATH`), with
  narrowing negatives that fail when the bound is widened
  (`test_widened_permissions_enum_fails`, `test_permissions_admitting_yolo_fails`,
  `test_permissions_without_a_closed_value_set_fails`,
  `test_section_12_2_drift_against_schema_fails`).
- Decision choices vs the brief: 0017 options 1–3 verbatim; Q1–Q7 recorded
  and fail-closed except Q4's `auto` clause (F2). 0018 adopted with the
  2026-09-16 amendment; Q1–Q7 as ruled (alias same increment; tracked-yolo
  refused until a versioned `ax` capability; unknown native forms refused;
  effective-native-policy line + launch record named; negative rows named;
  capability/token values "fixed by the implementing revision" as explicit
  follow-ups). Both Adoption-choices tables match results.md's tables in
  substance (results.md condenses wording).
- Normative amendments are the ones the decisions call for: environments
  §7.4/§7.7/§8.4.1/§10.1/§10.2/§10.4/§12.1/§12.2/§13, manager
  §1/§12.3/§12.4/§12.5, `cli/curator.md` informative summary. Scope
  additions: the COMPATIBILITY.md system-schema-2 key list corrected to the
  actual eleven keys and four direction rules (verified against the schema
  and `TestSystemConfigV2IsSchemaOnePlusTheLockableEnvironmentsKeys`) — a
  truthfulness fix within R3; and the N1 non-empty-directory bound.
- UNRESOLVED_QUESTIONS.md now carries the Adopted decisions index; CHANGELOG
  Unreleased entry present; every local link resolves
  (`validate_local_links` covers all `*.md`).
- Follow-up leaves map 1:1 to the deferrals: manager repairs + migration +
  tests (F-C1/F-C2/F-C3), launcher permission interface (F-L1), marker
  schema (F-S1), fragment transport + token (F-S2), fleet `isolated` policy
  (F-S3), `ax` capability (F-A1), macOS `claude_code` and Codex keyring
  experiments (E1/E2).

## Checklist disposition

- Tests green: yes (own reruns above).
- Implementation matches AC: no — "amended normative text consistent"
  fails on F1/F3; the fail-closed ruling fails on F2.
- Solution fits project architecture: no until F1 (prose ahead of a frozen
  closed schema) is resolved.
