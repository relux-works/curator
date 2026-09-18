# TASK-260910-1tvf2t evidence — spec-bootstrap-checkpoint (S2)

Story `STORY-260910-6bo7ej` (tofu-and-equivocation-mitigations), wave 3 of the
2026-09 security-audit remediation. Producer brief: signed bootstrap checkpoint
interchange and the TOFU/equivocation residuals (finding S2, Medium).

- Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-6bo7ej/worktree`
  (branch `task-board/story/STORY-260910-6bo7ej`)
- Base commit: `4a2fa3e` (curator-spec `main`, pre-existing checkout; no commit,
  push, branch, or PR by the producer)
- Role: doc-writer (normative spec text + spec-repo conformance machinery only;
  no implementation code, no LOGBOOK.md edits, no writes outside the worktree)

## Settled decisions honoured (brief, not reopened)

- **Bootstrap checkpoint = signed `registry-snapshot-v1` supplied out of band.**
  Closed form chosen: **path** (`bootstrap_checkpoint` next to `public_keys`),
  as the brief recommends. Rationale recorded in the spec: the path form keeps
  the checkpoint byte-identical to the R3/P2 operator's file so one file serves
  the client's first use and the service's startup comparison; there is no
  inline-object form. The file is manager-protected state under the S5
  discipline (ownership / private mutation permissions / containment / link
  safety validated before reading; symlinks never followed to reach it).
- **First use with checkpoint:** verify against pinned keys, persist as initial
  high-water BEFORE any network response is accepted; first network
  snapshot/page boundary must satisfy §5 against it (below → tampered,
  equal-different → tampered). No TOFU for that registry.
- **First use without checkpoint:** TOFU stays, reported once as posture
  (`registry_bootstrap_tofu`, warning, naming the registry and the hint to pin
  a checkpoint); status lists per registry whether the high-water came from a
  checkpoint or first use.
- **Rebootstrap after loss:** same object; a checkpoint below a still-present
  persisted high-water is refused (`registry_checkpoint_regression`) — a
  checkpoint never lowers state. Extended by the §5 equal-version rule to
  equal-with-different-body (see "Resolved while writing" below).
- **Equivocation:** residual defined precisely (§5.1: per-client monotonic but
  divergent views; the protocol detects it only when two views meet).
  Optional detection (MAY): closed **`mirror_group`** form chosen over a
  `mirrors_of` relation — registries sharing one identifier value form a group,
  symmetric by construction with no reference graph to validate. With 2+
  enabled registries of one group exposing the same `log_size`, an
  implementing client compares `merkle_root` and reports a difference as
  `registry_view_divergence` (warning under advisory, error under strict)
  without changing resolution; no quorum. Policy mapping references "registry
  policy" generically (S1/S3 hardened profile not restated).
- **Rollout direct** (impact row "S2"): no behaviour change without a
  checkpoint or a mirror group. Closed diagnostics exactly the three above.

## What changed per file

Normative prose (RFC 2119 keywords, closed lists, identical spellings in text,
tables, schema, vectors, CHANGELOG):

- `protocol/registry.md`
  - §2.1: cross-reference only — a bootstrap checkpoint verifies against the
    same pinned set; rotation never revalidates or resets checkpoint-persisted
    high-water.
  - §5: bootstrap-checkpoint rule (object, path form, protected-state
    handling, verify-then-persist-before-network, first-network-must-satisfy-§5
    with snapshot-tampered / `registry_page_boundary_stale` split, no TOFU
    with checkpoint); TOFU-without-checkpoint rule (fixation + once-posture +
    status row); rebootstrap rule (fresh checkpoint re-establishes;
    below/equal-different refused with `registry_checkpoint_regression`,
    state unchanged; equal-same no-op; above advances); closed diagnostics
    table (condition / diagnostic / severity) with "no other bootstrap
    diagnostic exists"; bad-signature/missing/unreadable/malformed checkpoint
    = configuration error naming the path, failing closed (registry
    unavailable on first use, state unchanged on rebootstrap).
  - §5.1 (new): equivocation residual + optional view-divergence detection
    (MAY implement; mirror-group definition; MUST compare at shared
    `log_size`; MUST report difference as `registry_view_divergence` with
    advisory/strict severity; report-only, no exclusion, no quorum; agree /
    <2 members / unshared sizes are not divergence; status row).
  - §8: cross-reference — cache/offline grace never bootstrap or lower state;
    cached/stale views still satisfy §5 (including against
    checkpoint-persisted high-water); TOFU posture reported once per registry
    whatever the fixing view's source.
  - §10: cross-reference — publication never bootstraps state; only an
    accepted snapshot/chain boundary or a verified checkpoint does.
- `profiles/registry-service.md` §10: one sentence — equivocation stays a
  stated residual; a client MAY compare roots at a shared size across a
  mirror group and report `registry_view_divergence` without changing
  resolution (pointer to registry §5.1); no service-side quorum. §5/§6
  untouched (operator checkpoint reused, not respecified).
- `profiles/manager.md`
  - §1: knob rows for `bootstrap_checkpoint` (OPTIONAL path, 1–4096 chars,
    relative resolves against the config directory, absent by default,
    schema-2-only, path-form rationale, protected-state MUSTs, fail-closed
    configuration error) and `mirror_group` (OPTIONAL portable identifier,
    absent by default, schema-2-only, symmetric-by-construction rationale);
    both live inside `audit_registries`, covered by the existing system lock,
    no new lock key.
  - §10: posture rows — `curator status` (and `curator env status` where the
    environments capability is implemented) list, per registry, high-water
    (`version`, `log_size`), bootstrap source (`checkpoint` / `first-use`),
    and last mirror-group comparison (`agree` / `diverged` / `not-compared`);
    `registry_checkpoint_regression` names the registry; `--check` treats a
    strict-policy diverged group as non-current, other bootstrap rows as
    warnings.
- `SECURITY.md` ("Registry security state"): checkpoint interchange paragraph
  (verify-then-persist, no TOFU with checkpoint, TOFU reported once without,
  rebootstrap-fresh-checkpoint, never-lowers-or-forks) + equivocation
  residual paragraph (MAY compare, report-only, no quorum), with pointers to
  `protocol/registry.md` §5/§5.1 and `profiles/manager.md` §1/§10.
- `schemas/v1/README.md`: S2 paragraph — v2-only extension, v1 byte-frozen.
- `CHANGELOG.md`: Unreleased → Added entry "S2: …" (finding id, rule,
  warn-first n/a — direct rollout, impact row "S2", story id).

Schema (frozen-schema rule: v1 byte-identical):

- `schemas/v1/manager-config-v2.schema.json`: `audit_registries` no longer
  `$ref`s the schema-1 shape; new closed v2 `$defs/registry` restates the
  schema-1 entry exactly and adds exactly two OPTIONAL members —
  `bootstrap_checkpoint` (string, 1–4096) and `mirror_group` (portable
  identifier) — `additionalProperties: false` unchanged; description updated.
- `schemas/v1/manager-config-v1.schema.json`: untouched (verified: zero diff).

Conformance machinery (generator-owned; hand edits only in generator sources):

- `tools/generate-vectors/main.go`: `registry-client.json` gains
  `bootstrap_cases` (15 cases: first-use accept, first-network
  below/equal-different tampered, TOFU posture, bad-signature fail-closed,
  rebootstrap advance/no-op/regression/equal-inconsistent/bad-signature,
  divergence detected advisory+strict/agreeing/sizes-skipped/single-skipped);
  `registry-behavior.json` gains `bootstrap` + `divergence` summary objects.
  Every pre-existing case object is byte-identical (regenerated diffs are
  insertion-only; see validation transcript).
- `tools/generate-vectors/manager_config.go`: two v2 vectors
  (`schema2-registry-bootstrap-members` valid, `schema2-registry-unknown-field`
  invalid) + five v2 schema-cases (one valid, four invalid: empty checkpoint,
  non-string checkpoint, mirror-group grammar, unknown field).
- `tools/generate-vectors/manager_config_test.go`: the v2 "schema-1 shape
  reuse" assertion now covers `skills_root`/`projects`/`audit` only;
  `audit_registries` must reference `#/$defs/registry`, which must restate
  the schema-1 entry exactly plus the two closed members with exact grammars;
  `registry` added to the closed-`$defs` check.
- `tools/validate.py`: new gate `validate_registry_bootstrap_vectors`
  (registered in `main()`), with `expected_bootstrap_verdict` (recomputes all
  eight outputs per phase) and `require_bootstrap_scenario` (pins every
  required name to its discriminating inputs; rule 7), plus closed-set
  constants (`BOOTSTRAP_DIAGNOSTICS`, `BOOTSTRAP_POSTURE`,
  `BOOTSTRAP_PHASES`, `BOOTSTRAP_POLICIES`, `BOOTSTRAP_SEVERITIES`).
- `tools/test_validate.py`: new `RegistryBootstrapVectorTests` (22 tests:
  published-vectors pass; 11 verdict-narrowing negatives; dropped case; 8
  self-consistent scenario replacements under required names; main-entry
  rejection).
- Regenerated (via `make regenerate`, exit 0): `vectors/registry-client.json`,
  `vectors/registry-behavior.json`, `vectors/manager-config-v2.json`,
  `schema-cases/index.json`, `manifest.json`, `release/1.0.0-rc.9.json`, and
  five new files under `schema-cases/manager-config-v2/`.

## Resolved while writing (brief-consistent completions, no new diagnostics)

1. **Equal-version-different-body checkpoint vs surviving state.** The brief
   MUSTs refusal only for the below arm. Accepting an equal-different
   checkpoint would fork established state in exactly the way §5's
   equal-version rule forbids for snapshots, so the spec refuses it with the
   same `registry_checkpoint_regression` ("a checkpoint never lowers or forks
   established state"); equal-same is a no-op accept. No fourth diagnostic.
2. **Checkpoint signature failure / missing / unreadable / malformed file.**
   The brief fixes exactly three diagnostics, so these fail closed as a
   configuration error naming the path (first use: registry unavailable;
   rebootstrap: state unchanged; registry usable on old state) — stated in
   §5, the §5 table coda, manager §1, and the vectors (both bad-signature
   cases carry `diagnostic: null` with consistent versions to prove
   verification precedes comparison).
3. **Strict-policy divergence stays report-only.** The brief orders error
   severity under strict "without changing resolution": the strict case pins
   `accepted: true`, `registry_excluded: false`, `resolution_changed: false`
   with `severity: error`. Severity affects reporting prominence (and the
   `--check` non-current mapping in manager §10), never resolution.
4. **Status surface.** The brief names "`env status`/`curator status`"; the
   spec requires the rows on `curator status` and, where the environments
   capability is implemented, `curator env status` (the §8.6 dual-surface
   precedent).

## Deliberately out of scope

- Implementation (`TASK-260910-2vnjej`, optional Merkle-root comparison), R3/P2
  (landed; object reused), S1/S3 (`TASK-260910-2qtiho`; policy mapping
  references "registry policy" generically), P4 import high-water, tags /
  releases, `ax`, proposals 0014–0018, anything the brief does not name.
- No v1 artifact touched: schema, vectors, and schema-cases for
  `manager-config-v1` are byte-identical (v1 rejection of the new members
  follows from frozen `additionalProperties: false`, already pinned by the
  existing `unknown-registry-field` v1 vector).
- No implementation code, no LOGBOOK.md edits, no commits/pushes/branches/PRs.

## Validation transcript

Shell: `bash` (login not required). Python: the repository venv
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python3`
(pinned `jsonschema==4.25.1` per `requirements-dev.txt`), selected via
`PATH=.../venv/bin:$PATH` from the worktree, because the system `python3`
has no `jsonschema` module. Go: system `go1.26.0`. Working directory for every
command: the worktree root above.

`make validate` was not run as one end-to-end call: the Python suite alone
takes ~11 minutes and a single shell call is time-bounded, so each of its
three phases ran green in bounded sequential calls (every test exactly once,
no narrowing): `tools/validate.py` in one call, the `unittest discover`
content split into five class/file slices covering all 461 tests, and
`go test ./tools/...` in one call. An early full-suite `make validate`
(passed `yield_time_ms` 300000, backgrounded anyway) was terminated once the
tree changed under it, and a full `unittest discover` was terminated as too
slow for one call and replaced by the slices below.

1. `go run ./tools/generate-vectors -root .` → exit 0 (`REGENERATE_EXIT:0`).
   Regenerated: `vectors/registry-client.json`,
   `vectors/registry-behavior.json`, `vectors/manager-config-v2.json`,
   `schema-cases/index.json`, `manifest.json`, `release/1.0.0-rc.9.json`,
   five new files under `schema-cases/manager-config-v2/`.
2. Frozen + insertion-only proof (all exit 0):
   - `git diff --stat -- schemas/v1/manager-config-v1.schema.json
     conformance/v1/vectors/manager-config.json
     conformance/v1/schema-cases/manager-config-v1/` → 0 lines (v1 untouched).
   - Removed-line counts (`git diff | grep -c "^-[^-]"`) → 0 in each of
     `registry-client.json`, `registry-behavior.json`,
     `manager-config-v2.json`, `schema-cases/index.json` (existing cases
     byte-identical; additions only).
3. `python3 tools/validate.py` → exit 0:
   `validated 62 schemas and 1108 vector files`.
4. `python3 -B -m unittest test_validate.RegistryBootstrapVectorTests` →
   exit 0: `Ran 22 tests in 26.873s / OK`.
5. `python3 -B -m unittest` slices of `test_validate` (all exit 0, OK):
   - classes 1–8: `Ran 83 tests in 70.838s / OK`;
   - classes 9–16: `Ran 168 tests in 198.107s / OK`;
   - classes 17–22: `Ran 110 tests in 322.605s / OK`.
   Together with slice 4: 383/383 `test_validate` tests green.
6. `python3 -B -m unittest test_implementation_coverage test_release_gate
   test_verify_release_commit test_verify_release_merge_policy` → exit 0:
   `Ran 78 tests in 264.523s / OK`.
   Total Python: 461/461 green (36+32+383+5+5 per-file `def test_` counts).
7. `go test ./tools/...` → exit 0:
   `ok github.com/relux-works/curator-spec/tools/generate-vectors 1.284s`.
8. Regeneration proof: sha256 snapshot of `manifest.json`, `index.json`,
   the three touched vector files, `release/1.0.0-rc.9.json`, and all 101
   `schema-cases/manager-config-v2/*.json`, then `go run
   ./tools/generate-vectors -root .` again, then `sha256sum -c` → every
   file `OK`, exit 0 (second run changed zero bytes).
9. `make regenerate-check` → exit 1 (expected-red in this worktree, not a
   gate failure of the change): it ends with `git diff --exit-code --
   conformance/v1 release/…` against HEAD, and this revision's regenerated
   files are deliberately uncommitted (the producer must not commit), so a
   diff is the correct outcome. The idempotence proof in (8) is the
   substantive regeneration evidence.

Closed-set spelling was additionally grepped: the three diagnostics appear
identically in `protocol/registry.md`, `profiles/manager.md`,
`profiles/registry-service.md` (divergence), `SECURITY.md`, `CHANGELOG.md`,
both vector files, `tools/validate.py`, and the generator; both knobs
appear identically in prose, schema, CHANGELOG, and schema-cases. No
`mirrors_of` remnants; no implementation code touched (`git status` shows
only spec-repo files); no LOGBOOK.md change.

---

# Revision 2 (rework after `TASK-260910-1tvf2t_review-verdict-rev1.md`, F1–F3)

Rework brief: `TASK-260910-1tvf2t_rework-rev2.md`. Everything the
brief-conformance matrix passed in round 1 is kept byte-identical except
the corrections below; the 15 `bootstrap_cases` scenario names are kept
and their pre-existing input/output values unchanged (one new output
field added per F2, see below).

## F1 — pins require present, typed evidence (rule 7)

`tools/validate.py` gained `bootstrap_bool(case, field)`, which raises
`ValidationFailure` unless the field exists with an exact `bool` type.
Both `expected_bootstrap_verdict` and `require_bootstrap_scenario` now
read every discriminating boolean through it
(`checkpoint_configured`, `signature_valid`, `candidate_same_body` on
the bootstrap/rebootstrap arms; `same_log_size`, `roots_equal` on the
compare arm), and every false-discriminator pin requires explicit
`False`. A missing, null, or mistyped input is refused — never read as
false.

`tools/test_validate.py` `RegistryBootstrapVectorTests` gained: six
missing/null/wrong-type negatives (one per discriminating field, each
trying `None`, `0`, `1`, `"true"`, `"false"`, `"yes"`, `[]`, plus field
removal), two self-consistent replacement negatives
(equal-different→consistent, sizes-skipped→compared), and a main-entry
negative replaying the reviewer's exact five-removal mutant through
`validate.main()`.

Measured refusal coverage for the five review probes (each removal
tested individually through `validate_registry_bootstrap_vectors`):
**5/5 refused** (was 0/5); the combined five-removal mutant through
`validate.main()` returns **1** (was 0). Valid vectors unchanged and
passing.

## F2 — status severity explicit per row

`profiles/manager.md` §10 no longer blanket-labels bootstrap rows. It
now states per row: checkpoint bootstrap success → current,
informational; TOFU → warning row carrying `registry_bootstrap_tofu`,
stays current; refused checkpoint → error row carrying
`registry_checkpoint_regression`, `--check` non-current; divergence →
warning staying current under advisory, error non-current under
strict; first-use checkpoint configuration error → error row,
non-current; ignored rebootstrap bad signature → persisted row stays
current; tampered first network → operation exclusion, persisted row
stays current. No new diagnostic.

Vector pin: every `bootstrap_cases` entry carries a new boolean output
`check_current` (`false` for the two regression refusals, the strict
divergence, and the first-use bad-signature configuration error;
`true` for the other 11), emitted by `tools/generate-vectors/main.go`
and recomputed by `expected_bootstrap_verdict` from the case inputs
(not from the case's own diagnostic field). Five flip negatives pin
it (regression→true, strict→true, TOFU→false, bad-sig-first-use→true,
advisory→false refused). The six replacement-scenario constants carry
`check_current: true` (all are current-row scenarios).

## F3 — cache-backed first fixation stated once

`protocol/registry.md` §5, §8, and the §5 diagnostics table now agree:
first fixation comes from a verified checkpoint or from the network
only. A cached entry fixes first-use high-water only when those exact
bytes were themselves accepted online under §5 — in which case they
reflect persisted state rather than a fresh fixation — and a cache
entry never accepted online MUST NOT fix state, bypass checkpoint
validation, or recover known-lost state. The `registry_bootstrap_tofu`
posture is reported once when the network fixes checkpoint-less first
use; serving a cached view thereafter does not re-report it.

Vector pin: `vectors/registry-behavior.json` `bootstrap` summary gains
`first_fixation_source: "checkpoint-or-network-only"`, and
`validate_registry_bootstrap_vectors` now pins the full S2
bootstrap/divergence summaries to their exact values (previously only
byte-integrity covered). Three summary negatives pin it (wrong source,
missing source, wrong TOFU posture refused). The permitted network
first-fixation posture itself remains pinned by
`no-checkpoint-first-use-tofu` (`test_tofu_posture_dropped_is_rejected`).

## Minor — CHANGELOG rollout sentence

The direct-rollout sentence now reads "no behavior change without a
checkpoint or a mirror group, except the new once-per-registry
`registry_bootstrap_tofu` warning on checkpoint-less first use."

## What did NOT change in revision 2

- The 15 scenario names; all pre-existing input/output values of the
  15 cases; `rollback_state_cases` and `snapshot_transitions`
  byte-identical to HEAD (verified per top-level key); v1 schema,
  v1 vectors, and v1 schema-cases zero diff.
- `profiles/registry-service.md`, `SECURITY.md`, schemas, schema-cases,
  manager-config vectors: untouched since rev1 (same 23 changed paths;
  `git status` lists only spec, schema, vector, manifest, release, and
  generator/validator/test files; no LOGBOOK.md, no implementation).
- The verdict's scope note (system-config-v2 lock claim vs frozen v1
  registry shape) is acknowledged but intentionally not edited: the
  rework brief orders F1–F3 plus the CHANGELOG sentence and
  byte-identical otherwise, and any clarification there would widen
  prose beyond the authorized corrections without changing the
  settled bootstrap design.

## Validation transcript (revision 2)

Shell `bash`; Python the repository venv
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin/python3`
via `PATH=.../venv/bin:$PATH`; Go system `go1.26.0`; worktree root for
every command. `make validate`'s three phases ran green in bounded
sequential calls (no single end-to-end call; every test exactly once):

1. `go run ./tools/generate-vectors -root .` → exit 0.
2. Frozen + pre-existing proof (all exit 0): `git diff --stat` over
   `manager-config-v1.schema.json`, `vectors/manager-config.json`,
   `schema-cases/manager-config-v1/` → 0 lines; per-key comparison of
   `registry-client.json` (`rollback_state_cases`,
   `snapshot_transitions`, `page_boundary_cases`, `key_rotation_resets_state`,
   `pagination_rejections`, `retry_cases`, `retry_policy`, `state_key`)
   and `registry-behavior.json` (`artifact_hash`, `cache`,
   `pagination`, `snapshot`, `submission`) against HEAD → identical.
3. `python3 tools/validate.py` → exit 0:
   `validated 62 schemas and 1108 vector files`.
4. `python3 -B -m unittest test_validate.RegistryBootstrapVectorTests` →
   exit 0: `Ran 39 tests in 44.204s / OK` (22 carried + 17 new).
5. `python3 -B -m unittest` slices of the remaining `test_validate`
   classes (all exit 0, OK): classes 1–8 `Ran 83 tests in 77.553s`;
   classes 9–16 `Ran 168 tests in 215.510s`; classes 17–22
   `Ran 110 tests in 305.938s`.
6. `python3 -B -m unittest test_implementation_coverage test_release_gate
   test_verify_release_commit test_verify_release_merge_policy` → exit 0:
   `Ran 78 tests in 212.095s / OK`.
   Total Python: 478/478 green (461 carried + 17 new).
7. `go test ./tools/...` → exit 0:
   `ok github.com/relux-works/curator-spec/tools/generate-vectors 1.394s`.
8. Adversarial reproduction of the reviewer's F1 probe (five
   discriminating inputs removed, through `validate.main()`):
   exit **1** (refused); per-probe gate checks: **5/5 refused**.
9. Regeneration proof: sha256 snapshot of `manifest.json`, `index.json`,
   the three touched vector files, `release/1.0.0-rc.9.json`, and all
   101 `schema-cases/manager-config-v2/*.json`, then regenerate, then
   `sha256sum -c` → every file `OK` (107/107), exit 0.
10. `make regenerate-check` not run as a literal gate: as in rev1 it
    compares regenerated files against HEAD while this revision is
    deliberately uncommitted, so a diff is the correct outcome; the
    idempotence proof in (9) is the substantive regeneration evidence.

Closed-set spelling re-grepped: `registry_bootstrap_tofu` (8),
`registry_view_divergence` (7), `registry_checkpoint_regression` (5)
spelled identically across prose, schema, vectors, validator, and
generator; `check_current` boolean in all 15 cases; no `mirrors_of`
remnants.
