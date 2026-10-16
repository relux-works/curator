# TASK-261016-i5gwaf — M2 qualification-record review handoff

Rework implements the convener's requested deliverable: an exact-build
[qualification record](../../docs/qualification/muse-1.4.8-R6256.1.md), with
exec and interactive results separated. **Exec loader observations pass;
interactive skill exclusion FAILS on this vendor build. The §7.9 interactive
release gate remains pending.** Green regression witnesses assert that
observed failure; they are not successful exclusion qualification.
Model-visible behavior and credential refresh remain unmeasured.

## Candidate and scope

- Curator checkpoint: `0d00a109ad1da77a5506747c89f24af9442eb341`; unchanged Story
  HEAD, no producer commit, branch switch or rebase. `HEAD..main` count: 0.
- Normative main: `d2a942cf17e4eb017c45475442da1456d8dfbe6c`, rechecked via
  GitHub; fetched environments bytes equal the pinned normative archive.
- Muse: `1.4.8-R6256.1`, binary SHA-256
  `5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c`.
  Exact private locator remains in ignored task-local `identity.json`;
  public artifacts use `$MUSE148_BINARY` for path hygiene.
- Environment: macOS 27.0 arm64, Go 1.25.5, tmux 3.6a, sandbox-exec network
  and Keychain deny, disposable HOME/XDG, offline echo, auto-update disabled.
- Only the evidence/probe paths and required README, LOGBOOK and
  UNRESOLVED_QUESTIONS documentation change. Canonical record added under
  `docs/qualification/` as explicitly required by rework. No runtime source,
  constants, gates, consumer tests, credential policy or CI config changes.
- All original test names remain. `TestMuse148InteractiveExclusionRegression`
  explicitly binds both native-root representations to the observed retained
  skill and points to the record when its observations change.
- Probe exit-status publication is now atomic: the first rework replay hit
  an empty status-file read between shell redirection and printf. That run
  is reported red, not reused as green evidence. No vendor assertion was
  weakened to handle this harness race.

## Coverage and bounds

**6 of 6 in-scope AC observation rows driven; 5 of 6 meet the narrow loader
contract.** The canonical record's coverage map names every production call
site and test for every supplied observation row. There is no catalog-keyed
surface table in the brief; this gap is declared. Declared surface:
vendor-build evidence authorizing form-specific exclusion.

The 36-case matrix covers both real forms, absent/link/snapshot roots,
separate Claude rules/Codex rules/foreign skills and both exclusion states.
There are two independent parser refusals, four workspace calls and two
explicit interactive-regression calls: 44 vendor entry cases. Native roots
continue loading while exec removes the foreign skill; interactive retains
it on both roots. The matrix records observations, and the dedicated witness
must fail on a corrected loader. Checker tests continue to reject retained
skills as evidence of zero skills. These expectations do not qualify a build.

Out-of-contract rows and authorizing clauses:

| Row | Clause / bound |
|---|---|
| Credential backend/login/refresh; model-visible foreign removal | M2 “Separate evidence”; no native credentials accessed |
| Model-visible workspace precedence | M2 workspace residual: enumeration/order/counts only; behavior unknown |
| Export/capture/publication/destination admission/runtime-B | M2 narrow frozen snapshot-fixture scope |
| Actual `env materialize -> Muse` | M2 separate required integration when materialize ships |
| Consumer build/serve refusals and registry constants | M1/M3 ownership; vendor replay is not their substitute |

## Exact commands and exits

Commands run directly, with file redirection, no tee pipeline. Logs reside
under ignored `.temp/TASK-261016-i5gwaf/`; sanitized task evidence is attached.
The large unrelated local-default diagnostic log is board evidence only;
probe and mutant logs remain beside the reproduction fixtures.
Vendor commands additionally set the three required `MUSE148_*` locators and
private `TMPDIR` as in the probe README. Each command uses a fresh run root.

| Command | Exit | Evidence / meaning |
|---|---:|---|
| `go build -o .temp/TASK-261016-i5gwaf/curator ./cmd/curator` | 0 | `build-rework-01.log` |
| Exact sandboxed binary `--provider echo --version` | 0 | `version-rework-01.log`, exact build above |
| `go test ./.research/muse148 -run '^(TestObservationChecker\|TestFrozenSnapshot)$' -count=1 -v` | 0 | `focused-rework-01.log` |
| First rework `go test ./.research/muse148 -count=1 -v` | **1** | `replay-rework-01.log`: empty status-file read; witness itself passed; superseded by atomic-publication replay |
| Reworked `go test ./.research/muse148 -count=1 -v` | 0 | `replay-rework-03.log` (final source; `replay-rework-02.log` also green) |
| `go vet ./.research/muse148` | 0 | `vet-rework-03.log` |
| Initial `golangci-lint run ./.research/muse148/...` | **1** | `lint-rework-01.log`: unused request parameter; fixed by `_` |
| `golangci-lint run ./.research/muse148/...` | 0 | `lint-rework-04.log`, candidate after atomic-publication edit |
| `go build ./...` | 0 | `build-all-rework-01.log` |
| `go test ./...` | **1** | `default-test-rework-01.log`; ambient installed provider causes status failures; cmd/curator, envprofile, install also hit default 10m timeout; no vendor probe in this suite |
| Clean-PATH `go test ./cmd/curator -run '^(TestDraftStatusReportsUpToDateThroughCLI\|TestDraftStatusNeedsInstallAfterRefreshWithoutInstall\|TestDraftStatusLiveMutationWithoutRefreshStaysCurrent\|TestEnvStatusMatrix\|TestEnvStatusRegistryBoundaryPostureAndCheck)$' -count=1 -timeout=4m -v` | 0 | `default-status-isolated-01.log`; all five status failures pass when the ambient installed provider is excluded |
| `sh scripts/remote-gate.sh` | 0 | `remote-gate-rework-01.log`; CI run 38835589308 |
| Exec omission: `MUSE148_OMIT_EXCLUSION=exec go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/exec/linked/skills' -count=1 -v` | **1** | `plant-exec-rework-03.log`; intended skill-count failure, valid control green |
| Interactive omission: `MUSE148_OMIT_EXCLUSION=interactive go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/interactive/linked/skills' -count=1 -v` | 0 | `plant-interactive-rework-03.log`; **survivor**, not exclusion coverage |
| `go test ./.temp/TASK-261016-i5gwaf/checker-mutant-rework-03 -run '^TestObservationChecker$' -count=1 -v` | **1** | `checker-mutant-rework-03.log`; intended retained-skill attestation failure |
| `$SPEC_PYTHON -B .research/muse148/reference-snapshot-check.py $PINNED_SPEC .research/muse148/testdata/snapshot` | 0 | `snapshot-reference-rework-01.log` |
| Same reference check with `MUSE148_SNAPSHOT_PLANT=extra` | **1** | `snapshot-plant-rework-01.log`; intended `ValueError: entry-set` |
| `go test ./.temp/TASK-261016-i5gwaf/snapshot-mutant-rework-03 -run '^TestFrozenSnapshot$' -count=1 -v` | **1** | `snapshot-mutant-rework-04.log`; admits only `home/extra`, intended `extra entry admitted` failure |
| Offline loopback-only policy exception: `go test ./.temp/TASK-261016-i5gwaf/offline-mutant-rework-03 -run '^TestMuse148OfflineBoundary$' -count=1 -v` | **1** | `offline-mutant-rework-05.log`; intended network-deny failure, synthetic endpoint becomes reachable |
| Snapshot exec omission: `MUSE148_OMIT_EXCLUSION=exec go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/exec/snapshot/skills' -count=1 -v` | **1** | `plant-exec-snapshot-rework-01.log`; intended retained-skill failure, no-exclusion control green |
| Snapshot interactive omission: `MUSE148_OMIT_EXCLUSION=interactive go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/interactive/snapshot/skills' -count=1 -v` | 0 | `plant-interactive-snapshot-rework-01.log`; **survivor** on this representation too |
| `git diff --check` | 0 | `diff-check-rework-01.log` |

The reference checker is the existing published snapshot implementation, not
another host-test language. A task-local Python 3.14.7 venv installs the spec's
`requirements-dev.txt`: `jsonschema==4.25.1`; exact commands are in the probe
README and resolved dependencies in `reference-environment-rework-01.log`.
The installation is confined to ignored task-local scratch state.

The isolated status retry uses `PATH="$GOROOT/bin:/opt/homebrew/bin:/usr/bin:/bin"`;
all five named status tests pass without the ambient installed provider.
The local 10m timeout results remain unverified, not green. Hosted gates use
the repository's configured 60m/120m package budgets and conformance pins;
no local mirror or fabricated hosted status is claimed.

Hosted snapshot: `9e6b55f4128ac8e92f11398ce7a23089fcd8a72d`, CI
[run 38835589308](https://github.com/relux-works/curator/actions/runs/38835589308).
`git diff --exit-code 9e6b55f4128ac8e92f11398ce7a23089fcd8a72d -- cmd internal go.mod go.sum .github scripts Makefile`
exits 0 (`hosted-source-identity-01.log`). The snapshot precedes only
probe-harness/comment and evidence-document updates. Default source/tests/configuration remain byte-identical; the final
opt-in source is separately replayed on the exact vendor binary. Evidence-only
updates after a gate do not imply execution of the excluded vendor package.

## Mutant evidence

| Mutant | What it narrows the gate to | Named failing test | Survivor bound |
|---|---|---|---|
| Omit exec exclusion only | Admit exactly the synthetic foreign skill while native root remains | `TestMuse148LoaderMatrix/exec/linked/skills/exclude-true` and `TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true` | Killed on both roots, exit 1; no-exclusion control green |
| Omit interactive exclusion only | Same retained skill with the accepted boolean omitted | None — **survivor**, exit 0 | Linked and snapshot roots both retain the skill with or without the boolean; no interactive protection established |
| Permit one user skill to satisfy zero | Keep event tokens, record/type/all other count checks; admit exactly one forbidden skill | `TestObservationChecker` | Killed, exit 1; behavioral suite executes, no static token scan |
| Ignore only the extra `home/extra` entry | Retain byte/type/mode/inventory checks for every other entry | `TestFrozenSnapshot` | Killed with `extra entry admitted`, exit 1 |
| Allow only outbound loopback TCP in otherwise denied policy | Network deny remains; synthetic local endpoint is reachable | `TestMuse148OfflineBoundary` | Killed with `network deny exit=0`, exit 1; no vendor/model process runs in this control |
| Add one unlisted snapshot file | Challenge exact single-root closure admission | Published `managed_home_snapshots.check_bundle` | Rejected with `entry-set`, exit 1; input plant, not a claimed source mutant |

Two initial additional-control attempts did not establish kills: snapshot
mutation without copied `testdata` failed setup (exit 1, `snapshot-mutant-rework-03.log`);
network policy attempts using an unsupported host address failed policy
parsing (exit 1, `offline-mutant-rework-03.log` and `offline-mutant-rework-04.log`).
Only the corrected, applied controls above count as evidence.

Reproducible narrowing patches: [mutants](mutants/README.md). All three
apply cleanly to the final replay source (`git apply --check`, exit 0).

No source-text scanner is introduced. No reviewer tests are removed/renamed.
Historical red qualification assertions and first-run sanitized observations
remain available beside the rework evidence, without relabeling them green.

## DoD disposition

- [x] Exact version, build, digest, platform, fixtures, commands and channels recorded.
- [x] Disposable offline boundary paired controls executed; native stores untouched.
- [x] Real exec acceptance and independent interactive flag/string-false refusals driven.
- [x] Root-absent/present foreign rule/skill controls driven on both forms.
- [x] Both forms and linked/regular-file native roots recorded separately.
- [x] Interactive boolean without native root retains rules fallback.
- [x] Real Curator-generated link matches independently checked frozen snapshot.
- [x] Workspace order/count transition recorded; model precedence remains unknown.
- [x] Checker, exec omission and snapshot plant controls fail for intended properties.
- [x] Interactive omission survivor explicitly declared; no false exclusion claim.
- [x] Model-visible behavior/refresh not measured; §7.9 interactive gate pending.
- [x] New exact-build regression witness added; default suite excludes vendor replay.
- [x] Original reviewer test names retained; no runtime constants/gates changed.
- [x] Code/probe paths bounded; required documentation exceptions justified above.
- [x] Focused vendor observations green; local default-suite red explicitly recorded.
- [x] Full remote gate green: 20 successful jobs; 2 configured optional jobs skipped.
  Default test suites pass on Linux/macOS/Windows; race passes on Linux/macOS.
  Local unfiltered default command remains red as reported; no silent relabeling.
- [x] Build, focused lint and vet pass after code changes.
- [x] Candidate remains uncommitted on supplied Story checkpoint.
- [x] New task-scoped qualification record, probe packet, local diagnostic and
  handoff evidence attached before role handoff.
- [x] Findings/release effect recorded in worktree LOGBOOK and unresolved questions.
- [ ] Required interactive exclusion actually removes skills — **vendor failure**;
  superseded as a producer requirement by convener rework, never represented as green.

## Hosted gate and attached resources

`sh scripts/remote-gate.sh` exited **0**. GitHub run 38835589308 reports
`completed/success` on exact snapshot `9e6b55f4128ac8e92f11398ce7a23089fcd8a72d`.
All 20 executed jobs passed. The configured optional native-platform test
and candidate-conformance job were skipped; neither is claimed as executed.
The working-tree HEAD remains the supplied checkpoint. The hosted source
identity comparison above and exact final-source native replay cover the
candidate's two test lanes without claiming that CI ran the opt-in vendor probe.

New task-scoped outcome resources:

- `TASK-261016-i5gwaf_rework1-qualification.md`
- `TASK-261016-i5gwaf_rework1-probe-evidence.md`
- `TASK-261016-i5gwaf_rework1-local-default-diagnostics.md`
- `TASK-261016-i5gwaf_rework1-handoff.md`

The last artifact includes this note, actual hosted result, sanitized gate
log and both supplemental snapshot omission logs. Prior blocked/red evidence
is preserved as historical evidence, not overwritten or relabeled green.

Release effect: keep qualification pending, coordinate any future corrected
build with M1/M3, and retain D10 credential admission as independent. This
qualification record is the review deliverable; no vendor fix is claimed.


## hosted-result-rework.json

````text
{
  "run": 38835589308,
  "head": "9e6b55f4128ac8e92f11398ce7a23089fcd8a72d",
  "status": "completed",
  "conclusion": "success",
  "url": "https://github.com/relux-works/curator/actions/runs/38835589308",
  "jobs": [
    {
      "name": "Test (ubuntu-latest)",
      "conclusion": "success"
    },
    {
      "name": "Gate self-test (ubuntu-latest)",
      "conclusion": "success"
    },
    {
      "name": "Gate self-test (macos-latest)",
      "conclusion": "success"
    },
    {
      "name": "Go 1.27.1 driver (ubuntu-latest)",
      "conclusion": "success"
    },
    {
      "name": "Test (windows-latest)",
      "conclusion": "success"
    },
    {
      "name": "Race (macos-latest)",
      "conclusion": "success"
    },
    {
      "name": "Go 1.26.0 driver (macos-latest)",
      "conclusion": "success"
    },
    {
      "name": "Go 1.26.0 driver (ubuntu-latest)",
      "conclusion": "success"
    },
    {
      "name": "Test (macos-latest)",
      "conclusion": "success"
    },
    {
      "name": "Go 1.26.0 driver (windows-latest)",
      "conclusion": "success"
    },
    {
      "name": "Interop conformance gate",
      "conclusion": "success"
    },
    {
      "name": "Go 1.25.5 driver (macos-latest)",
      "conclusion": "success"
    },
    {
      "name": "Gate self-test (windows-latest)",
      "conclusion": "success"
    },
    {
      "name": "Lint",
      "conclusion": "success"
    },
    {
      "name": "Naming gate",
      "conclusion": "success"
    },
    {
      "name": "Race (ubuntu-latest)",
      "conclusion": "success"
    },
    {
      "name": "Go 1.25.5 driver (windows-latest)",
      "conclusion": "success"
    },
    {
      "name": "Go 1.25.5 driver (ubuntu-latest)",
      "conclusion": "success"
    },
    {
      "name": "Go 1.27.1 driver (macos-latest)",
      "conclusion": "success"
    },
    {
      "name": "Go 1.27.1 driver (windows-latest)",
      "conclusion": "success"
    },
    {
      "name": "Test (optional native platform)",
      "conclusion": "skipped"
    },
    {
      "name": "Candidate suite (${{ matrix.os }})",
      "conclusion": "skipped"
    }
  ]
}

````


## remote-gate-rework-01.log

````text
remote gate: attempt 1/3: pushing 9e6b55f4128ac8e92f11398ce7a23089fcd8a72d as gate/STORY-261016-5y3h2r/261016-013130-40174-1
remote: 
remote: Create a pull request for 'gate/STORY-261016-5y3h2r/261016-013130-40174-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator/pull/new/gate/STORY-261016-5y3h2r/261016-013130-40174-1        
remote: 
remote gate: run 38835589308 (https://github.com/relux-works/curator/actions/runs/38835589308)
remote gate: run 38835589308 finished: success
  Test (ubuntu-latest): success  
  Gate self-test (ubuntu-latest): success  
  Gate self-test (macos-latest): success  
  Go 1.27.1 driver (ubuntu-latest): success  
  Test (windows-latest): success  
  Race (macos-latest): success  
  Go 1.26.0 driver (macos-latest): success  
  Go 1.26.0 driver (ubuntu-latest): success  
  Test (macos-latest): success  
  Go 1.26.0 driver (windows-latest): success  
  Interop conformance gate: success  
  Go 1.25.5 driver (macos-latest): success  
  Gate self-test (windows-latest): success  
  Lint: success  
  Naming gate: success  
  Race (ubuntu-latest): success  
  Go 1.25.5 driver (windows-latest): success  
  Go 1.25.5 driver (ubuntu-latest): success  
  Go 1.27.1 driver (macos-latest): success  
  Go 1.27.1 driver (windows-latest): success  
  Test (optional native platform): skipped  
  Candidate suite (${{ matrix.os }}): skipped  

````


## plant-exec-snapshot-rework-01.log

````text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:617: skills.catalog_load observed user_skills=1; want 0
--- FAIL: TestMuse148LoaderMatrix (3.43s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (1.29s)
    --- FAIL: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (1.29s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	3.843s
FAIL

````


## plant-interactive-snapshot-rework-01.log

````text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:619: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (5.24s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (2.19s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (1.85s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	5.427s

````
