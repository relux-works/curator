# TASK-260922-1hla8q results — F-S2: fragment permission members + transport token

Status: ready for review (developer role handoff).

## 1. Normative decision (for consumers — create follow-ups verbatim from this)

Minimum transport version token: `launch-env-fragment-v2` — the value of the
fragment's REQUIRED `fragment` member. A launcher establishes permission-policy
transport support iff the fragment revision is v2 or later.

Fragment member (REQUIRED in every v2 fragment; closed object, readers reject
unknown fields):

```json
"permissions": { "mode": "native", "locked": false, "source": "profile" }
```

- `mode`: REQUIRED, exactly `native` or `yolo` — the
  `permissions.<profile>` knob value for the resolved profile when the knob
  names one, `native` otherwise (lock override, or the silence placeholder).
- `locked`: REQUIRED, boolean — true iff the environments §12.2 force-`native`
  lock is engaged for the resolved profile.
- `source`: REQUIRED, exactly `profile`, `global`, or `default` — which
  Curator-side input fixed `mode`: `profile` = the machine knob named it;
  `global` = the fleet-global system-file lock fixed it (`global` never means
  the launcher-global `defaults.json` default, which the launcher reads from
  its own file and which never appears in the fragment); `default` = the knob
  was absent and no lock is engaged.
- Consistency (schema-enforced): `locked` is true iff `source` is `global`;
  `mode` is `native` whenever `source` is `global` or `default`. Hence `yolo`
  occurs only as `{yolo, locked:false, source:profile}`; `yolo` with
  `locked:true` is contradictory and invalid.
- Silence: `source:default` means the profile level is SILENT — the launcher
  falls through to the launcher-global default, never treating the `native`
  placeholder as an explicit naming. `source:profile` means the profile level
  names `mode`.

Fail-closed rule (cites the token): a fragment that predates
`launch-env-fragment-v2` cannot carry the policy or the lock and is never
silence — any launch that would otherwise resolve `yolo` (flag, profile,
global, or built-in default) is refused with `permission_policy_unsupported`;
a launch that resolves `native` (explicit, or headless/CI/tracked silence)
proceeds as `native`.

Headless markers: exactly {`CI`, `GITHUB_ACTIONS`}, stated once normatively in
environments §10.1, closed, additions by specification revision only, mirrored
in launcher SPEC §4.6. The launcher SPEC §4.1 transport-precondition mirror
lands with F-L1.

## 2. Follow-up leaves (for the orchestrator)

1. Curator emission leaf (manager side, curator repo — create verbatim): make
   `env resolve --format json` emit `launch-env-fragment-v2` with the REQUIRED
   `permissions` member from §1, computed for the resolved profile from the
   effective `permissions.<profile>` knob (system-file lock already applied
   per §12.2 with the manager §1 warning) and the lock engagement. Until that
   leaf lands, managers emit v1 (stated normatively in §10.2).
2. F-L1 (launcher, STORY-260922-39hxog): launcher SPEC §4.1 precondition
   mirror, mode resolution (flag over profile over global over built-in,
   `source:default` = silent level), transport of the resolved mode,
   provenance. No launcher edit in this task.

## 3. What changed (worktree, uncommitted)

- `protocol/environments.md`: §10.1 Permission-mode paragraph (token, member,
  single-place marker set + mirrors, fail-closed citing the token); §10.2
  retitled "The launch environment fragment" (two revisions, token rule,
  reader rules, v2 member grammar); §13 (v2 schema + cases listed).
- `profiles/manager.md` §12.5 (fragment reference, token, transport rule).
- `decisions/0018-curator-run-permission-interface.md` choice 7: appended the
  fixed token/member values only; adoption history untouched.
- `schemas/v1/launch-env-fragment-v2.schema.json`: NEW (v1 + REQUIRED closed
  `permissions` with three if/then consistency rules). v1 file byte-untouched.
- `schemas/v1/README.md`: v2 paragraph. `CHANGELOG.md`: Unreleased/Added
  entry naming F-L1 (STORY-260922-39hxog) and the curator emission follow-up.
- `tools/generate-vectors/environments.go`: `validLaunchEnvFragmentV2()`,
  `launchEnvFragmentV2SchemaExamples()` (3 valid + 9 invalid), v1 examples
  gain `invalid-permissions-member`. `main.go`: v2 wiring.
  `environments_test.go`: v2 map entry (≥3/≥8 + unknown-field enforced).
- `tools/validate.py`: wire-semantics channel/root checks dispatch
  v1 + v2. `tools/test_validate.py`: 3 new v2 rows.
- Regenerated (`go run ./tools/generate-vectors -root .`, no collateral):
  `conformance/v1/schema-cases/launch-env-fragment-v2/` (14 files),
  `launch-env-fragment-v1/invalid-permissions-member.json`, `index.json` (+15
  entries), `manifest.json` (1124 → 1139 files), `release/1.0.0-rc.9.json`
  (manifest pin only).

## 4. Cases and evidence

v2 valid (4): `valid.json` (native unlocked, source profile),
`valid-permissions-yolo-unlocked`, `valid-permissions-native-locked`,
`valid-permissions-native-silent`. v2 invalid (10): `invalid.json` (missing
env), `invalid-permissions-absent`, `-unknown-mode`, `-yolo-locked`,
`-yolo-silent`, `-locked-profile-source`, `-global-unlocked`,
`-unknown-source`, `-unknown-field`, `-missing-mode`. v1: new
`invalid-permissions-member`; all existing v1 valids unchanged
(`valid-minimal` re-verified valid).

Gate (project venv python, tree identical for every run below):

```text
$ make validate
python3 tools/validate.py
validated 63 schemas and 1139 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
Ran 579 tests in 1334.240s — OK
go test ./tools/...
ok  github.com/relux-works/curator-spec/tools/generate-vectors  3.595s
MAKE_VALIDATE_EXIT: 0
```

- Focused: `EnvironmentVectorTests.test_environment_schema_semantics_fail_closed`
  OK; `ManagerConfigVectorTests` + `SystemConfigV2SchemaTests` 53 tests OK;
  `go test -run TestEnvironmentSchemaCasesCoverTheClosedSurfaces` ok.
- Ad-hoc narrowing probes (`/tmp/v2mutants.py`, kept out of the repo):
  40+ checks, 0 failures — the full 12-combination mode×locked×source truth
  table (4 valid / 8 invalid), cross-schema rejections (v1-minimal vs v2
  schema, v2-full vs v1 schema, token swaps both directions), type/arity
  narrowings (locked-as-string, mode null, missing source/locked, array
  member, unknown top-level field).
- Anomaly (reported, not hidden): the FIRST end-to-end `make validate` run's
  unittest phase failed while this host ran several agents' suites plus a
  release rust build concurrently, and its tail was lost to `| tail`. The
  rerun of the full unittest suite on the identical tree passed 579/579
  (1908s, OK), and the definitive end-to-end `make validate` above then
  passed all three phases. No code changed between the red and green runs;
  assessed as load-induced flake. Production call site of the gate:
  `tools/validate.py:validate_wire_semantics` (fragment branch) +
  `validate_schemas`, driven by every committed case above.
- `gofmt -l` clean, `go vet` clean. No separate lint gate in this repo
  (Makefile: validate / regenerate / release-check only).

## 5. Bounds

- No launcher/curator code; no emission cutover (curator follow-up §2.1).
  `cli/curator.md` still names v1 (accurate until the cutover).
- `manager-config-v2` / `system-config-v2` untouched (member names forced no
  rename). Frozen v1 protocol schemas untouched; the v1 fragment file is
  byte-untouched (it is an environments-spec object, not a frozen v1
  protocol schema — no freeze list names it).
- `go-testing-tools` skill loaded per the assignment but inapplicable
  (bubbletea TUI toolkit); followed repo conventions (generated conformance
  cases + stdlib table tests) instead.
- No board logbook mechanism found (`task-board --help` has no logbook
  surface; LOGBOOK.md edits are out of scope per campaign rules) — decisions
  and the anomaly are recorded in this outcome instead.
