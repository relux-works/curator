# TASK-261016-i5gwaf — sanitized evidence packet

Archive upload was refused by the board binary-resource policy; this packet preserves the actual text artifacts and fixtures.

## README.md

`````
# Muse 1.4.8 exclusion replay — release qualification blocked

Task: TASK-261016-i5gwaf — implement-rc7-muse-148-exclusion-replay (M2).

The real `1.4.8-R6256.1` loader does **not** satisfy the required interactive
exclusion contrast in this offline echo replay. Both linked and regular-file
native roots load, but `-c context.foreign_personal_skills=false` leaves the
synthetic foreign skill loaded. The banner includes that skill and the vendor
`skills.catalog_load` record reports `user_skills=1`. The same cases with exec's
`--no-foreign-personal-context` report zero. Qualification and the §7.9 release
gate remain pending. No qualification constant, runtime gate or credentials
policy was changed.

## Evidence identities and limits

- Curator source/checkpoint: `0d00a109ad1da77a5506747c89f24af9442eb341`, equal to
  freshly fetched `origin/main` at start; candidate remains uncommitted.
- Normative source: curator-spec `main`, fetched as
  `d2a942cf17e4eb017c45475442da1456d8dfbe6c`. Read environments §§7.1, 7.9,
  10.2, CIP-0017 leaf 2/test plan, the Muse v4 conformance README and the
  published managed-home snapshot protocol. The new measurement does not
  reproduce the historical interactive contrast; §7.9 requires an erratum
  rather than a silent implementation workaround.
- Muse reports `Muse Code 1.4.8 (1.4.8-R6256.1)`; binary SHA-256
  `5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c`;
  macOS `27.0`, `arm64`. `$MUSE148_BINARY` identifies the separately scoped
  vendor binary, not the auto-updating launcher. Its exact private locator is
  retained under the task-local ignored `.temp/` directory, not published.
- Go `1.25.5`, tmux `3.6a`; Apple sandbox-exec denies all network and Keychain
  service/file access. A reachable loopback endpoint is blocked by that
  policy; a synthetic Keychains file is denied beside a readable control.
- All vendor runs use `env -i`, disposable HOME and four XDG parents,
  `MUSE_NO_AUTO_UPDATE=1`, `TBH_CREDENTIAL_BACKEND=file`, and `--provider echo`.
  No login/logout/auth/store command or active credential read was performed.
  Tests preserve fixture source/root bytes and require credential files to
  remain absent. This is no evidence of login, refresh or backend migration.
- Production observation channels are vendor `rules.context_load` and
  `skills.catalog_load` diagnostics, plus the real interactive `/rules` output
  and startup banner. TUI receives one offline echo sentinel to bind/flush its
  logs, then `/rules` and `/exit`; no remote inference is possible under the
  tested network deny. Echo's built-in reminder children can report that the
  provider lacks base-instruction support; that is not a loader verdict.

## Coverage and negative controls

**6 of 6 in-scope observation rows driven; 5 of 6 satisfy the narrow loader
contract.** The seventh table row explicitly requires separate credential or
model evidence and is outside this replay. This brief has an observation
table, but no catalog-keyed attack-surface table: that is a brief gap, not an
implicit waiver of coverage. No review handoff or release acceptance is claimed.

| Required observation | Production entry / named test | Result / bound |
|---|---|---|
| Exec flag accepted; interactive flag rejected; JSON boolean accepted | Real Muse exec / TUI parser; `TestMuse148ParserRefusals`, `TestMuse148LoaderMatrix` | Accepted flag/boolean reach the real loader; rejected flag and string `"false"` each exit 2 with their specific parser refusal, before loader |
| Native root suppresses foreign rules; foreign skills persist | Both real forms; `TestMuse148LoaderMatrix` | Claude and Codex rules tested separately; native root suppresses either, skill still loads without exclusion |
| Form-specific exclusions remove foreign loader sources with root | Both real forms; `TestMuse148LoaderMatrix/{exec,interactive}/{linked,snapshot}/skills/exclude-true` | Exec passes; interactive fails twice with `user_skills=1`, wanted 0; no model-visible removal claim |
| Override without native root leaves fallback | Real TUI; `TestMuse148LoaderMatrix/interactive/absent` | Claude/Codex fallback remains 1; foreign skill remains 1 |
| Real Curator link and frozen regular-file root load | Curator `profile install`, `env resolve muse --repair`; both real Muse forms; matrix | Actual generated link matches frozen bytes; regular snapshot has independently checked metadata/closure; this does not execute `env materialize` |
| Workspace conflict residual | Both real forms and both roots; `TestMuse148WorkspaceResidual` | `/rules` enumerates user then project; TUI diagnostics report project count 1 at startup and 0 in a subsequent reload. Behavioral precedence remains unknown |
| Credential backend/login/refresh or model-visible removal | Not driven | Outside M2's “Separate evidence” clause; no borrowed credential facts are qualification authority |

The matrix drives 36 cases: two forms × three root states × three isolated
foreign sources × exclusion absent/present. It preserves every no-exclusion
positive control. There are two independent parser refusals and four workspace
observations. `TestObservationChecker` refuses missing, failed, truncated and
contradictory records beside explicit zero-source controls. `TestFrozenSnapshot`
rejects altered root bytes, an extra entry and a link beside its regular-file
positive control. Runtime-consumer refusal tests remain M1/M3-owned; this leaf
does not replace them with vendor evidence.

| Mutant / plant | Narrows the gate to | Named failing test | Survival bound |
|---|---|---|---|
| Omit required exec exclusion only | Real exec admits the one synthetic foreign skill while the native root remains | `TestMuse148LoaderMatrix/exec/linked/skills/exclude-true` | See attached plant log; baseline counterpart is green |
| Omit interactive exclusion only | Real interactive admits the synthetic skill | `TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true` | **Non-discriminating survivor:** baseline already fails the same assertion; this cannot establish omission-specific coverage |
| Permit exactly one foreign skill in observation checker | Keep all record/type/count checks, but attest `user_skills=1` as zero | `TestObservationChecker` | See attached narrowing-mutant log; event tokens are retained and behavioral checker tests execute |
| Add one unlisted snapshot file | Exact entry-set gate admits an extra closure member | Published `managed_home_snapshots.check_bundle` | Reference checker raises `ValueError: entry-set`; standalone planted command exits 1 |

## Frozen snapshot fixture

`testdata/source` is a synthetic context-only profile, installed/resolved by the
real Curator CLI. Its normal link targets Curator's generated store document.
`testdata/root.md` and `testdata/snapshot/home/config/muse/AGENTS.md` freeze those
exact verified bytes, SHA-256
`8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0`.
The snapshot has one content entry, no credentials or state, declared XDG
placement and CCJ-1 metadata with LF. Runtime copies use private directories
and independent `0600` regular files; git does not preserve private mode bits.

Manifest SHA-256:
`924378006ba6e6e7db9d0277375456a958c14cbaff03d1ebd63dfa6e4b7ceac8`.
Descriptor SHA-256:
`2991bc9c97904d77aef70bb5c3ea7a7044ff462382310ae613bbd31f461fd5f1`.
The manifest's source-fragment hash binds the actual private resolve output;
it is not a portable-equivalence claim or proof of export provenance.

The existing Python snapshot checker from the pinned spec independently accepts
these metadata bytes, entry set, size, hash, mode and adapter binding. Reusing
that normative checker is the reason for the reference-check Python invocation;
host regression tests remain Go with existing dependencies. Capture locks,
destination admission, publication, runtime-B protection and the actual
`env materialize -> Muse` integration are explicitly outside this fixture's
claim, under M2's snapshot-fixture clause.

## Tools and reproduction

Run from repository root. Set `$MUSE148_BINARY` to the exact frozen binary;
the harness rejects another digest. Do not use an active native login or the
auto-updating launcher. All outputs stay under the task-local run root.

```sh
mkdir -p .temp/TASK-261016-i5gwaf/tmp
go build -o .temp/TASK-261016-i5gwaf/curator ./cmd/curator
export MUSE148_CURATOR="$PWD/.temp/TASK-261016-i5gwaf/curator"
export MUSE148_RUN_ROOT="$PWD/.temp/TASK-261016-i5gwaf/replay-new"
export TMPDIR="$PWD/.temp/TASK-261016-i5gwaf/tmp"
go test ./.research/muse148 -count=1 -v
go vet ./.research/muse148
```

With an archive of the pinned spec and its existing Python dependencies,
independently validate the fixture using the published reference checker:

```sh
chmod 600 .research/muse148/testdata/snapshot/home/config/muse/AGENTS.md
"$SPEC_PYTHON" -B .research/muse148/reference-snapshot-check.py \
  "$PINNED_SPEC" .research/muse148/testdata/snapshot
```

The test command intentionally remains red on the recorded artifact; it must
not be relabeled green or rewritten to expect successful exclusion. Ordinary
`go test ./...` does not discover this hidden evidence package. With no binary
configured, its vendor tests skip explicitly and establish no qualification.
Go runs the replay and assertion checker; Curator builds the linked fixture;
tmux supplies a private, background TUI terminal and records owned PIDs;
sandbox-exec provides the tested network/store boundary. No UI focus changes.
Use a **fresh** run root for every replay, including mutants. To plant an
omission, set `MUSE148_OMIT_EXCLUSION=exec` or `interactive` and restrict `-run`
to `^TestMuse148LoaderMatrix$/FORM/linked/skills`.

The task outcome records exact executed commands, exits, sanitized observations
and logs. The full repository gate is `sh scripts/remote-gate.sh`; its results
cannot override this independently red vendor qualification test.

## Stop boundary and required external resolution

The failed assumption is that the historical 1.4.7 interactive boolean excludes
foreign personal skills on this exact 1.4.8 artifact. Both independently
represented roots reproduce the failure; switching argv through the production
plugin or lying about ToolRelease cannot fix a vendor-loader result.

Recommended: retain required interactive-exclusion refusal and the pending
v4-pair release gate, report the exact regression to the vendor/spec owners,
and replay a corrected exact artifact after the owner resolves §7.9's erratum.
An alternative needs an explicit normative owner decision to restrict supported
forms/builds or adopt another verified control; it cannot be invented here.
Required external input: a vendor fix or owner-approved normative resolution
for this interactive loader behavior. D10's credential-mode admission and the
unmeasured refresh bound remain independent.

`````

## TASK-261016-i5gwaf_outcome.md

`````
# TASK-261016-i5gwaf — M2 exclusion replay: blocked

The exact real vendor build `1.4.8-R6256.1` retains the synthetic foreign
personal skill on interactive `-c context.foreign_personal_skills=false`.
Both Curator-linked and snapshot regular-file roots reproduce it. Keep
exclusion qualification and the §7.9 v4-pair release gate pending.

The candidate contains an opt-in Go vendor replay, a frozen synthetic
context-only profile/snapshot, sanitized observations and logs. No production
qualification constants, consumer gates, credential policy or native state
were changed. README, LOGBOOK and UNRESOLVED_QUESTIONS changes are documentation
required by repository instructions; all implementation/fixtures stay in
`.research/muse148`. No existing reviewer tests were removed or renamed.

## Evidence and source identity

See [record and reproduction](README.md), [observations](observations.json),
[full focused run](replay-03.log), [exec omission](plant-exec-01.log) and
[checker narrowing mutant](checker-mutant-01.log).

- Curator checkpoint: `0d00a109ad1da77a5506747c89f24af9442eb341`; zero commits
  behind freshly fetched main at task start; no commit/switch/rebase performed.
- Normative main: `d2a942cf17e4eb017c45475442da1456d8dfbe6c`.
- Binary SHA-256: `5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c`.
- Replay source SHA-256: `f4c1aa7a4e745bfc07d9b4536ab1173ea04d6c122af3849280af20360f80d3b1`.
- Environment: macOS 27.0 arm64, Go 1.25.5, tmux 3.6a, sandbox-exec network
  and Keychain deny; disposable HOME and four XDG parents; offline echo only.
- Exact private binary locator and raw case logs remain only under ignored
  `.temp/TASK-261016-i5gwaf/`; public evidence uses `$MUSE148_BINARY` and relative
  case labels. No active Keychain item was read or touched.

**6 of 6 in-scope AC observation rows driven, 5 of 6 satisfied.**
The seventh (credential/model) row is expressly separate evidence. Forty-two
vendor entry cases executed: 36 loader contrasts, two parser refusals and
four workspace observations. The loader matrix is **34 of 36 green**; both
failed tests name the required interactive root-present skill exclusion.
Both vendor processes exit 0; it is their observation assertion that fails.
All parser/fixture/boundary/checker/workspace test groups are green.

## Exact command results

Before Go tests, `TMPDIR=$PWD/.temp/TASK-261016-i5gwaf/tmp` was set where shown
in the replay instructions. Vendor runs additionally set `MUSE148_BINARY`,
`MUSE148_CURATOR` and `MUSE148_RUN_ROOT` as documented in README, each replay
using a fresh task-local root. Shell gate commands ran directly with log
redirection, no tee pipeline; non-zero exits below are genuine failures.

| Command | Exit | Evidence / meaning |
|---|---:|---|
| `go build -o .temp/TASK-261016-i5gwaf/curator ./cmd/curator` | 0 | Private `build-02.log`; actual Curator source-built CLI, rerun after edits |
| Exact binary `--provider echo --version` under disposable env/sandbox | 0 | Private `version-01.log`; exact build recorded above |
| `go test ./.research/muse148 -count=1 -v` | **1** | `replay-03.log`; mandatory interactive qualification fails, not a green suite |
| `go vet ./.research/muse148` | 0 | Private `vet-03.log`; checked after last Go edit |
| `MUSE148_OMIT_EXCLUSION=exec go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/exec/linked/skills' -count=1 -v` | **1** | `plant-exec-01.log`; no-exclusion control passes, required-exclusion test observes one forbidden skill |
| `MUSE148_OMIT_EXCLUSION=interactive go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/interactive/linked/skills' -count=1 -v` | **1** | Private `plant-interactive-01.log`; non-discriminating with the already red baseline |
| `go test ./.temp/TASK-261016-i5gwaf/checker-mutant -run '^TestObservationChecker$' -count=1 -v` | **1** | `checker-mutant-01.log`; exact-one-skill allowance is killed |
| `$SPEC_PYTHON -B .research/muse148/reference-snapshot-check.py $PINNED_SPEC .research/muse148/testdata/snapshot` | 0 | Existing pinned spec `managed_home_snapshots.check_bundle` accepts actual fixture schema, CCJ-1 bytes, inventory and identities; private `snapshot-reference-02.log` |
| `MUSE148_SNAPSHOT_PLANT=extra $SPEC_PYTHON -B $TASK_ROOT/reference-snapshot-check.py $PINNED_SPEC .research/muse148/testdata/snapshot` | **1** | Published checker raises precisely `ValueError: entry-set` |
| `sh scripts/remote-gate.sh` | **unrun** | Mandatory focused vendor qualification hit the external stop boundary first; no full-gate or review-handoff claim |

The reference invocation reuses the existing normative Python checker rather
than adding a second snapshot implementation. `$SPEC_PYTHON` is an existing
verified environment with the spec's dependencies, `$PINNED_SPEC` is an archive
of the exact recorded spec commit, and `$TASK_ROOT` is the task-local scratch
root. Its binding script is included in the public plain-text evidence packet for replay.
No package installation was needed.

## Coverage map and bounds

The README table covers every supplied observation row, including production
entrypoints and test names. A catalog-keyed surface table was **not supplied**;
this is recorded as a brief gap. The following map uses the declared leaf
attack surface without claiming that it is an assigned catalog key.

| Surface row | Attacking tests | Killed narrowing mutants | Out-of-contract inputs / clause |
|---|---|---|---|
| Vendor-build evidence authorizing form-specific exclusion | `TestMuse148LoaderMatrix`, `TestMuse148ParserRefusals`, `TestObservationChecker`, `TestFrozenSnapshot`, `TestMuse148WorkspaceResidual`, `TestMuse148OfflineBoundary` | Exec omission → matrix exec/linked/skills/exclude-true; exact-one-skill checker allowance → TestObservationChecker; extra snapshot member → reference check_bundle | Credential/login/refresh and model-visible removal: M2 “Separate evidence”; materialize/capture/publication/runtime-B: M2 snapshot-fixture scope; runtime-consumer constants/refusals: M1/M3 ownership |

| Mutant | What it narrows the gate to | Named failing test | Survivor bound |
|---|---|---|---|
| Omit exec exclusion only | Required exclusion admits the one foreign skill with native root still present | `TestMuse148LoaderMatrix/exec/linked/skills/exclude-true` | Killed; same baseline case green |
| Omit interactive exclusion only | Same synthetic foreign skill in interactive | `TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true` | **Non-discriminating survivor:** baseline already red, so no omission-specific coverage established |
| Allow exactly one user skill to satisfy zero | Retains record/type/other count checks and event tokens | `TestObservationChecker` | Killed by behavioral test, no static source-token attestation |
| Extra snapshot content member | Accept an unlisted file in the single-root closure | Published `managed_home_snapshots.check_bundle` | Intended `entry-set` failure, exit 1 |

No source-text scanner gate is introduced. Credentials and model-visible
precedence were not tested. Workspace `/rules` lists native user then project;
its loader diagnostics transition from project count 1 to 0 during the TUI
run, so enumeration is not behavioral precedence authority. Actual materialize
is a separate required acceptance row when that command ships.

## DoD disposition, item by item

- [ ] Every paired behavior satisfies AC: cases execute, but interactive
  root-present exclusion fails twice.
- [x] Bounded plant controls executed: exec omission, checker narrowing and
  snapshot extra-entry all fail for the intended property.
- [x] Paths bounded: evidence fixtures/tests plus required documentation;
  no runtime code or consumer constants modified.
- [ ] Focused tests and full gate green: focused exit 1; full gate unrun.
- [ ] Full task behavior/AC satisfied: vendor interactive behavior blocks it.
- [ ] All behavior tests passing: retained regression witnesses are red.
- [x] Candidate uncommitted on the supplied Story branch.
- [x] Every observation row driven or expressly bounded, ratio/call sites above.
- [x] Checker refusal/attestation has executable negative and positive controls.
- [ ] Every required gate qualified by killed mutant: interactive remains red
  and its omission plant is non-discriminating; no waiver claimed.
- [x] No source-text gate introduced; checker mutant retains event tokens and
  executes behavioral assertions.
- [x] Go formatting/vet clean, exact standalone checks recorded.
- [x] Curator builds; replay test package compiles and executes on intended host.
- [x] Task-scoped outcome and plain-text evidence packet attached through board CLI.
- [x] Regression, release effect and external resolution recorded in LOGBOOK
  and UNRESOLVED_QUESTIONS in this worktree.

## External blocker, options and recommendation

Constraint: required interactive boolean does not remove the foreign source
in the real loader on the exact tested build. Failed assumption: historical
1.4.7 interactive behavior transfers to 1.4.8. The boolean is accepted and the
string is rejected; real loader records and banner show the remaining skill,
so this is neither a parser failure nor a timeout, auth error or quota signal.
Links and regular-file roots reproduce it independently. No alternate flag,
faked ToolRelease, disabled gate or credential access was used to compensate.

Recommend vendor correction followed by exact-build replay and §7.9's
owner-managed erratum. Alternatively the normative owner may explicitly
restrict forms/builds or approve a different verified exclusion control, with
its compatibility tradeoff; this producer cannot choose that scope change.
Required external input: corrected vendor behavior or an explicit normative
owner resolution of this interactive exclusion regression. Keep the task
blocked and qualification/release pending; D10 admission remains independent.

`````

## checker-mutant-01.log

`````
=== RUN   TestObservationChecker
    replay_test.go:94: checker admitted one foreign skill
--- FAIL: TestObservationChecker (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/TASK-261016-i5gwaf/checker-mutant	0.286s
FAIL

`````

## observations.json

`````
{
  "version": "Muse Code 1.4.8 (1.4.8-R6256.1)",
  "binary_sha256": "5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c",
  "binary_locator": "$MUSE148_BINARY (exact private locator retained in ignored task-local identity.json)",
  "platform": "macOS 27.0 arm64",
  "curator_source": "0d00a109ad1da77a5506747c89f24af9442eb341",
  "spec_source": "d2a942cf17e4eb017c45475442da1456d8dfbe6c",
  "replay_test_sha256": "f4c1aa7a4e745bfc07d9b4536ab1173ea04d6c122af3849280af20360f80d3b1",
  "snapshot_manifest_sha256": "924378006ba6e6e7db9d0277375456a958c14cbaff03d1ebd63dfa6e4b7ceac8",
  "qualification": "blocked: interactive foreign skill remains loaded",
  "observations": [
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_claude-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your Claude Code personal rules \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-4834-76b1-8735-4ae2bff2035b\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-49f3-7483-b0a0-ab4f58674be1\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-49f3-7483-b0a0-ab4f58674be1\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-49f3-7483-b0a0-ab4f58674be1\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-4834-76b1-8735-4ae2bff2035b\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-49f3-7483-b0a0-ab4f58674be1\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-49f3-7483-b0a0-ab4f58674be1\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-49f3-7483-b0a0-ab4f58674be1\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_claude-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-4dc3-7fd2-bbb8-e0d7ffa633e8\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-4ed9-79b3-ba41-c14e7318c192\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-4ed9-79b3-ba41-c14e7318c192\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-4ed9-79b3-ba41-c14e7318c192\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-4dc3-7fd2-bbb8-e0d7ffa633e8\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-4ed9-79b3-ba41-c14e7318c192\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-4ed9-79b3-ba41-c14e7318c192\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-4ed9-79b3-ba41-c14e7318c192\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_codex-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your Codex personal rules \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-51bf-7793-9879-3dd11a7da7b0\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-5370-73d3-a8e7-18ca1feffcf1\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-5370-73d3-a8e7-18ca1feffcf1\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-5370-73d3-a8e7-18ca1feffcf1\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-51bf-7793-9879-3dd11a7da7b0\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-5370-73d3-a8e7-18ca1feffcf1\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-5370-73d3-a8e7-18ca1feffcf1\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-5370-73d3-a8e7-18ca1feffcf1\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_codex-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-5693-71a1-86ab-d61f62534d1e\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-578a-7501-bb96-23a50e56a238\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-578a-7501-bb96-23a50e56a238\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-578a-7501-bb96-23a50e56a238\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-5693-71a1-86ab-d61f62534d1e\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-578a-7501-bb96-23a50e56a238\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-578a-7501-bb96-23a50e56a238\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-578a-7501-bb96-23a50e56a238\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_skills_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-5a5f-72f0-8612-1aa2ffd1cf01\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-5bee-76f3-8dd6-51457583ce89\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-5bee-76f3-8dd6-51457583ce89\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-5bee-76f3-8dd6-51457583ce89\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-5a5f-72f0-8612-1aa2ffd1cf01\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-5bee-76f3-8dd6-51457583ce89\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-5bee-76f3-8dd6-51457583ce89\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-5bee-76f3-8dd6-51457583ce89\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_absent_skills_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-5eba-7a50-924c-97647dc75d30\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-5fab-7df1-92c7-a0ebd73eb6fe\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-5fab-7df1-92c7-a0ebd73eb6fe\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-5fab-7df1-92c7-a0ebd73eb6fe\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-5eba-7a50-924c-97647dc75d30\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-5fab-7df1-92c7-a0ebd73eb6fe\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-5fab-7df1-92c7-a0ebd73eb6fe\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-5fab-7df1-92c7-a0ebd73eb6fe\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_claude-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-6274-7192-8a2d-97157f7d755f\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-63f0-7a60-8fd4-6532ae7e552e\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-63f0-7a60-8fd4-6532ae7e552e\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-63f0-7a60-8fd4-6532ae7e552e\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-6274-7192-8a2d-97157f7d755f\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-63f0-7a60-8fd4-6532ae7e552e\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-63f0-7a60-8fd4-6532ae7e552e\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-63f0-7a60-8fd4-6532ae7e552e\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_claude-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-66ea-78b0-8b35-677d1271b913\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-67fc-7821-add2-0ffb73e69469\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-67fc-7821-add2-0ffb73e69469\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-67fc-7821-add2-0ffb73e69469\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-66ea-78b0-8b35-677d1271b913\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-67fc-7821-add2-0ffb73e69469\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-67fc-7821-add2-0ffb73e69469\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-67fc-7821-add2-0ffb73e69469\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_codex-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-6a9e-7080-b2ab-a3fdfe2d88da\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-6c21-70c1-bc0d-d6d9277fa115\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-6c21-70c1-bc0d-d6d9277fa115\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-6c21-70c1-bc0d-d6d9277fa115\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-6a9e-7080-b2ab-a3fdfe2d88da\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-6c21-70c1-bc0d-d6d9277fa115\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-6c21-70c1-bc0d-d6d9277fa115\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-6c21-70c1-bc0d-d6d9277fa115\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_codex-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-6ed5-7032-b367-00e0bb2a40b5\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-6ff6-7610-87ba-9cef6ca6f887\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-6ff6-7610-87ba-9cef6ca6f887\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-6ff6-7610-87ba-9cef6ca6f887\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-6ed5-7032-b367-00e0bb2a40b5\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-6ff6-7610-87ba-9cef6ca6f887\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-6ff6-7610-87ba-9cef6ca6f887\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-6ff6-7610-87ba-9cef6ca6f887\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_skills_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-72e3-7bc1-9b5b-23b2976a38cd\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-7409-7ee1-a753-e2d7d906fdda\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-7409-7ee1-a753-e2d7d906fdda\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-7409-7ee1-a753-e2d7d906fdda\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-72e3-7bc1-9b5b-23b2976a38cd\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-7409-7ee1-a753-e2d7d906fdda\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-7409-7ee1-a753-e2d7d906fdda\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-7409-7ee1-a753-e2d7d906fdda\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_linked_skills_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-769f-7123-904e-3dfd6666eb81\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-77d3-7c11-b6e1-3f495666a402\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-77d3-7c11-b6e1-3f495666a402\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-77d3-7c11-b6e1-3f495666a402\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-769f-7123-904e-3dfd6666eb81\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-77d3-7c11-b6e1-3f495666a402\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-77d3-7c11-b6e1-3f495666a402\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-77d3-7c11-b6e1-3f495666a402\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_claude-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-7a6c-7b43-a7e0-61b3bed97c04\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-7b8c-7691-a2a1-24fde94c391a\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-7b8c-7691-a2a1-24fde94c391a\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-7b8c-7691-a2a1-24fde94c391a\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-7a6c-7b43-a7e0-61b3bed97c04\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-7b8c-7691-a2a1-24fde94c391a\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-7b8c-7691-a2a1-24fde94c391a\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-7b8c-7691-a2a1-24fde94c391a\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_claude-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-7eac-7fb0-bc11-f7aee8564b6a\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-7f9d-7a53-a392-bff242d51377\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-7f9d-7a53-a392-bff242d51377\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-7f9d-7a53-a392-bff242d51377\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-7eac-7fb0-bc11-f7aee8564b6a\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-7f9d-7a53-a392-bff242d51377\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-7f9d-7a53-a392-bff242d51377\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-7f9d-7a53-a392-bff242d51377\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_codex-rules_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-8238-7611-ab43-25439134fb78\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-8349-78f1-b53a-fe8eb4368585\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-8349-78f1-b53a-fe8eb4368585\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-8349-78f1-b53a-fe8eb4368585\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-8238-7611-ab43-25439134fb78\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-8349-78f1-b53a-fe8eb4368585\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-8349-78f1-b53a-fe8eb4368585\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-8349-78f1-b53a-fe8eb4368585\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_codex-rules_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-864f-78f1-8e9b-2792143348e8\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-875d-7b83-8ad0-9bd883ece3a9\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-875d-7b83-8ad0-9bd883ece3a9\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-875d-7b83-8ad0-9bd883ece3a9\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-864f-78f1-8e9b-2792143348e8\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-875d-7b83-8ad0-9bd883ece3a9\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-875d-7b83-8ad0-9bd883ece3a9\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-875d-7b83-8ad0-9bd883ece3a9\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_skills_exclude-false",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "muse: Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-8b18-7840-af8e-adf4afd0f157\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-8c68-7721-b37f-151310813c87\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-8c68-7721-b37f-151310813c87\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-8c68-7721-b37f-151310813c87\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-8b18-7840-af8e-adf4afd0f157\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-8c68-7721-b37f-151310813c87\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-8c68-7721-b37f-151310813c87\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-8c68-7721-b37f-151310813c87\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_exec_snapshot_skills_exclude-true",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-8f45-75d2-afed-b3c761194bf7\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424a-90cf-7912-b20c-da5423f879a8\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424a-90cf-7912-b20c-da5423f879a8\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-90cf-7912-b20c-da5423f879a8\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424a-8f45-75d2-afed-b3c761194bf7\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424a-90cf-7912-b20c-da5423f879a8\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424a-90cf-7912-b20c-da5423f879a8\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424a-90cf-7912-b20c-da5423f879a8\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_claude-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0"
      ],
      "tui_observations": [
        "Including your Claude Code personal rules \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: ~/.claude/CLAUDE.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_claude-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=325 duration_ms=0"
      ],
      "tui_observations": [
        "Including your Claude Code personal rules \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: ~/.claude/CLAUDE.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_codex-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0"
      ],
      "tui_observations": [
        "Including your Codex personal rules \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: ~/.codex/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_codex-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=1 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=317 duration_ms=0"
      ],
      "tui_observations": [
        "Including your Codex personal rules \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: ~/.codex/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_skills_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: none",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_absent_skills_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"empty\" reason=\"none\" sources=0 user_sources=0 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=0 rendered_bytes=0 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: none",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_claude-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_claude-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_codex-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_codex-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_skills_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_linked_skills_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_claude-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_claude-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_codex-rules_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_codex-rules_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_skills_exclude-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148LoaderMatrix_interactive_snapshot_skills_exclude-true",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: none"
      ]
    },
    {
      "case": "TestMuse148ParserRefusals_interactive-exec-flag",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context"
      ],
      "vendor_exit": 2,
      "loader_records": [],
      "tui_observations": []
    },
    {
      "case": "TestMuse148ParserRefusals_interactive-string-false",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=\"false\""
      ],
      "vendor_exit": 2,
      "loader_records": [],
      "tui_observations": []
    },
    {
      "case": "TestMuse148WorkspaceResidual_exec_linked",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=2 user_sources=1 project_sources=1 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=2 rendered_bytes=1059 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424b-4d43-7580-9026-2f0e16784e38\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424b-4e88-7072-8439-1cffc5238cc5\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424b-4e88-7072-8439-1cffc5238cc5\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424b-4e88-7072-8439-1cffc5238cc5\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424b-4d43-7580-9026-2f0e16784e38\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424b-4e88-7072-8439-1cffc5238cc5\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424b-4e88-7072-8439-1cffc5238cc5\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424b-4e88-7072-8439-1cffc5238cc5\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148WorkspaceResidual_exec_snapshot",
      "argv": [
        "exec",
        "--json",
        "--provider",
        "echo",
        "--trust-workspace",
        "--no-foreign-personal-context",
        "MUSE148_OFFLINE_CONTROL"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=2 user_sources=1 project_sources=1 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=2 rendered_bytes=1059 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=0 user_skills=0 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0"
      ],
      "tui_observations": [
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c36f\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424b-5146-70e2-b944-e92af484812d\"},\"sequence\":18,\"recorded_at\":1780531400000031,\"record_type\":\"status\",\"durability\":\"ephemeral\",\"causation_id\":\"01a1424b-5263-7202-950e-da3c59e97b01\",\"payload_type\":\"run.output.delta\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_output_delta\",\"command_id\":\"01a1424b-5263-7202-950e-da3c59e97b01\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424b-5263-7202-950e-da3c59e97b01\"},\"text\":\"echo: MUSE148_OFFLINE_CONTROL\"}}",
        "{\"schema_version\":1,\"id\":\"018f0000-0000-7000-8000-00000000c391\",\"stream\":{\"kind\":\"session\",\"id\":\"01a1424b-5146-70e2-b944-e92af484812d\"},\"sequence\":35,\"recorded_at\":1780531400000065,\"record_type\":\"event\",\"durability\":\"durable\",\"causation_id\":\"01a1424b-5263-7202-950e-da3c59e97b01\",\"payload_type\":\"run.terminal.completed\",\"payload_schema_version\":1,\"payload\":{\"kind\":\"run_terminal\",\"command_id\":\"01a1424b-5263-7202-950e-da3c59e97b01\",\"run_stream\":{\"kind\":\"run\",\"id\":\"01a1424b-5263-7202-950e-da3c59e97b01\"},\"terminal\":\"completed\",\"text\":\"echo: MUSE148_OFFLINE_CONTROL\",\"reason\":null}}"
      ]
    },
    {
      "case": "TestMuse148WorkspaceResidual_interactive_linked",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=2 user_sources=1 project_sources=1 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=2 rendered_bytes=1059 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: AGENTS.md"
      ]
    },
    {
      "case": "TestMuse148WorkspaceResidual_interactive_snapshot",
      "argv": [
        "--provider",
        "echo",
        "--trust-workspace",
        "-c",
        "context.foreign_personal_skills=false"
      ],
      "vendor_exit": 0,
      "loader_records": [
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=2 user_sources=1 project_sources=1 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=2 rendered_bytes=1059 duration_ms=0",
        "event=\"skills.catalog_load\" outcome=\"ready\" reason=\"none\" skills=1 user_skills=1 project_skills=0 diagnostics=0 invalid=0 untrusted=0 shadowed=0 marker_invalid=0 duration_ms=0",
        "event=\"rules.context_load\" outcome=\"ready\" reason=\"none\" sources=1 user_sources=1 project_sources=0 foreign_sources=0 skipped_oversize=0 skipped_untrusted=0 authorization_blocks=1 rendered_bytes=971 duration_ms=0"
      ],
      "tui_observations": [
        "Including your 1 Claude Code personal skill \u2014 manage with /settings.",
        "\u25c6 echo: MUSE148_OFFLINE_CONTROL",
        "Rules loaded this session (in precedence order):",
        "user: $CONFIG_DIR/AGENTS.md",
        "project: AGENTS.md"
      ]
    }
  ]
}

`````

## plant-exec-01.log

`````
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:615: skills.catalog_load observed user_skills=1; want 0
--- FAIL: TestMuse148LoaderMatrix (3.73s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (1.05s)
    --- FAIL: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (1.32s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	3.908s
FAIL

`````

## plant-interactive-01.log

`````
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:615: skills.catalog_load observed user_skills=1; want 0
--- FAIL: TestMuse148LoaderMatrix (5.27s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (2.27s)
    --- FAIL: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (2.17s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	5.451s
FAIL

`````

## reference-snapshot-check.py

`````
import os, sys
from pathlib import Path
spec=Path(sys.argv[1]).resolve()
sys.path.insert(0,str(spec/'tools'))
from test_validate import validate
import managed_home_snapshots as snapshots
root=Path(sys.argv[2]).resolve()
manifest=validate.load_json(root/'manifest.json')
launch=validate.load_json(root/'launch.json')
content={p.relative_to(root).as_posix():(p.read_bytes(),f'{p.stat().st_mode & 0o777:04o}') for p in (root/'home').rglob('*') if p.is_file()}
if os.environ.get('MUSE148_SNAPSHOT_PLANT')=='extra': content['home/extra']=(b'extra','0600')
snapshots.check_bundle(manifest,launch,content,snapshots.validators(spec,validate.load_json),validate.ccj1_bytes,validate.validate_wire_semantics,(root/'manifest.json').read_bytes(),(root/'launch.json').read_bytes())
print('Reference check_bundle accepted exact frozen metadata and complete regular-file closure')

`````

## replay-03.log

`````
=== RUN   TestObservationChecker
--- PASS: TestObservationChecker (0.00s)
=== RUN   TestMuse148OfflineBoundary
--- PASS: TestMuse148OfflineBoundary (0.30s)
=== RUN   TestFrozenSnapshot
--- PASS: TestFrozenSnapshot (0.00s)
=== RUN   TestMuse148ParserRefusals
=== RUN   TestMuse148ParserRefusals/interactive-exec-flag
=== RUN   TestMuse148ParserRefusals/interactive-string-false
--- PASS: TestMuse148ParserRefusals (0.40s)
    --- PASS: TestMuse148ParserRefusals/interactive-exec-flag (0.11s)
    --- PASS: TestMuse148ParserRefusals/interactive-string-false (0.10s)
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:615: skills.catalog_load observed user_skills=1; want 0
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:617: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:615: skills.catalog_load observed user_skills=1; want 0
--- FAIL: TestMuse148LoaderMatrix (66.77s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false (1.42s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true (1.02s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false (1.24s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true (0.97s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-false (1.12s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-true (0.95s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false (1.15s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true (0.95s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false (1.08s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true (1.04s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (0.96s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (0.97s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false (1.08s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true (0.92s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false (1.05s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true (1.21s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (1.08s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (1.21s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false (2.21s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true (2.67s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false (2.18s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true (2.39s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false (2.62s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true (4.85s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false (4.36s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true (2.10s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false (2.09s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true (2.07s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (2.41s)
    --- FAIL: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (2.41s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false (2.69s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true (2.61s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false (2.18s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true (2.19s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (2.29s)
    --- FAIL: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (2.17s)
=== RUN   TestMuse148WorkspaceResidual
=== RUN   TestMuse148WorkspaceResidual/exec/linked
=== RUN   TestMuse148WorkspaceResidual/exec/snapshot
=== RUN   TestMuse148WorkspaceResidual/interactive/linked
=== RUN   TestMuse148WorkspaceResidual/interactive/snapshot
--- PASS: TestMuse148WorkspaceResidual (7.71s)
    --- PASS: TestMuse148WorkspaceResidual/exec/linked (1.03s)
    --- PASS: TestMuse148WorkspaceResidual/exec/snapshot (1.07s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/linked (2.39s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/snapshot (2.30s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	75.472s
FAIL

`````

## replay_test.go

`````
// Package muse148 records vendor-loader evidence; it never changes runtime admission.
package muse148

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const binarySHA = "5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c"
const rootSHA = "8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0"

func read(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func write(t *testing.T, p string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0600); err != nil {
		t.Fatal(err)
	}
}
func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func quote(s string) string  { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }

// checkCounts requires a real, successful loader record, including zero counts.
// Missing, failed, truncated and contradictory records cannot prove absence.
func checkCounts(log, event string, want map[string]int) error {
	seen := false
	for _, line := range strings.Split(log, "\n") {
		if !strings.Contains(line, `event="`+event+`"`) {
			continue
		}
		seen = true
		if !strings.Contains(line, `outcome="ready"`) && !strings.Contains(line, `outcome="empty"`) {
			return fmt.Errorf("%s failed loader record", event)
		}
		for key, n := range want {
			m := regexp.MustCompile(`(?:^| )` + key + `=([0-9]+)(?: |$)`).FindStringSubmatch(line)
			if m == nil {
				return fmt.Errorf("%s missing %s", event, key)
			}
			got, _ := strconv.Atoi(m[1])
			if got != n {
				return fmt.Errorf("%s observed %s=%d; want %d", event, key, got, n)
			}
		}
	}
	if !seen {
		return fmt.Errorf("missing %s loader observation", event)
	}
	return nil
}

// These controls prove that explicit clean records pass while a foreign source,
// missing/truncated record or failed read cannot be attested as clean.
func TestObservationChecker(t *testing.T) {
	clean := `event="rules.context_load" outcome="ready" user_sources=1 foreign_sources=0`
	want := map[string]int{"user_sources": 1, "foreign_sources": 0}
	if err := checkCounts(clean, "rules.context_load", want); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"", "timeout", strings.Replace(clean, "foreign_sources=0", "foreign_sources=1", 1), strings.Replace(clean, "foreign_sources=0", "foreign_sources=", 1), strings.Replace(clean, `outcome="ready"`, `outcome="failed"`, 1), clean + "\n" + strings.Replace(clean, "foreign_sources=0", "foreign_sources=1", 1)} {
		if checkCounts(bad, "rules.context_load", want) == nil {
			t.Fatalf("checker admitted %q", bad)
		}
	}
	if err := checkCounts(`event="skills.catalog_load" outcome="ready" user_skills=0`, "skills.catalog_load", map[string]int{"user_skills": 0}); err != nil {
		t.Fatal(err)
	}
	if checkCounts(`event="skills.catalog_load" outcome="ready" user_skills=1`, "skills.catalog_load", map[string]int{"user_skills": 0}) == nil {
		t.Fatal("checker admitted one foreign skill")
	}
}

type probe struct{ root, binary, curator, sandbox string }

func setup(t *testing.T) *probe {
	t.Helper()
	binary := os.Getenv("MUSE148_BINARY")
	if binary == "" {
		t.Skip("vendor replay requires MUSE148_BINARY, MUSE148_CURATOR and MUSE148_RUN_ROOT; no qualification from a skip")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatal("this evidence identity requires darwin/arm64")
	}
	base, err := filepath.Abs(os.Getenv("MUSE148_RUN_ROOT"))
	if err != nil || os.Getenv("MUSE148_RUN_ROOT") == "" {
		t.Fatal("explicit task-local run root required")
	}
	if !strings.Contains(base, string(filepath.Separator)+".temp"+string(filepath.Separator)) {
		t.Fatal("run root must be under .temp")
	}
	root := filepath.Join(base, strings.ReplaceAll(t.Name(), "/", "_"))
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	if digest(read(t, binary)) != binarySHA {
		t.Fatal("unqualified binary digest; this replay does not accept another build")
	}
	p := &probe{root: root, binary: binary, curator: os.Getenv("MUSE148_CURATOR"), sandbox: filepath.Join(root, "probe.sb")}
	if p.curator == "" {
		t.Fatal("MUSE148_CURATOR required")
	}
	// The real account's native stores are denied in addition to env -i and
	// disposable HOME/XDG. No credential is read to establish this boundary.
	native, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	sb := "(version 1)\n(allow default)\n(deny network*)\n(deny mach-lookup (global-name-regex #\"^com\\.apple\\.(security|Security)\"))\n(deny file-read* file-write* (regex #\"/Keychains(/|$)\"))\n"
	for _, rel := range []string{".claude", ".codex", ".config", ".local", ".muse", "Library"} {
		sb += "(deny file-read* file-write* (subpath " + strconv.Quote(filepath.Join(native, rel)) + "))\n"
	}
	write(t, p.sandbox, []byte(sb))
	return p
}

func (p *probe) env(dir string) []string {
	return []string{"HOME=" + filepath.Join(dir, "native"), "XDG_CONFIG_HOME=" + filepath.Join(dir, "config"), "XDG_DATA_HOME=" + filepath.Join(dir, "data"), "XDG_STATE_HOME=" + filepath.Join(dir, "state"), "XDG_CACHE_HOME=" + filepath.Join(dir, "cache"), "TMPDIR=" + filepath.Join(dir, "tmp"), "PATH=/usr/bin:/bin", "TERM=xterm-256color", "MUSE_NO_AUTO_UPDATE=1", "TBH_CREDENTIAL_BACKEND=file"}
}
func command(t *testing.T, dir string, env []string, name string, args ...string) ([]byte, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	b, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("bounded command timed out: %s", name)
	}
	if err == nil {
		return b, 0
	}
	if e, ok := err.(*exec.ExitError); ok {
		return b, e.ExitCode()
	}
	t.Fatal(err)
	return nil, -1
}
func (p *probe) fixture(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(p.root, "curator")
	env := p.env(dir)
	for _, sub := range []string{"native", "tmp", "config/muse", "work/.git"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"agent-context.json", "context/root.md"} {
		write(t, filepath.Join(dir, "source", f), read(t, filepath.Join("testdata/source", f)))
	}
	config := filepath.Join(dir, "manager/config.json")
	cfg, _ := json.Marshal(map[string]any{"schema_version": 1, "skills_root": filepath.Join(dir, "skills"), "projects": map[string]any{}})
	write(t, config, cfg)
	write(t, filepath.Join(dir, "system.json"), []byte(`{"schema_version":1}`))
	env = append(env, "CURATOR_CONFIG="+config, "CURATOR_SYSTEM_CONFIG="+filepath.Join(dir, "system.json"))
	for i, args := range [][]string{{"profile", "install", filepath.Join(dir, "source")}, {"env", "resolve", "muse", "--profile", "muse148-probe", "--repair", "--format", "json"}} {
		// Synthetic target permits the historical v3 resolver's metadata checks.
		// Muse never executes in this credential-bearing fixture home.
		if i == 1 {
			write(t, filepath.Join(dir, "config/muse/auth.json"), []byte("{}\n"))
		}
		b, code := command(t, filepath.Join(dir, "work"), env, p.curator, args...)
		write(t, filepath.Join(dir, fmt.Sprintf("curator-%d.log", i)), b)
		if code != 0 {
			t.Fatalf("Curator fixture command exit=%d: %s", code, b)
		}
	}
	link := filepath.Join(dir, "manager/environments/muse148-probe/muse/config/muse/AGENTS.md")
	s, err := os.Lstat(link)
	if err != nil || s.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("Curator did not produce a real link: %v", err)
	}
	if got := read(t, link); digest(got) != rootSHA || !bytes.Equal(got, read(t, "testdata/root.md")) {
		t.Fatal("Curator rendered root differs from frozen fixture")
	}
	return link
}

type observation struct {
	Log    string
	Output string
	Exit   int
}

func (p *probe) run(t *testing.T, form, representation, foreign string, exclude, workspace bool, link string, extra ...string) observation {
	t.Helper()
	dir := filepath.Join(p.root, strings.ReplaceAll(t.Name(), "/", "_"))
	for _, sub := range []string{"native", "config/muse", "data", "state", "cache", "tmp", "work/.git"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0700); err != nil {
			t.Fatal(err)
		}
	}
	root := filepath.Join(dir, "config/muse/AGENTS.md")
	launchEnv := p.env(dir)
	if representation == "linked" {
		if err := os.Symlink(link, root); err != nil {
			t.Fatal(err)
		}
	}
	if representation == "snapshot" {
		for _, f := range []string{"manifest.json", "launch.json", "home/config/muse/AGENTS.md"} {
			write(t, filepath.Join(dir, "snapshot", f), read(t, filepath.Join("testdata/snapshot", f)))
		}
		root = filepath.Join(dir, "snapshot/home/config/muse/AGENTS.md")
		for i, value := range launchEnv {
			if strings.HasPrefix(value, "XDG_") {
				parts := strings.SplitN(value, "=", 2)
				launchEnv[i] = parts[0] + "=" + filepath.Join(dir, "snapshot/home", filepath.Base(parts[1]))
				if err := os.MkdirAll(strings.SplitN(launchEnv[i], "=", 2)[1], 0700); err != nil {
					t.Fatal(err)
				}
			}
		}
		s, err := os.Lstat(root)
		if err != nil || !s.Mode().IsRegular() || s.Mode().Perm() != 0600 {
			t.Fatal("snapshot root is not an independent private regular file")
		}
	}
	switch foreign {
	case "claude-rules":
		write(t, filepath.Join(dir, "native/.claude/CLAUDE.md"), []byte("MUSE148_FOREIGN_CLAUDE_RULE\n"))
	case "codex-rules":
		write(t, filepath.Join(dir, "native/.codex/AGENTS.md"), []byte("MUSE148_FOREIGN_CODEX_RULE\n"))
	case "skills":
		write(t, filepath.Join(dir, "native/.claude/skills/muse148-foreign/SKILL.md"), []byte("---\nname: muse148-foreign\ndescription: Synthetic foreign personal skill for offline loader qualification.\n---\nMUSE148_FOREIGN_SKILL\n"))
	}
	if workspace {
		write(t, filepath.Join(dir, "work/AGENTS.md"), []byte("MUSE148_WORKSPACE_CONFLICT\n"))
	}
	before := map[string][]byte{}
	if err := filepath.WalkDir(filepath.Join(dir, "native"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			before[path] = read(t, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	args := []string{}
	if form == "exec" {
		args = append(args, "exec", "--json")
	}
	args = append(args, "--provider", "echo", "--trust-workspace")
	if exclude && os.Getenv("MUSE148_OMIT_EXCLUSION") != form {
		if form == "exec" {
			args = append(args, "--no-foreign-personal-context")
		} else {
			args = append(args, "-c", "context.foreign_personal_skills=false")
		}
	}
	args = append(args, extra...)
	if form == "exec" {
		args = append(args, "MUSE148_OFFLINE_CONTROL")
	}
	wrapped := append([]string{"-f", p.sandbox, "/usr/bin/env", "-i"}, launchEnv...)
	wrapped = append(wrapped, p.binary)
	wrapped = append(wrapped, args...)
	var out []byte
	code := 0
	if form == "exec" || len(extra) > 0 {
		out, code = command(t, filepath.Join(dir, "work"), p.env(dir), "/usr/bin/sandbox-exec", wrapped...)
	} else {
		out, code = p.tui(t, dir, wrapped)
	}
	write(t, filepath.Join(dir, "output.log"), out)
	for path, b := range before {
		if !bytes.Equal(read(t, path), b) {
			t.Fatalf("fixture source changed: %s", filepath.Base(path))
		}
	}
	if representation != "absent" {
		if digest(read(t, root)) != rootSHA {
			t.Fatal("native root bytes changed during vendor run")
		}
	} else if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatal("root-absent control changed")
	}
	for _, path := range []string{"config/muse/auth.json", "snapshot/home/config/muse/auth.json", "native/.config/muse/auth.json"} {
		if _, err := os.Lstat(filepath.Join(dir, path)); !os.IsNotExist(err) {
			t.Fatalf("credential store created or unreadable: %s: %v", path, err)
		}
	}
	var logs strings.Builder
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "cli-") && strings.HasSuffix(path, ".log") {
			logs.Write(read(t, path))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Persist only synthetic, sanitized loader observations for public evidence.
	safe := strings.ReplaceAll(string(out), dir, "$CASE")
	write(t, filepath.Join(dir, "argv.json"), mustJSON(t, args))
	write(t, filepath.Join(dir, "exit.txt"), []byte(strconv.Itoa(code)+"\n"))
	return observation{Log: logs.String(), Output: safe, Exit: code}
}

// A reachable local endpoint becomes unreachable only inside the probe policy;
// a synthetic Keychains file is denied while an ordinary file stays readable.
// No active credential store is accessed by this control.
func TestMuse148OfflineBoundary(t *testing.T) {
	p := setup(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte("MUSE148_NETWORK_CONTROL"))
	}))
	defer server.Close()
	args := []string{"--silent", "--show-error", "--noproxy", "*", "--connect-timeout", "2", server.URL}
	if b, c := command(t, p.root, os.Environ(), "/usr/bin/curl", args...); c != 0 || string(b) != "MUSE148_NETWORK_CONTROL" {
		t.Fatalf("network positive exit=%d: %s", c, b)
	}
	blocked := append([]string{"-f", p.sandbox, "/usr/bin/curl"}, args...)
	if b, c := command(t, p.root, os.Environ(), "/usr/bin/sandbox-exec", blocked...); c != 7 {
		t.Fatalf("network deny exit=%d: %s", c, b)
	}
	if calls.Load() != 1 {
		t.Fatal("denied network request reached the server")
	}
	for _, rel := range []string{"ordinary/control", "synthetic/Library/Keychains/control"} {
		write(t, filepath.Join(p.root, rel), []byte("MUSE148_SYNTHETIC_STORE_CONTROL"))
	}
	for _, tc := range []struct {
		path string
		code int
	}{{"ordinary/control", 0}, {"synthetic/Library/Keychains/control", 1}} {
		b, c := command(t, p.root, os.Environ(), "/usr/bin/sandbox-exec", "-f", p.sandbox, "/bin/cat", filepath.Join(p.root, tc.path))
		if c != tc.code {
			t.Fatalf("store boundary exit=%d want=%d: %s", c, tc.code, b)
		}
		if c != 0 && !strings.Contains(string(b), "Operation not permitted") {
			t.Fatalf("unrelated store refusal: %s", b)
		}
		if string(read(t, filepath.Join(p.root, tc.path))) != "MUSE148_SYNTHETIC_STORE_CONTROL" {
			t.Fatal("control bytes changed")
		}
	}
}

// Frozen metadata binds the single root-only closure; altered bytes, extra
// files and links refuse independently of successful vendor parsing.
// Published-schema validation is separately recorded in the evidence packet.
func TestFrozenSnapshot(t *testing.T) {
	var manifest struct {
		Files []struct {
			Path string
			Size int
			SHA  string `json:"sha256"`
			Mode string
		}
		Launch string `json:"launch_sha256"`
	}
	if err := json.Unmarshal(read(t, "testdata/snapshot/manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 1 || manifest.Files[0].Path != "home/config/muse/AGENTS.md" || manifest.Files[0].Mode != "0600" {
		t.Fatal("unexpected frozen closure")
	}
	if digest(read(t, "testdata/snapshot/launch.json")) != manifest.Launch {
		t.Fatal("descriptor identity differs")
	}
	entry := manifest.Files[0]
	check := func(root string) error {
		count := 0
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			count++
			info, err := d.Info()
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			if filepath.ToSlash(rel) != entry.Path || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
				return fmt.Errorf("fixture entry type/mode/set")
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if len(b) != entry.Size || digest(b) != entry.SHA || digest(b) != rootSHA {
				return fmt.Errorf("fixture byte identity")
			}
			return nil
		})
		if err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("fixture entry set")
		}
		return nil
	}
	root := t.TempDir()
	file := filepath.Join(root, entry.Path)
	frozen := read(t, "testdata/root.md")
	write(t, file, frozen)
	if err := check(root); err != nil {
		t.Fatal(err)
	}
	write(t, file, bytes.Replace(frozen, []byte("MUSE148_NATIVE_ROOT"), []byte("MUSE148_FORGED_ROOT"), 1))
	if check(root) == nil {
		t.Fatal("altered bytes admitted")
	}
	write(t, file, frozen)
	extra := filepath.Join(root, "home/extra")
	write(t, extra, []byte("extra"))
	if check(root) == nil {
		t.Fatal("extra entry admitted")
	}
	if err := os.Remove(extra); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs("testdata/root.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(abs, file); err != nil {
		t.Fatal(err)
	}
	if check(root) == nil {
		t.Fatal("linked snapshot admitted")
	}
}
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return append(b, '\n')
}

func (p *probe) tui(t *testing.T, dir string, args []string) ([]byte, int) {
	t.Helper()
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		t.Fatal(err)
	}
	// Relative AF_UNIX address avoids the platform path-length limit, while
	// keeping the socket inside the task-local worktree.
	socket := "tui.sock"
	call := func(a ...string) []byte {
		t.Helper()
		b, c := command(t, dir, os.Environ(), tmux, append([]string{"-S", socket}, a...)...)
		if c != 0 {
			t.Fatalf("tmux exit=%d: %s", c, b)
		}
		return b
	}
	var quoted []string
	for _, arg := range append([]string{"/usr/bin/sandbox-exec"}, args...) {
		quoted = append(quoted, quote(arg))
	}
	cmd := "cd " + quote(filepath.Join(dir, "work")) + " && " + strings.Join(quoted, " ") + "; printf '%s\\n' \"$?\" > " + quote(filepath.Join(dir, "tui-exit.txt"))
	call("new-session", "-d", "-x", "220", "-y", "70", "-s", "probe", cmd)
	pid := call("display-message", "-p", "#{pid} #{pane_pid}")
	write(t, filepath.Join(dir, "pids.txt"), pid)
	t.Cleanup(func() { command(t, dir, os.Environ(), tmux, "-S", socket, "kill-server") })
	deadline := time.Now().Add(15 * time.Second)
	var pane []byte
	for time.Now().Before(deadline) {
		pane = call("capture-pane", "-p", "-S", "-1000")
		if strings.Contains(string(pane), "Voice input") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(string(pane), "Voice input") {
		t.Fatalf("TUI not ready: %s", pane)
	}
	// Start a local echo turn so Muse binds and flushes its startup loader
	// diagnostics. The OS network deny remains active for the whole process.
	call("send-keys", "-l", "MUSE148_OFFLINE_CONTROL")
	time.Sleep(200 * time.Millisecond)
	call("send-keys", "Enter")
	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		pane = call("capture-pane", "-p", "-S", "-1000")
		if strings.Contains(string(pane), "echo: MUSE148_OFFLINE_CONTROL") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(string(pane), "echo: MUSE148_OFFLINE_CONTROL") {
		t.Fatalf("missing offline echo control: %s", pane)
	}
	call("send-keys", "-l", "/rules")
	time.Sleep(200 * time.Millisecond)
	call("send-keys", "Enter")
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		pane = call("capture-pane", "-p", "-S", "-1000")
		if strings.Contains(string(pane), "Rules loaded this session") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(string(pane), "Rules loaded this session") {
		t.Fatalf("missing real /rules observation: %s", pane)
	}
	call("send-keys", "-l", "/exit")
	time.Sleep(200 * time.Millisecond)
	call("send-keys", "Enter")
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if b, e := os.ReadFile(filepath.Join(dir, "tui-exit.txt")); e == nil {
			n, e := strconv.Atoi(strings.TrimSpace(string(b)))
			if e != nil {
				t.Fatal(e)
			}
			return pane, n
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("TUI did not exit normally")
	return nil, -1
}

// The real parser rejects the exec-only flag and a JSON string separately;
// the matrix below supplies the accepted boolean and exec positive controls.
func TestMuse148ParserRefusals(t *testing.T) {
	p := setup(t)
	for name, arg := range map[string][]string{"interactive-exec-flag": {"--no-foreign-personal-context"}, "interactive-string-false": {"-c", `context.foreign_personal_skills="false"`}} {
		t.Run(name, func(t *testing.T) {
			o := p.run(t, "interactive", "absent", "none", false, false, "", arg...)
			refusal := "invalid TUI options"
			if name == "interactive-string-false" {
				refusal = "the JSON type does not fit this setting"
			}
			if o.Exit != 2 || !strings.Contains(o.Output, refusal) {
				t.Fatalf("expected parser refusal exit=2, got %d: %s", o.Exit, o.Output)
			}
			if strings.Contains(o.Log, `event="rules.context_load"`) {
				t.Fatal("rejected parser input reached loader")
			}
		})
	}
}

// Each no-exclusion control must load its isolated foreign source; its paired
// exclusion case tests the actual form, native root representation and source.
func TestMuse148LoaderMatrix(t *testing.T) {
	p := setup(t)
	link := p.fixture(t)
	for _, form := range []string{"exec", "interactive"} {
		for _, root := range []string{"absent", "linked", "snapshot"} {
			for _, foreign := range []string{"claude-rules", "codex-rules", "skills"} {
				for _, exclude := range []bool{false, true} {
					name := fmt.Sprintf("%s/%s/%s/exclude-%t", form, root, foreign, exclude)
					t.Run(name, func(t *testing.T) {
						o := p.run(t, form, root, foreign, exclude, false, link)
						if o.Exit != 0 {
							t.Fatalf("vendor exit=%d: %s", o.Exit, o.Output)
						}
						users, fr, skills := 0, 0, 0
						if root != "absent" {
							users = 1
						}
						suppressed := exclude && (form == "exec" || root != "absent")
						if strings.HasSuffix(foreign, "rules") && root == "absent" && !suppressed {
							users = 1
							fr = 1
						}
						if foreign == "skills" && !suppressed {
							skills = 1
						}
						if err := checkCounts(o.Log, "rules.context_load", map[string]int{"user_sources": users, "foreign_sources": fr, "project_sources": 0}); err != nil {
							t.Fatal(err)
						}
						if err := checkCounts(o.Log, "skills.catalog_load", map[string]int{"user_skills": skills, "project_skills": 0}); err != nil {
							t.Fatal(err)
						}
						t.Logf("vendor exit=0; user_sources=%d foreign_sources=%d user_skills=%d; no model claim", users, fr, skills)
					})
				}
			}
		}
	}
}

// /rules and loader counts establish enumeration only. No model precedence
// or conflict resolution claim follows from both native and workspace loading.
func TestMuse148WorkspaceResidual(t *testing.T) {
	p := setup(t)
	link := p.fixture(t)
	for _, form := range []string{"exec", "interactive"} {
		for _, root := range []string{"linked", "snapshot"} {
			t.Run(form+"/"+root, func(t *testing.T) {
				o := p.run(t, form, root, "skills", true, true, link)
				if o.Exit != 0 {
					t.Fatalf("exit=%d", o.Exit)
				}
				if err := checkCounts(o.Log, "rules.context_load", map[string]int{"user_sources": 1, "foreign_sources": 0}); err != nil {
					t.Fatal(err)
				}
				// Interactive startup lists both sources, then a subsequent echo
				// reload reports project_sources=0. Preserve that residual rather
				// than asserting invariant workspace inclusion or precedence.
				observedProject := false
				for _, line := range strings.Split(o.Log, "\n") {
					if strings.Contains(line, `event="rules.context_load"`) && strings.Contains(line, " project_sources=1 ") {
						observedProject = true
					}
				}
				if !observedProject {
					t.Fatal("no observed native/workspace loader conflict")
				}
				if form == "interactive" && (!strings.Contains(o.Output, "project: AGENTS.md") || !strings.Contains(o.Output, "$CONFIG_DIR/AGENTS.md")) {
					t.Fatalf("missing /rules native/workspace order: %s", o.Output)
				}
			})
		}
	}
}

`````

## snapshot-plant.log

`````
Traceback (most recent call last):
  File "$TASK_ROOT/reference-snapshot-check.py", line 12, in <module>
    snapshots.check_bundle(manifest,launch,content,snapshots.validators(spec,validate.load_json),validate.ccj1_bytes,validate.validate_wire_semantics,(root/'manifest.json').read_bytes(),(root/'launch.json').read_bytes())
    ~~~~~~~~~~~~~~~~~~~~~~^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "$TASK_ROOT/spec-main/tools/managed_home_snapshots.py", line 81, in check_bundle
    raise ValueError("entry-set")
ValueError: entry-set

`````

## snapshot-reference.log

`````
Reference check_bundle accepted exact frozen metadata and complete regular-file closure

`````

## testdata/root.md

`````
<!--
curator-root-context-v2
root: muse148-probe 1.0.0 state sha256:730819a169c050ab2dbeabe75f0fea93395e67d14ca8d7d240cbebc1a326e968
member: muse148-probe 1.0.0 state sha256:730819a169c050ab2dbeabe75f0fea93395e67d14ca8d7d240cbebc1a326e968 weight 0
precedence: winner=higher-weight placement=winner-last
lock: sha256:458646149b0ef02eb037c74cb6fe38ca74b325c4be05bfbe5d9f8e302fb5f6fc
generated: Curator Protocol environments revision 1 (https://github.com/relux-works/curator-spec)
notice: generated file; direct edits are unsupported and are detected as drift; update the source profile repository or its composed profiles instead
-->

---

## Context: muse148-probe 1.0.0

MUSE148_NATIVE_ROOT

`````

## testdata/snapshot/home/config/muse/AGENTS.md

`````
<!--
curator-root-context-v2
root: muse148-probe 1.0.0 state sha256:730819a169c050ab2dbeabe75f0fea93395e67d14ca8d7d240cbebc1a326e968
member: muse148-probe 1.0.0 state sha256:730819a169c050ab2dbeabe75f0fea93395e67d14ca8d7d240cbebc1a326e968 weight 0
precedence: winner=higher-weight placement=winner-last
lock: sha256:458646149b0ef02eb037c74cb6fe38ca74b325c4be05bfbe5d9f8e302fb5f6fc
generated: Curator Protocol environments revision 1 (https://github.com/relux-works/curator-spec)
notice: generated file; direct edits are unsupported and are detected as drift; update the source profile repository or its composed profiles instead
-->

---

## Context: muse148-probe 1.0.0

MUSE148_NATIVE_ROOT

`````

## testdata/snapshot/launch.json

`````
{"descriptor":"snapshot-launch-v1","env":{"XDG_CACHE_HOME":"home/cache","XDG_CONFIG_HOME":"home/config","XDG_DATA_HOME":"home/data","XDG_STATE_HOME":"home/state"},"environment":"muse","permissions":{"locked":false,"mode":"native","source":"default"},"precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"458646149b0ef02eb037c74cb6fe38ca74b325c4be05bfbe5d9f8e302fb5f6fc","name":"muse148-probe"},"root_context":{"exclusive":true}}

`````

## testdata/snapshot/manifest.json

`````
{"adapter":"muse","curator_version":"0.15.0-rc.6.0.20261015232206-0d00a109ad1d+dirty","files":[{"mode":"0600","path":"home/config/muse/AGENTS.md","sha256":"8c9af1ca52b75aaa1e67af6365f1e39dfcc3c56a274d3571270079ef669ac9d0","size":692}],"fragment_sha256":"9eda0ac5bed34653262ee49b75b8e0bf7a74a29979ae9dfc272319b3333dfce9","launch_sha256":"2991bc9c97904d77aef70bb5c3ea7a7044ff462382310ae613bbd31f461fd5f1","profile":{"lock_sha256":"458646149b0ef02eb037c74cb6fe38ca74b325c4be05bfbe5d9f8e302fb5f6fc","name":"muse148-probe"},"snapshot":"managed-home-snapshot-v1"}

`````

## testdata/source/agent-context.json

`````
{"schema_version":1,"name":"muse148-probe","version":"1.0.0","context":{"modules":[{"path":"root.md"}]}}

`````

## testdata/source/context/root.md

`````
MUSE148_NATIVE_ROOT

`````
