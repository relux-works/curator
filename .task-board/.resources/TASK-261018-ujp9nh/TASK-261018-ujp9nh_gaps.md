# TASK-261018-ujp9nh — Muse 1.4.10 exclusion replay handoff

## Verdicts and bounds

| Form | P1–P4 loader result | Environments §7.9 gate | Qualification |
|---|---|---|---|
| exec | Exact exclusion flag removes foreign rules/skills; linked and snapshot native roots load | Measured exclusion replay passes | Qualified for these macOS arm64 loader rows |
| interactive | Valid boolean retains one foreign skill with either native root | Pending: required P4 exclusion fails | Refused with evidence; 1.4.8 failure persists |

Binary: `Muse Code 1.4.10 (1.4.10-R6427.1)`, SHA-256
`11cc1ae887b4975bbbe1de7358ee9260a612444df9ce4dbfffa1dd785826dbcf`.
Locator: `$MUSE1410_BINARY`, task-board Muse pin store, copied privately
without changing the manager's store. Every invocation sets
`MUSE_NO_AUTO_UPDATE=1`. Absolute locators remain only in ignored identity
and raw artifacts. Round 1 correctly stopped because the updater replaced
the named 1.4.10 executable with 1.4.11. **1.4.11 is the next build to qualify.**

Curator checkpoint: `e88fdf2dbec7f94c48b0232a350a0a9f9484fdeb`, zero commits
behind local main when measured. Spec main advertised/fetched/read:
`e41f96387287f5e76604627f9e63662f7d6d86d4`. Sources reread: environments
§§7.1 P1–P4, 7.9, 10.2; CIP-0017; Muse v4 conformance README.

The [qualification record](../../docs/qualification/muse-1.4.10-R6427.1.md)
contains every P row and its actual vendor exit. [Observations](observations.json)
bind 45 loader/parser cases (43 exit 0, 2 exit 2) to the final Go/Python
source digests. The separate identity test has version exit 0. The second
full opt-in run exercised the final harness source after the parser
observation correction. The first replay is historical evidence on the
pre-correction parser source; it is not substituted for the final run.

Contrary to the brief's anticipated refusal, the exec parser accepts
`-c context.foreign_personal_skills=false` and reaches the loader (exit 0).
The initial expected-refusal assertion failed, exit 1; a named observation
test now records acceptance. Its exclusion semantics on exec are unmeasured.
The two genuine interactive parser refusals remain exit 2 and never reach
rules loading. This parser difference does not establish an alternate
consumer spelling or change the required form-specific mapping.

## Coverage and excluded rows

**6 of 6 task AC rows driven**, mapped to named tests and production call
sites in the qualification record. No catalog-keyed surface table was
supplied: this is a brief gap, not an implicit completeness claim.

The real fixture call sites are `curator profile install`, `curator env
resolve muse --profile muse148-probe --repair --format json`, the vendor's
exec/TUI parsers and rules/skills loader, and the published snapshot
`managed_home_snapshots.check_bundle`. Unchanged 1.4.8 synthetic fixture
bytes are intentionally reused; the current resolver's link reproduces
them. Historical manifest metadata is not current export provenance.

Out-of-contract rows and acceptance boundaries:

- Credential login/backend qualification, refresh and broker keys: task
  description limits this leaf to P1–P4 loader replay; CIP-0017 minimal
  probes 3–5 require separate credential/broker evidence. No auth command
  or remote inference runs here.
- Model-visible foreign removal and workspace precedence: brief explicitly
  requests loader observations and a workspace residual, not a precedence
  guarantee. Native/project enumeration is measured for both forms.
- Capture locks, export provenance, publication, destination admission,
  runtime-B and materialize integration: brief reuses a frozen snapshot
  fixture; only its physical closure and loader representation are tested.
- Serve, runtime consumer gates/constants, registry pins and credential
  policy: explicitly excluded by task scope.
- Linux/Windows native vendor behavior: exact binary/platform identity is
  macOS arm64. No portable vendor qualification is claimed.

No source-token scanning gate exists in this module; the additional
source-token-preserving mutant condition is inapplicable. Behavioral checker,
physical closure, binary intake and sandbox boundaries have narrowing controls.
Every previously committed reviewer test remains unchanged in its original
module; the new module also retains `TestReviewContradictoryObservation`
under its original name and all cases. The fixture/reference regressions
are retained. This round is the resumed original intake, not reviewer rework;
all new files stay within `.research/muse1410/` and the named qualification
record. No root README, CHANGELOG, runtime, registry or policy edits.

## Commands and real exits

All gates ran as standalone processes, with direct file redirection and no
`tee`/status-masking pipe. Logs are sanitized copies under [evidence](evidence/).
The locator setup is documented in the module README; use fresh run roots.

| Command | Exit | Evidence / interpretation |
|---|---:|---|
| `go build -o .temp/TASK-261018-ujp9nh/curator ./cmd/curator` | 0 | `build-curator-01.log`; current production fixture binary |
| `go build ./...` | 0 | `build-01.log` |
| `go vet ./...` | 0 | `vet-01.log` |
| `go test ./...` | 1 | `default-test-01.log`: 77 packages pass, 5 have no tests, 3 hit the default 10-minute per-package timeout. Three CLI status tests also expose ambient user-bin providers; this run is failing, not green |
| `env -u MUSE1410_BINARY go test ./.research/muse1410 -count=1 -v` | 0 | `ungated-01.log`; vendor and unconfigured reference tests skip, non-vendor tests pass |
| `go test ./.research/muse1410 -run 'Test(VendorIntakeBoundary\|Muse1410Identity\|Muse1410OfflineBoundary\|Muse1410ParserRefusals\|Frozen\|Observation\|Review)' -count=1 -v` | 1 | `controls-01.log`; initial extra exec-key assertion actually ran and failed because vendor exit was 0 |
| `go test ./.research/muse1410 -run 'Test(VendorIntakeBoundary\|Muse1410Identity\|Muse1410OfflineBoundary\|Muse1410ParserRefusals\|Muse1410ExecInteractiveKeyObservation\|Frozen\|Reference\|Observation\|Review)' -count=1 -v` | 0 | `controls-02.log`; corrected observation plus published reference CLI and controls |
| `go test ./.research/muse1410 -run 'TestMuse1410(LoaderMatrix\|InteractiveExclusionRegression\|WorkspaceResidual)$' -count=1 -v` | 0 | `replay-01.log`; initial 42-case loader replay, failure reproduced |
| `go test ./.research/muse1410 -count=1 -v` | 0 | `replay-02.log`; final-source full gated run, all 45 cases plus version/reference/controls |
| `go vet ./.research/muse1410` | 0 | `harness-vet-01.log` |
| `golangci-lint run ./.research/muse1410/...` | 0 | `harness-lint-01.log`, zero issues |
| `MUSE1410_OMIT_EXCLUSION=exec go test ./.research/muse1410 -run '^TestMuse1410LoaderMatrix$/exec/linked/skills' -count=1 -v` | 1 | `plant-exec-01.log`; intended failure, forbidden skill retained |
| `MUSE1410_OMIT_EXCLUSION=interactive go test ./.research/muse1410 -run '^TestMuse1410LoaderMatrix$/interactive/linked/skills' -count=1 -v` | 0 | `plant-interactive-01.log`; survivor, no interactive protection proved |
| `git diff --check` | 0 | No tracked whitespace violations; new-file hygiene separately checked |
| Hosted default package stage, run `39193852508`, `test-gate.sh` → `go test -json -count=1 -timeout 60m PACKAGE...` | 0 | Original macOS job log explicitly records served stage exit 0, overall Go exit 0 and platform gate exit 0; all 85 default package names accounted for, see `hosted-default-summary.json` |

The reference CLI ran in a task-local venv using the spec's existing
`jsonschema==4.25.1` dependency. Installation exited 0. No new testing language
or library is introduced beyond the mirrored Go harness and published Python
checker wrapper. Tools: Go 1.25.5, tmux 3.6a, golangci-lint 2.12.2,
macOS 27.0 arm64, Apple sandbox-exec.

The unpartitioned default run is retained as failing evidence. The three
ambient-PATH failures are `TestDraftStatusReportsUpToDateThroughCLI`,
`TestDraftStatusNeedsInstallAfterRefreshWithoutInstall` and
`TestDraftStatusLiveMutationWithoutRefreshStaysCurrent`: status discovered
untrusted installed providers and returned 1. A standalone control using
only the Go toolchain and system/Homebrew bins ran all three green (exit 0,
`default-path-control-01.log`). No user installation was changed.

A bounded reproduction derives the complete test lists with real
`go test PACKAGE -list .` (exit 0 per package), then round-robin partitions
every listed Test/Example/Fuzz entry into four exact-name masks. The plan
contains 434 CLI, 393 envprofile and 371 install entries: 1,198 total,
with each assigned exactly once. Every slice is `go test PACKAGE -run MASK
-timeout=8m -count=1`. The other 82 packages are driven in one standalone
`go test -timeout=8m PACKAGE...` process. All bounded calls use the controlled
PATH; plan, commands and individual exits are recorded separately. This
attempt is diagnostic for changed environment/time partitioning, not a claim
that the original failing command passed. The shared host-GOROOT lock made
concurrent package slices unsuitable: CLI holds it for the whole process;
other CLI slices refused in TestMain after five minutes, and fingerprint
operations in other packages waited behind it. The owned waiters were stopped
by recorded PIDs. No lock or test guard was disabled. Remaining local slices
are **unrun/incomplete**, not green. The existing hosted full-suite artifact
below supplies accepted evidence for the unchanged default inputs; no full
local pass is claimed.

## Narrowing controls

These expected-red gates **failed with real exit 1**. A named assertion, not
an unrelated setup failure, killed each mutant. Individual commands and
exits are also preserved in `evidence/mutants-results.json` (the checker
one-skill and actual omission plants have their own logs).

| Mutant | What it narrows the gate to | Named failing test | Exit / bound |
|---|---|---|---|
| `checker-one-skill` | Accept exactly one skill when zero is required | `TestObservationChecker` | 1; killed |
| `checker-duplicate-count` | Accept contradictory zero then one | `TestReviewContradictoryObservation/count-zero-then-one` | 1; killed |
| `checker-duplicate-outcome` | Accept failed then ready | `TestReviewContradictoryObservation/outcome-failed-then-ready` | 1; killed |
| `snapshot-extra-entry` | Ignore exactly `home/extra` | `TestFrozenSnapshot/extra` | 1; killed |
| `snapshot-go-listed-link` | Follow only the listed root symlink | `TestFrozenSnapshot/listed-symlink`, `TestFrozenSnapshotSourceInputs/listed-symlink` | 1; killed |
| `snapshot-reference-listed-link` | Reference wrapper follows listed root only | `TestReferenceSnapshotPhysicalClosure/listed-symlink` | 1; killed |
| `snapshot-reference-dangling-link` | Reference wrapper omits hidden dangling link only | `TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link` | 1; killed |
| `intake-one-foreign-build` | Accept one synthetic wrong digest | `TestVendorIntakeBoundary/wrong-digest` | 1; killed |
| `offline-loopback` | Permit only outbound localhost TCP | `TestMuse1410OfflineBoundary` | 1; killed, no Muse invocation under weakened policy |
| `offline-one-keychain-file` | Permit reads of only the synthetic Keychains control file | `TestMuse1410OfflineBoundary` | 1; killed at store exit 0 vs required 1; no Muse invocation |
| Exec omission plant | Omit required flag for exec only | `TestMuse1410LoaderMatrix/exec/linked/skills/exclude-true` | 1; killed; no-exclusion control passes |
| Interactive omission plant | Omit boolean for interactive only | None — survivor | 0; broken vendor exclusion already retains one skill; this establishes no interactive protection |

Offline positive/negative controls: curl exit 0 / 7 with exactly one server
hit, ordinary/synthetic-Keychains cat exit 0 / 1 with unchanged bytes. Sandbox
also denies security-service lookups and real native config/library paths.
Every vendor argv selects echo and every vendor environment is cleared and
rebuilt with auto-update disabled. Vendor credential files stay absent.

## Landing evidence reuse

Unchanged default production sources, default tests, validation inputs and
workflow/configuration are bound to checkpoint
`e88fdf2dbec7f94c48b0232a350a0a9f9484fdeb`. The hosted workflow's own toolchain
and environment configuration is unchanged. Exact checkpoint CI
[39193852508](https://github.com/relux-works/curator/actions/runs/39193852508)
is completed/success; test, race, lint, interop, gate-self-test, naming and
Go-driver jobs are green. The optional self-hosted and candidate jobs were
skipped, not passing. `evidence/hosted-jobs.json` preserves the actual job
results with the private self-hosted runner label redacted. No hosted status is synthesized and no external-CI fallback is used.

Default-input tree-manifest SHA-256:
`baad451f52d27e33920e1aee5d1b91a53686416fa0fcd5c0c94af280ea4594a9`.
`evidence/reuse-identity.json` records the comparison. Added paths are only
the hidden opt-in module and this record; default Go package selection and
its source/test/config inputs remain identical. The new harness's changed
identity is covered by the local final gated run and lint, not borrowed CI.

## Changelog line and handoff

Changelog: Add an opt-in Muse 1.4.10-R6427.1 offline P1–P4 qualification record:
exec loader rows qualified; interactive exclusion refused with the 1.4.8
failure reproduced. No runtime admission change.

See [logbook](TASK-261018-ujp9nh_logbook.md) for decisions and anomalies.
The candidate stays uncommitted in the managed Story worktree for snapshot
handoff. No work is claimed accepted or landed by the developer.

## Bounded local default-suite diagnostics

These attempts do not replace the green hosted full-suite evidence. A
SIGTERM result means the owned concurrent waiter was stopped; its tests are
incomplete/unrun. Wrapper exit 241 is the shell representation of Python's
negative subprocess exit; the actual Go subprocess was signalled (-15).

| Package / slice | Real Go result | Observation |
|---|---:|---|
| `./cmd/curator` / 0 | 1 | 8-minute package timeout; partial body execution |
| `./cmd/curator` / 1 | 1 | 5-minute TestMain GOROOT-lock refusal; body unrun |
| `./cmd/curator` / 2 | 1 | 5-minute TestMain GOROOT-lock refusal; body unrun |
| `./cmd/curator` / 3 | 1 | 5-minute TestMain GOROOT-lock refusal; body unrun |
| `./internal/envprofile` / 0 | -15 | Incomplete: stopped owned lock waiter; wrapper exit 241 |
| `./internal/envprofile` / 1 | -15 | Incomplete: stopped owned lock waiter; wrapper exit 241 |
| `./internal/envprofile` / 2 | -15 | Incomplete: stopped owned lock waiter; wrapper exit 241 |
| `./internal/envprofile` / 3 | -15 | Incomplete: stopped owned lock waiter; wrapper exit 241 |
| `./internal/install` / 0 | 0 | Green, all 93 selected entries ran |
| `./internal/install` / 1 | -15 | Incomplete: stopped owned lock waiter; wrapper exit 241 |
| `./internal/install` / 2 | -15 | Incomplete: stopped owned lock waiter; wrapper exit 241 |
| `./internal/install` / 3 | -15 | Incomplete: stopped owned lock waiter; wrapper exit 241 |
| Remaining 82 packages | -15 | Stopped owned concurrent process; wrapper exit 241; incomplete/unrun |

The three ambient-provider regression controls were independently green
with the controlled PATH (exit 0). Local full-suite success is not asserted.
The original hosted artifact covers **85 of 85 default packages**: 80 pass,
5 with no tests, zero fail events. Its raw stream is SHA-256
`3592b90d813b35f565b63fd1d0e5a71586f79b8b9a40418c4d148bb48381aca8`.
`evidence/hosted-default-summary.json` records the verified stream and exact
checkpoint. The source/test/configuration is unchanged for those packages;
the original CI environment is preserved as that evidence's boundary, not
misrepresented as equivalent to the contended local host. This is exact
existing-evidence reuse under the assignment's standing order 10, not a local
CI fallback or a claim that unknown/interrupted results passed.
