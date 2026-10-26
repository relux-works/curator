# TASK-261026-jk47fn — Muse 1.4.13 exclusion replay handoff

## Per-form verdicts

| Form | Measured P1–P4 | Environments §7.9 | Versus 1.4.11 |
|---|---|---|---|
| exec | Exact flag removes foreign rules/skills; native linked/snapshot roots load | Qualified for measured macOS arm64 loader rows | Same |
| interactive | Valid boolean retains one foreign skill with both native roots | Refused; P4 fails, gate pending | Same |

Exact identity: Muse Code 1.4.13 (1.4.13-R6996.1), SHA-256
`9fd40ae9c6dfb847e3ef58b61405c004d0ebbf19ceb0297be966f750b3b84aa0`.
Curator checkpoint `c7b39b5474f68db1064256a91e59a9cdfb66b49b`; fetched spec
main `88b0d8615a5df8728e0e9540a9edd708c0ca5660`. Tools: Go 1.25.5,
tmux 3.8, golangci-lint 2.12.2, macOS 27.0.1 arm64 and sandbox-exec.
Only a read-only task-local copy was executed; installed source and updater
were untouched. Public locator is `$MUSE1413_BINARY`.

1.4.12-R6909.1 is skipped because intake reports it superseded on all known
hosts. It was not tested. The 45-case comparison uses 1.4.11-R6667.1,
not an unmeasured 1.4.12 baseline. All 45 semantic cases match.

## Coverage, retained tests and bounds

The [qualification record](../../docs/qualification/muse-1.4.13-R6996.1.md)
maps all six task AC rows to tests, real production entrypoints and controls.
The main package preserves intake, replay, contradictory-observation reviewer
regressions, physical source closure attacks and reference checker tests.
Current Curator drives real profile install and env resolve; the published
spec checker validates the frozen snapshot. Historical metadata remains
fixture provenance, not a current export claim.

The isolated P7 package explicitly answers the extra requirement:
**the 1.4.11 CIP-0002 resume/trust finding remains open and is reproduced
on 1.4.13.** Fresh untrusted exec excludes project rules/skill/hook/MCP,
while selected home sources work. Real `resume --last` offers trust/quit
without continue-untrusted. Granting fixture-only trust revives project
skill/hook/MCP and executes the project marker hook. `/rules` still lists
`project: none`; saved root context does not prove live catalog exclusion.
See [native excerpts](evidence/resume-observations.txt) and final-source
[resume replay](evidence/resume-04.log). This is distinct from foreign
personal P1–P4 qualification and from trusted-mode delivery.

Out of scope: real credentials/login/refresh, remote inference, model-visible
removal, live MCP handshake/tools, lazy discovery/native child execution,
permission-policy enforcement, full project-source exclusion, snapshot
export/publication/destination admission, serve/runtime consumer gates,
registry pins and other platforms. The exec config-key case measures only
parser acceptance. No broad isolation or consumer admission claim follows.

## Commands and real exits

Each command runs directly with stdout/stderr captured and its real process
exit recorded. Complete sanitized streams accompany
[commands.json](evidence/commands.json). JSON test streams have readable log
companions. Exact opt-in locators/run roots remain private; README documents
the invocation. Main and P7 source identities are in observations.json.

| Command / run | Real exit | Evidence |
|---|---:|---|
| `go build -o .temp/TASK-261026-jk47fn/curator ./cmd/curator` | 0 | [`build-curator-01.log`](evidence/build-curator-01.log) |
| `go test ./.research/muse1413/resume-trust -count=1 -v -json` | 1 | [`resume-01.log`](evidence/resume-01.log) |
| `go test ./.research/muse1413/... -count=1 -v` | 0 | [`ungated-01.log`](evidence/ungated-01.log) |
| `go vet ./.research/muse1413/...` | 0 | [`harness-vet-01.log`](evidence/harness-vet-01.log) |
| `golangci-lint run ./.research/muse1413/...` | 1 | [`harness-lint-01.log`](evidence/harness-lint-01.log) |
| `go test ./.research/muse1413/resume-trust -count=1 -v -json` | 0 | [`resume-02.log`](evidence/resume-02.log) |
| `go test ./.research/muse1413/resume-trust -count=1 -v -json` | 0 | [`resume-03.log`](evidence/resume-03.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-checker-one-skill -run ^(TestObservationChecker)$ -count=1 -v` | 1 | [`mutant-checker-one-skill-01.log`](evidence/mutant-checker-one-skill-01.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-checker-duplicate-count -run ^(TestReviewContradictoryObservation)$ -count=1 -v` | 1 | [`mutant-checker-duplicate-count-01.log`](evidence/mutant-checker-duplicate-count-01.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-checker-duplicate-outcome -run ^(TestReviewContradictoryObservation)$ -count=1 -v` | 1 | [`mutant-checker-duplicate-outcome-01.log`](evidence/mutant-checker-duplicate-outcome-01.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-snapshot-extra-entry -run ^(TestFrozenSnapshot)$ -count=1 -v` | 1 | [`mutant-snapshot-extra-entry-01.log`](evidence/mutant-snapshot-extra-entry-01.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-snapshot-go-listed-link -run ^(TestFrozenSnapshot\|TestFrozenSnapshotSourceInputs)$ -count=1 -v` | 1 | [`mutant-snapshot-go-listed-link-01.log`](evidence/mutant-snapshot-go-listed-link-01.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-snapshot-reference-listed-link -run ^(TestReferenceSnapshotPhysicalClosure)$ -count=1 -v` | 1 | [`mutant-snapshot-reference-listed-link-01.log`](evidence/mutant-snapshot-reference-listed-link-01.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-snapshot-reference-dangling-link -run ^(TestReferenceSnapshotPhysicalClosure)$ -count=1 -v` | 1 | [`mutant-snapshot-reference-dangling-link-01.log`](evidence/mutant-snapshot-reference-dangling-link-01.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-intake-one-foreign-build -run ^(TestVendorIntakeBoundary)$ -count=1 -v` | 1 | [`mutant-intake-one-foreign-build-01.log`](evidence/mutant-intake-one-foreign-build-01.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-offline-loopback -run ^(TestMuse1413OfflineBoundary)$ -count=1 -v` | 1 | [`mutant-offline-loopback-01.log`](evidence/mutant-offline-loopback-01.log) |
| `go test ./.temp/TASK-261026-jk47fn/mutant-offline-one-keychain-file -run ^(TestMuse1413OfflineBoundary)$ -count=1 -v` | 1 | [`mutant-offline-one-keychain-file-01.log`](evidence/mutant-offline-one-keychain-file-01.log) |
| `go test ./.research/muse1413/resume-trust -count=1 -v -json` | 0 | [`resume-04.log`](evidence/resume-04.log) |
| `go test ./.research/muse1413 -run ^TestMuse1413LoaderMatrix$/exec/linked/skills -count=1 -v -json` | 1 | [`plant-exec-01.log`](evidence/plant-exec-01.log) |
| `go test ./.research/muse1413/resume-trust -run ^TestOfflineBoundary$ -count=1 -v` | 1 | [`resume-control-network-01.log`](evidence/resume-control-network-01.log) |
| `go test ./.research/muse1413/resume-trust -run ^TestOfflineBoundary$ -count=1 -v` | 1 | [`resume-control-keychains-01.log`](evidence/resume-control-keychains-01.log) |
| `go test ./.research/muse1413 -run ^TestMuse1413LoaderMatrix$/interactive/linked/skills -count=1 -v -json` | 0 | [`plant-interactive-01.log`](evidence/plant-interactive-01.log) |
| `go build ./...` | 0 | [`build-01.log`](evidence/build-01.log) |
| `go vet ./.research/muse1413/...` | 0 | [`harness-vet-02.log`](evidence/harness-vet-02.log) |
| `go test ./.research/muse1413 -count=1 -v -json` | 0 | [`replay-01.log`](evidence/replay-01.log) |

Final-source checks additionally pass: [harness lint](evidence/harness-lint-03.log)
and [ungated tests](evidence/ungated-02.log), both exit 0. Their exact argv and
unset opt-ins are in commands.json.

The full main replay drives 45 vendor loader/parser cases (43 exit 0,
two exit 2) plus version inspection. Main offline controls and published
reference-checker positive/negative cases pass. The ungated explicit package
run verifies that vendor tests skip before fixtures; ordinary default Go
package selection does not include this hidden evidence tree.

Initial resume-01 failed only the port's mismatched synthetic hook literal;
resume-02 corrected it. Resume-03 added stronger pre-trust/home positives.
After removal of unused port code, resume-04 is the passing final-source
witness. Initial lint failures and corrected results remain recorded.
No failed setup or timeout is presented as passing evidence. See logbook for
the earlier discarded external-spec snapshot/dependency setup error.

## Narrowing controls

All ten patches were applied only to new disposable module copies, never
the candidate. Every exit 1 was inspected at its intended assertion;
no setup error, skip or unrelated crash counts as a kill.

| Mutant | Named failing test | Bound |
|---|---|---|
| `checker-one-skill` | `TestObservationChecker` | Exit 1; killed, [log](evidence/mutant-checker-one-skill-01.log) |
| `checker-duplicate-count` | `TestReviewContradictoryObservation` | Exit 1; killed, [log](evidence/mutant-checker-duplicate-count-01.log) |
| `checker-duplicate-outcome` | `TestReviewContradictoryObservation` | Exit 1; killed, [log](evidence/mutant-checker-duplicate-outcome-01.log) |
| `snapshot-extra-entry` | `TestFrozenSnapshot` | Exit 1; killed, [log](evidence/mutant-snapshot-extra-entry-01.log) |
| `snapshot-go-listed-link` | `TestFrozenSnapshot\|TestFrozenSnapshotSourceInputs` | Exit 1; killed, [log](evidence/mutant-snapshot-go-listed-link-01.log) |
| `snapshot-reference-listed-link` | `TestReferenceSnapshotPhysicalClosure` | Exit 1; killed, [log](evidence/mutant-snapshot-reference-listed-link-01.log) |
| `snapshot-reference-dangling-link` | `TestReferenceSnapshotPhysicalClosure` | Exit 1; killed, [log](evidence/mutant-snapshot-reference-dangling-link-01.log) |
| `intake-one-foreign-build` | `TestVendorIntakeBoundary` | Exit 1; killed, [log](evidence/mutant-intake-one-foreign-build-01.log) |
| `offline-loopback` | `TestMuse1413OfflineBoundary` | Exit 1; killed, [log](evidence/mutant-offline-loopback-01.log) |
| `offline-one-keychain-file` | `TestMuse1413OfflineBoundary` | Exit 1; killed, [log](evidence/mutant-offline-one-keychain-file-01.log) |
| Exec omission | `TestMuse1413LoaderMatrix/exec/linked/skills/exclude-true` | Exit 1, `user_skills=1; want 0`; no-exclusion control passes |
| Interactive omission | No failing assertion | Exit 0; declared survivor, vendor boolean already retains skill |
| P7 network deny removed | `TestOfflineBoundary` | Exit 1, curl succeeds with two hits instead of denial |
| P7 synthetic Keychains deny removed | `TestOfflineBoundary` | Exit 1, synthetic Keychains denial assertion |

The two offline patches and P7 denial controls run only curl/cat probes;
Muse is never run under weakened policy. The intake patch allows only one
synthetic digest and never executes those synthetic bytes. The checker
still refuses one skill when asked to attest zero.

## Hosted validation and handoff

The feature PR is the requested delivery boundary. Hosted CI must be green
on the delivered signed head; no old hosted result, local mirror or fabricated
status substitutes for it. Default hosted jobs validate normal repository
inputs; installed-vendor behavior is covered only by the explicit local
opt-in suites recorded here. Final hosted checks, review verdict, PR URL
and task/topic handoff are attached to TASK-261026-jk47fn.

Changelog: Add the Muse 1.4.13-R6996.1 offline qualification record: exec
loader rows qualify, interactive exclusion fails unchanged from 1.4.11;
CIP-0002 resume/trust finding remains open and reproduces on 1.4.13;
superseded 1.4.12 skipped. No runtime admission change.
