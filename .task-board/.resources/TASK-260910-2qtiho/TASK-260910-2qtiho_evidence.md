# Evidence — TASK-260910-2qtiho spec-hardened-defaults-profile (S1+S3), rev1

Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2qmrb8/worktree`
(branch `task-board/story/STORY-260910-2qmrb8`, base `e8b53a0`, own HEAD — NOT `origin/main`).
Role: doc-writer (normative spec text only). No implementation code touched.
No LOGBOOK.md change anywhere (spec repo has none in the diff; curator repo untouched).

## What changed per file (24 files, +3992/−32 vs HEAD)

Normative prose:

- `profiles/manager.md`
  - §1: `security_posture` top-level knob paragraph (schema-2 only,
    `permissive`/`hardened`, revision-dependent default; schema-1 runs
    `permissive`); `locked` gains `security_posture`, lockable only to
    `hardened` (fails closed otherwise).
  - new §7.1: hardened-defaults profile — closed effective-policy table
    (audit mode/policy, source allowlist, MCP allowlist, passthrough,
    transitive modules, signer requirement, unreachable-registry gate),
    lock > explicit > profile-default precedence with the three refusals
    as the profile's meaning, warn-first revisions A (knob + once-per-op
    `security_posture_permissive` warning with migration hint) and B
    (default flips to `hardened`), closed 2-code §7.1 diagnostics table,
    downstream execution binding.
  - §10: `security_posture` header row + closed 4 manager-gate rows with
    value + provenance; `--check` hardened-contradiction rule.
  - §12.7: `env status` posture rows; hardened-contradicting state is
    non-current.
- `protocol/registry.md` §4: S3 residual paragraph (revocation is
  network-dependent under advisory policy, up to the offline grace),
  `registry_unreachable_during_install` gate notice (warning permissive /
  error hardened, names artifacts), closed 1-code §4 table, per-query
  warning unchanged. §8: grace cross-reference.
- `SECURITY.md`: hardened-profile paragraph ("Security model"), S3
  residual paragraph ("Registry security state").
- `protocol/environments.md`: §2.2 escalation rule + null pointer; §2.1
  and §9.7 tables qualify the warning (permissive) and add the hardened
  error row; §10.3 explicit-null refusal rule; §10.4 new
  `passable_env_names_unbounded_refused` row; §12 header + 8
  environments-gate rows, hardened-contradiction currency, warning-list
  qualification; §12.1 prose pointer (table untouched); §13 family line.

Schemas (`schemas/v1/`):

- `manager-config-v2.schema.json`: top-level `security_posture`
  enum [`permissive`,`hardened`], default `permissive` (revision A).
- `system-config-v2.schema.json`: top-level `security_posture` enum
  [`hardened`]; `locked` enum gains `security_posture` after `audit`.

Generator + tests (`tools/generate-vectors/`):

- `manager_config.go`: 4 top-level schema examples (2 valid, value/type
  negatives). `system_config.go`: 3 examples (valid hardened-locked,
  permissive-direction and type negatives). Bases untouched, so all
  pre-existing schema-case files are byte-identical.
- `manager_config_test.go`, `main_test.go`: property sets +1, locked enum
  +1, hardened-direction assertion.

Vectors (new family, handwritten like shell-hook-trust):

- `conformance/v1/vectors/security-posture.json`: 17 cases — revision-A
  default + warning, revision-B flip, effective defaults per posture,
  explicit-beats-profile, locked-beats-explicit (value + posture), the
  three refusals (explicit-empty sources, absent-MCP-with-declarations,
  explicit-null passthrough), no-declaration warning positive,
  absent-passthrough-follows-s4-warn positive, unreachable warning vs
  error, schema-1-is-permissive, status-check contradiction, flipped
  revision rows. Every case pins profile, effective values + provenance,
  diagnostics, outcome, and full posture rows.

Validator + tests (`tools/`):

- `validate.py`: new `validate_security_posture_vectors` gate —
  closed vocabularies, per-case scenario pins (rule 7), full derivation
  of profile/values/diagnostics/outcome/rows from inputs; every code
  must appear ≥ once. Extended `validate_system_config_v2_schema` for
  the top-level member + locked enum. Dispatched in `main()`.
- `test_validate.py`: `SecurityPostureVectorTests`, 15 tests (pass +
  14 negatives incl. two rule-7 same-name rewrite substitutions).

Generated (via `make regenerate`):

- 7 new schema-case files, `security-posture.json` registered in
  `manifest.json`, `schema-cases/index.json` + `release/1.0.0-rc.9.json`
  pins updated. No other conformance/release file changed — all
  pre-existing vectors byte-identical (`git diff HEAD --stat` shows only
  the files above).

CHANGELOG: Unreleased entry "S1/S3: …" naming both rollout revisions,
the residual, and the new codes.

## Why

Brief `TASK-260910-2qtiho_brief.md` (S1+S3) under
`remediation-spec-producer-rules.md`. Deliberately last of wave 3 so the
posture report covers every gate landed so far.

## Validation transcript (all exit 0; venv python)

Shell: `/bin/sh` from the worktree; python =
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python`
(jsonschema 4.25.1; system `python3` has no jsonschema, so `make validate`
was executed as its three legs with the venv interpreter):

- `python tools/validate.py` → `validated 62 schemas and 1103 vector
  files`, exit 0 (24s)
- `python -B -m unittest discover -s tools -p 'test_*.py'`, sharded
  (runner backgrounds >~3min calls; shards are the same suite):
  test_validate 54+128+49+98+15(SecurityPosture)+29(main-based re-run
  post-fix) OK; test_implementation_coverage 36 OK; test_release_gate 32
  OK; test_verify_release_commit 5 OK;
  test_verify_release_merge_policy 5 OK — every shard exit 0
- `go test ./tools/...` → `ok .../tools/generate-vectors`, exit 0
- `go run ./tools/generate-vectors -root .` → exit 0
- `make regenerate-check` (after staging generated files) → exit 0
  (idempotent; `git diff --exit-code` clean on generated paths)

## Deliberately out of scope

- Implementation (`TASK-260910-1sapuy`, manager posture work), S2
  bootstrap checkpoint, changing any existing gate's own default, tags,
  proposals 0014–0018. No `manager-config-v2.json` vector cases added
  (schema cases + the new family pin the knob; keeps that file
  byte-identical). No `cli/curator.md` change (brief does not name it;
  its status row list is non-exhaustive).

## Decisions needing no reopen but worth naming

1. Provenance gains a 4th value, `shipped`, for revision/always-on gates
   (hook trust, passthrough, provider roots, update confirmation, store
   boundary, write discipline): the brief's three provenances
   (profile-default/explicit/locked) cannot describe a shipped revision.
   Stated in manager §10, pinned in vectors.
2. Header row spells the knob (`security_posture`, per the brief); the
   12 per-gate names use the hyphenated row vocabulary. Fixed during the
   run after a consistency grep caught `security-posture` in vectors.
3. Schema-1 managers run `permissive` under BOTH revisions (v1 schemas
   are byte-frozen without the knob). Pinned by
   `schema1-machine-is-permissive` (rollout B).
4. `security_posture_permissive` fires whenever the effective posture is
   `permissive` (both revisions); only the default flips in B.
5. Absent `passable_env_names` under hardened keeps following the shipped
   S4 profile (only explicit `null` is refused) — S4's own revision
   track is untouched per the brief. Pinned by
   `hardened-absent-passthrough-follows-s4-warn`.

---

# Revision 2 (rework after `changes_requested`, 2026-09-18)

Rework brief `TASK-260910-2qtiho_rework-rev2.md` (two verdict findings F1/F2;
everything else passed and is kept byte-identical). Base still `e8b53a0`;
worktree and role unchanged. EMPTY curator delta re-verified
(`git -C <curator-story-worktree> status --short` empty, `git diff --stat`
empty). No LOGBOOK.md change anywhere; no implementation code touched.

## F1 — every required scenario pinned to posture and precedence (rule 7)

`tools/validate.py`, `SECURITY_POSTURE_SCENARIOS`: each of the 17 scenarios
now pins its rollout revision, `machine.schema_version`, the
`security_posture` value or its absence, the `machine.audit` /
`allowed_sources` / `environments` presence, values and absences, the
exact `system.locked` list or its absence plus every system member
absence, all four `shipped_revisions`, the full operation shape (kind,
MCP declarations, unreachable registries and artifacts as non-empty or
exactly empty), and — new pin keys `effective.profile` /
`effective.profile_source` — the posture the inputs MUST resolve to
under the §7.1 merge rules. `_posture_check_scenario` takes the resolved
profile/profile-source (resolve now runs before the scenario check), so
a same-name rewrite that flips the effective posture or its precedence
source fails even when its expected block is internally consistent.

`tools/test_validate.py`, `SecurityPostureVectorTests`: two existing
tests updated to the renamed row fields (both status outputs); four new
tests — `test_mcp_refusal_rewritten_as_permissive_warning_fails` (the
reviewer's probe shape: explicit `permissive` + recomputed profile,
effective, diagnostics, outcome and both row arrays via
`_posture_resolve` / `_posture_diagnostics` / `_posture_status_rows`),
`test_revision_a_default_swapped_with_locked_case_fails`,
`test_revision_b_flip_swapped_with_schema1_case_fails`, and the
exhaustive `test_no_whole_case_substitution_survives` (every other
case's whole body under every name: 272/272 MUST raise).

Probe result: the MCP rewrite is refused
(`... must leave machine.security_posture absent`);
whole-case substitutions 272/272 rejected (rev1: 269/272 — the two
branch-erasing survivals plus the benign warning-once case now all
rejected via the empty-vs-non-empty unreachable pins).

## F2 — the full twelve-gate inventory on BOTH status commands

One shared closed vocabulary: top-level `status_gates` (the twelve gates
in order: `hook-trust`, `registry-policy`, `audit-mode`,
`source-allowlist`, `env-passthrough`, `transitive-system-modules`,
`provider-trust-roots`, `source-signers`, `update-confirmation`,
`store-boundary`, `write-discipline`, `mcp-package-allowlist`) plus
`schema1_gates` (the four-gate manager subset); per-case
`curator_status_rows` / `env_status_rows` (renamed from
`manager_rows`/`environments_rows`). Schema-2 cases pin BOTH outputs to
the header row plus all twelve gates; the schema-1 case pins
`curator_status_rows` to the header plus the four manager-subset rows
and `env_status_rows: null`. Provenance spellings unchanged
(`profile-default`, `explicit`, `locked`, `shipped`).

- `profiles/manager.md` §10: the closed twelve-row table (gate, value,
  provenance) carried identically by `curator status` and `env status`
  when the environments capability exists ("No other posture row exists
  on either command"); schema-1 bound (header + first four rows only,
  no `env status` posture section). §12.7: the `env status` posture
  paragraph now names the twelve §10 rows.
- `protocol/environments.md` §12: `env status` carries the twelve
  posture rows of the manager §10 table, "identical in gate name,
  order, value, and provenance", with the closed twelve-name list
  ("No other `env status` posture row exists"); the per-source and
  per-profile row clarifications kept.
- `tools/validate.py`: `SECURITY_POSTURE_STATUS_GATES` /
  `SECURITY_POSTURE_SCHEMA1_GATES`, shared `_posture_status_rows`
  derivation, both-output row checks, `status_gates`/`schema1_gates`
  vocabulary checks.
- `conformance/v1/vectors/security-posture.json`: transformed by
  `/tmp/posture_rev2_transform.py` (dump settings proven byte-identical
  on the untransformed file first); inputs untouched, only the
  vocabulary fields, scope note and row arrays changed.
- `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`:
  regenerated digests only.

Byte-identity vs rev1: per-file diff comparison of `git diff HEAD`
against `TASK-260910-2qtiho_spec-patch_rev1.patch` — 17 of 24 files
byte-identical (CHANGELOG, SECURITY, registry, both schemas, all four
generator files, index + all seven schema cases); the 7 changed files
are exactly the F1/F2 set above plus the two regenerated digest files.
`git diff HEAD --name-only -- conformance/v1/vectors` lists only
`security-posture.json` (pre-existing vectors byte-identical).
`git diff HEAD --check` clean.

## Revision 2 validation transcript (all exit 0)

Shell `/bin/sh`-family (`bash`) from the worktree; python =
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python`
(`make validate` needs the venv interpreter for `jsonschema`, so its
three legs ran with the venv on PATH / directly):

- `python tools/validate.py` → `validated 62 schemas and 1103 vector
  files`, exit 0
- `go test ./tools/...` →
  `ok github.com/relux-works/curator-spec/tools/generate-vectors`, exit 0
- `python -B -m unittest discover -s tools -p 'test_*.py'`, sharded
  3+heavy (same suite, bounded calls; shard runner
  `/tmp/posture_rev2_shard.py`): shard0 142 tests OK (310s), shard1 142
  OK (266s), shard2 141 OK (273s), heavy
  `WriteNofollowVectorTests.test_substituted_scenario_rejected_through_main`
  whole (11 scenarios) OK (243s) — 426/426 tests, every shard exit 0.
  (19 posture tests incl. the 4 new pinning tests pass standalone in ~2s.)
- `go run ./tools/generate-vectors -root .` → exit 0; `make
  regenerate-check` (after staging `conformance/v1` + `release/`, same
  procedure as rev1) → exit 0, generator idempotent.
- `/tmp/posture_rev2_probe.py` (MCP rewrite + exhaustive 272
  substitutions) → all refused, exit 0.
- Worktree/patch identity: `git diff HEAD | git patch-id --stable` and
  the attached `TASK-260910-2qtiho_spec-patch_rev2.patch` both yield
  `d0399de31edf57b13be39b9d12c5f96a62a9a1da`; patch sha256
  `bc2363422da4f59340f41cb90fc1167be4bfe648386bd94bd534e7e2fbdbcc0f`.

## Revision 2 scope notes

- CHANGELOG Unreleased entry unchanged from rev1 (it names no row
  counts; "one row per gate" covers the shared twelve).
- No new diagnostics, knobs, or schema fields in rev2; no existing gate
  default changed.
- Out of scope (unchanged): implementation (`TASK-260910-1sapuy`,
  manager posture work), S2 bootstrap checkpoint, tags, proposals
  0014–0018.
- Reviewer note: the round-1 verdict's `git diff origin/main` patch
  comparison does not apply — the rework brief (like the round-1
  round-specifics) pins the comparison to `git diff HEAD`, base
  `e8b53a0`, which the patch-id above verifies.

---

# Revision 3 (rework after `changes_requested`, 2026-09-18)

Rework brief `TASK-260910-2qtiho_rework-rev3.md` (one remaining correction
from `TASK-260910-2qtiho_review-verdict-rev2.md`: the F2 provenance
vocabulary; F1 pinning and the twelve-gate inventory already passed).
Base still `e8b53a0`; worktree and role unchanged. EMPTY curator delta
re-verified (`git -C <curator-story-worktree> status --short` empty,
`git diff --stat` empty). No LOGBOOK.md change anywhere; no implementation
code touched.

## Provenance rename: `profile-default`/`locked` to `profile`/`lock`

The settled closed set is exactly `profile`, `explicit`, `lock`,
`shipped` (rev-2 rework brief). The rev-2 candidate spelled
`profile-default`, `explicit`, `locked`, `shipped`. Renamed mechanically
in the six brief-named surfaces plus regenerated digests; configuration
fields such as `system.locked` and the `locked` config key are NOT
renamed (those are not provenance values). Case names
(`explicit-knob-beats-profile-default`, `locked-value-beats-explicit`,
`locked-posture-beats-explicit-permissive`) are kept: they name the lock
mechanism in English, not the provenance token.

- `profiles/manager.md` §10: provenance definition plus the six
  knob-gate rows (`registry-policy`, `audit-mode`, `source-allowlist`,
  `transitive-system-modules`, `source-signers`,
  `mcp-package-allowlist`) now read `` `profile`, `explicit`, or `lock` ``.
- `protocol/environments.md` §12: knob gates report `` `profile`,
  `explicit`, or `lock` ``; the §1 `locked` set reference untouched.
- `CHANGELOG.md` Unreleased S1/S3 entry: provenance now `` (`profile`,
  `explicit`, `lock`, `shipped`) ``.
- `conformance/v1/vectors/security-posture.json`: top-level
  `provenance` list, all 268 `profile_source` / `sources` / row `source`
  values `profile-default` to `profile`; all lock provenance values
  `locked` to `lock` (config `"locked": [...]` keys at the two system
  blocks preserved). JSON valid.
- `tools/validate.py`: `SECURITY_POSTURE_PROVENANCE` tuple, the eight
  `effective.profile_source` `profile` pins plus the one `lock` pin,
  and the `_posture_resolve` derivation (`permissive` fallback,
  revision default, lock branch, `resolve()` helper, null-explicit
  check). Config `locked` (schema properties, `system.get("locked")`,
  `system.locked` pins, `lock_key in locked`) untouched.
- `tools/test_validate.py`: the two existing provenance tests updated
  to the new spellings (`profile` for the absent rewrite,
  `lock` for the explicit-claimed-locked case); two NEW negative tests
  `test_superseded_provenance_profile_default_fails` and
  `test_superseded_provenance_locked_fails`, each pinning all four
  value locations (top-level `provenance`, `profile_source`,
  `sources`, row `source`) to reject the superseded spelling.
- `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`:
  regenerated digests only (`make regenerate`).

Byte-identity vs rev2: per-file diff comparison of `git diff HEAD`
against `TASK-260910-2qtiho_spec-patch_rev2.patch` — 16 of 24 files
byte-identical (SECURITY, registry, both schemas, all four generator
files, index + all seven schema cases); the 8 changed files are exactly
the rename set above plus the two regenerated digest files.
`git diff HEAD --name-only -- conformance/v1/vectors` lists only
`security-posture.json` (pre-existing vectors byte-identical).
`git diff HEAD --check` clean.

## Revision 3 validation transcript (all exit 0)

Shell `bash` from the worktree; python =
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python`:

- `go run ./tools/generate-vectors -root .` → exit 0
- `python tools/validate.py` → `validated 62 schemas and 1103 vector
  files`, exit 0
- `go test ./tools/...` →
  `ok github.com/relux-works/curator-spec/tools/generate-vectors`, exit 0
- `python -B -m unittest test_validate.SecurityPostureVectorTests`
  (from `tools/`) → 21 tests OK (19 rev-2 + 2 new superseded-spelling
  negatives), exit 0, ~1s
- Full suite via `/tmp/posture_rev2_shard.py` (same sharding as rev2):
  shard0 143 tests OK (204s), shard1 142 OK (222s), shard2 142 OK
  (211s), heavy
  `WriteNofollowVectorTests.test_substituted_scenario_rejected_through_main`
  whole OK (231s) — 428/428 tests (426 rev-2 + 2 new), every shard
  exit 0
- `make regenerate-check` (after staging `conformance/v1` + `release/`,
  same procedure as rev2) → exit 0, generator idempotent
- Worktree/patch identity: `git diff HEAD | git patch-id --stable` and
  the attached `TASK-260910-2qtiho_spec-patch_rev3.patch` both yield
  `d34f8329fc83f9b83a5d0b53f784bc0d7e4fcbea`; patch sha256
  `94d759b8cc8565b7c0c285304f00dc85098905eb20f9dfb3d628c5ef7efda84b`.

## Revision 3 scope notes

- No new diagnostics, knobs, or schema fields in rev3; no existing gate
  default changed; no F1/F2 logic changed beyond the spelling.
- Out of scope (unchanged): implementation (`TASK-260910-1sapuy`,
  manager posture work), S2 bootstrap checkpoint, tags, proposals
  0014–0018.
- Comparison base remains `git diff HEAD`, base `e8b53a0` (not
  `origin/main`), per the rework brief.
