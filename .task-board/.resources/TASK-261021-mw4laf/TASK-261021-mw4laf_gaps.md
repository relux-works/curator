# TASK-261021-mw4laf — Muse 1.4.11 exclusion replay handoff

## Per-form verdicts

| Form | P1–P4 observation | Environments §7.9 exclusion gate | Versus 1.4.10 |
|---|---|---|---|
| exec | Exact flag removes foreign rules/skills; linked and snapshot native roots load | Qualified for measured macOS arm64 loader rows | Same |
| interactive | Valid boolean retains one foreign skill on both root representations | Refused with evidence; required P4 exclusion fails, gate pending | Same; 1.4.8/1.4.10 failure persists |

Binary: `Muse Code 1.4.11 (1.4.11-R6667.1)`, SHA-256
`6c04ce1719844685153ef6f04b98113cb61b942d97f0a8b0bfb3932c343984d9`.
The read-only pin-store source and task-local copy both matched. Only ignored
task-local `identity.json` contains their absolute locators. Public locator:
`$MUSE1411_BINARY`. The manager's pin store and updating launcher were untouched.

Curator checkpoint: `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`, zero commits
behind local main when measured. Normative spec main fetched and reread:
`e41f96387287f5e76604627f9e63662f7d6d86d4`. A task-local snapshot avoids
depending on the older active spec checkout. Sources: environments §§7.1
P1–P4, 7.9, 10.2; CIP-0017; Muse v4 conformance README.

The [qualification record](../../docs/qualification/muse-1.4.11-R6667.1.md)
records every P row, vendor exit and 1.4.10 comparison. The full local gated
run exited **0**, driving **45 loader/parser cases** (43 exit 0, two exit 2),
plus version inspection exit 0. [Observations](observations.json) bind the
cases to exact Go/Python source digests. The
[comparison](comparison-1.4.10.json) finds 45 of 45 cases unchanged for
normalized argv, exits, ordered loader event/outcome/count fields and
selected visible marker sets. Timings, marker extraction order and
sentinel-dependent byte lengths are not qualification verdicts.

Exec still accepts `-c context.foreign_personal_skills=false`, exit 0;
the anticipated parser refusal is not observed. Its exclusion semantics
under foreign inputs remain unmeasured. Interactive rejects the exec flag
and string false separately, exit 2, before rules loading. No loader-relevant
change was observed in the exercised interface. Additive MSP session/leave,
inbox notifications and recoveryMode were not qualified by this loader replay.

## Coverage, retained tests and bounds

**6 of 6 task AC rows driven**, mapped to named candidate tests and real
production call sites in the qualification record. No catalog-keyed surface
table was supplied: this is a brief gap. It is not an implicit waiver or a
claim to cover every runtime-consumer vector.

Entry points: current `curator profile install SOURCE`, `curator env resolve
muse --profile muse148-probe --repair --format json`, the vendor exec/TUI
parsers and rules/skills loaders, and published snapshot
`managed_home_snapshots.check_bundle`. `testdata` is byte-identical to
1.4.10/1.4.8; historical snapshot metadata does not attest a current export.

Out-of-contract rows and their acceptance boundaries:

- Credential login/backend/refresh, broker key delivery and model-visible
  foreign removal: the brief limits execution to offline echo P1–P4;
  CIP-0017 minimal probes 3–5 require separate evidence. No auth or remote
  inference command runs.
- Model-visible workspace precedence: the §7.9 workspace residual is
  enumeration evidence only. Both forms load native/project sources;
  interactive project count changes 1 → 0 on reload, as before.
- Capture locks, export provenance, publication, destination admission,
  runtime-B and materialize integration: the brief reuses a frozen snapshot;
  this leaf checks physical closure, representation and loading.
- Serve, runtime/launcher gates, registry pins and credential policy:
  expressly excluded by Scope. Their authority remains with their owners.
- Linux/Windows native vendor behavior: the exact executable is macOS arm64.
- Exec config-key exclusion semantics: parser acceptance is measured with
  no foreign input; no alternate consumer exclusion mapping is inferred.

Every previously committed reviewer test remains untouched in its original
module. The new module retains `TestReviewContradictoryObservation` under
its original name and all cases, plus both physical-link regression classes.
The local green suite executes all of them, including the reference CLI.
This is original development, not reviewer rework. All repository changes
are inside `.research/muse1411/` and the requested qualification record.
No root README, CHANGELOG, runtime, registry or policy changes.

No new source-token-scanning gate is introduced. The behavioral observation,
intake, closure and sandbox guards have narrowing controls. The conditional
source-token-preserving scanner mutant row is inapplicable to these guards.

## Commands and real exits

Each gate ran as a standalone process with direct file redirection; no
`tee` or status-masking pipe. [Commands](evidence/commands.json) and complete
sanitized local logs preserve results. README documents the opt-in locators
and fresh run roots. Expected-red mutation results are failing gates,
not passing test runs.

| Command | Real exit | Evidence / interpretation |
|---|---:|---|
| `go build -o .temp/TASK-261021-mw4laf/curator ./cmd/curator` | 0 | `evidence/build-curator-01.log`; production fixture binary |
| `go build ./...` | 0 | `evidence/build-01.log` |
| `go vet ./...` | 0 | `evidence/vet-01.log` |
| `env -u MUSE1411_BINARY go test ./.research/muse1411 -count=1 -v` | 0 | `evidence/ungated-01.log`; all seven vendor tests and unconfigured reference test skip; non-vendor controls pass |
| `go test ./.research/muse1411 -count=1 -v` with README locators | 0 | `evidence/replay-01.log`; full candidate-source gated suite, all 45 cases, version, offline and reference controls |
| `go vet ./.research/muse1411` | 0 | `evidence/harness-vet-01.log` |
| `golangci-lint run ./.research/muse1411/...` | 0 | `evidence/harness-lint-01.log`; zero issues |
| `MUSE1411_OMIT_EXCLUSION=exec go test ./.research/muse1411 -run '^TestMuse1411LoaderMatrix$/exec/linked/skills' -count=1 -v` | 1 | `evidence/plant-exec-01.log`; intended P4 failure, positive control passes |
| `MUSE1411_OMIT_EXCLUSION=interactive go test ./.research/muse1411 -run '^TestMuse1411LoaderMatrix$/interactive/linked/skills' -count=1 -v` | 0 | `evidence/plant-interactive-01.log`; survivor, proves no interactive protection |
| Ten patch-mutant behavioral commands | 1 each | `evidence/mutants-results.json`; each named failure inspected, table below |
| Existing hosted default stage: `go test -json -count=1 -timeout 60m PACKAGE...` | 0 | Exact unchanged-input reuse, not a local replay; `evidence/hosted-default-summary.json`, `evidence/hosted-default-exits.log` |
| Initial task-local hygiene audits | 1 each | Audit-only false positives: mandatory blank patch context, then relative `snapshot/home/config` misread as an absolute home path. First shell wrapper returned 0 after a subsequent status read; that did not make its failed Python audit green |
| Corrected task-local hygiene audit | 0 | `evidence/hygiene-03.log`; preserves valid patch encoding and distinguishes relative fixture paths from absolute personal paths; source/fixture identity and scope checks pass |
| `git diff --check` | 0 | `evidence/diff-check-01.log`; existing tracked paths unchanged; new-file whitespace separately audited |
| `gofmt -l .research/muse1411/*.go` | 0 | `evidence/gofmt-01.log`; empty output |
| `bash .github/ci/naming-gate.sh .research/muse1411` | 0 | `evidence/naming-module-01.log` |
| `bash .github/ci/naming-gate.sh docs/qualification` | 0 | `evidence/naming-records-01.log` |

Task-local reference venv creation and installation from the exact spec's
`requirements-dev.txt` both exited 0. This retains the template's existing
Python `jsonschema==4.25.1` reference-checker dependency; host tests stay in Go.
Tools: Go 1.25.5, tmux 3.6a, golangci-lint 2.12.2, macOS 27.0 arm64,
Apple sandbox-exec. Tool readiness and private raw logs stay under ignored
task-local `.temp/`.

Offline controls: curl exit 0 / 7 with exactly one server hit, ordinary versus
synthetic Keychains cat exit 0 / 1 with unchanged bytes. Security-service
lookups and real native config/library directories are denied. Every vendor
call uses env -i, four XDG parents, disposable HOME, file backend, echo and
auto-update disabled. Vendor credential files remain absent. No login,
logout, auth command, real Keychain item or remote inference is used.

## Narrowing controls

All patch mutants preserve their guard and admit one forbidden input class.
They were applied only in disposable task-local module copies; candidate
source was unchanged. Each exit 1 below is the intended assertion failure,
not an unrelated setup error. The offline mutants run only curl/cat controls,
never Muse under a weakened sandbox.

| Mutant | What it narrows the gate to | Named failing test | Real exit / bound |
|---|---|---|---|
| `checker-one-skill` | Admit exactly one skill where zero is required | `TestObservationChecker` | 1; killed |
| `checker-duplicate-count` | Admit zero then contradictory one | `TestReviewContradictoryObservation/count-zero-then-one` | 1; killed |
| `checker-duplicate-outcome` | Admit failed then ready | `TestReviewContradictoryObservation/outcome-failed-then-ready` | 1; killed |
| `snapshot-extra-entry` | Ignore only `home/extra` | `TestFrozenSnapshot/extra` | 1; killed |
| `snapshot-go-listed-link` | Follow only the listed root link | `TestFrozenSnapshot/listed-symlink`, `TestFrozenSnapshotSourceInputs/listed-symlink` | 1; killed |
| `snapshot-reference-listed-link` | Reference inventory follows listed root only | `TestReferenceSnapshotPhysicalClosure/listed-symlink` | 1; killed |
| `snapshot-reference-dangling-link` | Reference inventory omits hidden dangling link only | `TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link` | 1; killed |
| `intake-one-foreign-build` | Allow one synthetic wrong digest | `TestVendorIntakeBoundary/wrong-digest` | 1; killed; synthetic bytes are never executed |
| `offline-loopback` | Permit outbound localhost TCP only | `TestMuse1411OfflineBoundary` | 1; killed at observed curl exit 0 instead of 7 |
| `offline-one-keychain-file` | Permit only the synthetic Keychains control file | `TestMuse1411OfflineBoundary` | 1; killed at observed cat exit 0 instead of 1 |
| Exec omission plant | Omit only the required exec flag | `TestMuse1411LoaderMatrix/exec/linked/skills/exclude-true` | 1; killed, `user_skills=1; want 0`; no-exclusion control passes |
| Interactive omission plant | Omit only the interactive boolean | None — survivor | 0; real vendor exclusion already retains one skill. This establishes no interactive protection |

## Exact default-gate evidence reuse

Standing order 10 requires reuse of exact green evidence on unchanged
source/test/config/environment identity. Default production sources, tests,
validation inputs and workflows match baseline
`e88fdf2dbec7f94c48b0232a350a0a9f9484fdeb`: **1,671 of 1,671 manifest entries**
match; default-input manifest SHA-256
`baad451f52d27e33920e1aee5d1b91a53686416fa0fcd5c0c94af280ea4594a9`.
The only new package is hidden and excluded from default package selection.
[Reuse identity](evidence/reuse-identity.json) binds this comparison to the
candidate checkpoint. The new harness is separately exercised locally.

The original [hosted macOS job](https://github.com/relux-works/curator/actions/runs/39193852508/job/117877582459)
and its complete test-evidence artifact were retrieved again, exit 0 for
both downloads. The original full Go stream matches SHA-256
`3592b90d813b35f565b63fd1d0e5a71586f79b8b9a40418c4d148bb48381aca8`:
**85 of 85 default packages**, 80 pass, five no-test skips, zero failure
events. Hosted stage, overall Go command and platform gate each record
exit 0. Existing full gate evidence also records race, lint, interop,
gate-self-test, naming and Go-driver jobs green. Optional jobs were skipped.

The environment boundary is the original unchanged hosted workflow and
toolchain, not asserted equivalence to this local host. No local full default
suite was rerun, and no prior local timeout is presented as passing. This
uses already-green evidence, not an external-CI fallback or a forged status.
Full downloaded raw evidence remains private and ignored; the remote job,
stream digest, preserved prior record and new verification summary are auditable.

## Changelog line and handoff

Changelog: Add an opt-in Muse 1.4.11-R6667.1 offline P1–P4 record: exec loader rows qualified; interactive exclusion refused, unchanged from 1.4.10; no runtime admission change.

See [logbook](TASK-261021-mw4laf_logbook.md) for findings and workflow decisions.
The candidate remains uncommitted in the managed Story worktree for snapshot
handoff. Developer delivery is ready for review; no acceptance or landing is claimed.
