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
