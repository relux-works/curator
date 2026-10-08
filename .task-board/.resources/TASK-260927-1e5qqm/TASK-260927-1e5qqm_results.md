# TASK-260927-1e5qqm — flip-codex-seed-to-revision-b

## Implementation findings

Landing remains held until the operator schedules the B release. Changes remain uncommitted in the assigned Story worktree.

Revision A prerequisite verified directly: GitHub release v0.15.0-rc.3 is published (2026-10-04T03:03:40Z), not a draft, and its tagged envregistry source selects CodexSeedRevisionA. Release: https://github.com/relux-works/curator/releases/tag/v0.15.0-rc.3.

Fresh origin main advertisement, exact main fetch, worktree HEAD and checkpoint all equal e5489b6ba22c9e9cf7a925e03f3653d115188999. No branch switch, local commit, rebase, merge, or landing performed.

Registry selects Codex seed B. All B vectors run Resolve / StatusOf with an empty revision override (production default); A vectors retain explicit historical regression coverage. Added A-to-B status and repair preservation checks. CLI env resolve / env status tests assert stripping, B record, names and warning without a migration hint. Schema-1 A fixtures now identify their historical record explicitly.

Removed exactly 36 rows owned by this flip leaf: nine cases across each of four manifest identities. Other gaps remain untouched. rc.14 suite checked out at 43bf0a2506d5c354a73bbc3ea4623d4653db10c7; manifest SHA-256 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5. Intended family counts: provisioning 7/7 (B production 4/4, A historical 3/3); posture 8/8 (B production 5/5, A historical 3/3). Measured results will be appended after execution.

SecurityPostureRevision remains A and is separately owned. No LOGBOOK.md or CHANGELOG.md edit: the task-specific seedB brief explicitly prohibits both. Important findings are recorded here and in board notes instead.

Initial status command exited 1 because an estimate was required. Set Fibonacci estimate 3, then development transition exited 0. This is lifecycle repair, not a validation pass.

## Bounded seed validation

GOFLAGS=-work and the shared build lock applied to every Go invocation. The broader local command `go test -count=1 -timeout=5m -v ./internal/envprofile ./internal/envregistry` exited 1: envprofile reached its 5-minute package deadline in unrelated TestResolveDriftRepair. Its completed seed cases were green and envregistry passed; the remaining envprofile scope did not run, so this is not a passing package-suite claim.

Three subsequent lock acquisition attempts exited 75 without starting a validation child. These are not test passes or test failures. A longer bounded acquisition succeeded, and `go test -count=1 -timeout=3m -v ./internal/envprofile ./internal/envregistry -run 'Test(CodexSeed|EnvironmentsCodexSeedVectors|RegistryClosed)'` exited 0. Exact rc.14 tally: provisioning 7 driven / 7 published; posture 8 driven / 8 published; zero known gaps, bounds, or skips. Includes invalid-TOML refusal before publishing a home, inline-table stripping, immutable B repair snapshot, and historical A-home preservation under the B default.

`git diff --check`, formatting over changed Go files (zero unformatted), the protected-file diff check, and `bash .github/ci/no-broad-suppression.sh` each exited 0. A ledger comparison confirms exactly the 36 flip-owned removals, with no added rows and every foreign-owned row unchanged.

The configured hosted gate is `sh scripts/remote-gate.sh`. Per the managed-run lifecycle, Change Request construction/validation runs after this producer ends its turn, once the completion guard sees the role handoff and a new/updated outcome. Hosted CI has not run in this session yet and is not claimed green. The runtime-owned CR gate is the full-suite arbiter; do not land before its verdict and the operator's B release schedule.

## Validation scheduling note

The first CLI targeted invocation also exited 75 on bounded build-lock acquisition and never started its Go child. A second bounded invocation is pending. The lock belongs to another live validator; no foreign lock or process was removed. The three seed acquisition refusals and this CLI acquisition refusal are recorded separately from actual command verdicts.

## CLI production-entry validation

The targeted cmd/curator command (full mask recorded in cli-targeted.log) exited 0. Five Codex-related CLI checks passed: A formatter regression, B table seed and status, B inline seed and status, historical schema-1 A repair byte preservation, and pre-rule repair byte preservation with a B re-provision hint. Security-posture CLI vectors also passed with their existing stated bound: 13/17 driven, 4/17 bound, zero gaps/skips; this task does not claim the separately owned posture B flip. No full local cmd package run is claimed; the configured hosted suite must provide package-wide isolation evidence.

## Adversarial proof at Resolve

Both mutants reran `go test -count=1 -timeout=2m -v ./internal/envprofile -run TestCodexSeedStripsInlineMCPTable` directly under the build lock and GOFLAGS=-work. Each test command exited 1 (expected red), not a pass.

1. Revert the registry selector from B to A: the test failed because production emitted the A warning and no not-inherited warning. This proves the behavioral test reaches the shipped registry selector, rather than merely asserting a constant.
2. Narrow B stripping to configurations containing a `[mcp_servers.` header, retaining the original config otherwise: the test failed because the production seed still contained its inline mcp_servers table. This is a narrowing mutant, not a delete-only proof.

Each mutation restored the source from its exact pre-mutation byte copy in a finally block, and verified restoration. Managed source SHA-256 after restoration: 8b425ae08134cacb2e0cd22518e74eeaefa8c8d0b7a811a7a800170dc49b53d1. Registry source after restoration: aaa84a5676ebd0b1c6c6ea09037ee7ec52df227e85500e0d29412dba468f195a. All six candidate-file digests still match the pre-mutation candidate. The expected-red logs are retained explicitly as failures.

The first bounded build-check group exited 75 before any child command started because the shared build lock was held. Restored-positive/lint/vet/build remain unrun at that checkpoint; a retry is pending. Environment identity observed directly: go version go1.26.0 darwin/amd64 (command exit 0).

## Build checks and lint cache anomaly

After lock acquisition the restored positive inline-stripping test exited 0; scoped `go vet ./cmd/curator ./internal/envprofile ./internal/envregistry` exited 0; `go build -o .temp/TASK-260927-1e5qqm/curator ./cmd/curator` exited 0. Each ran as a standalone subprocess under the same shared lock with GOFLAGS=-work.

The initial `golangci-lint run --timeout=2m` invocation exited 1. Its shared cache returned findings whose filenames pointed into deleted, unrelated Story worktrees, with explicit file-not-found warnings. This is not a green lint claim and not a current-source fix request. Preserved as lint-shared-cache.log/json; rerunning the exact lint command with task-private GOLANGCI_LINT_CACHE. No global cache was cleared and no source suppression was added.

A mistaken retry invoked `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint --version` instead of the installed lint binary. It exited 1 (missing module), with no dependency or repository changes. It is preserved as lint-command-error.log/json and is not lint validation. Corrected to the installed `golangci-lint run --timeout=2m` binary with the task-private cache.

## Review handoff — current candidate

The corrected installed `golangci-lint run --timeout=2m` invocation with task-private GOLANGCI_LINT_CACHE exited 0 and reported zero issues. The earlier validation group exited 1 because its shared-cache lint command failed; scoped vet/build in that group individually exited 0. Neither the group nor the stale-cache lint failure is represented as green.

Current checks: targeted seed tests 0; targeted CLI tests 0; restored positive 0; task-private lint 0; scoped vet 0; CLI build 0; formatting 0; suppression check 0; diff check 0. Both mutants remain expected-red exit 1. The broader local package run remains failing exit 1 due its 5-minute deadline, with its remaining envprofile cases unrun. Five lock-acquisition refusals exited 75 without starting children. No prior attached green evidence was substituted for these checks.

All six candidate file digests equal the tested candidate; HEAD remains the recorded checkpoint. Dependency files, LOGBOOK.md and CHANGELOG.md are unchanged. Changes remain uncommitted. The public evidence archive normalizes home/worktree/temporary directory prefixes and contains command transcripts, actual exit codes, durations, suite identity and candidate file digests, without copying the normative corpus.

Hosted CR gate is pending runtime publication after this producer exits; no hosted-green or full-local-suite claim. Ready for independent review after the runtime gate verdict. LANDING HELD until the operator schedules the B release.

Evidence archive: TASK-260927-1e5qqm_validation-evidence.tar.gz; SHA-256 13fd8c1f01ffeba40ebd75cc6a4ebedc2ece1f0684c4b05092dba63dabd54d93.

## Re-applied after converge — revision 2 (2026-10-08)

Revision 1 was ACCEPTED and went stale only because main advanced (afcf1898, the v2 writer flip). The orchestrator converged this Story workspace onto main 75ab9a71; three files were carried unchanged (docs/ci-gates.md, internal/envprofile/codex_seed_test.go, internal/envregistry/envregistry.go) and the three conflicting files were reverted to the fresh base. This run performed no checkout, reset, or converge, and changed exactly those three files:

1. `.github/ci/conformance-gaps.tsv`: removed the same 36 rows owned by TASK-260927-1e5qqm (nine cases across each of four manifest identities), nothing else. Verified: zero owned rows remain (grep exit 1, no matches, as expected); working-tree diff is exactly 36 deletions with 0 additions; header, comments, blank separators, and every foreign-owned row are byte-identical. The fresh base had already dropped the unrelated rc.14 byte-exact-snapshot row, so the file now ends at the rc.14 comment.
2. `cmd/curator/env_credential_marker_test.go`: re-applied the 4-line intent onto the rewritten file — the historical-A `codex_seed_record` pin in TestEnvResolveKeepsSchema1BytesForMetadataOnly and the `seed revision A` → `seed revision B` expectation in TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome. Every new v1/v2 assertion is intact (pinV1MarkerWriters, schema-1 downgrade with hash_version deletion, the empty-A-record check, the pre-rule nil-record check, marker/config byte preservation). The B expectation matches production: internal/envprofile/status.go formats the unstripped warning with the shipped registry revision.
3. `cmd/curator/envstatus_test.go`: re-applied verbatim. The fresh-base file is byte-identical to revision 1's pre-image (blob b9f63b5d; main never touched this path — last change 7444178d), and the re-applied file hashes to blob 2919fb6d, exactly revision 1's post-image: B not-inherited warnings with no migration hint, stripped-seed TOML assertions, B record checks, and the assertCLISeedStripped helper. The A formatter regression test is untouched.

Semantic delta versus revision 1: none. Same 36 gap removals, same test intents; envstatus_test.go is byte-identical to the accepted revision.

Verification in this run (all standalone processes, real exit codes): `mini-build-lock run seedB -- env GOFLAGS=-work go test ./internal/envprofile ./internal/envregistry -run 'Seed|Codex' -count=1 -timeout=6m` exited 0 — envprofile ok in 57.942s, envregistry ok with no tests under the mask (69.81s real). The TestEnvironmentsCodexSeedVectors B-registry assertion passed; vector driving skipped locally because CURATOR_CONFORMANCE_ROOT is unset (codex_seed_test.go skips by construction). `mini-build-lock run seedB -- env GOFLAGS=-work go vet ./cmd/curator ./internal/envprofile ./internal/envregistry` exited 0 in 10.41s, compiling all edited test files without executing any test. `git diff --check` exited 0; `gofmt -l` over the four changed Go files reported nothing. syspolicyd successive crashes held at 40 across all three checkpoints (no waits). cmd/curator tests were NOT run locally per R194 — the hosted CR gate is the arbiter. Full package suites, lint, and the revision-1 mutants were not re-run: behavior is unchanged from the accepted revision and the republish brief scopes local verification to the prescribed command plus compile checks. No LOGBOOK.md, CHANGELOG.md, or scripts/remote-gate.sh edits. Six modified files, all intended. LANDING STILL HELD until the operator schedules the B release.

## Re-applied after converge (2) — revision 3 (2026-10-08)

Revision 2 (the re-application on 75ab9a71) passed the hosted gate. Before it could be reviewed, posture-B (TASK-260927-25hk87) landed on main (35cac659), so the orchestrator converged this Story workspace onto the new main. Five of the six files were carried unchanged (the test files merged cleanly); ONE file conflicted and was reverted to the fresh base: docs/ci-gates.md. This run performed no checkout, reset, or converge, and changed exactly that one file:

1. `docs/ci-gates.md`: re-applied the seed-B intent from seedB-cigates-delta.patch onto the post-posture-B text. Replaced only the sentence "The nine rc.14 Codex seed gaps remain owned by their separate flip task." with the patch's seed-B sentence ("CodexSeedRevision selects B after the A warning release in v0.15.0-rc.3. The seed families account for 7/7 provisioning and 8/8 posture cases with no gaps, bounds, or skips: all four B provisioning and five B posture cases use the shipped registry through Resolve and StatusOf; the three A cases in each family retain historical coverage through the A test seam."). Posture-B's "SecurityPostureRevision ships B: ..." text is kept verbatim, byte for byte. Working-tree diff for the doc is exactly that one-sentence substitution.
2. `cmd/curator/env_credential_marker_test.go` and `cmd/curator/envstatus_test.go`: read, not rewritten. Both still carry the seed-B intent after the merge: the historical-A record pin and B unstripped expectation in the marker test; the renamed not-inherited/strip tests, B record checks, and assertCLISeedStripped helper in envstatus_test.go. No weakening of any main assertion; both diffs match revision 2's post-images.
3. Verified no collateral revision change: SecurityPostureRevision lives in internal/config and cmd/curator/security_posture*.go, none of which appear in this six-file diff; `git diff` names no SecurityPosture code line. Gap ledger holds zero TASK-260927-1e5qqm rows (36 deletions, 0 additions).

Semantic delta versus revision 2: none. Same flip, same tests, same gap removals; only the doc sentence was re-seated beside posture-B's new text.

Verification in this run (standalone processes, real exit codes): `mini-build-lock run seedB -- env GOFLAGS=-work go test ./internal/envprofile ./internal/envregistry -run 'Seed|Codex' -count=1 -timeout=6m` exited 0 — envprofile ok in 46.096s, envregistry ok with no tests under the mask. syspolicyd successive crashes held at 40 before and after (no waits). cmd/curator tests were NOT run locally per R194 — the hosted CR gate is the arbiter. No LOGBOOK.md, CHANGELOG.md, or scripts/remote-gate.sh edits. Six modified files, all intended. LANDING STILL HELD until the operator schedules the B release.
